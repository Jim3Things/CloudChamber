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

	err = store.Connect()
	assert.NotNilf(t, err, "Unexpectedly connected to store again - error: %v", err)
	assert.Equal(t, ErrStoreConnected, err, "Unexpected error response - expected: %v got: %v", ErrStoreConnected, err)

	store.Disconnect()

	// Try a second disconnect. Benign but should trigger different execution
	// path for coverage numbers.
	//
	store.Disconnect()

	store = nil

	return
}

func TestStoreConnectDisconnectWithInitialize(t *testing.T) {

	endpoints := defaultEndpoints
	timeoutConnect := defaultTimeoutConnect
	timeoutRequest := defaultTimeoutRequest

	store, err := New(endpoints, timeoutConnect, timeoutRequest)
	assert.Nilf(t, err, "Failed to allocate new store - error: %v", err)
	assert.NotNilf(t, store, "Failed to get the store as expected - error: %v", err)

	err = store.Initialize(endpoints, timeoutConnect, timeoutRequest)
	assert.Nilf(t, err, "Failed to re-initialize store - error: %v", err)

	err = store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.Initialize(endpoints, timeoutConnect, timeoutRequest)
	assert.NotNilf(t, err, "Unexpectedly re-initialized store after connect - error: %v", err)
	assert.Equal(t, ErrStoreConnected, err, "Unexpected error response - expected: %v got: %v", ErrStoreConnected, err)

	store.Disconnect()

	// Try a second disconnect. Benign but should trigger different execution
	// path for coverage numbers.
	//
	store.Disconnect()

	store = nil

	return
}

func TestStoreConnectDisconnectWithSet(t *testing.T) {

	endpoints := defaultEndpoints
	timeoutConnect := defaultTimeoutConnect
	timeoutRequest := defaultTimeoutRequest

	store, err := New(endpoints, timeoutConnect, timeoutRequest)
	assert.Nilf(t, err, "Failed to allocate new store - error: %v", err)
	assert.NotNilf(t, store, "Failed to get the store as expected - error: %v", err)

	err = store.SetAddress(endpoints)
	assert.Nilf(t, err, "Failed to update the address - error: %v", err)

	err = store.SetTimeoutConnect(timeoutConnect)
	assert.Nilf(t, err, "Failed to update the connect timeout - error: %v", err)

	err = store.SetTimeoutRequest(timeoutRequest)
	assert.Nilf(t, err, "Failed to update the request timeout - error: %v", err)

	err = store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.SetAddress(endpoints)
	assert.NotNilf(t, err, "Unexpectedly succeeded to update the address - error: %v", err)
	assert.Equal(t, ErrStoreConnected, err, "Unexpected error response - expected: %v got: %v", ErrStoreConnected, err)

	err = store.SetTimeoutConnect(timeoutConnect)
	assert.NotNilf(t, err, "Unexpectedly succeeded to  to update the connect timeout - error: %v", err)
	assert.Equal(t, ErrStoreConnected, err, "Unexpected error response - expected: %v got: %v", ErrStoreConnected, err)

	err = store.SetTimeoutRequest(timeoutRequest)
	assert.Nilf(t, err, "Failed to update the request timeout - error: %v", err)

	err = store.Connect()
	assert.NotNilf(t, err, "Unexpectedly connected to store again - error: %v", err)
	assert.Equal(t, ErrStoreConnected, err, "Unexpected error response - expected: %v got: %v", ErrStoreConnected, err)

	store.Disconnect()

	// Try a second disconnect. Benign but should trigger different execution
	// path for coverage numbers.
	//
	store.Disconnect()

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

	// Look for a name we do not expect to be present
	//
	invalidKey := key + "invalidname"
	response, err := store.Read(invalidKey)
	assert.NotNilf(t, err, "Succeeded to read non-existing key/value from store - error: %v key: %v value: %v", err, invalidKey, string(response))
	assert.Nilf(t, response, "Failed to get a nil response as expected - error: %v key: %v value: %v", err, invalidKey, string(response))

	// Now try to read a key which should be there.
	//
	response, err = store.Read(key)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.NotNilf(t, response, "Failed to get a response as expected - error: %v", err)
	assert.Equal(t, value, string(response), "response does not match written value - value: %v response: %v", value, string(response))

	store.Disconnect()

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
	keyValueSet := make([]KeyValueArg, keyValueSetSize)

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
		kvValue := string(kv.value)
		fmt.Printf("[%v/%v] %v: %v\n", i, len(response), kv.key, kvValue)
		assert.Equal(t, keyValueSet[i].key, kv.key, "Unexpected key - expected: %s received: %s", keyValueSet[i].key, kv.key)
		assert.Equal(t, keyValueSet[i].value, kvValue, "Unexpected value - expected: %s received: %s", keyValueSet[i].value, kvValue)
	}

	store.Disconnect()

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

	keyValueSet := make([]KeyValueArg, keyValueSetSize)

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

	err = store.WriteMultiple(keyValueSet)
	assert.Nilf(t, err, "Failed to write to store - error: %v", err)

	// Look for a prefix name we do not expect to be present.
	//
	// We expect success with a non-nil but empty set.
	//
	invalidKey := prefixKey + "invalidname"
	response, err := store.ReadWithPrefix(invalidKey)
	assert.Nilf(t, err, "Succeeded to read non-existing prefix key from store - error: %v prefixKey: %v ", err, invalidKey)
	assert.NotNilf(t, response, "Failed to get a non-nil response as expected - error: %v prefixKey: %v", err, invalidKey)
	assert.Equal(t, 0, len(response), "Got more results than expected")

	if nil != response && len(response) > 0 {
		for i, kv := range response {
			fmt.Printf("Unexpected key/value pair [%v/%v] key: %v value: %v\n", i, len(response), kv.key, string(kv.value))
		}
	}

	// Now look for a set of prefixed key/value pairs which we do expect to be present.
	//
	response, err = store.ReadWithPrefix(prefixKey)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.NotNilf(t, response, "Failed to get a response as expected - error: %v", err)
	assert.Equal(t, len(keyValueSet), len(response), "Failed to get the expected number of response values")

	for i, kv := range response {
		kvValue := string(kv.value)
		fmt.Printf("[%v/%v] %v: %v\n", i, len(response), kv.key, kvValue)
		assert.Equal(t, keyValueSet[i].key, kv.key, "Unexpected key - expected: %s received: %s", keyValueSet[i].key, kv.key)
		assert.Equal(t, keyValueMap[kv.key], kvValue, "Unexpected value - expected: %s received: %s", keyValueMap[kv.key], kvValue)
	}

	store.Disconnect()

	store = nil

	return
}

