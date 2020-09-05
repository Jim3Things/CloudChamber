// Unit tests for the web service store package
//
package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
	st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
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
	assert.Equal(t, ErrStoreConnected("already connected"), err, "Unexpected error response - expected: %v got: %v", ErrStoreConnected("already connected"), err)

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
	assert.Equal(t, ErrStoreConnected("already connected"), err, "Unexpected error response - expected: %v got: %v", ErrStoreConnected("already connected"), err)

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
	assert.Equal(t, ErrStoreConnected("already connected"), err, "Unexpected error response - expected: %v got: %v", ErrStoreConnected("already connected"), err)

	err = store.SetTimeoutConnect(timeoutConnect)
	assert.NotNilf(t, err, "Unexpectedly succeeded to update the connect timeout - error: %v", err)
	assert.Equal(t, ErrStoreConnected("already connected"), err, "Unexpected error response - expected: %v got: %v", ErrStoreConnected("already connected"), err)

	err = store.SetTimeoutRequest(timeoutRequest)
	assert.Nilf(t, err, "Failed to update the request timeout - error: %v", err)

	err = store.SetNamespaceSuffix(namespaceSuffix)
	assert.NotNilf(t, err, "Unexpectedly succeeded to update the namespace suffix - error: %v", err)
	assert.Equal(t, ErrStoreConnected("already connected"), err, "Unexpected error response - expected: %v got: %v", ErrStoreConnected("already connected"), err)

	err = store.Connect()
	assert.NotNilf(t, err, "Unexpectedly connected to store again - error: %v", err)
	assert.Equal(t, ErrStoreConnected("already connected"), err, "Unexpected error response - expected: %v got: %v", ErrStoreConnected("already connected"), err)

	store.Disconnect()

	// Try a second disconnect. Benign but should trigger different execution
	// path for coverage numbers.
	//
	store.Disconnect()

	store = nil

	return
}

func TestStoreWriteReadTxn(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreWriteReadTxn"
		invalidKey := testGenerateKeyFromNames(testName, "InvalidName")

		writeRequest := testGenerateRequestForWrite(1, testName)
		readRequest := testGenerateRequestForRead(1, testName)
		readInvalidRequest := testGenerateRequestForRead(1, invalidKey)

		assert.Equal(t, 1, len(writeRequest.Records))
		assert.Equal(t, 1, len(readRequest.Records))
		assert.Equal(t, 1, len(readInvalidRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		writeResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		require.Equal(t, len(writeRequest.Records), len(writeResponse.Records))

		// Look for a name we do not expect to be present
		//
		readResponse, err := store.ReadTxn(ctx, readInvalidRequest)
		assert.NotNilf(t, err, "Succeeded to read non-existing key/value from store - error: %v key: %v", err, invalidKey)
		assert.Equal(t, ErrStoreKeyNotFound(invalidKey), err, "unexpected failure when looking for an invalid key - error %v", err)
		assert.Nilf(t, readResponse, "Unexpected response for read of invalid key - error: %v key: %v", err, invalidKey)

		// Now try to read a key which should be there.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")

		testCompareReadResponseToWrite(t, readResponse, writeRequest, writeResponse)

		store.Disconnect()

		store = nil

		return nil
	})

	return
}

func TestStoreWriteReadMultipleTxn(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreWriteReadMultipleTxn"
		invalidKey := testGenerateKeyFromNames(testName, "InvalidName")

		writeRequest := testGenerateRequestForWrite(keySetSize, testName)
		readRequest := testGenerateRequestForRead(keySetSize, testName)
		readInvalidRequest := testGenerateRequestForRead(1, invalidKey)

		assert.Equal(t, keySetSize, len(writeRequest.Records))
		assert.Equal(t, keySetSize, len(readRequest.Records))
		assert.Equal(t, 1, len(readInvalidRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		writeResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		require.Equal(t, len(writeRequest.Records), len(writeResponse.Records))

		// Look for a name we do not expect to be present
		//
		readResponse, err := store.ReadTxn(ctx, readInvalidRequest)
		assert.NotNilf(t, err, "Succeeded to read non-existing key/value from store - error: %v key: %v", err, invalidKey)
		assert.Equal(t, ErrStoreKeyNotFound(invalidKey), err, "unexpected failure when looking for an invalid key - error %v", err)
		assert.Nilf(t, readResponse, "Unexpected response for read of invalid key - error: %v key: %v", err, invalidKey)

		// Now try to read the key/value pairs which should be there.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")

		testCompareReadResponseToWrite(t, readResponse, writeRequest, writeResponse)

		store.Disconnect()

		store = nil

		return nil
	})

	return
}

