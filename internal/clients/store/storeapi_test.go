package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
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

func TestCreate(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

	userName := admin + "." + tracing.MethodName(1)

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

	revCreate, err := store.CreateWithEncode(context.Background(), KeyRootUsers, userName, user)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, RevisionInvalid, revCreate, "Expected new store revision to be greater than initial revision")

	revCreate2, err := store.CreateWithEncode(context.Background(), KeyRootUsers, userName, user)
	assert.NotNilf(t, err, "Unexpected success attempting to (re-)create new user %q - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revCreate2, "Expected failure should result in no response")

	userRead := &pb.User{}

	revRead, err := store.ReadWithDecode(context.Background(), KeyRootUsers, userName, userRead)

	assert.Nilf(t, err, "Failed to read user %q - error: %v", userName, err)
	assert.Equalf(t, revCreate, revRead, "Unexpected difference in creation revision vs read revision")
	assert.Equalf(t, user, userRead, "Unexpected difference in creation user record and read user record")

	store.Disconnect()

	store = nil
	return
}

func TestReadNew(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

	userName := admin + "." + tracing.MethodName(1)

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

	revCreate, err := store.CreateWithEncode(context.Background(), KeyRootUsers, userName, user)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, RevisionInvalid, revCreate, "Expected new store revision to be greater than initial revision")

	readUser := &pb.User{}

	revRead, err := store.ReadWithDecode(context.Background(), KeyRootUsers, userName, readUser)
	assert.Nilf(t, err, "Unexpected failure attempting to read user %q - error: %v", userName, err)
	assert.Equalf(t, revCreate, revRead, "Expected read revision to be equal to create revision")
	assert.Equalf(t, user, readUser, "Unexpected difference in creation user record and read user record")

	readUserString, revReadValue, err := store.Read(context.Background(), KeyRootUsers, userName)
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

func TestReadNewValue(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

	userName := admin + "." + tracing.MethodName(1)

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

	userValue, err := Encode(user)
	assert.Nilf(t, err, "Failed to encode user record")

	revCreate, err := store.Create(context.Background(), KeyRootUsers, userName, userValue)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, RevisionInvalid, revCreate, "Expected new store revision to be greater than initial revision")

	readUserValue, revReadValue, err := store.Read(context.Background(), KeyRootUsers, userName)
	assert.Nilf(t, err, "Unexpected failure attempting to read user %q - error: %v", userName, err)
	assert.Equalf(t, revCreate, revReadValue, "Expected read revision to be equal to create revision")
	assert.Equalf(t, userValue, *readUserValue, "Unexpected difference in creation user record and read user record")

	readUser := &pb.User{}

	revRead, err := store.ReadWithDecode(context.Background(), KeyRootUsers, userName, readUser)
	assert.Nilf(t, err, "Unexpected failure attempting to read user %q - error: %v", userName, err)
	assert.Equalf(t, revCreate, revRead, "Expected read revision to be equal to create revision")
	assert.Equalf(t, user, readUser, "Unexpected difference in creation user record and read user record")

	readUserFromValue := &pb.User{}

	err = Decode(*readUserValue, readUserFromValue)
	assert.Nilf(t, err, "Unexpected failure attempting to decode string for user %q with value %q", userName, *readUserValue)
	assert.Equalf(t, user, readUserFromValue, "Unexpected difference in creation user record and read user record")

	store.Disconnect()

	store = nil
	return
}

func TestUpdate(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

	userName := admin + "." + tracing.MethodName(1)

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

	revCreate, err := store.CreateWithEncode(context.Background(), KeyRootUsers, userName, user)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, RevisionInvalid, revCreate, "Expected new store revision to be greater than initial revision")

	userRead := &pb.User{}

	revRead, err := store.ReadWithDecode(context.Background(), KeyRootUsers, userName, userRead)

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

	revUpdate, err := store.UpdateWithEncode(context.Background(), KeyRootUsers, userName, revRead, userUpdate)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, revRead, revUpdate, "Expected update revision to be greater than create revision")

	userReadUpdate := &pb.User{}

	revReadUpdate, err := store.ReadWithDecode(context.Background(), KeyRootUsers, userName, userReadUpdate)

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

	revReadUpdate2, err := store.UpdateWithEncode(context.Background(), KeyRootUsers, userName, revRead, userUpdate2)
	assert.NotNilf(t, err, "Unexpected success trying to update with wrong revision for user %q - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revReadUpdate2, "Expected update revision to be greater than create revision")

	store.Disconnect()

	store = nil
	return
}

