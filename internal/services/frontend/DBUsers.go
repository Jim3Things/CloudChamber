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
	"context"
	"strings"
	"sync"

	"github.com/golang/protobuf/jsonpb"

	"github.com/Jim3Things/CloudChamber/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/internal/config"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/admin"
)

// DBUsers is a container used to established synchronized access to
// the in-memory set of user records.
//
type DBUsers struct {
	Mutex sync.Mutex
	Users map[string]*pb.UserInternal

	Store *store.Store
}

// InitDBUsers is a method to initialize the users store.  For now this is only a map in memory.
func InitDBUsers(cfg *config.GlobalConfig) (err error) {
	if dbUsers == nil {
		dbUsers = &DBUsers{
			Mutex: sync.Mutex{},
			Users: make(map[string]*pb.UserInternal),
			Store: store.NewStore(),
		}
	}

	if err = dbUsers.Store.Connect(); err != nil {
		return err
	}

	_, err = userAdd(
		cfg.WebServer.SystemAccount,
		cfg.WebServer.SystemAccountPassword,
		true,
		true,
		true)

	// If the SystemAccount already exists, eat the failure
	//
	if err == store.ErrStoreRecordExists(cfg.WebServer.SystemAccount) {
		return nil
	}

	return err
}

func encode(u *pb.User) (s string, err error) {

	p := jsonpb.Marshaler{}

	if s, err = p.MarshalToString(u); err != nil {
		return "", err
	}

	return s, nil
}

func decode(s string) (*pb.User, error) {
	u := &pb.User{}

	if err := jsonpb.Unmarshal(strings.NewReader(s), u); err != nil {
		return nil, err
	}

	return u, nil
}

// Create a new user in the store
//
func (m *DBUsers) Create(u *pb.User) (int64, error) {

	rev, err := m.Store.CreateNew(context.Background(), store.KeyRootUsers, u.Name, u)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// Read the specified user from the store.
//
func (m *DBUsers) Read(name string) (*pb.User, int64, error) {

	u := &pb.User{}

	rev, err := m.Store.ReadNew(context.Background(), store.KeyRootUsers, name, u)

	if err == store.ErrStoreKeyNotFound(name) {
		return nil, InvalidRev, NewErrUserNotFound(name)
	}

	if err != nil {
		return nil, InvalidRev, err
	}

	return u, rev, nil
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
//
// NOTE: Currently, this uses a *VERY* clumsy read/modify/write action. What is
//       really needed here is to provide the ability to feed anaction into the
//       Update() routine itself to allow the caller to selectively update
//       individual fields from within the transaction.
//
func (m *DBUsers) Update(u *pb.User, match int64) (int64, error) {

	old := &pb.User{}

	//    rev, err := m.Store.ReadNewWithRevision(context.Background(), store.KeyRootUsers, key, old, match)
	rev, err := m.Store.ReadNew(context.Background(), store.KeyRootUsers, u.Name, old)

	if err == store.ErrStoreKeyNotFound(u.Name) {
		return InvalidRev, NewErrUserNotFound(u.Name)
	}

	if err != nil {
		return InvalidRev, err
	}

	if rev != match {
		return InvalidRev, NewErrUserStaleVersion(u.Name)
	}

	// Update the entry, retaining the fields from the old version that are
	// immutable
	//
	user := &pb.User{
		Name:              old.GetName(),
		PasswordHash:      u.GetPasswordHash(),
		UserId:            old.GetUserId(),
		Enabled:           u.GetEnabled(),
		CanManageAccounts: u.GetCanManageAccounts(),
		NeverDelete:       old.GetNeverDelete(),
	}

	rev, err = m.Store.UpdateNew(context.Background(), store.KeyRootUsers, u.Name, match, user)

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// Delete the entry
//
func (m *DBUsers) Delete(name string, match int64) error {

	n := store.GetNormalizedName(name)

	old := &pb.User{}

	rev, err := m.Store.ReadNew(context.Background(), store.KeyRootUsers, n, old)

	if err == store.ErrStoreKeyNotFound(n) {
		return NewErrUserNotFound(name)
	}

	if err != nil {
		return err
	}

	if old.GetNeverDelete() {
		return NewErrUserProtected(name)
	}

	if InvalidRev == match {

		// Requested an unconditional delete, at least as far as
		// the revision is concerned
		//
		rev = store.RevisionInvalid

	} else if rev != match {

		// Revision matters, so if it does not match then report
		// the problem
		//
		return NewErrUserStaleVersion(name)
	}

	_, err = m.Store.DeleteNew(context.Background(), store.KeyRootUsers, n, rev)

	if err == store.ErrStoreKeyNotFound(n) {
		return NewErrUserNotFound(name)
	}

	if err != nil {
		return err
	}

	return nil
}