func TestStoreWriteDeleteTxn(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreWriteReadMultipleTxn"
		invalidKey := testGenerateKeyFromNames(testName, "InvalidName")

		writeRequest := testGenerateRequestForWrite(1, testName)
		deleteRequest := testGenerateRequestForDelete(1, testName)
		deleteInvalidRequest := testGenerateRequestForDelete(1, invalidKey)

		assert.Equal(t, 1, len(writeRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		writeResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		require.Equal(t, len(writeRequest.Records), len(writeResponse.Records))

		// Delete the key we just wrote
		//
		var key string

		for key = range deleteRequest.Records {
		}

		deleteResponse, err := store.DeleteTxn(ctx, deleteRequest)
		assert.Nilf(t, err, "Failed to delete key from store - error: %v key: %v", err, key)
		require.NotNil(t, deleteResponse)
		require.Equal(t, len(deleteRequest.Records), len(deleteResponse.Records))

		// Try to delete the key we just wrote a second time
		//
		deleteResponse, err = store.DeleteTxn(ctx, deleteRequest)
		assert.NotNilf(t, err, "Unexpectedly deleted the key from store for a second time - error: %v key: %v", err, key)
		assert.Equal(t, ErrStoreKeyNotFound(key), err, "unexpected failure when looking for a previously deleted key - error %v", err)

		// Try to delete a name we do not expect to be present
		//
		deleteResponse, err = store.DeleteTxn(ctx, deleteInvalidRequest)
		assert.NotNilf(t, err, "Succeeded to delete a non-existing key/value from store - error: %v key: %v", err, invalidKey)
		assert.Equal(t, ErrStoreKeyNotFound(invalidKey), err, "unexpected failure when looking for an invalid key - error %v", err)
		assert.Nil(t, deleteResponse)

		store.Disconnect()

		store = nil

		return nil
	})

	return
}

func TestStoreWriteDeleteMultipleTxn(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreWriteDeleteMultipleTxn"
		invalidKey := testGenerateKeyFromNames(testName, "InvalidName")

		writeRequest := testGenerateRequestForWrite(1, testName)
		deleteRequest := testGenerateRequestForDelete(1, testName)
		deleteInvalidRequest := testGenerateRequestForDelete(1, invalidKey)

		assert.Equal(t, 1, len(writeRequest.Records))
		assert.Equal(t, 1, len(deleteRequest.Records))
		assert.Equal(t, 1, len(deleteInvalidRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		writeResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		require.Equal(t, len(writeRequest.Records), len(writeResponse.Records))

		// Delete the key we just wrote
		//
		var key string

		for key = range deleteRequest.Records {
		}

		deleteResponse, err := store.DeleteTxn(ctx, deleteRequest)
		assert.Nilf(t, err, "Failed to delete key from store - error: %v key: %v", err, key)
		require.NotNil(t, deleteResponse)
		require.Equal(t, len(deleteRequest.Records), len(deleteResponse.Records))

		// Try to delete the key we just wrote a second time
		//
		deleteResponse, err = store.DeleteTxn(ctx, deleteRequest)
		assert.NotNilf(t, err, "Unexpectedly deleted the key from store for a second time - error: %v key: %v", err, key)
		assert.Equal(t, ErrStoreKeyNotFound(key), err, "unexpected failure when looking for a previously deleted key - error %v", err)

		// Try to delete a name we do not expect to be present
		//
		deleteResponse, err = store.DeleteTxn(ctx, deleteInvalidRequest)
		assert.NotNilf(t, err, "Succeeded to delete a non-existing key/value from store - error: %v key: %v", err, invalidKey)
		assert.Equal(t, ErrStoreKeyNotFound(invalidKey), err, "unexpected failure when looking for an invalid key - error %v", err)
		assert.Nil(t, deleteResponse)

		store.Disconnect()

		store = nil

		return nil
	})

	return
}

