package store

import (
	"testing"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/admin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

const (
	userURI       = "/api/users/"
	admin         = "Admin"
	adminPassword = "AdminPassword"
	alice         = "Alice"
	bob           = "Bob"
	alicePassword = "AlicePassowrd"
	bobPassword   = "BobPassword"
)

func TestUserCreate(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)

	user := &pb.User{
		Name:           admin,
		PasswordHash:   passwordHash,
		UserId:         1,
		Enabled:        true,
		AccountManager: true,
		NeverDelete:    true,
	}

	rev, err := store.UserCreate(user)
	assert.Nilf(t, err, "Failed to create new user %q - error: %v", admin, err)
	assert.Greaterf(t, 0, rev, "Expected new store revision to be greater than zero")

	rev, err = store.UserCreate(user)
	assert.Nilf(t, err, "Unexpected succeeded to (re-)create new user %q - error: %v", admin, err)

	readUser, readRev, err := store.UserRead(admin)

	assert.Nilf(t, err, "Failed to read user %q - error: %v", admin, err)
	assert.Equalf(t, rev, readRev, "Unexpected difference in creation revision vs read revision")
	assert.Equalf(t, user, readUser, "Unexpected difference in creation user record and read user record")

	store.Disconnect()

	store = nil
	return
}
