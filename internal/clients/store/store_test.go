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

func TestStoreWriteReadMultiple(t *testing.T) {

	endpoints := defaultEndpoints
	timeoutConnect := defaultTimeoutConnect
	timeoutRequest := defaultTimeoutRequest

	keyValueSetSize := 100

	prefixKey := "TestStoreWriteReadMultiple/Key"
	prefixVal := "TestStoreWriteReadMultiple/Value"

	keySet := make([]string, keyValueSetSize)
	keyValueSet := make([]KeyValue, keyValueSetSize)

	for i := range keySet {
		keySet[i] = fmt.Sprintf("%s%04d", prefixKey, i)
	}

	for i := range keyValueSet {
		keyValueSet[i].key = fmt.Sprintf("%s%04d", prefixKey, i)
		keyValueSet[i].value = fmt.Sprintf("%s%04d", prefixVal, i)
	}

	store, err := New(endpoints, timeoutConnect, timeoutRequest)
	assert.Nilf(t, err, "Failed to allocate new store - error: %v", err)
	assert.NotNilf(t, store, "Failed to get the store as expected - error: %v", err)

	err = store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.WriteMultiple(keyValueSet)
	assert.Nilf(t, err, "Failed to write to store - error: %v", err)

	response, err := store.ReadMultiple(keySet)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.NotNilf(t, response, "Failed to get a response as expected - error: %v", err)

	for i, kv := range response {
		assert.Equal(t, keyValueSet[i].key, kv.key, "Unexpected key - expected: %s received: %s", keyValueSet[i].key, kv.key)
		assert.Equal(t, keyValueSet[i].value, kv.value, "Unexpected value - expected: %s received: %s", keyValueSet[i].value, kv.value)
	}

	err = store.Disconnect()
	assert.Nilf(t, err, "Failed to disconnect from store - error: %v", err)

	store = nil

	return
}

func TestStoreWriteReadWithPrefix(t *testing.T) {

	endpoints := defaultEndpoints
	timeoutConnect := defaultTimeoutConnect
	timeoutRequest := defaultTimeoutRequest

	keyValueSetSize := 100

	prefixKey := "TestStoreWriteReadWithPrefix/Key"
	prefixVal := "TestStoreWriteReadWithPrefix/Value"

	keyValueSet := make([]KeyValue, keyValueSetSize)

	for i := range keyValueSet {
		keyValueSet[i].key = fmt.Sprintf("%s%04d", prefixKey, i)
		keyValueSet[i].value = fmt.Sprintf("%s%04d", prefixVal, i)
	}

	keyValueMap := make(map[string]string, len(keyValueSet))

	for _, kv := range keyValueSet {
		keyValueMap[kv.key] = kv.value
	}

	store, err := New(endpoints, timeoutConnect, timeoutRequest)
	assert.Nilf(t, err, "Failed to allocate new store - error: %v", err)
	assert.NotNilf(t, store, "Failed to get the store as expected - error: %v", err)

	err = store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	for _, kv := range keyValueSet {
		err = store.Write(kv.key, kv.value)
		assert.Nilf(t, err, "Failed to write to store - error: %v key: %v value %v", err, kv.key, kv.value)
	}

	response, err := store.ReadWithPrefix(prefixKey)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.NotNilf(t, response, "Failed to get a response as expected - error: %v", err)

	for _, kv := range response {
		fmt.Printf("%s: %s\n", kv.key, kv.value)
		assert.Equal(t, keyValueMap[kv.key], kv.value, "Unexpected value - expected: %s received: %s", keyValueMap[kv.key], kv.value)
	}

	err = store.Disconnect()
	assert.Nilf(t, err, "Failed to disconnect from store - error: %v", err)

	store = nil

	return
}
