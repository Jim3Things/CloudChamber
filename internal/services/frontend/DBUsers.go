// This module encapsulates storage and retrieval of known users

// The full user entry contains attributes about what it can do, the password
// hash, and a revision number.  The password hash is never exposed outside of
// this module.  The revision number is returned, and used as a precondition
// on any update requests.

// Each user entry has an associated key which is the lowercased form of the
// username.  The supplied name is retained as an attribute in order to present
// the form that the caller originally used for display purposes.

package frontend

import (
    "strings"
    "sync"

    "github.com/Jim3Things/CloudChamber/internal/config"
    pb "github.com/Jim3Things/CloudChamber/pkg/protos/admin"
)

// DBUsers is a container used to established synchronized access to
// the in-memory set of user records.
//
type DBUsers struct {
    Mutex sync.Mutex
    Users map[string]*pb.UserInternal
}

// Initialize the users store.  For now this is only a map in memory.
func InitDBUsers(cfg *config.GlobalConfig) error {
    if dbUsers == nil {
        dbUsers = &DBUsers{
            Mutex: sync.Mutex{},
            Users: make(map[string]*pb.UserInternal),
        }
    }

    _, err := UserAdd(cfg.WebServer.SystemAccount, cfg.WebServer.SystemAccountPassword, true, true)
    return err
}

// Create a new user entry in the store.
func (m *DBUsers) Create(u *pb.User) (int64, error) {
    m.Mutex.Lock()
    defer m.Mutex.Unlock()

    key := strings.ToLower(u.Name)

    if _, ok := m.Users[key]; ok {
        return InvalidRev, NewErrUserAlreadyCreated(u.Name)
    }

    entry := &pb.UserInternal{
        User:                 &pb.User{
            Name:                 u.Name,
            PasswordHash:         u.PasswordHash,
            UserId:               u.UserId,
            Enabled:              u.Enabled,
            AccountManager:       u.AccountManager,
        },
        Revision:             1,
    }
    m.Users[key] = entry

    return 1, nil
}

// Get the specified user from the store.
func (m *DBUsers) Get(name string) (*pb.User, int64, error) {
    m.Mutex.Lock()
    defer m.Mutex.Unlock()

    key := strings.ToLower(name)

    entry, ok := m.Users[key]
    if !ok {
        return nil, InvalidRev, NewErrUserNotFound(name)
    }

    return entry.User, entry.Revision, nil
}

// Scan the set of known users in the store, invoking the supplied
// function with each entry.
func (m *DBUsers) Scan(action func(entry *pb.User) error) error {
    m.Mutex.Lock()
    defer m.Mutex.Unlock()

    for _, user := range dbUsers.Users {
        if err := action(user.User); err != nil {
            return err
        }
    }

    return nil
}

// Update an existing user entry, iff the current revision is the same as the
// expected (match) revision
func (m *DBUsers) Update(u *pb.User, match int64) (int64, error) {
    m.Mutex.Lock()
    defer m.Mutex.Unlock()

    key := strings.ToLower(u.Name)

    old, ok := m.Users[key]
    if !ok {
        return InvalidRev, NewErrUserNotFound(u.Name)
    }

    if old.Revision != match {
        return InvalidRev, NewErrUserStaleVersion(u.Name)
    }

    entry := &pb.UserInternal{
        User:                 &pb.User{
            Name:                 u.Name,
            PasswordHash:         u.PasswordHash,
            UserId:               u.UserId,
            Enabled:              u.Enabled,
            AccountManager:       u.AccountManager,
        },
        Revision:             match + 1,
    }
    m.Users[key] = entry

    return entry.Revision, nil
}

func (m *DBUsers) Remove(name string) error {
    m.Mutex.Lock()
    defer m.Mutex.Unlock()

    key := strings.ToLower(name)

    _, ok := m.Users[key]
    if !ok {
        return NewErrUserNotFound(name)
    }

    delete(m.Users, key)
    return nil
}
