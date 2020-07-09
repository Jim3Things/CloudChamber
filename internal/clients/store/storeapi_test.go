package store

import (
	"testing"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/admin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

const (
	userURI              = "/api/users/"
	admin                = "Admin"
	adminPassword        = "AdminPassword"
	adminUpdate          = "AdminUpdate"
	adminUpdatePassword  = "AdminUpdatePassword"
	adminUpdatePassword2 = "AdminUpdatePasswordUpdated"
	alice                = "Alice"
	bob                  = "Bob"
	alicePassword        = "AlicePassowrd"
	bobPassword          = "BobPassword"
)

func TestUserCreate(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	userName := admin

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)

	user := &pb.User{
		Name:           userName,
		PasswordHash:   passwordHash,
		UserId:         1,
		Enabled:        true,
		AccountManager: true,
		NeverDelete:    true,
	}

	revCreate, err := store.UserCreate(user)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, RevisionInvalid, revCreate, "Expected new store revision to be greater than initial revision")

	revRecreate, err := store.UserCreate(user)
	assert.NotNilf(t, err, "Unexpected success attempting to (re-)create new user %q - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revRecreate, "Expected failure should result in invalid revision")

	readUser, readRev, err := store.UserRead(userName)

	assert.Nilf(t, err, "Failed to read user %q - error: %v", userName, err)
	assert.Equalf(t, revCreate, readRev, "Unexpected difference in creation revision vs read revision")
	assert.Equalf(t, user, readUser, "Unexpected difference in creation user record and read user record")

	store.Disconnect()

	store = nil
	return
}

func TestUserUpdate(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	userName := adminUpdate

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminUpdatePassword), bcrypt.DefaultCost)

	user := &pb.User{
		Name:           userName,
		PasswordHash:   passwordHash,
		UserId:         1,
		Enabled:        true,
		AccountManager: true,
		NeverDelete:    true,
	}

	revCreate, err := store.UserCreate(user)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, RevisionInvalid, revCreate, "Expected new store revision to be greater than initial revision")

	revRecreate, err := store.UserCreate(user)
	assert.NotNilf(t, err, "Unexpected success attempting to (re-)create new user %q - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revRecreate, "Expected failure should result in invalid revision")

	readUser, readRev, err := store.UserRead(userName)
	assert.Nilf(t, err, "Failed to read user %q - error: %v", userName, err)
	assert.Equalf(t, revCreate, readRev, "Unexpected difference in creation revision vs read revision")
	assert.Equalf(t, user, readUser, "Unexpected difference in creation user record and read user record")

	// Now update the user record and see if the changes made it.
	//
	passwordHash, err = bcrypt.GenerateFromPassword([]byte(adminUpdatePassword2), bcrypt.DefaultCost)

	userUpdate := &pb.User{
		Name:           userName,
		PasswordHash:   passwordHash,
		UserId:         1,
		Enabled:        false,
		AccountManager: true,
		NeverDelete:    true,
	}

	revUpdate, err := store.UserUpdate(userUpdate, readRev)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, readRev, revUpdate, "Expected update revision to be greater than create revision")

	readUserUpdate, readRevUpdate, err := store.UserRead(userName)
	assert.Nilf(t, err, "Failed to read user %q - error: %v", userName, err)
	assert.Equalf(t, revUpdate, readRevUpdate, "Unexpected difference in update revision vs read revision")
	assert.Equalf(t, userUpdate, readUserUpdate, "Unexpected difference in updated user record and read user record")

	// Now try to update with the wrong revision
	//
	readRevUpdate, err = store.UserUpdate(userUpdate, readRev)
	assert.NotNilf(t, err, "Unecpected success trying to update with wrong revision for user %q - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, readRevUpdate, "Expected update revision to be greater than create revision")

	store.Disconnect()

	store = nil
	return
}