func TestStoreWriteDeleteWithPrefix(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreWriteDeleteWithPrefix"

		writeRequest := testGenerateRequestForWrite(keySetSize, testName)
		deleteRequest := testGenerateRequestForDelete(keySetSize, testName)

		assert.Equal(t, keySetSize, len(writeRequest.Records))
		assert.Equal(t, keySetSize, len(deleteRequest.Records))

		prefixKey := testGenerateKeyFromNames(testName, "")

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		writeResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		require.Equal(t, len(writeRequest.Records), len(writeResponse.Records))

		response, err := store.DeleteWithPrefix(ctx, prefixKey)
		assert.Nilf(t, err, "Failed to delete the prefix keys from the store - error: %v prefixKey: %v", err, prefixKey)
		require.NotNil(t, response)
		assert.Equal(t, 0, len(response.Records))

		response, err = store.DeleteWithPrefix(ctx, prefixKey)
		assert.Nilf(t, err, "Unexpected error when attmepting to delete prefix keys from store for a second - error: %v prefixKey: %v", err, prefixKey)

		store.Disconnect()

		store = nil

		return nil
	})

	return
}

func TestStoreWriteReadDeleteWithoutConnect(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreWriteReadDeleteWithoutConnect"

		writeRequest := testGenerateRequestForWrite(keySetSize, testName)
		readRequest := testGenerateRequestForRead(keySetSize, testName)
		deleteRequest := testGenerateRequestForRead(keySetSize, testName)
		deletePrefix := testGenerateKeyFromNames(testName, "")

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		response, err := store.WriteTxn(ctx, writeRequest)
		assert.NotNilf(t, err, "Unexpectedly succeeded to write to store - error: %v", err)
		assert.Equal(t, ErrStoreNotConnected("already disconnected"), err, "Unexpected error response - expected: %v got: %v", ErrStoreNotConnected("already disconnected"), err)
		assert.Nil(t, response)

		response, err = store.ReadTxn(ctx, readRequest)
		assert.NotNilf(t, err, "Unexpectedly succeeded to read from store - error: %v", err)
		assert.Equal(t, ErrStoreNotConnected("already disconnected"), err, "Unexpected error response - expected: %v got: %v", ErrStoreNotConnected("already disconnected"), err)
		assert.Nil(t, response)

		response, err = store.DeleteTxn(ctx, deleteRequest)
		assert.NotNilf(t, err, "Unexpectedly succeeded to delete from store - error: %v", err)
		assert.Equal(t, ErrStoreNotConnected("already disconnected"), err, "Unexpected error response - expected: %v got: %v", ErrStoreNotConnected("already disconnected"), err)
		assert.Nil(t, response)

		response, err = store.DeleteWithPrefix(ctx, deletePrefix)
		assert.NotNilf(t, err, "Unexpectedly succeeded to delete from store - error: %v", err)
		assert.Equal(t, ErrStoreNotConnected("already disconnected"), err, "Unexpected error response - expected: %v got: %v", ErrStoreNotConnected("already disconnected"), err)
		assert.Nil(t, response)

		store = nil

		return nil
	})

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

