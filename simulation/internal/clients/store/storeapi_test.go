package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"golang.org/x/crypto/bcrypt"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/namespace"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/admin"
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

type storeApiTestSuite struct {
	testSuiteCore

	store *Store
}

func (ts *storeApiTestSuite) SetupSuite() {
	require := ts.Require()

	ts.testSuiteCore.SetupSuite()

	ts.store = NewStore()
	require.NotNil(ts.store)
}

func (ts *storeApiTestSuite) SetupTest() {
	require := ts.Require()

	require.NoError(ts.utf.Open(ts.T()))
	require.NoError(ts.store.Connect())
}

func (ts *storeApiTestSuite) TearDownTest() {
	ts.store.Disconnect()
	ts.utf.Close()
}

func (ts *storeApiTestSuite) TestCreate() {
	assert  := ts.Assert()
	require := ts.Require()

	userName := admin + "." + tracing.MethodName(1)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	require.NoError(err)

	user := &pb.User{
		Name:         userName,
		PasswordHash: passwordHash,
		UserId:       1,
		Enabled:      true,
		Rights:       &pb.Rights{CanManageAccounts: true},
		NeverDelete:  true,
	}

	revCreate, err := ts.store.CreateWithEncode(context.Background(), namespace.KeyRootUsers, userName, user)
	require.NoError(err)
	require.Less(RevisionInvalid, revCreate)

	revCreate2, err := ts.store.CreateWithEncode(context.Background(), namespace.KeyRootUsers, userName, user)
	require.ErrorIs(errors.ErrStoreAlreadyExists(userName), err)
	assert.Equal(RevisionInvalid, revCreate2)

	userRead := &pb.User{}

	revRead, err := ts.store.ReadWithDecode(context.Background(), namespace.KeyRootUsers, userName, userRead)
	require.NoError(err)
	assert.Equal(revCreate, revRead)
	assert.Equal(user, userRead)
}

func (ts *storeApiTestSuite) TestReadNew() {
	assert  := ts.Assert()
	require := ts.Require()

	userName := admin + "." + tracing.MethodName(1)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	require.NoError(err)

	user := &pb.User{
		Name:         userName,
		PasswordHash: passwordHash,
		UserId:       1,
		Enabled:      true,
		Rights:       &pb.Rights{CanManageAccounts: true},
		NeverDelete:  true,
	}

	revCreate, err := ts.store.CreateWithEncode(context.Background(), namespace.KeyRootUsers, userName, user)
	require.NoError(err)
	assert.Less(RevisionInvalid, revCreate)

	readUser := &pb.User{}

	revRead, err := ts.store.ReadWithDecode(context.Background(), namespace.KeyRootUsers, userName, readUser)
	require.NoError(err)
	assert.Equal(revCreate, revRead)
	assert.Equal(user, readUser)

	readUserString, revReadValue, err := ts.store.Read(context.Background(), namespace.KeyRootUsers, userName)
	require.NoError(err)
	assert.Equal(revCreate, revReadValue)

	readUserValue := &pb.User{}

	err = Decode(*readUserString, readUserValue)

	require.NoError(err)
	assert.Equal(user, readUserValue)
}

func (ts *storeApiTestSuite) TestReadNewValue() {
	assert  := ts.Assert()
	require := ts.Require()

	userName := admin + "." + tracing.MethodName(1)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	require.NoError(err)

	user := &pb.User{
		Name:         userName,
		PasswordHash: passwordHash,
		UserId:       1,
		Enabled:      true,
		Rights:       &pb.Rights{CanManageAccounts: true},
		NeverDelete:  true,
	}

	userValue, err := Encode(user)
	require.NoError(err, "Failed to encode user record")

	revCreate, err := ts.store.Create(context.Background(), namespace.KeyRootUsers, userName, userValue)
	require.NoError(err)
	assert.Less(RevisionInvalid, revCreate)

	readUserValue, revReadValue, err := ts.store.Read(context.Background(), namespace.KeyRootUsers, userName)
	require.NoError(err)
	assert.Equal(revCreate, revReadValue)
	assert.Equal(userValue, *readUserValue)

	readUser := &pb.User{}

	revRead, err := ts.store.ReadWithDecode(context.Background(), namespace.KeyRootUsers, userName, readUser)
	require.NoError(err)
	assert.Equal(revCreate, revRead)
	assert.Equal(user, readUser)

	readUserFromValue := &pb.User{}

	err = Decode(*readUserValue, readUserFromValue)
	require.NoError(err)
	assert.Equal(user, readUserFromValue)
}

