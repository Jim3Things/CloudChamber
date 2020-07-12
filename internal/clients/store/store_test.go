// Unit tests for the web service store package
//
package store

import (
	"context"
	"fmt"
	"testing"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
	"github.com/stretchr/testify/assert"
)

const revStoreInitial = int64(0)

func TestNew(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	store := NewStore()

	assert.NotNilf(t, store, "Failed to get the store as expected")

	store = nil

	return
}

func TestInitialize(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	store := NewStore()

	assert.NotNilf(t, store, "Failed to get the store as expected")
	assert.Equal(t, getDefaultEndpoints(), store.Endpoints, "Mismatch in initialization of endpoints")
	assert.Equal(t, getDefaultTimeoutConnect(), store.TimeoutConnect, "Mismatch in initialization of connection timeout")
	assert.Equal(t, getDefaultTimeoutRequest(), store.TimeoutRequest, "Mismatch in initialization of request timeout")
	assert.Equal(t, getDefaultTraceFlags(), store.TraceFlags, "Mismatch in initialization of trace flags")
	assert.Equal(t, getDefaultNamespaceSuffix(), store.NamespaceSuffix, "Mismatch in initialization of namespace suffix")

	endpoints := []string{"localhost:8080", "localhost:8181"}
	timeoutConnect := getDefaultTimeoutConnect() * 2
	timeoutRequest := getDefaultTimeoutRequest() * 3
	traceFlags := traceFlagEnabled
	namespaceSuffix := getDefaultNamespaceSuffix() + "/Suffix"

	err := store.Initialize(endpoints, timeoutConnect, timeoutRequest, traceFlags, namespaceSuffix)

	assert.Nilf(t, err, "Failed to initialize new store - error: %v", err)
	assert.Equal(t, endpoints, store.Endpoints, "Mismatch in initialization of endpoints")
	assert.Equal(t, timeoutConnect, store.TimeoutConnect, "Mismatch in initialization of connection timeout")
	assert.Equal(t, timeoutRequest, store.TimeoutRequest, "Mismatch in initialization of request timeout")
	assert.Equal(t, traceFlags, store.TraceFlags, "Mismatch in initialization of trace flags")
	assert.Equal(t, namespaceSuffix, store.NamespaceSuffix, "Mismatch in initialization of namespace suffix")

	store = nil

	return
}

func TestNewWithArgs(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	// Use non-default values to ensure we get what we asked for and not the defaults.
	//
	endpoints := []string{"localhost:8282", "localhost:8383"}
	timeoutConnect := getDefaultTimeoutConnect() * 4
	timeoutRequest := getDefaultTimeoutRequest() * 5
	traceFlags := traceFlagExpandResults
	namespaceSuffix := getDefaultNamespaceSuffix()

	store := New(endpoints, timeoutConnect, timeoutRequest, traceFlags, namespaceSuffix)

	assert.NotNilf(t, store, "Failed to get the store as expected")
	assert.Equal(t, endpoints, store.Endpoints, "Mismatch in initialization of endpoints")
	assert.Equal(t, timeoutConnect, store.TimeoutConnect, "Mismatch in initialization of connection timeout")
	assert.Equal(t, timeoutRequest, store.TimeoutRequest, "Mismatch in initialization of request timeout")
	assert.Equal(t, traceFlags, store.TraceFlags, "Mismatch in initialization of trace flags")
	assert.Equal(t, namespaceSuffix, store.NamespaceSuffix, "Mismatch in initialization of namespace suffix")

	store = nil

	return
}