func TestStoreWriteMultipleTxnCreate(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreWriteMultipleTxnCreate"

		writeRequest := testGenerateRequestForWriteWithCondition(keySetSize, testName, ConditionCreate)
		readRequest := testGenerateRequestFromWriteRequest(writeRequest)

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		// Verify that none of the keys we care about exist in the store
		//
		readResponse, err := store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		assert.Lessf(t, RevisionInvalid, readResponse.Revision, "Unexpected value for store revision given expected failure")
		assert.Equalf(t, 0, len(readResponse.Records), "Unexpected numbers of records returned")

		createResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, createResponse)
		require.Equal(t, len(writeRequest.Records), len(createResponse.Records))
		assert.Lessf(t, revStoreInitial, createResponse.Revision, "Unexpected value for store revision on write(crerate) completion")

		// The write claimed to succeed, now go fetch the record(s) and verify
		// the revision(s) and value(s) are as expected
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNil(t, readResponse)
		assert.Equalf(t, createResponse.Revision, readResponse.Revision, "Unexpected value for store revision given no updates")
		assert.Equalf(t, len(createResponse.Records), len(readResponse.Records), "Unexpected numbers of records returned")

		testCompareReadResponseToWrite(t, readResponse, writeRequest, createResponse)

		// Try to re-create the same keys. These should fail and the original values and revisions should survive.
		//
		recreateResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.NotNilf(t, err, "Succeeded where we expected to get a failed store write - error: %v", err)
		require.NotNil(t, recreateResponse)
		assert.Equalf(t, RevisionInvalid, recreateResponse.Records, "Unexpected value for store revision on write(re-create) completion")

		readRecreateResponse, err := store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNil(t, readRecreateResponse)
		assert.Equalf(t, createResponse.Revision, readRecreateResponse.Revision, "Unexpected value for store revision given no updates")
		assert.Equalf(t, len(createResponse.Records), len(readRecreateResponse.Records), "Unexpected numbers of records returned")

		testCompareReadResponseToWrite(t, readResponse, writeRequest, createResponse)

		store.Disconnect()

		store = nil

		return nil
	})

	return
}

func TestStoreWriteMultipleTxnOverwrite(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreWriteMultipleTxnOverwrite"

		writeRequest := testGenerateRequestForWriteWithCondition(keySetSize, testName, ConditionCreate)
		readRequest := testGenerateRequestFromWriteRequest(writeRequest)

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		// Verify that none of the keys we care about exist in the store
		//
		readResponse, err := store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		assert.Lessf(t, RevisionInvalid, readResponse.Revision, "Unexpected value for store revision given expected failure")
		assert.Equalf(t, 0, len(readResponse.Records), "Unexpected numbers of records returned")

		createResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, createResponse)
		require.Equal(t, len(writeRequest.Records), len(createResponse.Records))
		assert.Lessf(t, revStoreInitial, createResponse.Revision, "Unexpected value for store revision on write(crerate) completion")

		// The write claimed to succeed, now go fetch the record(s) and verify
		// the revision(s) and value(s) are as expected
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		assert.Equalf(t, createResponse.Revision, readResponse.Revision, "Unexpected value for store revision given no updates")
		assert.Equalf(t, len(writeRequest.Records), len(readResponse.Records), "Unexpected numbers of records returned")

		testCompareReadResponseToWrite(t, readResponse, writeRequest, createResponse)

		// We verified the write worked, so try an unconditional overwrite. Set the
		// required condition and change the value so we can verify after the update.
		//
		updateRequest := testGenerateRequestForOverwriteFromWriteRequest(writeRequest, ConditionUnconditional)

		updateResponse, err := store.WriteTxn(ctx, updateRequest)
		assert.Nilf(t, err, "Failed to write unconditional update to store - error: %v", err)
		assert.Lessf(t, createResponse.Revision, updateResponse.Revision, "Expected new store revision to be greater than the earlier store revision")

		readResponseUpdate, err := store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		assert.Equalf(t, createResponse.Revision, readResponse.Revision, "Unexpected value for store revision given no updates")
		assert.Equalf(t, len(writeRequest.Records), len(readResponse.Records), "Unexpected numbers of records returned")

		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		assert.Equalf(t, updateResponse.Revision, readResponseUpdate.Revision, "Unexpected value for store revision given no updates")
		assert.Equalf(t, len(updateRequest.Records), len(readResponseUpdate.Records), "Unexpected numbers of records returned")

		testCompareReadResponseToWrite(t, readResponseUpdate, updateRequest, updateResponse)

		store.Disconnect()

		store = nil

		return nil
	})

	return
}

