// Unit tests for the web service store package
//
package store

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/embed"
)

const (
	defaultEmbeddedEtcdAddr       = string("127.0.0.1")
	defaultEmbeddedEtcdNode       = string("localhost")
	defaultEmbeddedEtcdPortClient = string("9379")
	defaultEmbeddedEtcdPortPeer   = string("9380")

	defaultEmbeddedEtcdPeer   = defaultEmbeddedEtcdNode + ":" + defaultEmbeddedEtcdPortPeer
	defaultEmbeddedEtcdClient = defaultEmbeddedEtcdNode + ":" + defaultEmbeddedEtcdPortClient

	useEmbeddedEtcd = true
)

var (
	baseURI     string
	initialized bool

	etcdConfig  *embed.Config
	etcdService *embed.Etcd
)

func etcdStart() {

	var err error
	var hostList = make(map[string]struct{})
	var urlPeer = "http://" + defaultEmbeddedEtcdPeer
	var urlClient = "http://" + defaultEmbeddedEtcdClient

	hostList[defaultEmbeddedEtcdNode] = struct{}{}
	hostList[defaultEmbeddedEtcdAddr] = struct{}{}

	etcdConfig = embed.NewConfig()

	// Set a location to place the underlying files for the store
	//
	etcdConfig.Dir = "c:\\temp\\CloudChamber\\default.etcd"

	// Only allow calls from clients on the current node
	//
	etcdConfig.HostWhitelist = hostList

	// Override the default listening and advertising endpoints for both clients and peers.
	//
	lpurl, _ := url.Parse(urlPeer)
	lcurl, _ := url.Parse(urlClient)
	apurl, _ := url.Parse(urlPeer)
	acurl, _ := url.Parse(urlClient)

	etcdConfig.LPUrls = []url.URL{*lpurl}
	etcdConfig.LCUrls = []url.URL{*lcurl}
	etcdConfig.APUrls = []url.URL{*apurl}
	etcdConfig.ACUrls = []url.URL{*acurl}

	// Override the default log level "info" which is very noisy
	//
	etcdConfig.LogLevel = "warn"

	etcdService, err = embed.StartEtcd(etcdConfig)
	if err != nil {
		log.Fatalf("Failed to start Etcd instance - error: %v", err)
	}

	select {
	case <-etcdService.Server.ReadyNotify():
		log.Printf("Server is ready!")

	case <-time.After(60 * time.Second):
		etcdService.Server.Stop() // trigger a shutdown
		log.Fatalf("Server took too long to start! - error: %v", <-etcdService.Err())
	}

	return
}

func etcdStop() {

	log.Printf("Server is stopping...")

	etcdService.Close()

	etcdService = nil
	etcdConfig = nil

	log.Printf("Server is stopped!")
}

func commonSetup() {

	Initialize()

	if useEmbeddedEtcd {

		// Override the normal default endpoint list.
		//
		storeRoot.DefaultEndpoints = []string{defaultEmbeddedEtcdClient}
	}

	etcdStart()
	return
}

func commonCleanup() {
	etcdStop()
}

func TestMain(m *testing.M) {

	commonSetup()

	result := m.Run()

	commonCleanup()

	os.Exit(result)
}

func TestNew(t *testing.T) {

	store := NewWithDefaults()

	assert.NotNilf(t, store, "Failed to get the store as expected")

	store = nil

	return
}

func TestInitialize(t *testing.T) {

	store := NewWithDefaults()

	assert.NotNilf(t, store, "Failed to get the store as expected")
	assert.Equal(t, getDefaultEndpoints(), store.Endpoints, "Mismatch in initialization of endpoints")
	assert.Equal(t, getDefaultTimeoutConnect(), store.TimeoutConnect, "Mismatch in initialization of connection timeout")
	assert.Equal(t, getDefaultTimeoutRequest(), store.TimeoutRequest, "Mismatch in initialization of request timeout")
	assert.Equal(t, getDefaultTraceFlags(), store.TraceFlags, "Mismatch in initialization of trace flags")

	endpoints := []string{"localhost:8080", "localhost:8181"}
	timeoutConnect := getDefaultTimeoutConnect() * 2
	timeoutRequest := getDefaultTimeoutRequest() * 3
	traceFlags := traceFlagEnabled

	err := store.Initialize(endpoints, timeoutConnect, timeoutRequest, traceFlags)

	assert.Nilf(t, err, "Failed to initialize new store - error: %v", err)
	assert.Equal(t, endpoints, store.Endpoints, "Mismatch in initialization of endpoints")
	assert.Equal(t, timeoutConnect, store.TimeoutConnect, "Mismatch in initialization of connection timeout")
	assert.Equal(t, timeoutRequest, store.TimeoutRequest, "Mismatch in initialization of request timeout")
	assert.Equal(t, traceFlags, store.TraceFlags, "Mismatch in initialization of trace flags")

	store = nil

	return
}

