// Unit tests for the web service store package
//
package store

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	baseURI     string
	initialized bool
)

func commonSetup() error {

	Initialize()
	return nil
}

func TestMain(m *testing.M) {

	commonSetup()

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {

	store, err := NewWithDefaults()

	assert.Nilf(t, err, "Failed to allocate new store - error: %v", err)
	assert.NotNilf(t, store, "Failed to get the store as expected - error: %v", err)

	store = nil

	return
}

func TestInitialize(t *testing.T) {

	endpoints := defaultEndpoints
	timeoutConnect := defaultTimeoutConnect
	timeoutRequest := defaultTimeoutRequest

	store, err := NewWithDefaults()

	assert.Nilf(t, err, "Failed to allocate new store - error: %v", err)
	assert.NotNilf(t, store, "Failed to get the store as expected - error: %v", err)

	err = store.Initialize(endpoints, timeoutConnect, timeoutRequest)

	assert.Nilf(t, err, "Failed to initialize new store - error: %v", err)

	store = nil

	return
}

func TestNewWithArgs(t *testing.T) {

	endpoints := defaultEndpoints
	timeoutConnect := defaultTimeoutConnect
	timeoutRequest := defaultTimeoutRequest

	store, err := New(endpoints, timeoutConnect, timeoutRequest)

	assert.Nilf(t, err, "Failed to allocate new store - error: %v", err)
	assert.NotNilf(t, store, "Failed to get the store as expected - error: %v", err)

	store = nil

	return
}

func TestStoreConnectDisconnect(t *testing.T) {

	endpoints := defaultEndpoints
	timeoutConnect := defaultTimeoutConnect
	timeoutRequest := defaultTimeoutRequest

	store, err := New(endpoints, timeoutConnect, timeoutRequest)

	assert.Nilf(t, err, "Failed to allocate new store - error: %v", err)
	assert.NotNilf(t, store, "Failed to get the store as expected - error: %v", err)

	err = store.Connect()

	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.Disconnect()

	assert.Nilf(t, err, "Failed to disconnect from store - error: %v", err)

	store = nil

	return
}

func TestStoreWriteRead(t *testing.T) {

	endpoints := defaultEndpoints
	timeoutConnect := defaultTimeoutConnect
	timeoutRequest := defaultTimeoutRequest

	key := "TestStoreWriteRead/Key"
	value := "TestStoreWriteRead/Value"

	store, err := New(endpoints, timeoutConnect, timeoutRequest)
	assert.Nilf(t, err, "Failed to allocate new store - error: %v", err)
	assert.NotNilf(t, store, "Failed to get the store as expected - error: %v", err)

	err = store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.Write(key, value)
	assert.Nilf(t, err, "Failed to write to store - error: %v", err)

	response, err := store.Read(key)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.NotNilf(t, response, "Failed to get a response as expected - error: %v", err)
	assert.Equal(t, value, response, "response does not match written value - value: %v response: %v", value, response)

	err = store.Disconnect()
	assert.Nilf(t, err, "Failed to disconnect from store - error: %v", err)

	store = nil

	return
}

func TestStoreWriteReadMultippleWithPrefix(t *testing.T) {

	endpoints := defaultEndpoints
	timeoutConnect := defaultTimeoutConnect
	timeoutRequest := defaultTimeoutRequest

	prefixKey := "TestStoreWriteRead/Key"
	prefixVal := "TestStoreWriteRead/Value"

	store, err := New(endpoints, timeoutConnect, timeoutRequest)
	assert.Nilf(t, err, "Failed to allocate new store - error: %v", err)
	assert.NotNilf(t, store, "Failed to get the store as expected - error: %v", err)

	err = store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.Write(prefixKey, prefixVal)
	assert.Nilf(t, err, "Failed to write to store - error: %v", err)

	for i := 0; i < 10; i++ {

		key := fmt.Sprintf("%s/%d", prefixKey, i)
		val := fmt.Sprintf("%s/%d", prefixVal, i)

		err = store.Write(key, val)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
	}

	response, err := store.ReadMultipleWithPrefix(prefixKey)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.NotNilf(t, response, "Failed to get a response as expected - error: %v", err)

	for _, vp := range response {
		fmt.Printf("%s: %s\n", vp.key, vp.value)
	}
	return
}
