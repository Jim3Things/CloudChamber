package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/admin"
)

const (
	admin                = "Admin"
	adminPassword        = "AdminPassword"
	adminUpdate          = "AdminUpdate"
	adminUpdatePassword  = "AdminUpdatePassword"
	adminUpdatePassword2 = "AdminUpdatePasswordUpdated"
	adminDelete          = "AdminDelete"
	adminDeletePassword  = "AdminDeletePassword"
	alice                = "Alice"
	bob                  = "Bob"
	eve                  = "Eve"
	alicePassword        = "AlicePassword"
	bobPassword          = "BobPassword"
	evePassword          = "EvePassword"
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
		Name:              userName,
		PasswordHash:      passwordHash,
		UserId:            1,
		Enabled:           true,
		CanManageAccounts: true,
		NeverDelete:       true,
	}

	revCreate, err := store.UserCreate(context.Background(), user)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, RevisionInvalid, revCreate, "Expected new store revision to be greater than initial revision")

	revRecreate, err := store.UserCreate(context.Background(), user)
	assert.NotNilf(t, err, "Unexpected success attempting to (re-)create new user %q - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revRecreate, "Expected failure should result in invalid revision")

	readUser, readRev, err := store.UserRead(context.Background(), userName)

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
		Name:              userName,
		PasswordHash:      passwordHash,
		UserId:            1,
		Enabled:           true,
		CanManageAccounts: true,
		NeverDelete:       true,
	}

	revCreate, err := store.UserCreate(context.Background(), user)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, RevisionInvalid, revCreate, "Expected new store revision to be greater than initial revision")

	revRecreate, err := store.UserCreate(context.Background(), user)
	assert.NotNilf(t, err, "Unexpected success attempting to (re-)create new user %q - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revRecreate, "Expected failure should result in invalid revision")

	readUser, readRev, err := store.UserRead(context.Background(), userName)
	assert.Nilf(t, err, "Failed to read user %q - error: %v", userName, err)
	assert.Equalf(t, revCreate, readRev, "Unexpected difference in creation revision vs read revision")
	assert.Equalf(t, user, readUser, "Unexpected difference in creation user record and read user record")

	// Now update the user record and see if the changes made it.
	//
	passwordHash, err = bcrypt.GenerateFromPassword([]byte(adminUpdatePassword2), bcrypt.DefaultCost)

	userUpdate := &pb.User{
		Name:              userName,
		PasswordHash:      passwordHash,
		UserId:            1,
		Enabled:           false,
		CanManageAccounts: true,
		NeverDelete:       true,
	}

	revUpdate, err := store.UserUpdate(context.Background(), userUpdate, readRev)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, readRev, revUpdate, "Expected update revision to be greater than create revision")

	readUserUpdate, readRevUpdate, err := store.UserRead(context.Background(), userName)
	assert.Nilf(t, err, "Failed to read user %q - error: %v", userName, err)
	assert.Equalf(t, revUpdate, readRevUpdate, "Unexpected difference in update revision vs read revision")
	assert.Equalf(t, userUpdate, readUserUpdate, "Unexpected difference in updated user record and read user record")

	// Now try to update with the wrong revision
	//
	readRevUpdate, err = store.UserUpdate(context.Background(), userUpdate, readRev)
	assert.NotNilf(t, err, "Unexpected success trying to update with wrong revision for user %q - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, readRevUpdate, "Expected update revision to be greater than create revision")

	store.Disconnect()

	store = nil
	return
}

