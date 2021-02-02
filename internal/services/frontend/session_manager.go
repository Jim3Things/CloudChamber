// This module contains the session management support methods

package frontend

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/sessions"

	"github.com/Jim3Things/CloudChamber/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/admin"
)

const (
	sessionCookieName = "CC-Session"

	sessionIDKey = "session-id"

	expirationTimeout = time.Duration(1) * time.Hour
	sessionLimit      = 100
)

var activeSessions = newSessionTable(sessionLimit, expirationTimeout)

// sessionState holds the current state for the session
//
type sessionState struct {
	name    string
	timeout time.Time
}

// managedSessions defines the functions that support a collection of sessions
// that can expire over time.
type managedSessions interface {
	add(state sessionState) (int64, error)
	delete(id int64) (sessionState, bool)
	get(id int64) (sessionState, bool)
	touch(id int64) (sessionState, bool)

	knownIDs() []int64
	count() int

	limit() int
	inactivity() time.Duration
}

// getSessionSummaryList returns the list of session IDs for all currently
// active sessions.
func getSessionSummaryList() []int64 {
	return activeSessions.knownIDs()
}

// getSessionTimeout returns the inactivity timer used to determine when a
// session should be removed due to lack of incoming requests.
func getSessionTimeout() time.Duration {
	return activeSessions.inactivity()
}

// newSession is a function that creates a new login session, so long as
// one is not currently active
func newSession(session *sessions.Session, state sessionState) error {
	// Fail if there is already a valid active session
	if id, ok := session.Values[sessionIDKey].(int64); ok {
		if _, ok2 := activeSessions.get(id); ok2 {
			return errors.ErrUserAlreadyLoggedIn
		}
	}

	// Create the new session
	id, err := activeSessions.add(state)
	if err != nil {
		return err
	}

	session.Values[sessionIDKey] = id
	return nil
}

// removeSession is a function to remove the designated session, or
// silently proceed if there is no active session
func removeSession(session *sessions.Session) {
	if id, ok := session.Values[sessionIDKey].(int64); ok {
		if _, ok2 := activeSessions.delete(id); ok2 {
			delete(session.Values, sessionIDKey)
		}
	}
}

// removeSessionById removes the active session specified by the id, if it is
// active.  It returns a copy of the deleted session state for informational
// use by the caller.
func removeSessionById(id int64) (sessionState, bool) {
	return activeSessions.delete(id)
}

// getSession is a function that returns the state associated with the current
// session.  It also returns a true/false flag indicating if the state was
// found in the active sessions, much like map lookup does.
func getSession(session *sessions.Session) (sessionState, bool) {
	if id, ok := session.Values[sessionIDKey].(int64); ok {
		// Bump timeout to account for the usage of the session
		if entry, ok2 := activeSessions.touch(id); ok2 {
			// .. and return the resulting entry
			return entry, true
		}

		// We have a key in the cookie, but that key is invalid, so
		// delete it.
		delete(session.Values, sessionIDKey)
	}

	return sessionState{}, false
}

// getSessionById returns the session state for an active session specified by
// the id.  This lookup does not count as an attempt to use the session, so does
// not reset the inactivity timer.  The second return value is true iff an
// active session with the specified id is found.
func getSessionById(id int64) (sessionState, bool) {
	entry, ok := activeSessions.get(id)
	return entry, ok
}

