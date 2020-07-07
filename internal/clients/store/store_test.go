// Unit tests for the web service store package
//
package store

import (
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
	"github.com/Jim3Things/CloudChamber/internal/tracing/setup"
	"github.com/stretchr/testify/assert"
	//	"go.opentelemetry.io/otel/api/kv"
)

// A number of tests use a pre-computed set of keys for the purposes of
// the test. This constant determines the standard size of these sets.
// The value chosen is largely arbitrary. Selecting a value that is too
// large may lead to problems with values to /from the underlying store,
// so be reasonable.
//
const keySetSize = 100

var (
	baseURI     string
	initialized bool

	testNamespaceSuffixRoot = "/Test"

	configPath *string
)

func commonSetup() {
	var testNamespace string

	setup.Init(exporters.UnitTest)

	configPath = flag.String("config", ".", "path to the configuration file")
	flag.Parse()

	cfg, err := config.ReadGlobalConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to process the global configuration: %v", err)
	}

	Initialize(cfg)

	// It is meaningless to have both a unique per-instance test namespace
	// and to clean the store before the tests are run
	//
	if cfg.Store.Test.UseUniqueInstance && cfg.Store.Test.PreCleanStore {
		log.Fatalf("invalid configuration: both UseUniqueInstance and PreCleanStore are enabled: %v", err)
	}

	// For test purposes, need to set an alternate namespace rather than
	// rely on the standard. From the configuration, we can either use the
	// standard, fixed, well-known prefix, or we can use a per-instance
	// unique prefix derived from the current time
	//
	if cfg.Store.Test.UseUniqueInstance {
		testNamespace = fmt.Sprintf("%s/%s/", testNamespaceSuffixRoot, time.Now().Format(time.RFC3339Nano))
	} else {
		testNamespace = testNamespaceSuffixRoot + "/Standard/"
	}

	if cfg.Store.Test.PreCleanStore {
		if err := cleanNamespace(testNamespace); err != nil {
			log.Fatalf("failed to pre-clean the store as requested - namespace: %s err: %v", testNamespace, err)
		}
	}

	setDefaultNamespaceSuffix(testNamespace)
	return
}

func cleanNamespace(testNamespace string) error {

	store := NewStore()

	if store == nil {
		log.Fatalf("unable to allocate store context for pre-cleanup")
	}

	if err := store.SetNamespaceSuffix(""); err != nil {
		return err
	}
	if err := store.Connect(); err != nil {
		return err
	}
	if err := store.DeleteWithPrefix(testNamespace); err != nil {
		return err
	}

	store.Disconnect()

	return nil
}

func testGenerateKeyValueSet(setSize int, setName string) []KeyValueArg {

	keyValueSet := make([]KeyValueArg, setSize)

	for i := range keyValueSet {
		keyValueSet[i].key = fmt.Sprintf("%s/Key%04d", setName, i)
		keyValueSet[i].value = fmt.Sprintf("%s/Val%04d", setName, i)
	}

	return keyValueSet
}

func testGenerateKeyValueMapFromKeyValueSet(keyValueSet []KeyValueArg) map[string]string {

	keyValueMap := make(map[string]string, len(keyValueSet))

	for _, kv := range keyValueSet {
		keyValueMap[kv.key] = kv.value
	}

	return keyValueMap
}

func testGenerateKeySetFromKeyValueSet(keyValueSet []KeyValueArg) []string {

	keySet := make([]string, len(keyValueSet))

	for i, kv := range keyValueSet {
		keySet[i] = kv.key
	}

	return keySet
}