func TestUserDelete(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	userName := adminDelete

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminDeletePassword), bcrypt.DefaultCost)

	user := &pb.User{
		Name:              userName,
		PasswordHash:      passwordHash,
		UserId:            1,
		Enabled:           true,
		CanManageAccounts: true,
		NeverDelete:       true,
	}

	revCreate, err := store.UserCreate(context.Background(), user)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, RevisionInvalid, revCreate, "Expected new store revision to be greater than initial revision")

	userRead, revRead, err := store.UserRead(context.Background(), userName)
	assert.Nilf(t, err, "Failed to read user %q - error: %v", userName, err)
	assert.Equalf(t, revCreate, revRead, "Unexpected difference in update revision vs read revision")
	assert.Equalf(t, user, userRead, "Unexpected difference in updated user record and read user record")

	revDelete, err := store.UserDelete(context.Background(), user, revRead-1)
	assert.NotNilf(t, err, "Unexpected success trying to update with wrong revision for user %q - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revDelete, "Expected post-delete revision to be greater than read revision")

	revDelete, err = store.UserDelete(context.Background(), user, revRead)
	assert.Nilf(t, err, "Failed to delete user %q - error: %v", userName, err)
	assert.Lessf(t, revRead, revDelete, "Expected post-delete revision to be greater than read revision")

	userReread, revReread, err := store.UserRead(context.Background(), userName)
	assert.NotNilf(t, err, "Unexpected success reading user %q after deletion - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revReread, "Unexpected difference in update revision vs read revision")
	assert.Nilf(t, userReread, "Unexpected success in reading user %q after delete", userName)

	revDeleteAgain, err := store.UserDelete(context.Background(), user, revRead)
	assert.NotNilf(t, err, "Unexpected success trying to update with wrong revision for user %q - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revDeleteAgain, "Expected post-re-delete revision invalid")

	store.Disconnect()

	store = nil
	return
}

func TestUserList(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	type urec struct {
		name string
		pwd  string
	}

	userName := adminDelete

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	userSet := []urec{
		{name: alice, pwd: alicePassword},
		{name: bob, pwd: bobPassword},
		{name: eve, pwd: evePassword},
	}

	users := make(map[string]*pb.User, len(userSet))

	for i, u := range userSet {
		pwdHash, err := bcrypt.GenerateFromPassword([]byte(u.pwd), bcrypt.DefaultCost)

		assert.Nilf(t, err, "Failed to create password hash for user %q - error: %v", u.name, err)

		users[u.name] = &pb.User{
			Name:              u.name,
			PasswordHash:      pwdHash,
			UserId:            int64(i + 1),
			Enabled:           true,
			CanManageAccounts: false,
			NeverDelete:       false,
		}
	}

	recordSet := UserRecordSet{StoreRevision: RevisionInvalid, Records: make(map[string]UserRecord, len(userSet))}

	for name, rec := range users {
		revCreate, err := store.UserCreate(context.Background(), rec)
		assert.Nilf(t, err, "Failed to create new user %q - error: %v", name, err)
		assert.Lessf(t, RevisionInvalid, revCreate, "Expected new store revision to be greater than initial revision")

		recordSet.Records[name] = UserRecord{Revision: revCreate, User: rec}
	}

	userList, err := store.UserList(context.Background())
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, RevisionInvalid, userList.StoreRevision, "Expected new store revision to be greater than initial revision")

	// Use "less than or equal" relationship to allow for the cases where all the
	// file tests are being executed and there are potentially user records left over
	// from tests running earlier in the set.
	//
	assert.LessOrEqualf(t, len(recordSet.Records), len(userList.Records), "Unexpected difference in count of records returned from user list")

	// Check that the records this test created are present. There may be others.
	//
	for n, ur := range recordSet.Records {
		assert.Equalf(t, ur.Revision, userList.Records[n].Revision, "Unexpected difference in revision from create for user %q", n)
		assert.Equalf(t, ur.User, userList.Records[n].User, "Unexpected difference in record from create for user %q", n)
	}

	store.Disconnect()

	store = nil
	return
}

func TestCreate(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	userName := admin

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)

	user := &pb.User{
		Name:              userName,
		PasswordHash:      passwordHash,
		UserId:            1,
		Enabled:           true,
		CanManageAccounts: true,
		NeverDelete:       true,
	}

	revCreate, err := store.CreateNew(context.Background(), KeyRootUsers, userName, user)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, RevisionInvalid, revCreate, "Expected new store revision to be greater than initial revision")

	revCreate2, err := store.CreateNew(context.Background(), KeyRootUsers, userName, user)
	assert.NotNilf(t, err, "Unexpected success attempting to (re-)create new user %q - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revCreate2, "Expected failure should result in no response")

	userRead := &pb.User{}

	revRead, err := store.ReadNew(context.Background(), KeyRootUsers, userName, userRead)

	assert.Nilf(t, err, "Failed to read user %q - error: %v", userName, err)
	assert.Equalf(t, revCreate, revRead, "Unexpected difference in creation revision vs read revision")
	assert.Equalf(t, user, userRead, "Unexpected difference in creation user record and read user record")

	store.Disconnect()

	store = nil
	return
}