func TestStoreSetAndGet(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	store := NewStore()

	assert.NotNilf(t, store, "Failed to get the store as expected")
	assert.Equal(t, getDefaultEndpoints(), store.Endpoints, "Mismatch in initialization of endpoints")
	assert.Equal(t, getDefaultTimeoutConnect(), store.TimeoutConnect, "Mismatch in initialization of connection timeout")
	assert.Equal(t, getDefaultTimeoutRequest(), store.TimeoutRequest, "Mismatch in initialization of request timeout")
	assert.Equal(t, getDefaultTraceFlags(), store.TraceFlags, "Mismatch in initialization of trace flags")
	assert.Equal(t, getDefaultNamespaceSuffix(), store.NamespaceSuffix, "Mismatch in initialization of namespace suffix")

	assert.Equal(t, store.Endpoints, store.GetAddress(), "Mismatch in fetch of endpoints")
	assert.Equal(t, store.TimeoutConnect, store.GetTimeoutConnect(), "Mismatch in fetch of connection timeout")
	assert.Equal(t, store.TimeoutRequest, store.GetTimeoutRequest(), "Mismatch in fetch of request timeout")
	assert.Equal(t, store.TraceFlags, store.GetTraceFlags(), "Mismatch in fetch of trace flags")

	endpoints := []string{"localhost:8484", "localhost:8585"}
	timeoutConnect := getDefaultTimeoutConnect() * 6
	timeoutRequest := getDefaultTimeoutRequest() * 7
	traceFlags := traceFlagExpandResults
	namespaceSuffix := getDefaultNamespaceSuffix() + "/Suffix2"

	err := store.Initialize(endpoints, timeoutConnect, timeoutRequest, traceFlags, namespaceSuffix)

	assert.Nilf(t, err, "Failed to update new store - error: %v", err)
	assert.Equal(t, endpoints, store.Endpoints, "Mismatch in update of endpoints")
	assert.Equal(t, timeoutConnect, store.TimeoutConnect, "Mismatch in update of connection timeout")
	assert.Equal(t, timeoutRequest, store.TimeoutRequest, "Mismatch in update of request timeout")
	assert.Equal(t, traceFlags, store.TraceFlags, "Mismatch in update of trace flags")
	assert.Equal(t, namespaceSuffix, store.NamespaceSuffix, "Mismatch in update of namespace suffix")

	assert.Equal(t, store.Endpoints, store.GetAddress(), "Mismatch in re-fetch of endpoints")
	assert.Equal(t, store.TimeoutConnect, store.GetTimeoutConnect(), "Mismatch in re-fetch of connection timeout")
	assert.Equal(t, store.TimeoutRequest, store.GetTimeoutRequest(), "Mismatch in re-fetch of request timeout")
	assert.Equal(t, store.TraceFlags, store.GetTraceFlags(), "Mismatch in re-fetch of trace flags")
	assert.Equal(t, store.NamespaceSuffix, store.GetNamespaceSuffix(), "Mismatch in re-fetch of namespace suffix")
}