func TestStoreWriteMultipleTxnCompareEqual(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreWriteMultipleTxnCompareEqual"
		keySetSize := 1

		writeRequest := testGenerateRequestForWriteWithCondition(keySetSize, testName, ConditionCreate)
		readRequest := testGenerateRequestFromWriteRequest(writeRequest)

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		// Verify that none of the keys we care about exist in the store
		//
		readResponse, err := store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		assert.Lessf(t, RevisionInvalid, readResponse.Revision, "Unexpected value for store revision given expected failure")
		assert.Equalf(t, 0, len(readResponse.Records), "Unexpected numbers of records returned")

		createResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, createResponse)
		require.Equal(t, len(writeRequest.Records), len(createResponse.Records))
		assert.Lessf(t, revStoreInitial, createResponse.Revision, "Unexpected value for store revision on write(crerate) completion")

		// The write claimed to succeed, now go fetch the record(s) and verify
		// the revision(s) and value(s) are as expected
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		assert.Equalf(t, createResponse.Revision, readResponse.Revision, "Unexpected value for store revision given no updates")
		assert.Equalf(t, len(writeRequest.Records), len(readResponse.Records), "Unexpected numbers of records returned")

		testCompareReadResponseToWrite(t, readResponse, writeRequest, createResponse)

		// We verified the write worked, so try a conditional update when the revisions
		// are equal. Set the required condition and change the value so we can verify
		// after the update.
		//
		updateRequest := testGenerateRequestForOverwriteFromWriteRequest(writeRequest, ConditionRevisionEqual)

		updateResponse, err := store.WriteTxn(ctx, updateRequest)
		assert.Nilf(t, err, "Failed to write conditional update to store - error: %v", err)
		assert.Lessf(t, createResponse.Revision, updateResponse.Revision, "Expected new store revision to be greater than the earlier store revision")

		// verify the update happened as expected
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		assert.Equalf(t, updateResponse.Revision, readResponse.Revision, "Unexpected value for store revision given no updates")
		assert.Equalf(t, len(updateRequest.Records), len(readResponse.Records), "Unexpected numbers of records returned")

		testCompareReadResponseToWrite(t, readResponse, updateRequest, updateResponse)

		store.Disconnect()

		store = nil

		return nil
	})

	return
}

func TestStoreListWithPrefix(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreListWithPrefix"
		invalidKey := testGenerateKeyFromNames(testName, "InvalidName")
		prefix := testGenerateKeyFromNames(testName, "")

		writeRequest := testGenerateRequestForWrite(keySetSize, testName)

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		writeResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		require.Equal(t, len(writeRequest.Records), len(writeResponse.Records))

		// Look for a prefix name we do not expect to be present.
		//
		// We expect success with a non-nil but empty set.
		//
		listResponse, err := store.ListWithPrefix(ctx, invalidKey)
		assert.Nilf(t, err, "Succeeded to read non-existing prefix key from store - error: %v prefixKey: %v", err, invalidKey)
		require.NotNilf(t, listResponse, "Failed to get a non-nil response as expected - error: %v prefixKey: %v", err, invalidKey)
		assert.Equal(t, 0, len(listResponse.Records), "Got more results than expected")

		if len(listResponse.Records) > 0 {
			for k, r := range listResponse.Records {
				t.Logf("Unexpected key/value pair key: %v value: %v", k, r.Value)
			}
		}

		// Now look for a set of prefixed key/value pairs which we do expect to be present.
		//
		listResponse, err = store.ListWithPrefix(ctx, prefix)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, listResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(writeResponse.Records), len(listResponse.Records), "Failed to get the expected number of response values")

		testCompareReadResponseToWrite(t, listResponse, writeRequest, writeResponse)

		// Check we got records for each key we asked for
		//
		for k, r := range writeRequest.Records {
			rec, present := listResponse.Records[k]

			assert.Truef(t, present, "Missing record for key - %v", k)

			if present {
				assert.Equal(t, r.Value, rec.Value, "Unexpected value - expected: %q received: %q", r.Value, rec.Value)
			}
		}

		// Check we ONLY got records for the keys we asked for
		//
		for k, r := range listResponse.Records {
			val, present := writeRequest.Records[k]
			assert.Truef(t, present, "Extra key: %v record: %v", k, r)
			if present {
				assert.Equalf(t, val, r.Value, "key: %v Expected: %q Actual %q", val, r.Value)
			}
		}

		store.Disconnect()

		store = nil

		return nil
	})

	return
}