func TestCreate2(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	userName := admin

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)

	user := &pb.User{
		Name:              userName,
		PasswordHash:      passwordHash,
		UserId:            1,
		Enabled:           true,
		CanManageAccounts: true,
		NeverDelete:       true,
	}

	revision, err := store.CreateNew(context.Background(), KeyRootUsers, userName, user)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, RevisionInvalid, revision, "Expected new store revision to be greater than initial revision")

	revision2, err := store.CreateNew(context.Background(), KeyRootUsers, userName, user)
	assert.NotNilf(t, err, "Unexpected success attempting to (re-)create new user %q - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revision2, "Expected failure should result in no response")

	readUser, readRev, err := store.UserRead(context.Background(), userName)

	assert.Nilf(t, err, "Failed to read user %q - error: %v", userName, err)
	assert.Equalf(t, revision, readRev, "Unexpected difference in creation revision vs read revision")
	assert.Equalf(t, user, readUser, "Unexpected difference in creation user record and read user record")

	store.Disconnect()

	store = nil
	return
}

func TestReadNew(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	userName := admin

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)

	user := &pb.User{
		Name:              userName,
		PasswordHash:      passwordHash,
		UserId:            1,
		Enabled:           true,
		CanManageAccounts: true,
		NeverDelete:       true,
	}

	revCreate, err := store.CreateNew(context.Background(), KeyRootUsers, userName, user)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, RevisionInvalid, revCreate, "Expected new store revision to be greater than initial revision")

	readUser := &pb.User{}

	revRead, err := store.ReadNew(context.Background(), KeyRootUsers, userName, readUser)
	assert.Nilf(t, err, "Unexpected failure attempting to read user %q - error: %v", userName, err)
	assert.Equalf(t, revCreate, revRead, "Expected read revision to be equal to create revision")
	assert.Equalf(t, user, readUser, "Unexpected difference in creation user record and read user record")

	readUserString, revReadValue, err := store.ReadNewValue(context.Background(), KeyRootUsers, userName)
	assert.Nilf(t, err, "Unexpected failure attempting to read user %q - error: %v", userName, err)
	assert.Equalf(t, revCreate, revReadValue, "Expected read revision to be equal to create revision")

	readUserValue := &pb.User{}

	err = Decode(*readUserString, readUserValue)

	assert.Nilf(t, err, "Unexpected failure attempting to decode string for user %q - error: %v", userName, err, *readUserString)
	assert.Equalf(t, user, readUserValue, "Unexpected difference in creation user record and read user record")

	store.Disconnect()

	store = nil
	return
}

func TestUpdate(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	userName := adminUpdate

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminUpdatePassword), bcrypt.DefaultCost)

	user := &pb.User{
		Name:              userName,
		PasswordHash:      passwordHash,
		UserId:            1,
		Enabled:           true,
		CanManageAccounts: true,
		NeverDelete:       true,
	}

	revCreate, err := store.CreateNew(context.Background(), KeyRootUsers, userName, user)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, RevisionInvalid, revCreate, "Expected new store revision to be greater than initial revision")

	userRead := &pb.User{}

	revRead, err := store.ReadNew(context.Background(), KeyRootUsers, userName, userRead)

	assert.Nilf(t, err, "Failed to read user %q - error: %v", userName, err)
	assert.Equalf(t, revCreate, revRead, "Unexpected difference in creation revision vs read revision")
	assert.Equalf(t, user, userRead, "Unexpected difference in creation user record and read user record")

	// Now update the user record and see if the changes made it.
	//
	passwordHash, err = bcrypt.GenerateFromPassword([]byte(adminUpdatePassword2), bcrypt.DefaultCost)

	userUpdate := &pb.User{
		Name:              userName,
		PasswordHash:      passwordHash,
		UserId:            1,
		Enabled:           false,
		CanManageAccounts: true,
		NeverDelete:       true,
	}

	revUpdate, err := store.UpdateNew(context.Background(), KeyRootUsers, userName, revRead, userUpdate)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, revRead, revUpdate, "Expected update revision to be greater than create revision")

	userReadUpdate := &pb.User{}

	revReadUpdate, err := store.ReadNew(context.Background(), KeyRootUsers, userName, userReadUpdate)

	assert.Nilf(t, err, "Failed to read user %q - error: %v", userName, err)
	assert.Equalf(t, revUpdate, revReadUpdate, "Unexpected difference in update revision vs read revision")
	assert.Equalf(t, userUpdate, userReadUpdate, "Unexpected difference in updated user record and read user record")

	// Now try to update with the wrong revision
	//
	userUpdate2 := &pb.User{
		Name:              userName,
		PasswordHash:      passwordHash,
		UserId:            1,
		Enabled:           true,
		CanManageAccounts: false,
		NeverDelete:       true,
	}

	revReadUpdate2, err := store.UpdateNew(context.Background(), KeyRootUsers, userName, revRead, userUpdate2)
	assert.NotNilf(t, err, "Unexpected success trying to update with wrong revision for user %q - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revReadUpdate2, "Expected update revision to be greater than create revision")

	store.Disconnect()

	store = nil
	return
}

