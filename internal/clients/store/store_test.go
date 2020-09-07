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
		key := testGenerateKeyFromName(testName)

		writeRequest := testGenerateRequestForWrite(1, key)
		readRequest := testGenerateRequestForRead(1, key)

		assert.Equal(t, 1, len(writeRequest.Records))
		assert.Equal(t, 1, len(readRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		writeResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		assert.Equal(t, 0, len(writeResponse.Records))
		assert.Lessf(t, revStoreInitial, writeResponse.Revision, "Unexpected value for store revision on write completion")

		readResponse, err := store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
		assert.Equal(t, writeResponse.Revision, readResponse.Revision)

		testCompareReadResponseToWrite(t, readResponse, writeRequest, writeResponse)

		store.Disconnect()

		store = nil

		return nil
	})

	return
}

func TestStoreWriteReadTxnRequired(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreWriteReadTxnRequired"
		missingName := "MissingName"
		key := testGenerateKeyFromNames(testName, missingName)

		writeRequest := testGenerateRequestForSimpleWrite(key)
		readRequest := testGenerateRequestForSimpleReadWithCondition(key, ConditionRequired)

		assert.Equal(t, 1, len(writeRequest.Records))
		assert.Equal(t, 1, len(readRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		// Attempt to read the key before it exists. As this is a "ConditionRequired" read
		// request, it should fail and produce no response.
		//
		readResponse, err := store.ReadTxn(ctx, readRequest)
		assert.NotNilf(t, err, "Unexpected success reading key supposedly absent - error: %v key: %v", err)
		assert.Equal(t, ErrStoreKeyNotFound(key), err, "unexpected failure when looking for an absent key - error %v", err)
		assert.Nilf(t, readResponse, "Unexpected response for read of invalid key - error: %v key: %v", err, key)

		// Now write the key
		//
		writeResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		assert.Equal(t, 0, len(writeResponse.Records))
		assert.Lessf(t, revStoreInitial, writeResponse.Revision, "Unexpected value for store revision on write completion")

		// Now try to read the key which should now be there.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
		assert.Equal(t, writeResponse.Revision, readResponse.Revision)

		testCompareReadResponseToWrite(t, readResponse, writeRequest, writeResponse)

		store.Disconnect()

		store = nil

		return nil
	})

	return
}