func TestMain(m *testing.M) {
	commonSetup()
	os.Exit(m.Run())
}

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
	response, err := store.ReadWithPrefix(invalidKey)
	assert.Nilf(t, err, "Succeeded to read non-existing prefix key from store - error: %v prefixKey: %v", err, invalidKey)
	assert.NotNilf(t, response, "Failed to get a non-nil response as expected - error: %v prefixKey: %v", err, invalidKey)
	assert.Equal(t, 0, len(response), "Got more results than expected")

	if nil != response && len(response) > 0 {
		for i, kv := range response {
			t.Logf("Unexpected key/value pair [%v/%v] key: %v value: %v", i, len(response), kv.key, string(kv.value))
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
		if store.trace(traceFlagExpandResultsInTest) {
			t.Logf("[%v/%v] %v: %v", i, len(response), kv.key, kvValue)
		}
		assert.Equal(t, keyValueSet[i].key, kv.key, "Unexpected key - expected: %s received: %s", keyValueSet[i].key, kv.key)
		assert.Equal(t, keyValueMap[kv.key], kvValue, "Unexpected value - expected: %s received: %s", keyValueMap[kv.key], kvValue)
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

	response, err := store.ReadMultipleTxn(recordKeySet)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.Equalf(t, len(recordKeySet.Keys), len(response.Records), "Unexpected numbers of records returned")

	for _, kv := range keyValueSet {
		record := response.Records[kv.key]
		assert.NotNilf(t, record, "Failed to retrieve record for key %q", kv.key)
		assert.Equalf(t, kv.value, response.Records[kv.key].Value, "Unexpected value for record %q retrieved for key %q", record, kv.key)
	}

	store.Disconnect()

	store = nil

	return
}

func TestStoreWriteMultipleTxn(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	keyValueSet := testGenerateKeyValueSet(keySetSize, "TestStoreWriteReadMultipleTxn")
	keySet := testGenerateKeySetFromKeyValueSet(keyValueSet)

	recordUpdateSet := RecordUpdateSet{Label: "TestStoreWriteMultipleTxn", Records: make(map[string]RecordUpdate)}

	// Build a set of key,value pairs to be unconditionally written to the store.
	//
	for _, kv := range keyValueSet {
		recordUpdateSet.Records[kv.key] =
			RecordUpdate{
				Compare: RevisionCompareUnconditional,
				Record: Record{
					Revision: RevisionUnconditional,
					Value:    kv.value,
				},
			}
	}

	store := NewStore()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	revStoreWrite, err := store.WriteMultipleTxn(&recordUpdateSet)
	assert.Nilf(t, err, "Failed to write to store - error: %v", err)
	assert.NotEqualf(t, 0, revStoreWrite, "Unexpected value for store revision on transaction completion")

	// Fetch the set of key,value pairs that we just wrote, along
	// with the revisions of the writes.
	//
	recordKeySet := RecordKeySet{Label: "TestStoreWriteReadMultipleTxn", Keys: keySet}

	readResponse, err := store.ReadMultipleTxn(recordKeySet)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.Equalf(t, len(recordKeySet.Keys), len(readResponse.Records), "Unexpected numbers of records returned")
	assert.Equalf(t, revStoreWrite, readResponse.Revision, "Unexpected value for store revision given no updates")

	for _, kv := range keyValueSet {
		record := readResponse.Records[kv.key]
		assert.NotNilf(t, record, "Failed to retrieve record for key %q", kv.key)
		assert.Equalf(t, kv.value, readResponse.Records[kv.key].Value, "Unexpected value for record %q retrieved for key %q", record, kv.key)
	}

	// We verified the write worked, so try some conditional updates.
	//
	revStoreCurrent := readResponse.Revision

	key := keyValueSet[0].key
	val := keyValueSet[0].value

	keySetUpdate := []string{key}
	recordKeySetUpdate := RecordKeySet{Label: "TestStoreWriteReadMultipleTxn", Keys: keySetUpdate}
	recordConditionalUpdate := RecordUpdateSet{Label: "TestStoreWriteMultipleTxn", Records: make(map[string]RecordUpdate)}

	// Verify condition "unconditional"
	//
	recordConditionalUpdate.Records[key] = RecordUpdate{
		Compare: RevisionCompareUnconditional,
		Record: Record{
			Revision: RevisionUnconditional,
			Value:    val + "Update-Unconditional",
		},
	}

	revStoreNew, err := store.WriteMultipleTxn(&recordConditionalUpdate)
	assert.Nilf(t, err, "Failed to write unconditional update to store - error: %v", err)
	assert.Lessf(t, revStoreCurrent, revStoreNew, "Expected new store revision to be greater than the earlier store revision")

	revStoreCurrent = revStoreNew

	readResponseUpdate, err := store.ReadMultipleTxn(recordKeySetUpdate)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.Equalf(t, len(recordKeySet.Keys), len(readResponseUpdate.Records), "Unexpected numbers of records returned")
	assert.Equalf(t, revStoreCurrent, readResponseUpdate.Revision, "Unexpected value for store revision given no updates")

	assert.Lessf(t, readResponse.Records[key].Revision, recordConditionalUpdate.Records[key].Record.Revision, "Unexpected revision for unconditional update")
	assert.Equalf(t, recordConditionalUpdate.Records[key].Record.Value, readResponseUpdate.Records[key].Value, "Unexpected value for unconditional update")

	// Verify condition "less than"
	//
	recordConditionalUpdate.Records[key] = RecordUpdate{
		Compare: RevisionCompareLess,
		Record: Record{
			Revision: readResponseUpdate.Records[key].Revision,
			Value:    val + "Update-ConditionalLess",
		},
	}

	revStoreNew, err = store.WriteMultipleTxn(&recordConditionalUpdate)
	assert.Nilf(t, err, "Failed to write condition less-than update to store - error: %v", err)
	assert.Lessf(t, revStoreCurrent, revStoreNew, "Expected new store revision to be greater than the earlier store revision")

	revStoreCurrent = revStoreNew

	readResponseUpdate, err := store.ReadMultipleTxn(recordKeySetUpdate)
	assert.Nilf(t, err, "Failed to read from store - error: %v", err)
	assert.Equalf(t, len(recordKeySet.Keys), len(readResponseUpdate.Records), "Unexpected numbers of records returned")
	assert.Equalf(t, revStoreCurrent, readResponseUpdate.Revision, "Unexpected value for store revision given no updates")

	assert.Lessf(t, readResponse.Records[key].Revision, recordConditionalUpdate.Records[key].Record.Revision, "Unexpected revision for less-than update")
	assert.Equalf(t, recordConditionalUpdate.Records[key].Record.Value, readResponseUpdate.Records[key].Value, "Unexpected value for less-than update")

	store.Disconnect()

	store = nil

	return
}