func TestUpdateUnconditional(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	userName := adminUpdate

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminUpdatePassword), bcrypt.DefaultCost)

	user := &pb.User{
		Name:              userName,
		PasswordHash:      passwordHash,
		UserId:            1,
		Enabled:           true,
		CanManageAccounts: true,
		NeverDelete:       true,
	}

	revCreate, err := store.CreateNew(context.Background(), KeyRootUsers, userName, user)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, RevisionInvalid, revCreate, "Expected create revision to be greater than initial revision")

	userRead := &pb.User{}

	revRead, err := store.ReadNew(context.Background(), KeyRootUsers, userName, userRead)

	assert.Nilf(t, err, "Failed to read user %q - error: %v", userName, err)
	assert.Equalf(t, revCreate, revRead, "Unexpected difference in creation revision vs read revision")
	assert.Equalf(t, user, userRead, "Unexpected difference in creation user record and read user record")

	// Now update the user record and see if the changes made it.
	//
	passwordHash, err = bcrypt.GenerateFromPassword([]byte(adminUpdatePassword2), bcrypt.DefaultCost)

	userUpdate := &pb.User{
		Name:              userName,
		PasswordHash:      passwordHash,
		UserId:            1,
		Enabled:           false,
		CanManageAccounts: true,
		NeverDelete:       true,
	}

	revUpdate, err := store.UpdateNew(context.Background(), KeyRootUsers, userName, revRead, userUpdate)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, revRead, revUpdate, "Expected update revision to be greater than first read revision")

	userReadUpdate := &pb.User{}

	revReadUpdate, err := store.ReadNew(context.Background(), KeyRootUsers, userName, userReadUpdate)

	assert.Nilf(t, err, "Failed to read user %q - error: %v", userName, err)
	assert.Equalf(t, revUpdate, revReadUpdate, "Unexpected difference in update revision vs read revision")
	assert.Equalf(t, userUpdate, userReadUpdate, "Unexpected difference in updated user record and read user record")

	// Now try to update with the wrong revision
	//
	userUpdate2 := &pb.User{
		Name:              userName,
		PasswordHash:      passwordHash,
		UserId:            1,
		Enabled:           true,
		CanManageAccounts: false,
		NeverDelete:       true,
	}

	revReadUpdate2, err := store.UpdateNew(context.Background(), KeyRootUsers, userName, revRead, userUpdate2)
	assert.NotNilf(t, err, "Unexpected success trying to update with wrong revision for user %q - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revReadUpdate2, "Expected update revision to be nil")

	// Now try to update unconditioanlly
	//
	userUpdate3 := &pb.User{
		Name:              userName,
		PasswordHash:      passwordHash,
		UserId:            1,
		Enabled:           false,
		CanManageAccounts: false,
		NeverDelete:       true,
	}

	revReadUpdate3, err := store.UpdateNew(context.Background(), KeyRootUsers, userName, RevisionInvalid, userUpdate3)
	assert.Nilf(t, err, "Failed trying to update upconditionally for user %q - error: %v", userName, err)
	assert.Lessf(t, revReadUpdate, revReadUpdate3, "Expected update revision to be greater than first update revision")

	store.Disconnect()

	store = nil
	return
}