func TestStoreConnectDisconnect(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
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
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Initialize(
		getDefaultEndpoints(),
		getDefaultTimeoutConnect(),
		getDefaultTimeoutRequest(),
		getDefaultTraceFlags(),
		getDefaultNamespaceSuffix())
	assert.Nilf(t, err, "Failed to re-initialize store - error: %v", err)

	err = store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.Initialize(
		getDefaultEndpoints(),
		getDefaultTimeoutConnect(),
		getDefaultTimeoutRequest(),
		getDefaultTraceFlags(),
		getDefaultNamespaceSuffix())
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
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	endpoints := getDefaultEndpoints()
	timeoutConnect := getDefaultTimeoutConnect()
	timeoutRequest := getDefaultTimeoutRequest()
	traceFlags := getDefaultTraceFlags()
	namespaceSuffix := getDefaultNamespaceSuffix()

	store := New(endpoints, timeoutConnect, timeoutRequest, traceFlags, namespaceSuffix)
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.SetAddress(endpoints)
	assert.Nilf(t, err, "Failed to update the address - error: %v", err)

	err = store.SetTimeoutConnect(timeoutConnect)
	assert.Nilf(t, err, "Failed to update the connect timeout - error: %v", err)

	err = store.SetTimeoutRequest(timeoutRequest)
	assert.Nilf(t, err, "Failed to update the request timeout - error: %v", err)

	err = store.SetNamespaceSuffix(namespaceSuffix)
	assert.Nilf(t, err, "Failed to update the namespace suffix - error: %v", err)

	store.SetTraceFlags(0)
	store.SetTraceFlags(traceFlagEnabled)
	store.SetTraceFlags(traceFlagExpandResults)
	store.SetTraceFlags(traceFlagEnabled | traceFlagExpandResults)

	err = store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.SetAddress(endpoints)
	assert.NotNilf(t, err, "Unexpectedly succeeded to update the address - error: %v", err)
	assert.Equal(t, ErrStoreConnected, err, "Unexpected error response - expected: %v got: %v", ErrStoreConnected, err)

	err = store.SetTimeoutConnect(timeoutConnect)
	assert.NotNilf(t, err, "Unexpectedly succeeded to update the connect timeout - error: %v", err)
	assert.Equal(t, ErrStoreConnected, err, "Unexpected error response - expected: %v got: %v", ErrStoreConnected, err)

	err = store.SetTimeoutRequest(timeoutRequest)
	assert.Nilf(t, err, "Failed to update the request timeout - error: %v", err)

	err = store.SetNamespaceSuffix(namespaceSuffix)
	assert.NotNilf(t, err, "Unexpectedly succeeded to update the namespace suffix - error: %v", err)
	assert.Equal(t, ErrStoreConnected, err, "Unexpected error response - expected: %v got: %v", ErrStoreConnected, err)

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
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	key := "TestStoreWriteRead/Key"
	value := "TestStoreWriteRead/Value"

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.Write(key, value)
	assert.Nilf(t, err, "Failed to write to store - error: %v", err)

	// Look for a name we do not expect to be present
	//
	invalidKey := key + "invalidname"
	response, err := store.Read(invalidKey)
	assert.NotNilf(t, err, "Succeeded to read non-existing key/value from store - error: %v key: %v value: %v", err, invalidKey, string(response))
	assert.Equal(t, ErrStoreKeyNotFound(invalidKey), err, "unexpected failure when looking for an invalid key - error %v", err)
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
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	keyValueSet := testGenerateKeyValueSet(keySetSize, "TestStoreWriteReadMultiple")
	keySet := testGenerateKeySetFromKeyValueSet(keyValueSet)

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.WriteMultiple(keyValueSet)
	assert.Nilf(t, err, "Failed to write to store - error: %v", err)

	response, err := store.ReadMultiple(keySet)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.NotNilf(t, response, "Failed to get a response as expected - error: %v", err)

	for i, kv := range response {
		kvValue := string(kv.value)
		if store.trace(traceFlagExpandResultsInTest) {
			t.Logf("[%v/%v] %v: %v", i, len(response), kv.key, kvValue)
		}
		assert.Equal(t, keyValueSet[i].key, kv.key, "Unexpected key - expected: %s received: %s", keyValueSet[i].key, kv.key)
		assert.Equal(t, keyValueSet[i].value, kvValue, "Unexpected value - expected: %s received: %s", keyValueSet[i].value, kvValue)
	}

	store.Disconnect()

	store = nil

	return
}

func TestStoreWriteReadWithPrefix(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	prefix := "TestStoreWriteReadWithPrefix"
	prefixKey := prefix + "/Key"

	keyValueSet := testGenerateKeyValueSet(keySetSize, prefix)
	keyValueMap := testGenerateKeyValueMapFromKeyValueSet(keyValueSet)

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.WriteMultiple(keyValueSet)
	assert.Nilf(t, err, "Failed to write to store - error: %v", err)

	// Look for a prefix name we do not expect to be present.
	//
	// We expect success with a non-nil but empty set.
	//
	invalidKey := prefixKey + "Invalidname"
	response, err := store.ReadWithPrefix(context.Background(), invalidKey)
	assert.Nilf(t, err, "Succeeded to read non-existing prefix key from store - error: %v prefixKey: %v", err, invalidKey)
	assert.NotNilf(t, response, "Failed to get a non-nil response as expected - error: %v prefixKey: %v", err, invalidKey)
	assert.Equal(t, 0, len(response.Records), "Got more results than expected")

	if nil != response && len(response.Records) > 0 {
		for k, r := range response.Records {
			t.Logf("Unexpected key/value pair key: %v value: %v", k, r.Value)
		}
	}

	// Now look for a set of prefixed key/value pairs which we do expect to be present.
	//
	response, err = store.ReadWithPrefix(context.Background(), prefixKey)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.NotNilf(t, response, "Failed to get a response as expected - error: %v", err)
	assert.Equal(t, len(keyValueSet), len(response.Records), "Failed to get the expected number of response values")

	// Check we got records for each key we asked for
	//
	for i, kv := range keyValueSet {
		if store.trace(traceFlagExpandResultsInTest) {
			t.Logf("[%v/%v] %v: Expected: %v Actual: %v", i, len(keyValueSet), kv.key, kv.value, response.Records[kv.key].Value)
		}

		rec, present := response.Records[kv.key]

		assert.Truef(t, present, "Missing record for key - %v", kv.key)

		if present {
			assert.Equal(t, kv.value, rec.Value, "Unexpected value - expected: %q received: %q", kv.value, rec.Value)
		}
	}

	// Check we ONLY got records for the keys we asked for
	//
	for k, r := range response.Records {
		val, present := keyValueMap[k]
		assert.Truef(t, present, "Extra key: %v record: %v", k, r)
		if present {
			assert.Equalf(t, val, r.Value, "key: %v Expected: %q Actual %q", val, r.Value)
		}
	}

	store.Disconnect()

	store = nil

	return
}