func TestStoreWriteReadTxnOptional(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreWriteReadTxnOptional"
		missingName := "MissingName"
		key := testGenerateKeyFromNames(testName, missingName)

		writeRequest := testGenerateRequestForSimpleWrite(key)
		readRequest := testGenerateRequestForSimpleReadWithCondition(key, ConditionUnconditional)

		assert.Equal(t, 1, len(writeRequest.Records))
		assert.Equal(t, 1, len(readRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		// Attempt to read the key before it exists. As this is a "ConditionUnconditional" read
		// request, it should succeed and produce an empty response.
		//
		readResponse, err := store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Unexpected failure unconditionally reading key supposedly absent - error: %v key: %v", err, key)
		require.NotNilf(t, readResponse, "Unexpected missing response for unconditional read of absent key - error: %v key: %v", err, key)
		assert.Equalf(t, 0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
		assert.Lessf(t, revStoreInitial, readResponse.Revision, "Unexpected value for store revision on read completion")

		// Now write the key
		//
		writeResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		assert.Equal(t, 0, len(writeResponse.Records))
		assert.Lessf(t, readResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

		// Now try to read a key which should now be there.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Unexpected failure unconditionally reading key supposedly present - error: %v key: %v", err, key)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
		assert.Equal(t, writeResponse.Revision, readResponse.Revision)

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
		key := testGenerateKeyFromName(testName)

		writeRequest := testGenerateRequestForWrite(keySetSize, key)
		readRequest := testGenerateRequestForRead(keySetSize, key)

		assert.Equal(t, keySetSize, len(writeRequest.Records))
		assert.Equal(t, keySetSize, len(readRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		writeResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		assert.Equal(t, 0, len(writeResponse.Records))
		assert.Lessf(t, revStoreInitial, writeResponse.Revision, "Unexpected value for store revision on write completion")

		// Now try to read the key/value pairs which should be there.
		//
		readResponse, err := store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
		assert.Equal(t, writeResponse.Revision, readResponse.Revision)

		testCompareReadResponseToWrite(t, readResponse, writeRequest, writeResponse)

		store.Disconnect()

		store = nil

		return nil
	})

	return
}

func TestStoreWriteReadMultipleTxnRequired(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreWriteReadMultipleTxnRequired"
		key := testGenerateKeyFromName(testName)

		writeRequest := testGenerateRequestForWrite(keySetSize, key)
		readRequest := testGenerateRequestForReadWithCondition(keySetSize, key, ConditionRequired)

		assert.Equal(t, keySetSize, len(writeRequest.Records))
		assert.Equal(t, keySetSize, len(readRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		// Attempt to read from the set of keys before they exist. As this is a "ConditionRequired" read
		// request, it should fail and produce no response.
		//
		readResponse, err := store.ReadTxn(ctx, readRequest)
		assert.NotNilf(t, err, "Unexpectedly succeeded reading from store - error: %v", err)

		// We have a slight problem here in that we do not know which key of the set will be reported
		// in the error. That means it could be any of them, so we need to check for them all and only
		// assert if none of them match.
		//
		var foundError bool

		for k := range readRequest.Records {
			if err == ErrStoreKeyNotFound(k) {
				foundError = true
				break
			}
		}

		assert.Truef(t, foundError, "Returned error failed to match any of the expected values - err: %v", err)
		assert.Nilf(t, readResponse, "Unexpected response for read of absent keys - error: %v", err)

		// Now write the keys
		//
		writeResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		assert.Equal(t, 0, len(writeResponse.Records))
		assert.Lessf(t, revStoreInitial, writeResponse.Revision, "Unexpected value for store revision on write completion")

		// Finally try to read the key/value pairs which should now be there.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
		assert.Equal(t, writeResponse.Revision, readResponse.Revision)

		testCompareReadResponseToWrite(t, readResponse, writeRequest, writeResponse)

		store.Disconnect()

		store = nil

		return nil
	})

	return
}

func TestStoreWriteReadMultipleTxnOptional(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreWriteReadMultipleTxnOptional"
		key := testGenerateKeyFromName(testName)

		writeRequest := testGenerateRequestForWrite(keySetSize, key)
		readRequest := testGenerateRequestForReadWithCondition(keySetSize, key, ConditionUnconditional)

		assert.Equal(t, keySetSize, len(writeRequest.Records))
		assert.Equal(t, keySetSize, len(readRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		// Attempt to read from the set of keys before they exist. As this is a "ConditionUnconditional" read
		// request, it should succeed and produce an empty response.
		//
		readResponse, err := store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equalf(t, 0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
		assert.Lessf(t, revStoreInitial, readResponse.Revision, "Unexpected value for store revision on read completion")

		// Now write the keys
		//
		writeResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		assert.Equal(t, 0, len(writeResponse.Records))
		assert.Lessf(t, readResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

		// Finally try to read the key/value pairs which should now be there.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
		assert.Equal(t, writeResponse.Revision, readResponse.Revision)

		testCompareReadResponseToWrite(t, readResponse, writeRequest, writeResponse)

		store.Disconnect()

		store = nil

		return nil
	})

	return
}

func TestStoreWriteReadMultipleTxnPartial(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreWriteReadMultipleTxnPartial"
		key := testGenerateKeyFromName(testName)

		writeRequestPartial := testGenerateRequestForWrite(1, key)
		writeRequestComplete := testGenerateRequestForWrite(2, key)

		readRequest := testGenerateRequestForReadWithCondition(len(writeRequestComplete.Records), key, ConditionUnconditional)

		assert.Equal(t, 1, len(writeRequestPartial.Records))
		assert.Equal(t, 2, len(writeRequestComplete.Records))
		assert.Equal(t, len(writeRequestComplete.Records), len(readRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		// Attempt to read from the set of keys before they exist. As this is a "ConditionUnconditional" read
		// request, it should succeed and produce an empty response.
		//
		readResponse, err := store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equalf(t, 0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
		assert.Lessf(t, revStoreInitial, readResponse.Revision, "Unexpected value for store revision on read completion")

		// Now write a partial set of keys
		//
		writeResponse, err := store.WriteTxn(ctx, writeRequestPartial)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		assert.Equal(t, 0, len(writeResponse.Records))
		assert.Lessf(t, readResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

		// Try to read the complete set of key/value pairs, only some of which should be there.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(writeRequestPartial.Records), len(readResponse.Records), "Read returned unexpected number of records")
		assert.Equal(t, writeResponse.Revision, readResponse.Revision)

		testCompareReadResponseToWrite(t, readResponse, writeRequestPartial, writeResponse)

		// Now write a complete set of keys
		//
		writeResponse, err = store.WriteTxn(ctx, writeRequestComplete)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		assert.Equal(t, 0, len(writeResponse.Records))
		assert.Lessf(t, readResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

		// Finally try to read the complete set of key/value pairs all of which should now be there.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
		assert.Equal(t, writeResponse.Revision, readResponse.Revision)

		testCompareReadResponseToWrite(t, readResponse, writeRequestComplete, writeResponse)

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
		testName := "TestStoreWriteDeleteTxn"
		key := testGenerateKeyFromName(testName)

		writeRequest := testGenerateRequestForSimpleWrite(key)
		readRequest := testGenerateRequestForSimpleRead(key)
		deleteRequest := testGenerateRequestForSimpleDelete(key)

		assert.Equal(t, 1, len(writeRequest.Records))
		assert.Equal(t, 1, len(readRequest.Records))
		assert.Equal(t, 1, len(deleteRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		// Attempt to read the key before it exists. As this is a "ConditionRequired" read
		// request, it should fail and produce no response.
		//
		readResponse, err := store.ReadTxn(ctx, readRequest)
		assert.NotNilf(t, err, "Unexpected success reading key supposedly absent - error: %v key: %v", err)
		assert.Equal(t, ErrStoreKeyNotFound(key), err, "unexpected failure when looking for an absent key - error %v", err)
		assert.Nilf(t, readResponse, "Unexpected response for read of invalid key - error: %v key: %v", err, key)

		// Now write the key
		//
		writeResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		assert.Equal(t, 0, len(writeResponse.Records))
		assert.Lessf(t, revStoreInitial, writeResponse.Revision, "Unexpected value for store revision on write completion")

		// Verify the key is now there.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
		assert.Equal(t, writeResponse.Revision, readResponse.Revision)

		testCompareReadResponseToWrite(t, readResponse, writeRequest, writeResponse)

		// Delete the key we just wrote
		//
		deleteResponse, err := store.DeleteTxn(ctx, deleteRequest)
		assert.Nilf(t, err, "Failed to delete key from store - error: %v key: %v", err, key)
		require.NotNil(t, deleteResponse)
		assert.Equal(t, 0, len(deleteResponse.Records))
		assert.Less(t, writeResponse.Revision, deleteResponse.Revision)

		// Attempt to re-read the key which once again should be absent.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.NotNilf(t, err, "Unexpected success reading key supposedly absent - error: %v key: %v", err)
		assert.Equal(t, ErrStoreKeyNotFound(key), err, "unexpected failure when looking for an absent key - error %v", err)
		assert.Nilf(t, readResponse, "Unexpected response for read of invalid key - error: %v key: %v", err, key)

		store.Disconnect()

		store = nil

		return nil
	})

	return
}

func TestStoreWriteDeleteTxnDeleteAbsent(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreWriteDeleteTxnDeleteAbsent"
		key := testGenerateKeyFromName(testName)

		writeRequest := testGenerateRequestForSimpleWrite(key)
		readRequest := testGenerateRequestForSimpleRead(key)
		deleteRequest := testGenerateRequestForSimpleDelete(key)

		assert.Equal(t, 1, len(writeRequest.Records))
		assert.Equal(t, 1, len(readRequest.Records))
		assert.Equal(t, 1, len(deleteRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		// Attempt to read the key before it exists. As this is a "ConditionRequired" read
		// request, it should fail and produce no response.
		//
		readResponse, err := store.ReadTxn(ctx, readRequest)
		assert.NotNilf(t, err, "Unexpected success reading key supposedly absent - error: %v key: %v", err)
		assert.Equal(t, ErrStoreKeyNotFound(key), err, "unexpected failure when looking for an absent key - error %v", err)
		assert.Nilf(t, readResponse, "Unexpected response for read of invalid key - error: %v key: %v", err, key)

		// Try to delete the key we just verified is absent
		//
		deleteResponse, err := store.DeleteTxn(ctx, deleteRequest)
		assert.NotNilf(t, err, "Unexpectedly deleted the key from store for a second time - error: %v key: %v", err, key)
		assert.Equal(t, ErrStoreKeyNotFound(key), err, "unexpected failure when looking for a previously deleted key - error %v", err)
		assert.Nil(t, deleteResponse)

		// Now write the key
		//
		writeResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		assert.Equal(t, 0, len(writeResponse.Records))
		assert.Lessf(t, revStoreInitial, writeResponse.Revision, "Unexpected value for store revision on write completion")

		// Verify the key is now there.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
		assert.Equal(t, writeResponse.Revision, readResponse.Revision)

		testCompareReadResponseToWrite(t, readResponse, writeRequest, writeResponse)

		// Delete the key we just wrote
		//
		deleteResponse, err = store.DeleteTxn(ctx, deleteRequest)
		assert.Nilf(t, err, "Failed to delete key from store - error: %v key: %v", err, key)
		require.NotNil(t, deleteResponse)
		assert.Equal(t, 0, len(deleteResponse.Records))
		assert.Less(t, readResponse.Revision, deleteResponse.Revision)

		// Attempt to re-read the key which once again should be absent.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.NotNilf(t, err, "Unexpected success reading key supposedly absent - error: %v key: %v", err)
		assert.Equal(t, ErrStoreKeyNotFound(key), err, "unexpected failure when looking for an absent key - error %v", err)
		assert.Nilf(t, readResponse, "Unexpected response for read of invalid key - error: %v key: %v", err, key)

		store.Disconnect()

		store = nil

		return nil
	})

	return
}

func TestStoreWriteDeleteMultipleTxnRequired(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreWriteDeleteMultipleTxnRequired"
		key := testGenerateKeyFromName(testName)

		writeRequest := testGenerateRequestForWrite(keySetSize, key)
		readRequest := testGenerateRequestForReadWithCondition(keySetSize, key, ConditionUnconditional)
		deleteRequest := testGenerateRequestForDeleteWithCondition(keySetSize, key, ConditionRequired)

		assert.Equal(t, keySetSize, len(writeRequest.Records))
		assert.Equal(t, keySetSize, len(readRequest.Records))
		assert.Equal(t, keySetSize, len(deleteRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		// Verify none of the keys exist.
		//
		readResponse, err := store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equalf(t, 0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
		assert.Lessf(t, revStoreInitial, readResponse.Revision, "Unexpected value for store revision on read completion")

		// Attempt to delete the non-existent set of keys
		//
		deleteResponse, err := store.DeleteTxn(ctx, deleteRequest)
		assert.NotNilf(t, err, "Failed to read from store - error: %v", err)

		// We have a slight problem here in that we do not know which key of the set will be reported
		// in the error. That means it could be any of them, so we need to check for them all and only
		// assert if none of them match.
		//
		var foundError bool

		for k := range readRequest.Records {
			if err == ErrStoreKeyNotFound(k) {
				foundError = true
				break
			}
		}

		assert.Truef(t, foundError, "Returned error failed to match any of the expected values - err: %v", err)
		assert.Nilf(t, deleteResponse, "Unexpected response for read of absent keys - error: %v", err)

		// Now write the keys
		//
		writeResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		assert.Equal(t, 0, len(writeResponse.Records))
		assert.Lessf(t, readResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

		// Verify all the expected keys are present
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(writeRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
		assert.Equal(t, writeResponse.Revision, readResponse.Revision)

		testCompareReadResponseToWrite(t, readResponse, writeRequest, writeResponse)

		// Delete the keys we just wrote
		//
		deleteResponse, err = store.DeleteTxn(ctx, deleteRequest)
		assert.Nilf(t, err, "Failed to delete one or more keys from store - error: %v key: %v", err)
		require.NotNil(t, deleteResponse)
		require.Equal(t, 0, len(deleteResponse.Records))
		assert.Less(t, readResponse.Revision, deleteResponse.Revision)

		// and finally, verify none of the keys remain after the delete.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equalf(t, 0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
		assert.Equalf(t, deleteResponse.Revision, readResponse.Revision, "Unexpected value for store revision on read completion")

		store.Disconnect()

		store = nil

		return nil
	})

	return
}

func TestStoreWriteDeleteMultipleTxnOptional(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreWriteDeleteMultipleTxnOptional"
		key := testGenerateKeyFromName(testName)

		writeRequest := testGenerateRequestForWrite(keySetSize, key)
		readRequest := testGenerateRequestForReadWithCondition(keySetSize, key, ConditionUnconditional)
		deleteRequest := testGenerateRequestForDeleteWithCondition(keySetSize, key, ConditionUnconditional)

		assert.Equal(t, keySetSize, len(writeRequest.Records))
		assert.Equal(t, keySetSize, len(readRequest.Records))
		assert.Equal(t, keySetSize, len(deleteRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		// Verify none of the keys exist.
		//
		readResponse, err := store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equalf(t, 0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
		assert.Lessf(t, revStoreInitial, readResponse.Revision, "Unexpected value for store revision on read completion")

		// Attempt to delete the non-existent set of keys. This is an unconditional request and so
		// should succeed regardless of the presence or absence of the listed keys.
		//
		deleteResponse, err := store.DeleteTxn(ctx, deleteRequest)
		assert.Nilf(t, err, "Unexpected failure in unconditionally deleting absent keys - error: %v", err)
		require.NotNilf(t, deleteResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, deleteResponse.Revision, deleteResponse.Revision)

		// Now write the keys
		//
		writeResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		assert.Equal(t, 0, len(writeResponse.Records))
		assert.Lessf(t, deleteResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

		// Verify all the expected keys are present
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(writeRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
		assert.Equal(t, writeResponse.Revision, readResponse.Revision)

		testCompareReadResponseToWrite(t, readResponse, writeRequest, writeResponse)

		// Delete the keys we just wrote
		//
		deleteResponse, err = store.DeleteTxn(ctx, deleteRequest)
		assert.Nilf(t, err, "Failed to delete one or more keys from store - error: %v key: %v", err)
		require.NotNil(t, deleteResponse)
		require.Equal(t, 0, len(deleteResponse.Records))
		assert.Less(t, readResponse.Revision, deleteResponse.Revision)

		// and finally, verify none of the keys remain after the delete.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equalf(t, 0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
		assert.Equalf(t, deleteResponse.Revision, readResponse.Revision, "Unexpected value for store revision on read completion")

		store.Disconnect()

		store = nil

		return nil
	})

	return
}

func TestStoreWriteDeleteMultipleTxnPartialRequired(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreWriteDeleteMultipleTxnPartial"
		key := testGenerateKeyFromName(testName)

		writeRequestPartial := testGenerateRequestForWrite(1, key)
		writeRequestComplete := testGenerateRequestForWrite(2, key)

		readRequest := testGenerateRequestForReadWithCondition(len(writeRequestComplete.Records), key, ConditionUnconditional)
		deleteRequest := testGenerateRequestForDeleteWithCondition(len(writeRequestComplete.Records), key, ConditionRequired)

		assert.Equal(t, 1, len(writeRequestPartial.Records))
		assert.Equal(t, 2, len(writeRequestComplete.Records))
		assert.Equal(t, len(writeRequestComplete.Records), len(readRequest.Records))
		assert.Equal(t, len(writeRequestComplete.Records), len(deleteRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		// Verify none of the keys exist.
		//
		readResponse, err := store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equalf(t, 0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
		assert.Lessf(t, revStoreInitial, readResponse.Revision, "Unexpected value for store revision on read completion")

		// Now write a partial set of keys
		//
		writeResponse, err := store.WriteTxn(ctx, writeRequestPartial)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		assert.Equal(t, 0, len(writeResponse.Records))
		assert.Lessf(t, readResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

		// Verify we have what we expect by trying to read the complete set of
		// key/value pairs, only some of which should be there.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(writeRequestPartial.Records), len(readResponse.Records), "Read returned unexpected number of records")
		assert.Equal(t, writeResponse.Revision, readResponse.Revision)

		testCompareReadResponseToWrite(t, readResponse, writeRequestPartial, writeResponse)

		// Determine which key is in both the partial and complete write
		// request (keyPartial), and which key is only in the complete
		// write set (keyComplete)
		//
		var keyPartial string
		var keyComplete string

		for k := range writeRequestPartial.Records {
			keyPartial = k
			break
		}

		for k := range writeRequestComplete.Records {
			if k != keyPartial {
				keyComplete = k
				break
			}
		}

		// Attempt to delete the full set of keys. This should fail, with no
		// response, and all the keys should remain
		//
		deleteResponse, err := store.DeleteTxn(ctx, deleteRequest)
		assert.NotNilf(t, err, "Failed to delete one or more keys from store - error: %v key: %v", err)
		assert.Equal(t, ErrStoreKeyNotFound(keyComplete), err)
		assert.Nil(t, deleteResponse)

		// Verify we still have what we expect by trying to read the complete set of
		// key/value pairs, only some of which should be there.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(writeRequestPartial.Records), len(readResponse.Records), "Read returned unexpected number of records")
		assert.Equal(t, writeResponse.Revision, readResponse.Revision)

		testCompareReadResponseToWrite(t, readResponse, writeRequestPartial, writeResponse)

		// Now write a complete set of keys
		//
		writeResponse, err = store.WriteTxn(ctx, writeRequestComplete)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		assert.Equal(t, 0, len(writeResponse.Records))
		assert.Lessf(t, readResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

		// Verify we have the complete set of key/value pairs all of which should now be there.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(writeRequestComplete.Records), len(readResponse.Records), "Read returned unexpected number of records")
		assert.Equal(t, writeResponse.Revision, readResponse.Revision)

		testCompareReadResponseToWrite(t, readResponse, writeRequestComplete, writeResponse)

		// Once again attempt to delete the complete set of keys and this time, the delete should succeed.
		//
		deleteResponse, err = store.DeleteTxn(ctx, deleteRequest)
		assert.Nilf(t, err, "Failed to delete one or more keys from store - error: %v key: %v", err)
		require.NotNil(t, deleteResponse)
		require.Equal(t, 0, len(deleteResponse.Records))
		assert.Less(t, readResponse.Revision, deleteResponse.Revision)

		// and finally, verify none of the keys remain after the delete.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equalf(t, 0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
		assert.Equalf(t, deleteResponse.Revision, readResponse.Revision, "Unexpected value for store revision on read completion")

		store.Disconnect()

		store = nil

		return nil
	})

	return
}

func TestStoreWriteDeleteMultipleTxnPartialOptional(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreWriteDeleteMultipleTxnPartialOptional"
		key := testGenerateKeyFromName(testName)

		writeRequestPartial := testGenerateRequestForWrite(1, key)
		writeRequestComplete := testGenerateRequestForWrite(2, key)

		readRequest := testGenerateRequestForReadWithCondition(len(writeRequestComplete.Records), key, ConditionUnconditional)
		deleteRequest := testGenerateRequestForDeleteWithCondition(len(writeRequestComplete.Records), key, ConditionUnconditional)

		assert.Equal(t, 1, len(writeRequestPartial.Records))
		assert.Equal(t, 2, len(writeRequestComplete.Records))
		assert.Equal(t, len(writeRequestComplete.Records), len(readRequest.Records))
		assert.Equal(t, len(writeRequestComplete.Records), len(deleteRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		// Verify none of the keys exist.
		//
		readResponse, err := store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equalf(t, 0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
		assert.Lessf(t, revStoreInitial, readResponse.Revision, "Unexpected value for store revision on read completion")

		// Now write a partial set of keys
		//
		writeResponse, err := store.WriteTxn(ctx, writeRequestPartial)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		assert.Equal(t, 0, len(writeResponse.Records))
		assert.Lessf(t, readResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

		// Verify we have what we expect by trying to read the complete set of
		// key/value pairs, only some of which should be there.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(writeRequestPartial.Records), len(readResponse.Records), "Read returned unexpected number of records")
		assert.Equal(t, writeResponse.Revision, readResponse.Revision)

		testCompareReadResponseToWrite(t, readResponse, writeRequestPartial, writeResponse)

		// Attempt to delete the full set of keys. This should succeed and all the
		// keys should have been deleted.
		//
		deleteResponse, err := store.DeleteTxn(ctx, deleteRequest)
		assert.Nilf(t, err, "Failed to delete one or more keys from store - error: %v key: %v", err)
		require.NotNil(t, deleteResponse)
		assert.Equal(t, 0, len(deleteResponse.Records))
		assert.Less(t, readResponse.Revision, deleteResponse.Revision)

		// and finally, verify none of the keys remain after the delete.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equalf(t, 0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
		assert.Equalf(t, deleteResponse.Revision, readResponse.Revision, "Unexpected value for store revision on read completion")

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
		key := testGenerateKeyFromName(testName)

		writeRequest := testGenerateRequestForWrite(keySetSize, key)
		readRequest := testGenerateRequestForReadWithCondition(keySetSize, key, ConditionUnconditional)
		deleteRequest := testGenerateRequestForDelete(keySetSize, key)

		assert.Equal(t, keySetSize, len(writeRequest.Records))
		assert.Equal(t, keySetSize, len(readRequest.Records))
		assert.Equal(t, keySetSize, len(deleteRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		// Write the keys to the store
		//
		writeResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		assert.Equal(t, 0, len(writeResponse.Records))
		assert.Lessf(t, revStoreInitial, writeResponse.Revision, "Unexpected value for store revision on write completion")

		// Verify all the expected keys are present
		//
		readResponse, err := store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(writeRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
		assert.Equal(t, writeResponse.Revision, readResponse.Revision)

		testCompareReadResponseToWrite(t, readResponse, writeRequest, writeResponse)

		// Now delete the keys by prefix. Note that because of the way the request are
		// built, the supplied "key" argument is an effective prefix for the set of keys.
		//
		deleteResponse, err := store.DeleteWithPrefix(ctx, key)
		assert.Nilf(t, err, "Failed to delete the keys from the store - error: %v prefix: %v", err, key)
		require.NotNil(t, deleteResponse)
		assert.Equal(t, 0, len(deleteResponse.Records))
		assert.Less(t, writeResponse.Revision, deleteResponse.Revision)

		// and finally, verify none of the keys remain after the delete.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equalf(t, 0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
		assert.Equalf(t, deleteResponse.Revision, readResponse.Revision, "Unexpected value for store revision on read completion")

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

		writeRequest := testGenerateRequestForWriteCreate(keySetSize, testName)
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
		assert.Equal(t, 0, len(createResponse.Records))
		assert.Lessf(t, revStoreInitial, createResponse.Revision, "Unexpected value for store revision on write(create) completion")

		// The write claimed to succeed, now go fetch the record(s) and verify
		// the revision(s) and value(s) are as expected
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNil(t, readResponse)
		assert.Equalf(t, createResponse.Revision, readResponse.Revision, "Unexpected value for store revision given no updates")
		assert.Equalf(t, len(writeRequest.Records), len(readResponse.Records), "Unexpected numbers of records returned")

		testCompareReadResponseToWrite(t, readResponse, writeRequest, createResponse)

		// Try to re-create the same keys. These should fail and the original values and revisions should survive.
		//
		recreateResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.NotNilf(t, err, "Succeeded where we expected to get a failed store write - error: %v", err)
		require.Nil(t, recreateResponse)
		// TODO		assert.Equalf(t, RevisionInvalid, recreateResponse.Records, "Unexpected value for store revision on write(re-create) completion")

		readRecreateResponse, err := store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNil(t, readRecreateResponse)
		assert.Equalf(t, createResponse.Revision, readRecreateResponse.Revision, "Unexpected value for store revision given no updates")
		assert.Equalf(t, len(writeRequest.Records), len(readRecreateResponse.Records), "Unexpected numbers of records returned")

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

		writeRequest := testGenerateRequestForWrite(keySetSize, testName)
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
		assert.Equal(t, 0, len(createResponse.Records))
		assert.Lessf(t, readResponse.Revision, createResponse.Revision, "Unexpected value for store revision on write completion")

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
		updateRequest := testGenerateRequestFromReadResponse(readResponse)

		updateResponse, err := store.WriteTxn(ctx, updateRequest)
		assert.Nilf(t, err, "Failed to write unconditional update to store - error: %v", err)
		require.NotNil(t, updateResponse)
		assert.Equal(t, 0, len(updateResponse.Records))
		assert.Lessf(t, readResponse.Revision, updateResponse.Revision, "Expected new store revision to be greater than the earlier store revision")

		readResponseUpdate, err := store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNil(t, readResponseUpdate)
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
		key := testGenerateKeyFromName(testName)
		keySetSize := 1

		writeRequest := testGenerateRequestForWrite(keySetSize, key)
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
		require.Equal(t, 0, len(createResponse.Records))
		assert.Lessf(t, readResponse.Revision, createResponse.Revision, "Unexpected value for store revision on write(create) completion")

		// The write claimed to succeed, now go fetch the record(s) and verify
		// the revision(s) and value(s) are as expected
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNil(t, readResponse)
		assert.Equalf(t, createResponse.Revision, readResponse.Revision, "Unexpected value for store revision given no updates")
		assert.Equalf(t, len(writeRequest.Records), len(readResponse.Records), "Unexpected numbers of records returned")

		testCompareReadResponseToWrite(t, readResponse, writeRequest, createResponse)

		// We verified the write worked, so try a conditional update when the revisions
		// are equal. Set the required condition and change the value so we can verify
		// after the update.
		//
		updateRequest := testGenerateRequestFromWReadResponseWithCondition(readResponse, ConditionRevisionEqual)

		updateResponse, err := store.WriteTxn(ctx, updateRequest)
		assert.Nilf(t, err, "Failed to write conditional update to store - error: %v", err)
		require.NotNil(t, updateResponse)
		assert.Equal(t, 0, len(updateResponse.Records))
		assert.Lessf(t, readResponse.Revision, updateResponse.Revision, "Expected new store revision to be greater than the earlier store revision")

		// verify the update happened as expected
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNil(t, readResponse)
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
		key := testGenerateKeyFromName(testName)

		writeRequest := testGenerateRequestForWrite(keySetSize, key)

		assert.Equal(t, keySetSize, len(writeRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		writeResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		assert.Equal(t, 0, len(writeResponse.Records))
		assert.Lessf(t, revStoreInitial, writeResponse.Revision, "Unexpected value for store revision on write completion")

		// Look for a set of prefixed key/value pairs which we do expect to be present.
		//
		listResponse, err := store.ListWithPrefix(ctx, key)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, listResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(writeRequest.Records), len(listResponse.Records), "Failed to get the expected number of response values")

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
			_, present := writeRequest.Records[k]
			assert.Truef(t, present, "Extra key: %v record: %v", k, r)
			if present {
				val := writeRequest.Records[k].Value
				assert.Equalf(t, val, r.Value, "key: %v Expected: %q Actual %q", k, val, r.Value)
			}
		}

		store.Disconnect()

		store = nil

		return nil
	})

	return
}

func TestStoreListWithPrefixEmptySet(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	_ = st.WithSpan(context.Background(), func(ctx context.Context) (err error) {
		testName := "TestStoreListWithPrefixEmptySet"
		key := testGenerateKeyFromName(testName)

		writeRequest := testGenerateRequestForWrite(keySetSize, key)
		readRequest := testGenerateRequestForReadWithCondition(keySetSize, key, ConditionUnconditional)

		assert.Equal(t, keySetSize, len(writeRequest.Records))
		assert.Equal(t, keySetSize, len(readRequest.Records))

		store := NewStore()
		assert.NotNilf(t, store, "Failed to get the store as expected")

		err = store.Connect()
		assert.Nilf(t, err, "Failed to connect to store - error: %v", err)

		// Attempt to read from the set of keys before they exist. As this is a "ConditionUnconditional" read
		// request, it should succeed and produce an empty response.
		//
		readResponse, err := store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equalf(t, 0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
		assert.Lessf(t, revStoreInitial, readResponse.Revision, "Unexpected value for store revision on read completion")

		// Look for a prefix name after verifying the keys are absent.
		//
		// We expect success with a non-nil but empty set.
		//
		listResponse, err := store.ListWithPrefix(ctx, key)
		assert.Nilf(t, err, "Unexpected failure attempting to list non-existing key set - error: %v prefixKey: %v", err, key)
		require.NotNilf(t, listResponse, "Failed to get a non-nil response as expected - error: %v prefixKey: %v", err, key)
		assert.Equal(t, 0, len(listResponse.Records), "Got more results than expected")
		assert.Equal(t, readResponse.Revision, listResponse.Revision)

		if len(listResponse.Records) > 0 {
			for k, r := range listResponse.Records {
				t.Logf("Unexpected key/value pair key: %v value: %v", k, r.Value)
			}
		}

		// Now write the keys
		//
		writeResponse, err := store.WriteTxn(ctx, writeRequest)
		assert.Nilf(t, err, "Failed to write to store - error: %v", err)
		require.NotNil(t, writeResponse)
		assert.Equal(t, 0, len(writeResponse.Records))
		assert.Lessf(t, readResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

		// verify the existence of the keys we just wrote.
		//
		readResponse, err = store.ReadTxn(ctx, readRequest)
		assert.Nilf(t, err, "Failed to read from store - error: %v", err)
		require.NotNilf(t, readResponse, "Failed to get a response as expected - error: %v", err)
		assert.Equal(t, len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
		assert.Equal(t, writeResponse.Revision, readResponse.Revision)

		testCompareReadResponseToWrite(t, readResponse, writeRequest, writeResponse)

		// Now look for a set of prefixed key/value pairs which we now expect and have verified to be present.
		//
		listResponse, err = store.ListWithPrefix(ctx, key)
		assert.Nilf(t, err, "Unexpected failure attempting to list existing key set - error: %v prefixKey: %v", err, key)
		require.NotNilf(t, listResponse, "Failed to get a response as expected - error: %v prefixKey: %v", err, key)
		assert.Equal(t, len(writeRequest.Records), len(listResponse.Records), "Failed to get the expected number of response values")
		assert.Equal(t, readResponse.Revision, readResponse.Revision)

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
			_, present := writeRequest.Records[k]
			assert.Truef(t, present, "Extra key: %v record: %v", k, r)
			if present {
				val := writeRequest.Records[k].Value
				assert.Equalf(t, val, r.Value, "key: %v Expected: %q Actual %q", k, val, r.Value)
			}
		}

		store.Disconnect()

		store = nil

		return nil
	})

	return
}
