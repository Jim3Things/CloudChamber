// This module containes the routines etc to implement the frontend handlers for the users
// part of the API
//

package frontend

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"

	"golang.org/x/crypto/bcrypt"
)

//TODO This is just a placeholder until we have proper user records held in a persisted store (Etcd)
//
// User is a representation of an individual user
type User struct {
	Name         string
	PasswordHash []byte
	//	UserId   int64
	Enabled bool
}

type DbUsers struct {
	Mutex sync.Mutex
	Users map[string]User
}

var (
	dbUsers DbUsers
)

func userCreate(name string, password []byte) (*User, error) {

	passwordHash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	user := User{Name: name, PasswordHash: passwordHash}

	return &user, nil
}

func userAdd(name string, password []byte) error {

	newUser, err := userCreate(name, password)

	if nil != err {
		return err
	}

	dbUsers.Mutex.Lock()
	defer dbUsers.Mutex.Unlock()

	_, found := dbUsers.Users[name]

	if found {
		return ErrUserAlreadyExists
	}

	dbUsers.Users[name] = *newUser
	return nil
}

func userRemove(name string) error {
	dbUsers.Mutex.Lock()
	delete(dbUsers.Users, name)
	dbUsers.Mutex.Unlock()
	return nil
}

func userVerifyPassword(name string, password []byte) error {

	dbUsers.Mutex.Lock()
	defer dbUsers.Mutex.Unlock()

	return bcrypt.CompareHashAndPassword(dbUsers.Users[name].PasswordHash, password)
}

func userEnable(name string, enable bool) error {

	dbUsers.Mutex.Lock()
	defer dbUsers.Mutex.Unlock()

	user, found := dbUsers.Users[name]
	if !found {
		return ErrUserNotFound
	}

	user.Enabled = enable
	return nil
}

func usersAddRoutes(routeBase *mux.Router) {

	routeUsers := routeBase.PathPrefix("/users").Subrouter()

	routeUsers.HandleFunc("", handlerUsersList).Methods("GET")
	routeUsers.HandleFunc("/", handlerUsersList).Methods("GET")

	// In the following, the "GET" method is allowed just for the purposes of test and
	// evaluation. At somepoint, it will need to be removed, but in the meantime, leaving
	// it there allows simple experimentation with just a browser.
	//
	// As a reminder,
	//	 PUT is idempotent so translates to UPDATE in the CRUD methodolgy
	//   POST is NOT idempotent so translates to CREATE in the CRUD methodolgy
	//
	routeUsers.HandleFunc("/{username}", handlerUsersCreate).Methods("POST", "GET")
	routeUsers.HandleFunc("/{username}", handlerUsersRead).Methods("GET")
	routeUsers.HandleFunc("/{username}", handlerUsersUpdate).Methods("PUT", "GET")
	routeUsers.HandleFunc("/{username}", handlerUsersDelete).Methods("DELETE", "GET")

}

func usersDisplayArguments(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	username := vars["username"]

	fmt.Fprintf(w, "User: %s", username)
}

func handlerUsersList(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Users (List)")
}

func handlerUsersCreate(w http.ResponseWriter, r *http.Request) {

	usersDisplayArguments(w, r)
}

func handlerUsersRead(w http.ResponseWriter, r *http.Request) {

	usersDisplayArguments(w, r)
}

func handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {

	usersDisplayArguments(w, r)
}

func handlerUsersDelete(w http.ResponseWriter, r *http.Request) {

	usersDisplayArguments(w, r)
}

func secret(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Print secret message
	fmt.Fprintln(w, "secret message")
}

func login(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	// Authentication goes here
	// ...

	// Set user as authenticated
	session.Values["authenticated"] = true
	session.Save(r, w)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	// Revoke users authentication
	session.Values["authenticated"] = false
	session.Save(r, w)
}

//func handlerAuthenticateSession(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
//    return func(w http.ResponseWriter, r *http.Request) {
//        m := validPath.FindStringSubmatch(r.URL.Path)
//        if m == nil {
//            http.NotFound(w, r)
//            return
//        }
//        fn(w, r, m[2])
//    }
//}