func TestStoreWriteDelete(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	key := "TestStoreWriteDelete/Key"
	value := "TestStoreWriteDelete/Value"

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
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
	assert.Equal(t, ErrStoreKeyNotFound(key), err, "unexpected failure when looking for a previously deleted key - error %v", err)

	// Try to delete a name we do not expect to be present
	//
	invalidKey := key + "invalidname"
	err = store.Delete(invalidKey)
	assert.NotNilf(t, err, "Succeeded to delete a non-existing key/value from store - error: %v key: %v", err, invalidKey)
	assert.Equal(t, ErrStoreKeyNotFound(invalidKey), err, "unexpected failure when looking for an invalid key - error %v", err)

	store.Disconnect()

	store = nil

	return
}

func TestStoreWriteDeleteMultiple(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	keyValueSet := testGenerateKeyValueSet(keySetSize, "TestStoreWriteReadMultipleTxn")
	keySet := testGenerateKeySetFromKeyValueSet(keyValueSet)

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
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
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	keyValueSetSize := keySetSize

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

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
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
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	key := "TestStoreWriteReadDeleteWithoutConnect/Key"
	value := "TestStoreWriteReadDeleteWithoutConnect/Value"

	keySet := make([]string, 1)
	keySet[0] = key

	keyValueSet := make([]KeyValueArg, 1)
	keyValueSet[0].key = key
	keyValueSet[0].value = value

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Write(key, value)
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

	_, err = store.ReadWithPrefix(context.Background(), key)
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
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	key := "TestStoreSetWatch/Key"

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.SetWatch(key)
	assert.NotNilf(t, err, "Unexpectedly succeeded setting a watch point - error: %v", err)
	assert.Equal(t, ErrStoreNotImplemented("SetWatch"), err, "Unexpected error response - expected: %v got: %v", ErrStoreNotImplemented("SetWatch"), err)

	store.Disconnect()

	store = nil

	return
}

func TestStoreSetWatchMultiple(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	keySet := []string{"TestStoreSetWatchMultiple/Key"}

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.SetWatchMultiple(keySet)
	assert.NotNilf(t, err, "Unexpectedly succeeded setting a watch point - error: %v", err)
	assert.Equal(t, ErrStoreNotImplemented("SetWatchMultiple"), err, "Unexpected error response - expected: %v got: %v", ErrStoreNotImplemented("SetWatchMultiple"), err)

	store.Disconnect()

	store = nil

	return
}

func TestStoreSetWatchPrefix(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	key := "TestStoreSetWatchPrefix/Key"

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.SetWatchWithPrefix(key)
	assert.NotNilf(t, err, "Unexpectedly succeeded setting a watch point - error: %v", err)
	assert.Equal(t, ErrStoreNotImplemented("SetWatchWithPrefix"), err, "Unexpected error response - expected: %v got: %v", ErrStoreNotImplemented("SetWatchWithPrefix"), err)

	store.Disconnect()

	store = nil

	return
}

func TestStoreGetMemberList(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	response, err := store.GetClusterMembers()
	assert.Nilf(t, err, "Failed to fetch member list from store - error: %v", err)
	assert.NotNilf(t, response, "Failed to get a response as expected - error: %v", err)
	assert.GreaterOrEqual(t, 1, len(response.Members), "Failed to get the minimum number of response values")

	for i, node := range response.Members {
		t.Logf("node [%v] Id: %v Name: %v", i, node.ID, node.Name)
		for i, url := range node.ClientURLs {
			t.Logf("  client [%v] URL: %v", i, url)
		}
		for i, url := range node.PeerURLs {
			t.Logf("  peer [%v] URL: %v", i, url)
		}
	}

	store.Disconnect()

	store = nil

	return
}

func TestStoreSyncClusterConnections(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.UpdateClusterConnections()
	assert.Nilf(t, err, "Failed to update cluster connections - error: %v", err)

	store.Disconnect()

	store = nil

	return
}

