// This module contains the session management support methods

package frontend

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/sessions"

	ts "github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/admin"
)

const (
	sessionCookieName = "CC-Session"

	sessionIDKey = "session-id"

	expirationTimeout = time.Duration(1) * time.Hour
)

// SessionState holds the current state for the session
//
type SessionState struct {
	name    string
	timeout time.Time
}

var mutex = sync.Mutex{}

// activeSessions is the set of currently active logged in sessions, indexed
// by an id.  timeouts holds the mapping from expiration time to the active
// session id, and is used for purging stale sessions from the server.
var activeSessions = map[int64]SessionState{}
var timeouts = map[time.Time]int64{}

// lastID contains the last session ID used by this server
var lastID int64 = 0

// newSession is a function that creates a new login session, so long as
// one is not currently active
func newSession(session *sessions.Session, state SessionState) error {
	mutex.Lock()
	defer mutex.Unlock()

	purgeStaleSessions()

	// Fail if there is already a valid active session
	if id, ok := session.Values[sessionIDKey].(int64); ok {
		if _, ok = activeSessions[id]; ok {
			return ErrUserAlreadyLoggedIn
		}
	}

	// Create the new session
	lastID++
	state.timeout = time.Now().Add(expirationTimeout)

	activeSessions[lastID] = state
	timeouts[state.timeout] = lastID
	session.Values[sessionIDKey] = lastID

	return nil
}

// removeSession is a function to remove the designated session, or
// silently proceed if there is no active session
func removeSession(session *sessions.Session) {
	mutex.Lock()
	defer mutex.Unlock()

	purgeStaleSessions()

	if id, ok := session.Values[sessionIDKey].(int64); ok {
		if entry, ok := activeSessions[id]; ok {
			delete(activeSessions, id)
			delete(timeouts, entry.timeout)

			delete(session.Values, sessionIDKey)
		}
	}
}

// getSession is a function that returns the state associated with the current
// session.  It also returns a true/false flag indicating if the state was
// found in the active sessions, much like map lookup does.
func getSession(session *sessions.Session) (SessionState, bool) {
	mutex.Lock()
	defer mutex.Unlock()

	purgeStaleSessions()

	if id, ok := session.Values[sessionIDKey].(int64); ok {
		if entry, ok := activeSessions[id]; ok {
			// Bump timeout to account for the usage of the session
			delete(timeouts, entry.timeout)

			entry.timeout = time.Now().Add(expirationTimeout)
			activeSessions[id] = entry
			timeouts[entry.timeout] = id

			// .. and return the resulting entry
			return entry, true
		}

		// We have a key in the cookie, but that key is invalid, so
		// delete it.
		delete(session.Values, sessionIDKey)

	}

	return SessionState{}, false
}

// purgeStaleSessions removes stale sessions from the active session list.  It
// assumes that the number of sessions is not so large that the active session
// list cannot be scanned.
func purgeStaleSessions() {
	now := time.Now()

	for k, v := range timeouts {
		if k.Before(now) {
			delete(activeSessions, v)
			delete(timeouts, k)
		}
	}
}

// +++ logging helpers

// getActiveSessionCount gets the number of active sessions currently held
func getActiveSessionCount() int {
	mutex.Lock()
	defer mutex.Unlock()

	return len(activeSessions)
}

// dumpSessionState returns a string that has the current session's state
// formatted.  This is intended for use by tracing calls.
func dumpSessionState(session *sessions.Session) string {
	stateString := "session state not found"
	idString := "session ID not found"

	if id, ok := session.Values[sessionIDKey].(int64); ok {
		idString = fmt.Sprintf("ID: %d", id)
	}

	if state, ok := getSession(session); ok {
		stateString = fmt.Sprintf(
			"[Username: %s, expiry: %v",
			state.name,
			state.timeout)
	}

	return fmt.Sprintf(
		"Session state: [%s, %s], active session count: %d",
		idString,
		stateString,
		getActiveSessionCount())
}

// --- logging helpers

// getLoggedInUser returns the user definition for the current session,
// or an error, if no user can be found.
func getLoggedInUser(session *sessions.Session) (*pb.User, error) {
	entry, ok := getSession(session)
	if !ok {
		return nil, &HTTPError{
			SC:   http.StatusBadRequest,
			Base: http.ErrNoCookie,
		}
	}

	user, _, err := dbUsers.Read(context.Background(), entry.name)
	return user, err
}

// doSessionHeader wraps a handler action with the necessary code to retrieve any existing session state,
// and to attach that state to the response prior to returning.
//
// The session object is passed out for reference use by any later body processing.
func doSessionHeader(
	ctx context.Context, w http.ResponseWriter, r *http.Request,
	action func(ctx context.Context, session *sessions.Session) error) error {

	session, _ := server.cookieStore.Get(r, sessionCookieName)

	err := action(ctx, session)

	if errx := session.Save(r, w); errx != nil {
		return &HTTPError{
			SC:   http.StatusInternalServerError,
			Base: errx,
		}
	}

	return err
}

// ensureEstablishedSession verifies that the session is not new, and triggers
// an error if it is.
func ensureEstablishedSession(session *sessions.Session) error {
	if session.IsNew {
		return NewErrNoSessionActive()
	}

	if _, ok := getSession(session); !ok {
		return NewErrNoSessionActive()
	}

	return nil
}

// tick provides the current simulated time tick, or '-1' if the simulated time
// cannot be retrieved (e.g. during startup)
func tick() int64 {
	now, err := ts.Now()
	if err != nil {
		return -1
	}

	return now.Ticks
}