func TestNewWithArgs(t *testing.T) {

	// Use non-default values to ensure we get what we asked for and not the defaults.
	//
	endpoints := []string{"localhost:8282", "localhost:8383"}
	timeoutConnect := getDefaultTimeoutConnect() * 4
	timeoutRequest := getDefaultTimeoutRequest() * 5
	traceFlags := traceFlagExpandResults

	store := New(endpoints, timeoutConnect, timeoutRequest, traceFlags)

	assert.NotNilf(t, store, "Failed to get the store as expected")
	assert.Equal(t, endpoints, store.Endpoints, "Mismatch in initialization of endpoints")
	assert.Equal(t, timeoutConnect, store.TimeoutConnect, "Mismatch in initialization of connection timeout")
	assert.Equal(t, timeoutRequest, store.TimeoutRequest, "Mismatch in initialization of request timeout")
	assert.Equal(t, traceFlags, store.TraceFlags, "Mismatch in initialization of trace flags")

	store = nil

	return
}

func TestStoreSetAndGet(t *testing.T) {

	store := NewWithDefaults()

	assert.NotNilf(t, store, "Failed to get the store as expected")
	assert.Equal(t, getDefaultEndpoints(), store.Endpoints, "Mismatch in initialization of endpoints")
	assert.Equal(t, getDefaultTimeoutConnect(), store.TimeoutConnect, "Mismatch in initialization of connection timeout")
	assert.Equal(t, getDefaultTimeoutRequest(), store.TimeoutRequest, "Mismatch in initialization of request timeout")
	assert.Equal(t, getDefaultTraceFlags(), store.TraceFlags, "Mismatch in initialization of trace flags")

	assert.Equal(t, store.Endpoints, store.GetAddress(), "Mismatch in fetch of endpoints")
	assert.Equal(t, store.TimeoutConnect, store.GetTimeoutConnect(), "Mismatch in fetch of connection timeout")
	assert.Equal(t, store.TimeoutRequest, store.GetTimeoutRequest(), "Mismatch in fetch of request timeout")
	assert.Equal(t, store.TraceFlags, store.GetTraceFlags(), "Mismatch in fetch of trace flags")

	endpoints := []string{"localhost:8484", "localhost:8585"}
	timeoutConnect := getDefaultTimeoutConnect() * 6
	timeoutRequest := getDefaultTimeoutRequest() * 7
	traceFlags := traceFlagExpandResults

	err := store.Initialize(endpoints, timeoutConnect, timeoutRequest, traceFlags)

	assert.Nilf(t, err, "Failed to update new store - error: %v", err)
	assert.Equal(t, endpoints, store.Endpoints, "Mismatch in update of endpoints")
	assert.Equal(t, timeoutConnect, store.TimeoutConnect, "Mismatch in update of connection timeout")
	assert.Equal(t, timeoutRequest, store.TimeoutRequest, "Mismatch in update of request timeout")
	assert.Equal(t, traceFlags, store.TraceFlags, "Mismatch in update of trace flags")

	assert.Equal(t, store.Endpoints, store.GetAddress(), "Mismatch in re-fetch of endpoints")
	assert.Equal(t, store.TimeoutConnect, store.GetTimeoutConnect(), "Mismatch in re-fetch of connection timeout")
	assert.Equal(t, store.TimeoutRequest, store.GetTimeoutRequest(), "Mismatch in re-fetch of request timeout")
	assert.Equal(t, store.TraceFlags, store.GetTraceFlags(), "Mismatch in re-fetch of trace flags")
}