func TestUpdateUnconditional(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

	userName := admin + "." + tracing.MethodName(1)

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

	revCreate, err := store.CreateWithEncode(context.Background(), KeyRootUsers, userName, user)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, RevisionInvalid, revCreate, "Expected create revision to be greater than initial revision")

	userRead := &pb.User{}

	revRead, err := store.ReadWithDecode(context.Background(), KeyRootUsers, userName, userRead)

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

	revUpdate, err := store.UpdateWithEncode(context.Background(), KeyRootUsers, userName, revRead, userUpdate)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, revRead, revUpdate, "Expected update revision to be greater than first read revision")

	userReadUpdate := &pb.User{}

	revReadUpdate, err := store.ReadWithDecode(context.Background(), KeyRootUsers, userName, userReadUpdate)

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

	revReadUpdate2, err := store.UpdateWithEncode(context.Background(), KeyRootUsers, userName, revRead, userUpdate2)
	assert.NotNilf(t, err, "Unexpected success trying to update with wrong revision for user %q - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revReadUpdate2, "Expected update revision to be nil")

	// Now try to update unconditionally
	//
	userUpdate3 := &pb.User{
		Name:              userName,
		PasswordHash:      passwordHash,
		UserId:            1,
		Enabled:           false,
		CanManageAccounts: false,
		NeverDelete:       true,
	}

	revReadUpdate3, err := store.UpdateWithEncode(context.Background(), KeyRootUsers, userName, RevisionInvalid, userUpdate3)
	assert.Nilf(t, err, "Failed trying to update upconditionally for user %q - error: %v", userName, err)
	assert.Lessf(t, revReadUpdate, revReadUpdate3, "Expected update revision to be greater than first update revision")

	store.Disconnect()

	store = nil
	return
}

func TestDelete(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

	userName := alice + "." + tracing.MethodName(1)
	passWord := alicePassword

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(passWord), bcrypt.DefaultCost)

	user := &pb.User{
		Name:              userName,
		PasswordHash:      passwordHash,
		UserId:            1,
		Enabled:           true,
		CanManageAccounts: true,
		NeverDelete:       true,
	}

	revCreate, err := store.CreateWithEncode(context.Background(), KeyRootUsers, userName, user)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", userName, err)
	assert.Lessf(t, RevisionInvalid, revCreate, "Expected new store revision to be greater than initial revision")

	userRead := &pb.User{}

	revRead, err := store.ReadWithDecode(context.Background(), KeyRootUsers, userName, userRead)

	assert.Nilf(t, err, "Failed to read user %q - error: %v", userName, err)
	assert.Equalf(t, revCreate, revRead, "Unexpected difference in creation revision vs read revision")
	assert.Equalf(t, user, userRead, "Unexpected difference in creation user record and read user record")

	// Fiurst try to delete using the wrong revision
	//
	revDelete, err := store.Delete(context.Background(), KeyRootUsers, userName, revRead-1)
	assert.NotNilf(t, err, "Unexpected success trying to update with wrong revision for user %q - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revDelete, "Expected post-delete revision to be greater than read revision")

	// Now delete with the correct revision
	//
	revDelete, err = store.Delete(context.Background(), KeyRootUsers, userName, revRead)
	assert.Nilf(t, err, "Failed to delete user %q - error: %v", userName, err)
	assert.Lessf(t, revRead, revDelete, "Expected post-delete revision to be greater than read revision")

	// Try to read after delete
	//
	userReread := &pb.User{}

	revReread, err := store.ReadWithDecode(context.Background(), KeyRootUsers, userName, userReread)
	assert.NotNilf(t, err, "Unexpected success reading user %q after deletion - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revReread, "Unexpected difference in update revision vs read revision")

	// Try to delete a non-existing record.
	//
	revDeleteAgain, err := store.Delete(context.Background(), KeyRootUsers, userName, revRead)
	assert.NotNilf(t, err, "Unexpected success trying to update with wrong revision for user %q - error: %v", userName, err)
	assert.Equalf(t, RevisionInvalid, revDeleteAgain, "Expected post-re-delete revision invalid")

	store.Disconnect()

	store = nil
	return
}

func TestList(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

	type urec struct {
		name string
		pwd  string
	}

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	suffix := "." + tracing.MethodName(1)

	userSet := []urec{
		{name: alice + suffix, pwd: alicePassword},
		{name: bob + suffix, pwd: bobPassword},
		{name: eve + suffix, pwd: evePassword},
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

	revFirstCreate := RevisionInvalid

	userRecords := make(map[string]Record, len(userSet))

	for n, u := range users {
		v, err := Encode(u)
		assert.Nilf(t, err, "Failed to encode value for new user %q - error: %v", n, err)

		revCreate, err := store.Create(context.Background(), KeyRootUsers, n, v)
		assert.Nilf(t, err, "Failed to create new user %q - error: %v", n, err)
		assert.Lessf(t, RevisionInvalid, revCreate, "Expected new store revision to be greater than initial revision")

		userRecords[n] = Record{Revision: revCreate, Value: v}

		if revFirstCreate == RevisionInvalid {
			revFirstCreate = revCreate
		}
	}

	listRecs, listRev, err := store.List(context.Background(), KeyRootUsers)
	assert.Nilf(t, err, "Failed to list records")
	assert.LessOrEqualf(t, revFirstCreate, listRev, "Expected new store revision to be greater than initial revision")

	// Use "less than or equal" relationship to allow for the cases where all the
	// file tests are being executed and there are potentially user records left over
	// from tests running earlier in the set.
	//
	assert.LessOrEqualf(t, len(userRecords), len(*listRecs), "Unexpected difference in count of records returned from user list")

	// Check that the records this test created are present. There may be others.
	//
	for n, u := range userRecords {
		assert.Equalf(t, u.Revision, (*listRecs)[n].Revision, "Unexpected difference in revision from create for user %q", n)
		assert.Equalf(t, u.Value, (*listRecs)[n].Value, "Unexpected difference in value from create for user %q", n)
	}

	store.Disconnect()

	store = nil
	return
}