func TestStoreWriteDelete(t *testing.T) {

	endpoints := defaultEndpoints
	timeoutConnect := defaultTimeoutConnect
	timeoutRequest := defaultTimeoutRequest

	key := "TestStoreWriteDelete/Key"
	value := "TestStoreWriteDelete/Value"

	store, err := New(endpoints, timeoutConnect, timeoutRequest)
	assert.Nilf(t, err, "Failed to allocate new store - error: %v", err)
	assert.NotNilf(t, store, "Failed to get the store as expected - error: %v", err)

	err = store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.Write(key, value)
	assert.Nilf(t, err, "Failed to write to store - error: %v key: %v value: %v", err, key, value)

	// Delete the key we just wrote
	//
	err = store.Delete(key)
	assert.Nilf(t, err, "Failed to delete key from store - error: %v key: %v", err, key)

	// Try to delete the key we just wrote a second time
	//
	err = store.Delete(key)
	assert.NotNilf(t, err, "Unexpectedly deleted the key from store for a second time - error: %v key: %v", err, key)

	// Try to delete a name we do not expect to be present
	//
	invalidKey := key + "invalidname"
	err = store.Delete(invalidKey)
	assert.NotNilf(t, err, "Succeeded to delete a non-existing key/value from store - error: %v key: %v", err, invalidKey)

	store.Disconnect()

	store = nil

	return
}

func TestStoreWriteDeleteMultiple(t *testing.T) {

	endpoints := defaultEndpoints
	timeoutConnect := defaultTimeoutConnect
	timeoutRequest := defaultTimeoutRequest

	keyValueSetSize := 100

	prefixKey := "TestStoreWriteDeleteMultiple/Key"
	prefixVal := "TestStoreWriteDeleteMultiple/Value"

	keySet := make([]string, keyValueSetSize)
	keyValueSet := make([]KeyValueArg, keyValueSetSize)

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

	err = store.DeleteMultiple(keySet)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)

	store.Disconnect()

	store = nil

	return
}

func TestStoreWriteDeleteWithPrefix(t *testing.T) {

	endpoints := defaultEndpoints
	timeoutConnect := defaultTimeoutConnect
	timeoutRequest := defaultTimeoutRequest

	keyValueSetSize := 100

	prefixKey := "TestStoreWriteDeleteWithPrefix/Key"
	prefixVal := "TestStoreWriteDeleteWithPrefix/Value"

	keyValueSet := make([]KeyValueArg, keyValueSetSize)

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

	err = store.WriteMultiple(keyValueSet)
	assert.Nilf(t, err, "Failed to write to store - error: %v prefixKey: %v", err, prefixKey)

	err = store.DeleteWithPrefix(prefixKey)
	assert.Nilf(t, err, "Failed to delete the prefix keys from the store - error: %v prefixKey: %v", err, prefixKey)

	err = store.DeleteWithPrefix(prefixKey)
	assert.Nilf(t, err, "Unexpected error when attmepting to delete prefix keys from store for a second - error: %v prefixKey: %v", err, prefixKey)

	store.Disconnect()

	store = nil

	return
}