func TestStoreConnectDisconnect(t *testing.T) {

	endpoints := getDefaultEndpoints()
	timeoutConnect := getDefaultTimeoutConnect()
	timeoutRequest := getDefaultTimeoutRequest()
	traceFlags := getDefaultTraceFlags()

	store := New(endpoints, timeoutConnect, timeoutRequest, traceFlags)
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

	endpoints := getDefaultEndpoints()
	timeoutConnect := getDefaultTimeoutConnect()
	timeoutRequest := getDefaultTimeoutRequest()
	traceFlags := getDefaultTraceFlags()

	store := New(endpoints, timeoutConnect, timeoutRequest, traceFlags)
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Initialize(endpoints, timeoutConnect, timeoutRequest, traceFlags)
	assert.Nilf(t, err, "Failed to re-initialize store - error: %v", err)

	err = store.Connect()
	assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

	err = store.Initialize(endpoints, timeoutConnect, timeoutRequest, traceFlags)
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

	endpoints := getDefaultEndpoints()
	timeoutConnect := getDefaultTimeoutConnect()
	timeoutRequest := getDefaultTimeoutRequest()
	traceFlags := getDefaultTraceFlags()

	store := New(endpoints, timeoutConnect, timeoutRequest, traceFlags)
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.SetAddress(endpoints)
	assert.Nilf(t, err, "Failed to update the address - error: %v", err)

	err = store.SetTimeoutConnect(timeoutConnect)
	assert.Nilf(t, err, "Failed to update the connect timeout - error: %v", err)

	err = store.SetTimeoutRequest(timeoutRequest)
	assert.Nilf(t, err, "Failed to update the request timeout - error: %v", err)

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

	key := "TestStoreWriteRead/Key"
	value := "TestStoreWriteRead/Value"

	store := NewWithDefaults()
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

	store := NewWithDefaults()
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
		if store.trace(traceFlagExpandResults) {
			fmt.Printf("[%v/%v] %v: %v\n", i, len(response), kv.key, kvValue)
		}
		assert.Equal(t, keyValueSet[i].key, kv.key, "Unexpected key - expected: %s received: %s", keyValueSet[i].key, kv.key)
		assert.Equal(t, keyValueSet[i].value, kvValue, "Unexpected value - expected: %s received: %s", keyValueSet[i].value, kvValue)
	}

	store.Disconnect()

	store = nil

	return
}

func TestStoreWriteReadWithPrefix(t *testing.T) {

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

	store := NewWithDefaults()
	assert.NotNilf(t, store, "Failed to get the store as expected")

	err := store.Connect()
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
		if store.trace(traceFlagExpandResults) {
			fmt.Printf("[%v/%v] %v: %v\n", i, len(response), kv.key, kvValue)
		}
		assert.Equal(t, keyValueSet[i].key, kv.key, "Unexpected key - expected: %s received: %s", keyValueSet[i].key, kv.key)
		assert.Equal(t, keyValueMap[kv.key], kvValue, "Unexpected value - expected: %s received: %s", keyValueMap[kv.key], kvValue)
	}

	store.Disconnect()

	store = nil

	return
}

func TestStoreWriteDelete(t *testing.T) {

	key := "TestStoreWriteDelete/Key"
	value := "TestStoreWriteDelete/Value"

	store := NewWithDefaults()
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

	store := NewWithDefaults()
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

	store := NewWithDefaults()
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

	key := "TestStoreWriteReadDeleteWithoutConnect/Key"
	value := "TestStoreWriteReadDeleteWithoutConnect/Value"

	keySet := make([]string, 1)
	keySet[0] = key

	keyValueSet := make([]KeyValueArg, 1)
	keyValueSet[0].key = key
	keyValueSet[0].value = value

	store := NewWithDefaults()
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

	key := "TestStoreSetWatch/Key"

	store := NewWithDefaults()
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

	keySet := []string{"TestStoreSetWatchMultiple/Key"}

	store := NewWithDefaults()
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

	key := "TestStoreSetWatchPrefix/Key"

	store := NewWithDefaults()
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