func TestStoreWriteReadMultipleTxn(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	keyValueSet := testGenerateKeyValueSet(keySetSize, "TestStoreWriteReadMultipleTxn")
	keySet := testGenerateKeySetFromKeyValueSet(keyValueSet)

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.WriteMultiple(keyValueSet)
	assert.Nilf(t, err, "Failed to write to store - error: %v", err)

	recordKeySet := RecordKeySet{"TestStoreWriteReadMultipleTxn", keySet}

	readResponse, err := store.ReadMultipleTxn(context.Background(), recordKeySet)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.Lessf(t, revStoreInitial, readResponse.Revision, "Unexpected value for store revision on transaction completion")
	assert.Equalf(t, len(recordKeySet.Keys), len(readResponse.Records), "Unexpected numbers of records returned")

	for _, kv := range keyValueSet {
		record := readResponse.Records[kv.key]
		assert.NotNilf(t, record, "Failed to retrieve record for key %q", kv.key)
		assert.Lessf(t, revStoreInitial, readResponse.Records[kv.key].Revision, "Unexpected revision for record %q retrieved for key %q", record, kv.key)
		assert.Equalf(t, kv.value, readResponse.Records[kv.key].Value, "Unexpected value for record %q retrieved for key %q", record, kv.key)
	}

	store.Disconnect()

	store = nil

	return
}

func TestStoreWriteMultipleTxn(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	testName := "TestStoreWriteMultipleTxn"

	keyValueSet := testGenerateKeyValueSet(keySetSize, testName)
	keySet := testGenerateKeySetFromKeyValueSet(keyValueSet)
	recordUpdateSet := testGenerateRecordUpdateSetFromKeyValueSet(keyValueSet, testName, ConditionUnconditional)
	recordReadSet := RecordKeySet{Label: testName, Keys: keySet}

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	revStoreWrite, err := store.WriteMultipleTxn(context.Background(), &recordUpdateSet)
	assert.Nilf(t, err, "Failed to write to store - error: %v", err)
	assert.Lessf(t, revStoreInitial, revStoreWrite, "Unexpected value for store revision on transaction completion")

	// Fetch the set of key,value pairs that we just wrote, along
	// with the revisions of the writes.
	//
	readResponse, err := store.ReadMultipleTxn(context.Background(), recordReadSet)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.Equalf(t, revStoreWrite, readResponse.Revision, "Unexpected value for store revision given no updates")
	assert.Equalf(t, len(recordReadSet.Keys), len(readResponse.Records), "Unexpected numbers of records returned")

	// Verify all the keys in the update set were actually written
	//
	for _, kv := range keyValueSet {
		record := readResponse.Records[kv.key]
		assert.NotNilf(t, record, "Failed to retrieve record for key %q", kv.key)
		assert.Equalf(t, revStoreWrite, readResponse.Records[kv.key].Revision, "Unexpected revision for record %q retrieved for key %q", record, kv.key)
		assert.Equalf(t, kv.value, readResponse.Records[kv.key].Value, "Unexpected value for record %q retrieved for key %q", record, kv.key)
	}

	store.Disconnect()

	store = nil

	return
}