// getLoggedInUser returns the user definition for the current session,
// or an error, if no user can be found.
func getLoggedInUser(ctx context.Context, session *sessions.Session) (*pb.User, error) {
	entry, ok := getSession(session)
	if !ok {
		return nil, &HTTPError{
			SC:   http.StatusBadRequest,
			Base: http.ErrNoCookie,
		}
	}

	user, _, err := dbUsers.Read(ctx, entry.name)
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

	if err2 := session.Save(r, w); err2 != nil {
		return &HTTPError{
			SC:   http.StatusInternalServerError,
			Base: err2,
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

// +++ logging helpers

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
		activeSessions.count())
}

// --- logging helpers

// +++ managed session implementation

// sessionTable holds the information about logged in sessions and supports
// the managedSession interface.
type sessionTable struct {
	m           sync.Mutex
	known       map[int64]sessionState
	timeouts    map[time.Time][]int64
	lastID      int64
	maxInactive time.Duration
	maxCount    int
}

// newSessionTable creates a new sessionTable with the total session limit and
// inactivity timeout specified by the caller.
func newSessionTable(limit int, inactivity time.Duration) managedSessions {
	return &sessionTable{
		m:           sync.Mutex{},
		known:       make(map[int64]sessionState),
		timeouts:    make(map[time.Time][]int64),
		lastID:      0,
		maxInactive: inactivity,
		maxCount:    limit,
	}
}

// +++ managedSessions interface functions

// knownIDs returns the list of known session IDs
func (st *sessionTable) knownIDs() []int64 {
	st.m.Lock()
	defer st.m.Unlock()

	keys := make([]int64, 0, len(st.known))

	for key := range st.known {
		keys = append(keys, key)
	}

	return keys
}

// add, well, adds an entry into the sessionTable with the configured timeout.
// It returns the ID for this new entry.  This function returns an error if the
// maximum number of active sessions allowed is exceeded.
//
// Note that expired entries are first removed from the sessionTable prior to
// processing this request.
func (st *sessionTable) add(entry sessionState) (int64, error) {
	st.m.Lock()
	defer st.m.Unlock()

	st.purge()

	if len(st.known) >= st.limit() {
		return 0, errors.ErrMaxLenMap{
			Field:  "active sessions",
			Actual: int64(len(st.known)),
			Limit:  int64(st.limit()),
		}
	}

	st.lastID++
	entry.timeout = time.Now().Add(st.maxInactive)
	st.addToTables(st.lastID, entry)

	return st.lastID, nil
}

// delete removes the entry specified by the supplied id from the sessionTable.
// If the entry was found, it is deleted and a copy of the entry is returned
// with a boolean of true.  If it was not found, an empty entry is returned with
// a boolean of false.
//
// Note that expired entries are first removed from the sessionTable prior to
// processing this request.
func (st *sessionTable) delete(id int64) (sessionState, bool) {
	st.m.Lock()
	defer st.m.Unlock()

	st.purge()

	return st.deleteFromTables(id)
}

// get returns the entry associated with the supplied id.  It returns the value
// and true, if the entry was found.  It returns an empty entry and false, if
// it was not.
//
// Note that expired entries are first removed from the sessionTable prior to
// processing this request.
func (st *sessionTable) get(id int64) (sessionState, bool) {
	st.m.Lock()
	defer st.m.Unlock()

	st.purge()

	entry, ok := st.known[id]
	return entry, ok
}

// touch updates the inactivity expiration for the entry associated with the
// supplied id.  It returns the updated value and true, if the entry was found.
// It returns an empty entry and false, if it was not.
//
// Note that expired entries are first removed from the sessionTable prior to
// processing this request.
func (st *sessionTable) touch(id int64) (sessionState, bool) {
	st.m.Lock()
	defer st.m.Unlock()

	st.purge()

	if entry, ok := st.known[id]; ok {
		st.removeFromTimeoutList(id, entry.timeout)
		entry.timeout = time.Now().Add(st.maxInactive)
		st.addToTimeoutList(id, entry.timeout)

		return entry, true
	}

	return sessionState{}, false
}

// count returns the number of entries currently stored in this sessionTable
func (st *sessionTable) count() int {
	st.m.Lock()
	defer st.m.Unlock()

	return len(st.known)
}

// limit returns the maximum number of entries allowed in this sessionTable.
// It is a constant value specified when the sessionTable instance was created.
func (st *sessionTable) limit() int {
	return st.maxCount
}

// inactivity returns the timeout duration before an entry that has not been
// touched is removed.  It is a constant value specified when the sessionTable
// instance was created.
func (st *sessionTable) inactivity() time.Duration {
	return st.maxInactive
}

// --- managedSessions interface functions

// +++ internal sessionTable functions

// purge removes stale sessions from the active session list.  It assumes that
// the number of sessions is not so large that the active session list cannot
// be scanned.
func (st *sessionTable) purge() {
	now := time.Now()

	for k, v := range st.timeouts {
		if k.Before(now) {
			for _, id := range v {
				delete(st.known, id)
			}

			delete(st.timeouts, k)
		}
	}
}

// addToTables adds the entry with the specified id to the tables that make up
// the multi-index map that is the sessionTable
func (st *sessionTable) addToTables(id int64, entry sessionState) {
	st.known[id] = entry
	st.addToTimeoutList(id, entry.timeout)
}

// deleteFromTables removes the entry with the specified id from the tables that
// make up the multi-index map that is the sessionTable.  This function returns
// the removed entry and true, if found; an empty entry and false, if not.
func (st *sessionTable) deleteFromTables(id int64) (sessionState, bool) {
	if entry, ok := st.known[id]; ok {
		delete(st.known, id)

		st.removeFromTimeoutList(id, entry.timeout)

		return entry, true
	}

	return sessionState{}, false
}

// addToTimeoutList adds an entry to the map keyed by timeout.  Entries may
// collide on timeout, so this map handles lists of entries that are attached
// to a single timeout.
func (st *sessionTable) addToTimeoutList(id int64, expiry time.Time) {
	list, ok := st.timeouts[expiry]
	if !ok {
		list = []int64{}
	}

	list = append(list, id)
	st.timeouts[expiry] = list
}

// removeFromTimeoutList removes an entry from the appropriate timeout map
// list.  It cleans up map when the list of entries at a given timeout becomes
// empty.
func (st *sessionTable) removeFromTimeoutList(id int64, expiry time.Time) {
	list := st.timeouts[expiry]
	var l2 []int64
	for _, i := range list {
		if i != id {
			l2 = append(l2, i)
		}
	}

	if len(l2) == 0 {
		delete(st.timeouts, expiry)
	} else {
		st.timeouts[expiry] = l2
	}
}

// --- internal sessionTable functions
