// This module contains the session management support methods
package frontend

import (
    "context"
    "net/http"

    "github.com/gorilla/sessions"

    pb "github.com/Jim3Things/CloudChamber/pkg/protos/admin"
)

// getLoggedInUser returns the user definition for the current session,
// or an error, if no user can be found.
func getLoggedInUser(session *sessions.Session) (*pb.User, error) {
    key, ok := session.Values[UserNameKey].(string)
    if !ok {
        return nil, &HTTPError{
            SC:   http.StatusBadRequest,
            Base: http.ErrNoCookie,
        }
    }

    user, _, err := dbUsers.Get(key)
    return user, err
}

// doSessionHeader wraps a handler action with the necessary code to retrieve any existing session state,
// and to attach that state to the response prior to returning.
//
// The session object is passed out for reference use by any later body processing.
func doSessionHeader(
    ctx context.Context, w http.ResponseWriter, r *http.Request,
    action func(ctx context.Context, session *sessions.Session) error) error {

    session, _ := server.cookieStore.Get(r, SessionCookieName)

    err := action(ctx, session)

    if errx := session.Save(r, w); errx != nil {
        return &HTTPError{
            SC:   http.StatusInternalServerError,
            Base: errx,
        }
    }

    return err
}