func TestStoreWriteMultipleTxnCreate(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	testName := "TestStoreWriteMultipleTxnCreate"
	keySetSize := 1

	keyValueSet := testGenerateKeyValueSet(keySetSize, testName)
	keySet := testGenerateKeySetFromKeyValueSet(keyValueSet)
	recordUpdateSet := testGenerateRecordUpdateSetFromKeyValueSet(keyValueSet, testName, ConditionCreate)
	recordReadSet := RecordKeySet{Label: testName, Keys: keySet}

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	// Verify that none of the keys we care about exist in the store
	//
	readResponse, err := store.ReadMultipleTxn(context.Background(), recordReadSet)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.Lessf(t, RevisionInvalid, readResponse.Revision, "Unexpected value for store revision given expected failure")
	assert.Equalf(t, 0, len(readResponse.Records), "Unexpected numbers of records returned")

	revStoreCreate, err := store.WriteMultipleTxn(context.Background(), &recordUpdateSet)
	assert.Nilf(t, err, "Failed to write to store - error: %v", err)
	assert.Lessf(t, revStoreInitial, revStoreCreate, "Unexpected value for store revision on write(crerate) completion")

	// The write claimed to succeed, now go fetch the record(s) and verify
	// the revision(s) and value(s) are as expected
	//
	readResponse, err = store.ReadMultipleTxn(context.Background(), recordReadSet)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.Equalf(t, revStoreCreate, readResponse.Revision, "Unexpected value for store revision given no updates")
	assert.Equalf(t, len(recordReadSet.Keys), len(readResponse.Records), "Unexpected numbers of records returned")

	for k, r := range readResponse.Records {
		assert.Equalf(t, revStoreCreate, r.Revision, "read revision does not match earlier write revision")
		assert.Equalf(t, recordUpdateSet.Records[k].Record.Value, r.Value, "read value does not match earlier write value")
	}

	// Try to re-create the same keys. These should fail and the original values and revisions should survive.
	//
	revStoreRecreate, err := store.WriteMultipleTxn(context.Background(), &recordUpdateSet)
	assert.NotNilf(t, err, "Succeeded where we expected to get a failed store write - error: %v", err)
	assert.Equalf(t, RevisionInvalid, revStoreRecreate, "Unexpected value for store revision on write(re-create) completion")

	readResponseRecreate, err := store.ReadMultipleTxn(context.Background(), recordReadSet)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.Equalf(t, len(recordReadSet.Keys), len(readResponseRecreate.Records), "Unexpected numbers of records returned")
	assert.Equalf(t, revStoreCreate, readResponseRecreate.Revision, "Unexpected value for store revision given no updates")

	for k, r := range readResponseRecreate.Records {
		assert.Equalf(t, revStoreCreate, r.Revision, "read revision does not match earlier write revision")
		assert.Equalf(t, recordUpdateSet.Records[k].Record.Value, r.Value, "read value does not match earlier write value")
	}

	store.Disconnect()

	store = nil

	return
}

func TestStoreWriteMultipleTxnOverwrite(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	testName := "TestStoreWriteMultipleTxnOverwrite"
	keySetSize := 1

	keyValueSet := testGenerateKeyValueSet(keySetSize, testName)
	keySet := testGenerateKeySetFromKeyValueSet(keyValueSet)
	recordCreateSet := testGenerateRecordUpdateSetFromKeyValueSet(keyValueSet, testName, ConditionCreate)
	recordReadSet := RecordKeySet{Label: testName, Keys: keySet}

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	// Verify that none of the keys we care about exist in the store
	//
	readResponse, err := store.ReadMultipleTxn(context.Background(), recordReadSet)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.Lessf(t, RevisionInvalid, readResponse.Revision, "Unexpected value for store revision given expected failure")
	assert.Equalf(t, 0, len(readResponse.Records), "Unexpected numbers of records returned")

	revStoreCreate, err := store.WriteMultipleTxn(context.Background(), &recordCreateSet)
	assert.Nilf(t, err, "Failed to write to store - error: %v", err)
	assert.Lessf(t, revStoreInitial, revStoreCreate, "Unexpected value for store revision on write(create) completion")

	// The write claimed to succeed, now go fetch the record(s) and verify
	// the revision(s) and value(s) are as expected
	//
	readResponse, err = store.ReadMultipleTxn(context.Background(), recordReadSet)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.Equalf(t, revStoreCreate, readResponse.Revision, "Unexpected value for store revision given no updates")
	assert.Equalf(t, len(recordReadSet.Keys), len(readResponse.Records), "Unexpected numbers of records returned")

	for k, r := range readResponse.Records {
		assert.Equalf(t, revStoreCreate, r.Revision, "read revision does not match earlier write revision")
		assert.Equalf(t, recordCreateSet.Records[k].Record.Value, r.Value, "read value does not match earlier write value")
	}

	// We verified the write worked, so try an unconditional overwrite. Set the
	// required condition and change the value so we can verify after the update.
	//
	recordUpdateSet := RecordUpdateSet{Label: testName, Records: make(map[string]RecordUpdate)}

	for k, r := range recordCreateSet.Records {
		recordUpdateSet.Records[k] = RecordUpdate{
			Condition: ConditionUnconditional,
			Record: Record{
				Revision: RevisionInvalid,
				Value:    r.Record.Value + "+ConditionOverwrite",
			},
		}
	}

	revStoreUpdate, err := store.WriteMultipleTxn(context.Background(), &recordUpdateSet)
	assert.Nilf(t, err, "Failed to write unconditional update to store - error: %v", err)
	assert.Lessf(t, revStoreCreate, revStoreUpdate, "Expected new store revision to be greater than the earlier store revision")

	readResponseUpdate, err := store.ReadMultipleTxn(context.Background(), recordReadSet)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.Equalf(t, revStoreUpdate, readResponseUpdate.Revision, "Unexpected value for store revision given no updates")
	assert.Equalf(t, len(recordReadSet.Keys), len(readResponseUpdate.Records), "Unexpected numbers of records returned")

	for k, r := range readResponseUpdate.Records {
		assert.Equalf(t, revStoreUpdate, r.Revision, "read revision does not match earlier write revision")
		assert.Equalf(t, recordUpdateSet.Records[k].Record.Value, r.Value, "read value does not match earlier write value")
	}

	store.Disconnect()

	store = nil

	return
}