func TestStoreWriteReadDeleteWithoutConnect(t *testing.T) {

	endpoints := defaultEndpoints
	timeoutConnect := defaultTimeoutConnect
	timeoutRequest := defaultTimeoutRequest

	key := "TestStoreWriteReadDeleteWithoutConnect/Key"
	value := "TestStoreWriteReadDeleteWithoutConnect/Value"

	keySet := make([]string, 1)
	keySet[0] = key

	keyValueSet := make([]KeyValueArg, 1)
	keyValueSet[0].key = key
	keyValueSet[0].value = value

	store, err := New(endpoints, timeoutConnect, timeoutRequest)
	assert.Nilf(t, err, "Failed to allocate new store - error: %v", err)
	assert.NotNilf(t, store, "Failed to get the store as expected - error: %v", err)

	err = store.Write(key, value)
	assert.NotNilf(t, err, "Unexpectedly succeeded to write to store - error: %v", err)
	assert.Equal(t, ErrStoreNotConnected, err, "Unexpected error response - expected: %v got: %v", ErrStoreNotConnected, err)

	err = store.WriteMultiple(keyValueSet)
	assert.NotNilf(t, err, "Unexpectedly succeeded to write to store - error: %v", err)
	assert.Equal(t, ErrStoreNotConnected, err, "Unexpected error response - expected: %v got: %v", ErrStoreNotConnected, err)

	_, err = store.Read(key)
	assert.NotNilf(t, err, "Unexpectedly succeeded to read from store - error: %v", err)
	assert.Equal(t, ErrStoreNotConnected, err, "Unexpected error response - expected: %v got: %v", ErrStoreNotConnected, err)

	_, err = store.ReadMultiple(keySet)
	assert.NotNilf(t, err, "Unexpectedly succeeded to read from store - error: %v", err)
	assert.Equal(t, ErrStoreNotConnected, err, "Unexpected error response - expected: %v got: %v", ErrStoreNotConnected, err)

	_, err = store.ReadWithPrefix(key)
	assert.NotNilf(t, err, "Unexpectedly succeeded to read from store - error: %v", err)
	assert.Equal(t, ErrStoreNotConnected, err, "Unexpected error response - expected: %v got: %v", ErrStoreNotConnected, err)

	err = store.Delete(key)
	assert.NotNilf(t, err, "Unexpectedly succeeded to delete from store - error: %v", err)
	assert.Equal(t, ErrStoreNotConnected, err, "Unexpected error response - expected: %v got: %v", ErrStoreNotConnected, err)

	err = store.DeleteMultiple(keySet)
	assert.NotNilf(t, err, "Unexpectedly succeeded to delete from store - error: %v", err)
	assert.Equal(t, ErrStoreNotConnected, err, "Unexpected error response - expected: %v got: %v", ErrStoreNotConnected, err)

	err = store.DeleteWithPrefix(key)
	assert.NotNilf(t, err, "Unexpectedly succeeded to delete from store - error: %v", err)
	assert.Equal(t, ErrStoreNotConnected, err, "Unexpected error response - expected: %v got: %v", ErrStoreNotConnected, err)

	store = nil

	return
}

func TestStoreSetWatch(t *testing.T) {

	key := "TestStoreSetWatch/Key"

	store, err := New(defaultEndpoints, defaultTimeoutConnect, defaultTimeoutRequest)
	assert.Nilf(t, err, "Failed to allocate new store - error: %v", err)
	assert.NotNilf(t, store, "Failed to get the store as expected - error: %v", err)

	err = store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.SetWatch(key)
	assert.NotNilf(t, err, "Unexpectedly succeeded setting a watch point - error: %v", err)
	assert.Equal(t, ErrStoreNotImplemented, err, "Unexpected error response - expected: %v got: %v", ErrStoreNotImplemented, err)

	store.Disconnect()

	store = nil

	return
}

func TestStoreSetWatchMultiple(t *testing.T) {

	keySet := []string{"TestStoreSetWatchMultiple/Key"}

	store, err := New(defaultEndpoints, defaultTimeoutConnect, defaultTimeoutRequest)
	assert.Nilf(t, err, "Failed to allocate new store - error: %v", err)
	assert.NotNilf(t, store, "Failed to get the store as expected - error: %v", err)

	err = store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.SetWatchMultiple(keySet)
	assert.NotNilf(t, err, "Unexpectedly succeeded setting a watch point - error: %v", err)
	assert.Equal(t, ErrStoreNotImplemented, err, "Unexpected error response - expected: %v got: %v", ErrStoreNotImplemented, err)

	store.Disconnect()

	store = nil

	return
}

func TestStoreSetWatchPrefix(t *testing.T) {

	key := "TestStoreSetWatchPrefix/Key"

	store, err := New(defaultEndpoints, defaultTimeoutConnect, defaultTimeoutRequest)
	assert.Nilf(t, err, "Failed to allocate new store - error: %v", err)
	assert.NotNilf(t, store, "Failed to get the store as expected - error: %v", err)

	err = store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.SetWatchWithPrefix(key)
	assert.NotNilf(t, err, "Unexpectedly succeeded setting a watch point - error: %v", err)
	assert.Equal(t, ErrStoreNotImplemented, err, "Unexpected error response - expected: %v got: %v", ErrStoreNotImplemented, err)

	store.Disconnect()

	store = nil

	return
}
