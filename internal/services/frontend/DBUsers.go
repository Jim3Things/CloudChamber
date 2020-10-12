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

	"golang.org/x/crypto/bcrypt"

	"github.com/Jim3Things/CloudChamber/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/admin"
)

// DBUsers is a container used to established synchronized access to
// the in-memory set of user records.
//
type DBUsers struct {
	Store *store.Store
}

var dbUsers *DBUsers

// InitDBUsers is a method to initialize the users store.  For now this is only a map in memory.
func InitDBUsers(ctx context.Context, cfg *config.GlobalConfig) (err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Initialize User DB Connection"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	if dbUsers == nil {
		dbUsers = &DBUsers{
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

	_, err = dbUsers.Create(ctx, &pb.User{
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
		existingUser, _, err := dbUsers.Read(ctx, cfg.WebServer.SystemAccount)

		if err != nil {
			return tracing.Error(ctx, ErrUnableToVerifySystemAccount{
				Name: cfg.WebServer.SystemAccount,
				Err:  err,
			})
		}

		if err = bcrypt.CompareHashAndPassword(
			existingUser.GetPasswordHash(),
			[]byte(cfg.WebServer.SystemAccountPassword)); err != nil {
			tracing.Info(ctx, "CloudChamber: standard %q account is not using using configured password - error %v", cfg.WebServer.SystemAccount, err)
		}

		return nil
	}

	return err
}

// Create a new user in the store
//
func (m *DBUsers) Create(ctx context.Context, u *pb.User) (int64, error) {

	v, err := store.Encode(u)

	if err != nil {
		return InvalidRev, err
	}

	rev, err := m.Store.Create(ctx, store.KeyRootUsers, u.Name, v)

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
func (m *DBUsers) Read(ctx context.Context, name string) (*pb.User, int64, error) {

	val, rev, err := m.Store.Read(ctx, store.KeyRootUsers, name)

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
func (m *DBUsers) Update(ctx context.Context, name string, u *pb.UserUpdate, match int64) (*pb.User, int64, error) {

	val, rev, err := m.Store.Read(ctx, store.KeyRootUsers, name)

	if err == store.ErrStoreKeyNotFound(name) {
		return nil, InvalidRev, ErrUserNotFound(name)
	}

	if err != nil {
		return nil, InvalidRev, err
	}

	if rev != match {
		return nil, InvalidRev, ErrUserStaleVersion(name)
	}

	old := &pb.User{}

	if err = store.Decode(*val, old); err != nil {
		return nil, InvalidRev, err
	}

	// Update the entry, retaining the fields from the old version that are
	// immutable
	//
	user := &pb.User{
		Name:              old.GetName(),
		PasswordHash:      old.GetPasswordHash(),
		UserId:            old.GetUserId(),
		Enabled:           u.GetEnabled(),
		CanManageAccounts: u.GetCanManageAccounts(),
		NeverDelete:       old.GetNeverDelete(),
	}

	rev, err = m.Store.UpdateWithEncode(ctx, store.KeyRootUsers, name, match, user)

	if err != nil {
		return nil, InvalidRev, err
	}

	return user, rev, nil
}

// UpdatePassword is a function that updates the password hash field only in an
// existing user record.
//
// That this is split out from Update reflects the usage patterns for updating
// user entries.
func (m *DBUsers) UpdatePassword(ctx context.Context, name string, hash []byte, match int64) (*pb.User, int64, error) {

	val, rev, err := m.Store.Read(ctx, store.KeyRootUsers, name)

	if err == store.ErrStoreKeyNotFound(name) {
		return nil, InvalidRev, ErrUserNotFound(name)
	}

	if err != nil {
		return nil, InvalidRev, err
	}

	if rev != match {
		return nil, InvalidRev, ErrUserStaleVersion(name)
	}

	old := &pb.User{}

	if err = store.Decode(*val, old); err != nil {
		return nil, InvalidRev, err
	}

	// Update the entry, retaining all fields save the password hash
	//
	user := &pb.User{
		Name:              old.GetName(),
		PasswordHash:      hash,
		UserId:            old.GetUserId(),
		Enabled:           old.GetEnabled(),
		CanManageAccounts: old.GetCanManageAccounts(),
		NeverDelete:       old.GetNeverDelete(),
	}

	rev, err = m.Store.UpdateWithEncode(ctx, store.KeyRootUsers, name, match, user)

	if err != nil {
		return nil, InvalidRev, err
	}

	return user, rev, nil
}

// Delete the entry
//
func (m *DBUsers) Delete(ctx context.Context, name string, match int64) error {

	n := store.GetNormalizedName(name)

	val, rev, err := m.Store.Read(ctx, store.KeyRootUsers, n)

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

	_, err = m.Store.Delete(ctx, store.KeyRootUsers, n, rev)

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
func (m *DBUsers) Scan(ctx context.Context, action func(entry *pb.User) error) error {

	recs, _, err := m.Store.List(ctx, store.KeyRootUsers)

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

		if err = action(u); err != nil {
			return err
		}
	}

	return nil
}
