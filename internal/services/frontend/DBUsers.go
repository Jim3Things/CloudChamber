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
	"sync"

	"github.com/Jim3Things/CloudChamber/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/admin"
	"golang.org/x/crypto/bcrypt"
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

	err = st.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {
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

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(cfg.WebServer.SystemAccountPassword), bcrypt.DefaultCost)

		if err != nil {
			return err
		}

		_, err = dbUsers.Create(&pb.User{
			Name:              cfg.WebServer.SystemAccount,
			PasswordHash:      passwordHash,
			Enabled:           true,
			CanManageAccounts: true,
			NeverDelete:       true})

		// If the SystemAccount already exists we need to do a couple of things.
		//
		// First, check to see if the password is the same and if not, issue a warning message
		// to help troubleshoot the inability to use the SystemAccount with the expected password.
		//
		// Secondly, eat the "already exists" failure as there is no need to prevent startup
		// if the account is already present.
		//
		if err == ErrUserAlreadyExists(cfg.WebServer.SystemAccount) {

			existingUser, _, err := dbUsers.Read(cfg.WebServer.SystemAccount)

			if err != nil {
				return st.Errorf(ctx, -1, "CloudChamber: unable to verify the standard %q account is using configured password - error %v", cfg.WebServer.SystemAccount, err)
			}

			if err := bcrypt.CompareHashAndPassword(existingUser.GetPasswordHash(), []byte(cfg.WebServer.SystemAccountPassword)); err != nil {
				st.Infof(ctx, -1, "CloudChamber: standard %q account is not using using configured password - error %v", cfg.WebServer.SystemAccount, err)
			}

			return nil
		}

		return err
	})

	return err
}

// Create a new user in the store
//
func (m *DBUsers) Create(u *pb.User) (int64, error) {

	v, err := store.Encode(u)

	if err != nil {
		return InvalidRev, err
	}

	rev, err := m.Store.CreateNewValue(context.Background(), store.KeyRootUsers, u.Name, v)

	if err == store.ErrStoreAlreadyExists(u.Name) {
		return InvalidRev, ErrUserAlreadyExists(u.Name)
	}

	if err != nil {
		return InvalidRev, err
	}

	return rev, nil
}

// Read the specified user from the store.
//
func (m *DBUsers) Read(name string) (*pb.User, int64, error) {

	val, rev, err := m.Store.ReadNewValue(context.Background(), store.KeyRootUsers, name)

	if err == store.ErrStoreKeyNotFound(name) {
		return nil, InvalidRev, ErrUserNotFound(name)
	}

	if err != nil {
		return nil, InvalidRev, err
	}

	u := &pb.User{}

	if err = store.Decode(*val, u); err != nil {
		return nil, InvalidRev, err
	}

	if store.GetNormalizedName(name) != store.GetNormalizedName(u.GetName()) {
		return nil, InvalidRev, ErrUserBadRecordContent{name, *val}
	}

	return u, rev, nil
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

	val, rev, err := m.Store.ReadNewValue(context.Background(), store.KeyRootUsers, u.Name)

	if err == store.ErrStoreKeyNotFound(u.Name) {
		return InvalidRev, ErrUserNotFound(u.Name)
	}

	if err != nil {
		return InvalidRev, err
	}

	if rev != match {
		return InvalidRev, ErrUserStaleVersion(u.Name)
	}

	old := &pb.User{}

	if err = store.Decode(*val, old); err != nil {
		return InvalidRev, err
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

	val, rev, err := m.Store.ReadNewValue(context.Background(), store.KeyRootUsers, n)

	if err == store.ErrStoreKeyNotFound(n) {
		return ErrUserNotFound(name)
	}

	if err != nil {
		return err
	}

	old := &pb.User{}

	if err = store.Decode(*val, old); err != nil {
		return err
	}

	if old.GetNeverDelete() {
		return ErrUserProtected(name)
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
		return ErrUserStaleVersion(name)
	}

	_, err = m.Store.DeleteNew(context.Background(), store.KeyRootUsers, n, rev)

	if err == store.ErrStoreKeyNotFound(n) {
		return ErrUserNotFound(name)
	}

	if err != nil {
		return err
	}

	return nil
}

// Scan the set of known users in the store, invoking the supplied
// function with each entry.
//
func (m *DBUsers) Scan(action func(entry *pb.User) error) error {

	recs, _, err := m.Store.ListNew(context.Background(), store.KeyRootUsers)

	if err != nil {
		return err
	}

	for n, r := range *recs {

		u := &pb.User{}

		if err = store.Decode(r.Value, u); err != nil {
			return err
		}

		if n != store.GetNormalizedName(u.GetName()) {
			return ErrUserBadRecordContent{n, r.Value}
		}

		if err := action(u); err != nil {
			return err
		}
	}

	return nil
}