func TestStoreWriteMultipleTxnCompareEqual(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	testName := "TestStoreWriteMultipleTxnCompareEqual"
	keySetSize := 1

	keyValueSet := testGenerateKeyValueSet(keySetSize, testName)
	keySet := testGenerateKeySetFromKeyValueSet(keyValueSet)
	recordCreateSet := testGenerateRecordUpdateSetFromKeyValueSet(keyValueSet, testName, ConditionCreate)
	recordReadSet := RecordKeySet{Label: testName, Keys: keySet}

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	// Verify that none of the keys we care about exist in the store
	//
	readResponse, err := store.ReadMultipleTxn(context.Background(), recordReadSet)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.Lessf(t, RevisionInvalid, readResponse.Revision, "Unexpected value for store revision given expected failure")
	assert.Equalf(t, 0, len(readResponse.Records), "Unexpected numbers of records returned")

	revStoreCreate, err := store.WriteMultipleTxn(context.Background(), &recordCreateSet)
	assert.Nilf(t, err, "Failed to write to store - error: %v", err)
	assert.Lessf(t, revStoreInitial, revStoreCreate, "Unexpected value for store revision on write(create) completion")

	// The write claimed to succeed, now go fetch the record(s) and verify
	// the revision(s) and value(s) are as expected
	//
	readResponse, err = store.ReadMultipleTxn(context.Background(), recordReadSet)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.Equalf(t, revStoreCreate, readResponse.Revision, "Unexpected value for store revision given no updates")
	assert.Equalf(t, len(recordReadSet.Keys), len(readResponse.Records), "Unexpected numbers of records returned")

	for k, r := range readResponse.Records {
		assert.Equalf(t, revStoreCreate, r.Revision, "read revision does not match earlier write revision")
		assert.Equalf(t, recordCreateSet.Records[k].Record.Value, r.Value, "read value does not match earlier write value")
	}

	// We verified the write worked, so try a conditional update when the revisions
	// are equal. Set the required condition and change the value so we can verify
	// after the update.
	//
	recordUpdateSet := RecordUpdateSet{Label: testName, Records: make(map[string]RecordUpdate)}

	for k, r := range readResponse.Records {
		recordUpdateSet.Records[k] = RecordUpdate{
			Condition: ConditionRevisionEqual,
			Record: Record{
				Revision: r.Revision,
				Value:    r.Value + "+ConditionEqual",
			},
		}
	}

	revStoreUpdate, err := store.WriteMultipleTxn(context.Background(), &recordUpdateSet)
	assert.Nilf(t, err, "Failed to write conditional(equal) update to store - error: %v", err)
	assert.Lessf(t, revStoreCreate, revStoreUpdate, "Expected new store revision to be greater than the earlier store revision")

	readResponseUpdate, err := store.ReadMultipleTxn(context.Background(), recordReadSet)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.Equalf(t, revStoreUpdate, readResponseUpdate.Revision, "Unexpected value for store revision given no further updates")
	assert.Equalf(t, len(recordReadSet.Keys), len(readResponseUpdate.Records), "Unexpected numbers of records returned")

	for k, r := range readResponseUpdate.Records {
		assert.Equalf(t, revStoreUpdate, r.Revision, "read revision does not match earlier write revision")
		assert.Equalf(t, recordUpdateSet.Records[k].Record.Value, r.Value, "read value does not match earlier write value")
	}

	store.Disconnect()

	store = nil

	return
}