func (ts *storeApiTestSuite) TestUpdate() {
	assert  := ts.Assert()
	require := ts.Require()

	userName := admin + "." + tracing.MethodName(1)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminUpdatePassword), bcrypt.DefaultCost)
	require.NoError(err)

	user := &pb.User{
		Name:         userName,
		PasswordHash: passwordHash,
		UserId:       1,
		Enabled:      true,
		Rights:       &pb.Rights{CanManageAccounts: true},
		NeverDelete:  true,
	}

	revCreate, err := ts.store.CreateWithEncode(context.Background(), namespace.KeyRootUsers, userName, user)
	require.NoError(err)
	assert.Less(RevisionInvalid, revCreate)

	userRead := &pb.User{}

	revRead, err := ts.store.ReadWithDecode(context.Background(), namespace.KeyRootUsers, userName, userRead)

	require.NoError(err)
	assert.Equal(revCreate, revRead)
	assert.Equal(user, userRead)

	// Now update the user record and see if the changes made it.
	//
	passwordHash, err = bcrypt.GenerateFromPassword([]byte(adminUpdatePassword2), bcrypt.DefaultCost)
	require.NoError(err)

	userUpdate := &pb.User{
		Name:         userName,
		PasswordHash: passwordHash,
		UserId:       1,
		Enabled:      false,
		Rights:       &pb.Rights{CanManageAccounts: true},
		NeverDelete:  true,
	}

	revUpdate, err := ts.store.UpdateWithEncode(context.Background(), namespace.KeyRootUsers, userName, revRead, userUpdate)
	require.NoError(err)
	assert.Less(revRead, revUpdate)

	userReadUpdate := &pb.User{}

	revReadUpdate, err := ts.store.ReadWithDecode(context.Background(), namespace.KeyRootUsers, userName, userReadUpdate)

	require.NoError(err)
	assert.Equal(revUpdate, revReadUpdate)
	assert.Equal(userUpdate, userReadUpdate)

	// Now try to update with the wrong revision
	//
	userUpdate2 := &pb.User{
		Name:         userName,
		PasswordHash: passwordHash,
		UserId:       1,
		Enabled:      true,
		Rights:       &pb.Rights{CanManageAccounts: false},
		NeverDelete:  true,
	}

	revReadUpdate2, err := ts.store.UpdateWithEncode(context.Background(), namespace.KeyRootUsers, userName, revRead, userUpdate2)

	// See issue #254
	//
	require.ErrorIs(errors.ErrStoreConditionFail{
			Key:       namespace.GetKeyFromUsername(userName),
			Requested: revCreate,
			Condition: string(ConditionRevisionEqual),
			Actual:    revUpdate,
		},
		err)

	assert.Equal(RevisionInvalid, revReadUpdate2)
}

func (ts *storeApiTestSuite) TestUpdateUnconditional() {
	assert  := ts.Assert()
	require := ts.Require()

	userName := admin + "." + tracing.MethodName(1)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminUpdatePassword), bcrypt.DefaultCost)
	require.NoError(err)

	user := &pb.User{
		Name:         userName,
		PasswordHash: passwordHash,
		UserId:       1,
		Enabled:      true,
		Rights:       &pb.Rights{CanManageAccounts: true},
		NeverDelete:  true,
	}

	revCreate, err := ts.store.CreateWithEncode(context.Background(), namespace.KeyRootUsers, userName, user)
	require.NoError(err)
	assert.Less(RevisionInvalid, revCreate)

	userRead := &pb.User{}

	revRead, err := ts.store.ReadWithDecode(context.Background(), namespace.KeyRootUsers, userName, userRead)

	require.NoError(err)
	assert.Equal(revCreate, revRead)
	assert.Equal(user, userRead)

	// Now update the user record and see if the changes made it.
	//
	passwordHash, err = bcrypt.GenerateFromPassword([]byte(adminUpdatePassword2), bcrypt.DefaultCost)
	require.NoError(err)

	userUpdate := &pb.User{
		Name:         userName,
		PasswordHash: passwordHash,
		UserId:       1,
		Enabled:      false,
		Rights:       &pb.Rights{CanManageAccounts: true},
		NeverDelete:  true,
	}

	revUpdate, err := ts.store.UpdateWithEncode(context.Background(), namespace.KeyRootUsers, userName, revRead, userUpdate)
	require.NoError(err)
	assert.Less(revRead, revUpdate)

	userReadUpdate := &pb.User{}

	revReadUpdate, err := ts.store.ReadWithDecode(context.Background(), namespace.KeyRootUsers, userName, userReadUpdate)

	require.NoError(err)
	assert.Equal(revUpdate, revReadUpdate)
	assert.Equal(userUpdate, userReadUpdate)

	// Now try to update with the wrong revision
	//
	userUpdate2 := &pb.User{
		Name:         userName,
		PasswordHash: passwordHash,
		UserId:       1,
		Enabled:      true,
		Rights:       &pb.Rights{CanManageAccounts: false},
		NeverDelete:  true,
	}

	revReadUpdate2, err := ts.store.UpdateWithEncode(context.Background(), namespace.KeyRootUsers, userName, revRead, userUpdate2)

	// See issue #254
	//
	require.ErrorIs(errors.ErrStoreConditionFail{
			Key:       namespace.GetKeyFromUsername(userName),
			Requested: revCreate,
			Condition: string(ConditionRevisionEqual),
			Actual:    revUpdate,
		},
		err)

	assert.Equal(RevisionInvalid, revReadUpdate2)

	// Now try to update unconditionally
	//
	userUpdate3 := &pb.User{
		Name:         userName,
		PasswordHash: passwordHash,
		UserId:       1,
		Enabled:      false,
		Rights:       &pb.Rights{CanManageAccounts: false},
		NeverDelete:  true,
	}

	revReadUpdate3, err := ts.store.UpdateWithEncode(context.Background(), namespace.KeyRootUsers, userName, RevisionInvalid, userUpdate3)
	require.NoError(err)
	assert.Less(revReadUpdate, revReadUpdate3)
}