func TestDelete(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	userName := adminDelete

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminDeletePassword), bcrypt.DefaultCost)

	user := &pb.User{
		Name:              userName,
		PasswordHash:      passwordHash,
		UserId:            1,
		Enabled:           true,
		CanManageAccounts: true,
		NeverDelete:       true,
	}

	revCreate, err := store.CreateNew(context.Background(), KeyRootUsers, userName, user)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, RevisionInvalid, revCreate, "Expected new store revision to be greater than initial revision")

	userRead := &pb.User{}

	revRead, err := store.ReadNew(context.Background(), KeyRootUsers, userName, userRead)

	assert.Nilf(t, err, "Failed to read user %q - error: %v", userName, err)
	assert.Equalf(t, revCreate, revRead, "Unexpected difference in creation revision vs read revision")
	assert.Equalf(t, user, userRead, "Unexpected difference in creation user record and read user record")

	// Fiurst try to delete using the wrong revision
	//
	revDelete, err := store.DeleteNew(context.Background(), KeyRootUsers, userName, revRead-1)
	assert.NotNilf(t, err, "Unexpected success trying to update with wrong revision for user %q - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revDelete, "Expected post-delete revision to be greater than read revision")

	// Now delete with the correct revision
	//
	revDelete, err = store.DeleteNew(context.Background(), KeyRootUsers, userName, revRead)
	assert.Nilf(t, err, "Failed to delete user %q - error: %v", userName, err)
	assert.Lessf(t, revRead, revDelete, "Expected post-delete revision to be greater than read revision")

	// Try to read after delete
	//
	userReread := &pb.User{}

	revReread, err := store.ReadNew(context.Background(), KeyRootUsers, userName, userReread)
	assert.NotNilf(t, err, "Unexpected success reading user %q after deletion - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revReread, "Unexpected difference in update revision vs read revision")

	// Try to delete a non-existing record.
	//
	revDeleteAgain, err := store.DeleteNew(context.Background(), KeyRootUsers, userName, revRead)
	assert.NotNilf(t, err, "Unexpected success trying to update with wrong revision for user %q - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revDeleteAgain, "Expected post-re-delete revision invalid")

	store.Disconnect()

	store = nil
	return
}

func TestList(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	type urec struct {
		name string
		pwd  string
	}

	userName := adminDelete

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	userSet := []urec{
		{name: alice, pwd: alicePassword},
		{name: bob, pwd: bobPassword},
		{name: eve, pwd: evePassword},
	}

	users := make(map[string]*pb.User, len(userSet))

	for i, u := range userSet {
		pwdHash, err := bcrypt.GenerateFromPassword([]byte(u.pwd), bcrypt.DefaultCost)

		assert.Nilf(t, err, "Failed to create password hash for user %q - error: %v", u.name, err)

		users[GetNormalizedName(u.name)] = &pb.User{
			Name:              u.name,
			PasswordHash:      pwdHash,
			UserId:            int64(i + 1),
			Enabled:           true,
			CanManageAccounts: false,
			NeverDelete:       false,
		}
	}

	recordSet := UserRecordSet{StoreRevision: RevisionInvalid, Records: make(map[string]UserRecord, len(userSet))}

	for name, rec := range users {
		revCreate, err := store.UserCreate(context.Background(), rec)
		assert.Nilf(t, err, "Failed to create new user %q - error: %v", name, err)
		assert.Lessf(t, RevisionInvalid, revCreate, "Expected new store revision to be greater than initial revision")

		recordSet.Records[name] = UserRecord{Revision: revCreate, User: rec}
	}

	userList, err := store.UserList(context.Background())
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, RevisionInvalid, userList.StoreRevision, "Expected new store revision to be greater than initial revision")

	// Use "less than or equal" relationship to allow for the cases where all the
	// file tests are being executed and there are potentially user records left over
	// from tests running earlier in the set.
	//
	assert.LessOrEqualf(t, len(recordSet.Records), len(userList.Records), "Unexpected difference in count of records returned from user list")

	// Check that the records this test created are present. There may be others.
	//
	for n, ur := range recordSet.Records {
		assert.Equalf(t, ur.Revision, userList.Records[n].Revision, "Unexpected difference in revision from create for user %q", n)
		assert.Equalf(t, ur.User, userList.Records[n].User, "Unexpected difference in record from create for user %q", n)
	}

	store.Disconnect()

	store = nil
	return
}