func (ts *storeApiTestSuite) TestDelete() {
	assert  := ts.Assert()
	require := ts.Require()

	userName := alice + "." + tracing.MethodName(1)
	passWord := alicePassword

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(passWord), bcrypt.DefaultCost)
	require.NoError(err)

	user := &pb.User{
		Name:         userName,
		PasswordHash: passwordHash,
		UserId:       1,
		Enabled:      true,
		Rights:       &pb.Rights{CanManageAccounts: true},
		NeverDelete:  true,
	}

	revCreate, err := ts.store.CreateWithEncode(context.Background(), namespace.KeyRootUsers, userName, user)
	require.NoError(err)
	assert.Less(RevisionInvalid, revCreate)

	userRead := &pb.User{}

	revRead, err := ts.store.ReadWithDecode(context.Background(), namespace.KeyRootUsers, userName, userRead)

	require.NoError(err)
	assert.Equal(revCreate, revRead)
	assert.Equal(user, userRead)

	// Fiurst try to delete using the wrong revision
	//
	revDelete, err := ts.store.Delete(context.Background(), namespace.KeyRootUsers, userName, revRead-1)

	// See issue #254
	//
	require.ErrorIs(errors.ErrStoreConditionFail{
			Key:       namespace.GetKeyFromUsername(userName),
			Requested: revRead-1,
			Condition: string(ConditionRevisionEqual),
			Actual:    revRead,
		},
		err)

	assert.Equal(RevisionInvalid, revDelete)

	// Now delete with the correct revision
	//
	revDelete, err = ts.store.Delete(context.Background(), namespace.KeyRootUsers, userName, revRead)
	require.NoError(err)
	assert.Less(revRead, revDelete)

	// Try to read after delete
	//
	userReread := &pb.User{}

	revReread, err := ts.store.ReadWithDecode(context.Background(), namespace.KeyRootUsers, userName, userReread)
	require.ErrorIs(errors.ErrStoreKeyNotFound(userName), err)
	assert.Equal(RevisionInvalid, revReread)

	// Try to delete a non-existing record.
	//
	revDeleteAgain, err := ts.store.Delete(context.Background(), namespace.KeyRootUsers, userName, revRead)

	// See issue #254
	//
	require.ErrorIs(errors.ErrStoreConditionFail{
			Key:       namespace.GetKeyFromUsername(userName),
			Requested: revRead,
			Condition: string(ConditionRevisionEqual),
			Actual:    0,
		},
		err)
	assert.Equal(RevisionInvalid, revDeleteAgain)
}

func (ts *storeApiTestSuite) TestList() {
	assert  := ts.Assert()
	require := ts.Require()

	type urec struct {
		name string
		pwd  string
	}

	suffix := "." + tracing.MethodName(1)

	userSet := []urec{
		{name: alice + suffix, pwd: alicePassword},
		{name: bob + suffix, pwd: bobPassword},
		{name: eve + suffix, pwd: evePassword},
	}

	users := make(map[string]*pb.User, len(userSet))

	for i, u := range userSet {
		pwdHash, err := bcrypt.GenerateFromPassword([]byte(u.pwd), bcrypt.DefaultCost)
		require.NoError(err)

		users[namespace.GetNormalizedName(u.name)] = &pb.User{
			Name:         u.name,
			PasswordHash: pwdHash,
			UserId:       int64(i + 1),
			Enabled:      true,
			Rights:       &pb.Rights{CanManageAccounts: false},
			NeverDelete:  false,
		}
	}

	revFirstCreate := RevisionInvalid

	userRecords := make(map[string]Record, len(userSet))

	for n, u := range users {
		v, err := Encode(u)
		require.NoError(err)

		revCreate, err := ts.store.Create(context.Background(), namespace.KeyRootUsers, n, v)
		require.NoError(err)
		assert.Less(RevisionInvalid, revCreate)

		userRecords[n] = Record{Revision: revCreate, Value: v}

		if revFirstCreate == RevisionInvalid {
			revFirstCreate = revCreate
		}
	}

	listRecs, listRev, err := ts.store.List(context.Background(), namespace.KeyRootUsers, "")
	require.NoError(err)
	assert.LessOrEqual(revFirstCreate, listRev)

	// Use "less than or equal" relationship to allow for the cases where all the
	// file tests are being executed and there are potentially user records left over
	// from tests running earlier in the set.
	//
	assert.LessOrEqual(len(userRecords), len(*listRecs))

	// Check that the records this test created are present. There may be others.
	//
	for n, u := range userRecords {
		assert.Equal(u.Revision, (*listRecs)[n].Revision)
		assert.Equal(u.Value, (*listRecs)[n].Value)
	}
}

func TestStoreApiTestSuite(t *testing.T) {
	suite.Run(t, new(storeApiTestSuite))
}
