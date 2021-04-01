// Unit tests for the web service store package
//
package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

const revStoreInitial = int64(0)

type storeTestSuite struct {
	testSuiteCore

	store *Store
}

func (ts *storeTestSuite) SetupSuite() {
	require := ts.Require()

	ts.testSuiteCore.SetupSuite()

	ts.store = NewStore()
	require.NotNil(ts.store, "Failed to get the store as expected")
}

func (ts *storeTestSuite) SetupTest() {
	require := ts.Require()

	require.NoError(ts.utf.Open(ts.T()))
	require.NoError(ts.store.Connect())
}

func (ts *storeTestSuite) TearDownTest() {
	ts.store.Disconnect()
	ts.utf.Close()
}

func (ts *storeTestSuite) TestNew() {
	require := ts.Require()

	store := NewStore()

	require.NotNil(store)

	store = nil
}

func (ts *storeTestSuite) TestInitialize() {
	assert  := ts.Assert()
	require := ts.Require()

	store := NewStore()

	require.NotNil(store, "Failed to get the store as expected")
	assert.Equal(getDefaultEndpoints(), store.Endpoints, "Mismatch in initialization of endpoints")
	assert.Equal(getDefaultTimeoutConnect(), store.TimeoutConnect, "Mismatch in initialization of connection timeout")
	assert.Equal(getDefaultTimeoutRequest(), store.TimeoutRequest, "Mismatch in initialization of request timeout")
	assert.Equal(getDefaultTraceFlags(), store.TraceFlags, "Mismatch in initialization of trace flags")
	assert.Equal(getDefaultNamespaceSuffix(), store.NamespaceSuffix, "Mismatch in initialization of namespace suffix")

	endpoints := []string{"localhost:8080", "localhost:8181"}
	timeoutConnect := getDefaultTimeoutConnect() * 2
	timeoutRequest := getDefaultTimeoutRequest() * 3
	traceFlags := traceFlagEnabled
	namespaceSuffix := getDefaultNamespaceSuffix() + "/Suffix"

	err := store.Initialize(endpoints, timeoutConnect, timeoutRequest, traceFlags, namespaceSuffix)

	require.NoError(err, "Failed to initialize new store - error: %v", err)
	assert.Equal(endpoints, store.Endpoints, "Mismatch in initialization of endpoints")
	assert.Equal(timeoutConnect, store.TimeoutConnect, "Mismatch in initialization of connection timeout")
	assert.Equal(timeoutRequest, store.TimeoutRequest, "Mismatch in initialization of request timeout")
	assert.Equal(traceFlags, store.TraceFlags, "Mismatch in initialization of trace flags")
	assert.Equal(namespaceSuffix, store.NamespaceSuffix, "Mismatch in initialization of namespace suffix")

	store = nil
}

func (ts *storeTestSuite) TestNewWithArgs() {
	assert  := ts.Assert()
	require := ts.Require()

	// Use non-default values to ensure we get what we asked for and not the defaults.
	//
	endpoints := []string{"localhost:8282", "localhost:8383"}
	timeoutConnect := getDefaultTimeoutConnect() * 4
	timeoutRequest := getDefaultTimeoutRequest() * 5
	traceFlags := traceFlagExpandResults
	namespaceSuffix := getDefaultNamespaceSuffix()

	store := New(endpoints, timeoutConnect, timeoutRequest, traceFlags, namespaceSuffix)

	require.NotNil(store, "Failed to get the store as expected")
	assert.Equal(endpoints, store.Endpoints, "Mismatch in initialization of endpoints")
	assert.Equal(timeoutConnect, store.TimeoutConnect, "Mismatch in initialization of connection timeout")
	assert.Equal(timeoutRequest, store.TimeoutRequest, "Mismatch in initialization of request timeout")
	assert.Equal(traceFlags, store.TraceFlags, "Mismatch in initialization of trace flags")
	assert.Equal(namespaceSuffix, store.NamespaceSuffix, "Mismatch in initialization of namespace suffix")

	store = nil
}

func (ts *storeTestSuite) TestStoreSetAndGet() {
	assert  := ts.Assert()
	require := ts.Require()

	store := NewStore()

	require.NotNil(store, "Failed to get the store as expected")
	assert.Equal(getDefaultEndpoints(), store.Endpoints, "Mismatch in initialization of endpoints")
	assert.Equal(getDefaultTimeoutConnect(), store.TimeoutConnect, "Mismatch in initialization of connection timeout")
	assert.Equal(getDefaultTimeoutRequest(), store.TimeoutRequest, "Mismatch in initialization of request timeout")
	assert.Equal(getDefaultTraceFlags(), store.TraceFlags, "Mismatch in initialization of trace flags")
	assert.Equal(getDefaultNamespaceSuffix(), store.NamespaceSuffix, "Mismatch in initialization of namespace suffix")

	assert.Equal(store.Endpoints, store.GetAddress(), "Mismatch in fetch of endpoints")
	assert.Equal(store.TimeoutConnect, store.GetTimeoutConnect(), "Mismatch in fetch of connection timeout")
	assert.Equal(store.TimeoutRequest, store.GetTimeoutRequest(), "Mismatch in fetch of request timeout")
	assert.Equal(store.TraceFlags, store.GetTraceFlags(), "Mismatch in fetch of trace flags")

	endpoints := []string{"localhost:8484", "localhost:8585"}
	timeoutConnect := getDefaultTimeoutConnect() * 6
	timeoutRequest := getDefaultTimeoutRequest() * 7
	traceFlags := traceFlagExpandResults
	namespaceSuffix := getDefaultNamespaceSuffix() + "/Suffix2"

	err := store.Initialize(endpoints, timeoutConnect, timeoutRequest, traceFlags, namespaceSuffix)

	require.NoError(err, "Failed to update new store - error: %v", err)
	assert.Equal(endpoints, store.Endpoints, "Mismatch in update of endpoints")
	assert.Equal(timeoutConnect, store.TimeoutConnect, "Mismatch in update of connection timeout")
	assert.Equal(timeoutRequest, store.TimeoutRequest, "Mismatch in update of request timeout")
	assert.Equal(traceFlags, store.TraceFlags, "Mismatch in update of trace flags")
	assert.Equal(namespaceSuffix, store.NamespaceSuffix, "Mismatch in update of namespace suffix")

	assert.Equal(store.Endpoints, store.GetAddress(), "Mismatch in re-fetch of endpoints")
	assert.Equal(store.TimeoutConnect, store.GetTimeoutConnect(), "Mismatch in re-fetch of connection timeout")
	assert.Equal(store.TimeoutRequest, store.GetTimeoutRequest(), "Mismatch in re-fetch of request timeout")
	assert.Equal(store.TraceFlags, store.GetTraceFlags(), "Mismatch in re-fetch of trace flags")
	assert.Equal(store.NamespaceSuffix, store.GetNamespaceSuffix(), "Mismatch in re-fetch of namespace suffix")
}

func (ts *storeTestSuite) TestStoreConnectDisconnect() {
	require := ts.Require()

	store := NewStore()
	require.NotNil(store, "Failed to get the store as expected")

	err := store.Connect()
	require.NoError(err, "Failed to connect to store - error: %v", err)

	err = store.Connect()
	require.ErrorIs(errors.ErrStoreConnected("already connected"), err)

	store.Disconnect()

	// Try a second disconnect. Benign but should trigger different execution
	// path for coverage numbers.
	//
	store.Disconnect()

	store = nil
}

func (ts *storeTestSuite) TestStoreConnectDisconnectWithInitialize() {
	require := ts.Require()

	store := NewStore()
	require.NotNil(store, "Failed to get the store as expected")

	err := store.Initialize(
		getDefaultEndpoints(),
		getDefaultTimeoutConnect(),
		getDefaultTimeoutRequest(),
		getDefaultTraceFlags(),
		getDefaultNamespaceSuffix())
	require.NoError(err, "Failed to re-initialize store - error: %v", err)

	err = store.Connect()
	require.NoError(err, "Failed to connect to store - error: %v", err)

	err = store.Initialize(
		getDefaultEndpoints(),
		getDefaultTimeoutConnect(),
		getDefaultTimeoutRequest(),
		getDefaultTraceFlags(),
		getDefaultNamespaceSuffix())

	require.ErrorIs(errors.ErrStoreConnected("already connected"), err)

	store.Disconnect()

	// Try a second disconnect. Benign but should trigger different execution
	// path for coverage numbers.
	//
	store.Disconnect()

	store = nil
}

func (ts *storeTestSuite) TestStoreConnectDisconnectWithSet() {
	require := ts.Require()

	endpoints := getDefaultEndpoints()
	timeoutConnect := getDefaultTimeoutConnect()
	timeoutRequest := getDefaultTimeoutRequest()
	traceFlags := getDefaultTraceFlags()
	namespaceSuffix := getDefaultNamespaceSuffix()

	store := New(endpoints, timeoutConnect, timeoutRequest, traceFlags, namespaceSuffix)
	require.NotNil(store, "Failed to get the store as expected")

	err := store.SetAddress(endpoints)
	require.NoError(err, "Failed to update the address - error: %v", err)

	err = store.SetTimeoutConnect(timeoutConnect)
	require.NoError(err, "Failed to update the connect timeout - error: %v", err)

	err = store.SetTimeoutRequest(timeoutRequest)
	require.NoError(err, "Failed to update the request timeout - error: %v", err)

	err = store.SetNamespaceSuffix(namespaceSuffix)
	require.NoError(err, "Failed to update the namespace suffix - error: %v", err)

	store.SetTraceFlags(0)
	store.SetTraceFlags(traceFlagEnabled)
	store.SetTraceFlags(traceFlagExpandResults)
	store.SetTraceFlags(traceFlagEnabled | traceFlagExpandResults)

	err = store.Connect()
	require.NoError(err, "Failed to connect to store - error: %v", err)

	err = store.SetAddress(endpoints)
	require.ErrorIs(errors.ErrStoreConnected("already connected"), err)

	err = store.SetTimeoutConnect(timeoutConnect)
	require.ErrorIs(errors.ErrStoreConnected("already connected"), err)

	err = store.SetTimeoutRequest(timeoutRequest)
	require.NoError(err, "Failed to update the request timeout - error: %v", err)

	err = store.SetNamespaceSuffix(namespaceSuffix)
	require.ErrorIs(errors.ErrStoreConnected("already connected"), err)

	err = store.Connect()
	require.ErrorIs(errors.ErrStoreConnected("already connected"), err)

	store.Disconnect()

	// Try a second disconnect. Benign but should trigger different execution
	// path for coverage numbers.
	//
	store.Disconnect()

	store = nil
}

func (ts *storeTestSuite) TestStoreWriteReadTxn() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreWriteReadTxn"
	key := testGenerateKeyFromName(testName)

	writeRequest := testGenerateRequestForWrite(1, key)
	readRequest := testGenerateRequestForRead(1, key)

	assert.Equal(1, len(writeRequest.Records))
	assert.Equal(1, len(readRequest.Records))

	writeResponse, err := ts.store.WriteTxn(ctx, writeRequest)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(revStoreInitial, writeResponse.Revision, "Unexpected value for store revision on write completion")

	readResponse, err := ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
	assert.Equal(writeResponse.Revision, readResponse.Revision)

	ts.testCompareReadResponseToWrite(readResponse, writeRequest, writeResponse)
}

func (ts *storeTestSuite) TestStoreWriteReadTxnRequired() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreWriteReadTxnRequired"
	missingName := "MissingName"
	key := testGenerateKeyFromNames(testName, missingName)

	writeRequest := testGenerateRequestForSimpleWrite(key)
	readRequest := testGenerateRequestForSimpleReadWithCondition(key, ConditionRequired)

	assert.Equal(1, len(writeRequest.Records))
	assert.Equal(1, len(readRequest.Records))

	// Attempt to read the key before it exists. As this is a "ConditionRequired" read
	// request, it should fail and produce no response.
	//
	readResponse, err := ts.store.ReadTxn(ctx, readRequest)
	require.ErrorIs(errors.ErrStoreKeyNotFound(key), err)
	assert.Nil(readResponse, "Unexpected response for read of invalid key - error: %v key: %v", err, key)

	// Now write the key
	//
	writeResponse, err := ts.store.WriteTxn(ctx, writeRequest)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(revStoreInitial, writeResponse.Revision, "Unexpected value for store revision on write completion")

	// Now try to read the key which should now be there.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
	assert.Equal(writeResponse.Revision, readResponse.Revision)

	ts.testCompareReadResponseToWrite(readResponse, writeRequest, writeResponse)
}

func (ts *storeTestSuite) TestStoreWriteReadTxnOptional() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreWriteReadTxnOptional"
	missingName := "MissingName"
	key := testGenerateKeyFromNames(testName, missingName)

	writeRequest := testGenerateRequestForSimpleWrite(key)
	readRequest := testGenerateRequestForSimpleReadWithCondition(key, ConditionUnconditional)

	assert.Equal(1, len(writeRequest.Records))
	assert.Equal(1, len(readRequest.Records))

	// Attempt to read the key before it exists. As this is a "ConditionUnconditional" read
	// request, it should succeed and produce an empty response.
	//
	readResponse, err := ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Unexpected failure unconditionally reading key supposedly absent - error: %v key: %v", err, key)
	require.NotNil(readResponse, "Unexpected missing response for unconditional read of absent key - error: %v key: %v", err, key)
	assert.Equal(0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
	assert.Less(revStoreInitial, readResponse.Revision, "Unexpected value for store revision on read completion")

	// Now write the key
	//
	writeResponse, err := ts.store.WriteTxn(ctx, writeRequest)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(readResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

	// Now try to read a key which should now be there.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Unexpected failure unconditionally reading key supposedly present - error: %v key: %v", err, key)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
	assert.Equal(writeResponse.Revision, readResponse.Revision)

	ts.testCompareReadResponseToWrite(readResponse, writeRequest, writeResponse)
}

func (ts *storeTestSuite) TestStoreWriteReadMultipleTxn() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreWriteReadMultipleTxn"
	key := testGenerateKeyFromName(testName)

	writeRequest := testGenerateRequestForWrite(keySetSize, key)
	readRequest := testGenerateRequestForRead(keySetSize, key)

	assert.Equal(keySetSize, len(writeRequest.Records))
	assert.Equal(keySetSize, len(readRequest.Records))

	writeResponse, err := ts.store.WriteTxn(ctx, writeRequest)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(revStoreInitial, writeResponse.Revision, "Unexpected value for store revision on write completion")

	// Now try to read the key/value pairs which should be there.
	//
	readResponse, err := ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
	assert.Equal(writeResponse.Revision, readResponse.Revision)

	ts.testCompareReadResponseToWrite(readResponse, writeRequest, writeResponse)
}

func (ts *storeTestSuite) TestStoreWriteReadMultipleTxnRequired() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreWriteReadMultipleTxnRequired"
	key := testGenerateKeyFromName(testName)

	writeRequest := testGenerateRequestForWrite(keySetSize, key)
	readRequest := testGenerateRequestForReadWithCondition(keySetSize, key, ConditionRequired)

	assert.Equal(keySetSize, len(writeRequest.Records))
	assert.Equal(keySetSize, len(readRequest.Records))

	// Attempt to read from the set of keys before they exist. As this is a "ConditionRequired" read
	// request, it should fail and produce no response.
	//
	readResponse, err := ts.store.ReadTxn(ctx, readRequest)
	require.ErrorContains(err, errors.ErrStoreKeyNotFound(testName).Error())

	// We have a slight problem here in that we do not know which key of the set will be reported
	// in the error. That means it could be any of them, so we need to check for them all and only
	// assert if none of them match.
	//
	var foundError bool

	for k := range readRequest.Records {
		if err == errors.ErrStoreKeyNotFound(k) {
			foundError = true
			break
		}
	}

	assert.True(foundError, "Returned error failed to match any of the expected values - err: %v", err)
	assert.Nil(readResponse, "Unexpected response for read of absent keys - error: %v", err)

	// Now write the keys
	//
	writeResponse, err := ts.store.WriteTxn(ctx, writeRequest)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(revStoreInitial, writeResponse.Revision, "Unexpected value for store revision on write completion")

	// Finally try to read the key/value pairs which should now be there.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
	assert.Equal(writeResponse.Revision, readResponse.Revision)

	ts.testCompareReadResponseToWrite(readResponse, writeRequest, writeResponse)
}

func (ts *storeTestSuite) TestStoreWriteReadMultipleTxnOptional() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreWriteReadMultipleTxnOptional"
	key := testGenerateKeyFromName(testName)

	writeRequest := testGenerateRequestForWrite(keySetSize, key)
	readRequest := testGenerateRequestForReadWithCondition(keySetSize, key, ConditionUnconditional)

	assert.Equal(keySetSize, len(writeRequest.Records))
	assert.Equal(keySetSize, len(readRequest.Records))

	// Attempt to read from the set of keys before they exist. As this is a "ConditionUnconditional" read
	// request, it should succeed and produce an empty response.
	//
	readResponse, err := ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
	assert.Less(revStoreInitial, readResponse.Revision, "Unexpected value for store revision on read completion")

	// Now write the keys
	//
	writeResponse, err := ts.store.WriteTxn(ctx, writeRequest)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(readResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

	// Finally try to read the key/value pairs which should now be there.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
	assert.Equal(writeResponse.Revision, readResponse.Revision)

	ts.testCompareReadResponseToWrite(readResponse, writeRequest, writeResponse)
}

func (ts *storeTestSuite) TestStoreWriteReadMultipleTxnPartial() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreWriteReadMultipleTxnPartial"
	key := testGenerateKeyFromName(testName)

	writeRequestPartial := testGenerateRequestForWrite(1, key)
	writeRequestComplete := testGenerateRequestForWrite(2, key)

	readRequest := testGenerateRequestForReadWithCondition(len(writeRequestComplete.Records), key, ConditionUnconditional)

	assert.Equal(1, len(writeRequestPartial.Records))
	assert.Equal(2, len(writeRequestComplete.Records))
	assert.Equal(len(writeRequestComplete.Records), len(readRequest.Records))

	// Attempt to read from the set of keys before they exist. As this is a "ConditionUnconditional" read
	// request, it should succeed and produce an empty response.
	//
	readResponse, err := ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
	assert.Less(revStoreInitial, readResponse.Revision, "Unexpected value for store revision on read completion")

	// Now write a partial set of keys
	//
	writeResponse, err := ts.store.WriteTxn(ctx, writeRequestPartial)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(readResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

	// Try to read the complete set of key/value pairs, only some of which should be there.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(len(writeRequestPartial.Records), len(readResponse.Records), "Read returned unexpected number of records")
	assert.Equal(writeResponse.Revision, readResponse.Revision)

	ts.testCompareReadResponseToWrite(readResponse, writeRequestPartial, writeResponse)

	// Now write a complete set of keys
	//
	writeResponse, err = ts.store.WriteTxn(ctx, writeRequestComplete)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(readResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

	// Finally try to read the complete set of key/value pairs all of which should now be there.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
	assert.Equal(writeResponse.Revision, readResponse.Revision)

	ts.testCompareReadResponseToWrite(readResponse, writeRequestComplete, writeResponse)
}

func (ts *storeTestSuite) TestStoreWriteDeleteTxn() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreWriteDeleteTxn"
	key := testGenerateKeyFromName(testName)

	writeRequest := testGenerateRequestForSimpleWrite(key)
	readRequest := testGenerateRequestForSimpleRead(key)
	deleteRequest := testGenerateRequestForSimpleDelete(key)

	assert.Equal(1, len(writeRequest.Records))
	assert.Equal(1, len(readRequest.Records))
	assert.Equal(1, len(deleteRequest.Records))

	// Attempt to read the key before it exists. As this is a "ConditionRequired" read
	// request, it should fail and produce no response.
	//
	readResponse, err := ts.store.ReadTxn(ctx, readRequest)
	require.ErrorIs(errors.ErrStoreKeyNotFound(key), err)
	assert.Nil(readResponse, "Unexpected response for read of invalid key - error: %v key: %v", err, key)

	// Now write the key
	//
	writeResponse, err := ts.store.WriteTxn(ctx, writeRequest)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(revStoreInitial, writeResponse.Revision, "Unexpected value for store revision on write completion")

	// Verify the key is now there.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
	assert.Equal(writeResponse.Revision, readResponse.Revision)

	ts.testCompareReadResponseToWrite(readResponse, writeRequest, writeResponse)

	// Delete the key we just wrote
	//
	deleteResponse, err := ts.store.DeleteTxn(ctx, deleteRequest)
	require.NoError(err, "Failed to delete key from store - error: %v key: %v", err, key)
	require.NotNil(deleteResponse)
	assert.Equal(0, len(deleteResponse.Records))
	assert.Less(writeResponse.Revision, deleteResponse.Revision)

	// Attempt to re-read the key which once again should be absent.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.ErrorIs(errors.ErrStoreKeyNotFound(key), err)
	assert.Nil(readResponse, "Unexpected response for read of invalid key - error: %v key: %v", err, key)
}

func (ts *storeTestSuite) TestStoreWriteDeleteTxnDeleteAbsent() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreWriteDeleteTxnDeleteAbsent"
	key := testGenerateKeyFromName(testName)

	writeRequest := testGenerateRequestForSimpleWrite(key)
	readRequest := testGenerateRequestForSimpleRead(key)
	deleteRequest := testGenerateRequestForSimpleDelete(key)

	assert.Equal(1, len(writeRequest.Records))
	assert.Equal(1, len(readRequest.Records))
	assert.Equal(1, len(deleteRequest.Records))

	// Attempt to read the key before it exists. As this is a "ConditionRequired" read
	// request, it should fail and produce no response.
	//
	readResponse, err := ts.store.ReadTxn(ctx, readRequest)
	require.ErrorIs(errors.ErrStoreKeyNotFound(key), err)
	assert.Nil(readResponse, "Unexpected response for read of invalid key - error: %v key: %v", err, key)

	// Try to delete the key we just verified is absent
	//
	deleteResponse, err := ts.store.DeleteTxn(ctx, deleteRequest)
	require.ErrorIs(errors.ErrStoreKeyNotFound(key), err)
	assert.Nil(deleteResponse)

	// Now write the key
	//
	writeResponse, err := ts.store.WriteTxn(ctx, writeRequest)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(revStoreInitial, writeResponse.Revision, "Unexpected value for store revision on write completion")

	// Verify the key is now there.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
	assert.Equal(writeResponse.Revision, readResponse.Revision)

	ts.testCompareReadResponseToWrite(readResponse, writeRequest, writeResponse)

	// Delete the key we just wrote
	//
	deleteResponse, err = ts.store.DeleteTxn(ctx, deleteRequest)
	require.NoError(err, "Failed to delete key from store - error: %v key: %v", err, key)
	require.NotNil(deleteResponse)
	assert.Equal(0, len(deleteResponse.Records))
	assert.Less(readResponse.Revision, deleteResponse.Revision)

	// Attempt to re-read the key which once again should be absent.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.ErrorIs(errors.ErrStoreKeyNotFound(key), err)
	assert.Nil(readResponse, "Unexpected response for read of invalid key - error: %v key: %v", err, key)
}

func (ts *storeTestSuite) TestStoreWriteDeleteMultipleTxnRequired() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreWriteDeleteMultipleTxnRequired"
	key := testGenerateKeyFromName(testName)

	writeRequest := testGenerateRequestForWrite(keySetSize, key)
	readRequest := testGenerateRequestForReadWithCondition(keySetSize, key, ConditionUnconditional)
	deleteRequest := testGenerateRequestForDeleteWithCondition(keySetSize, key, ConditionRequired)

	assert.Equal(keySetSize, len(writeRequest.Records))
	assert.Equal(keySetSize, len(readRequest.Records))
	assert.Equal(keySetSize, len(deleteRequest.Records))

	// Verify none of the keys exist.
	//
	readResponse, err := ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
	assert.Less(revStoreInitial, readResponse.Revision, "Unexpected value for store revision on read completion")

	// Attempt to delete the non-existent set of keys
	//
	deleteResponse, err := ts.store.DeleteTxn(ctx, deleteRequest)
	require.ErrorContains(err, errors.ErrStoreKeyNotFound(key).Error())

	// We have a slight problem here in that we do not know which key of the set will be reported
	// in the error. That means it could be any of them, so we need to check for them all and only
	// assert if none of them match.
	//
	var foundError bool

	for k := range readRequest.Records {
		if err == errors.ErrStoreKeyNotFound(k) {
			foundError = true
			break
		}
	}

	assert.True(foundError, "Returned error failed to match any of the expected values - err: %v", err)
	assert.Nil(deleteResponse, "Unexpected response for read of absent keys - error: %v", err)

	// Now write the keys
	//
	writeResponse, err := ts.store.WriteTxn(ctx, writeRequest)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(readResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

	// Verify all the expected keys are present
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(len(writeRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
	assert.Equal(writeResponse.Revision, readResponse.Revision)

	ts.testCompareReadResponseToWrite(readResponse, writeRequest, writeResponse)

	// Delete the keys we just wrote
	//
	deleteResponse, err = ts.store.DeleteTxn(ctx, deleteRequest)
	require.NoError(err, "Failed to delete one or more keys from store - error: %v key: %v", err)
	require.NotNil(deleteResponse)
	require.Equal(0, len(deleteResponse.Records))
	assert.Less(readResponse.Revision, deleteResponse.Revision)

	// and finally, verify none of the keys remain after the delete.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
	assert.Equal(deleteResponse.Revision, readResponse.Revision, "Unexpected value for store revision on read completion")
}

func (ts *storeTestSuite) TestStoreWriteDeleteMultipleTxnOptional() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreWriteDeleteMultipleTxnOptional"
	key := testGenerateKeyFromName(testName)

	writeRequest := testGenerateRequestForWrite(keySetSize, key)
	readRequest := testGenerateRequestForReadWithCondition(keySetSize, key, ConditionUnconditional)
	deleteRequest := testGenerateRequestForDeleteWithCondition(keySetSize, key, ConditionUnconditional)

	assert.Equal(keySetSize, len(writeRequest.Records))
	assert.Equal(keySetSize, len(readRequest.Records))
	assert.Equal(keySetSize, len(deleteRequest.Records))

	// Verify none of the keys exist.
	//
	readResponse, err := ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
	assert.Less(revStoreInitial, readResponse.Revision, "Unexpected value for store revision on read completion")

	// Attempt to delete the non-existent set of keys. This is an unconditional request and so
	// should succeed regardless of the presence or absence of the listed keys.
	//
	deleteResponse, err := ts.store.DeleteTxn(ctx, deleteRequest)
	require.NoError(err, "Unexpected failure in unconditionally deleting absent keys - error: %v", err)
	require.NotNil(deleteResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(deleteResponse.Revision, deleteResponse.Revision)

	// Now write the keys
	//
	writeResponse, err := ts.store.WriteTxn(ctx, writeRequest)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(deleteResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

	// Verify all the expected keys are present
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(len(writeRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
	assert.Equal(writeResponse.Revision, readResponse.Revision)

	ts.testCompareReadResponseToWrite(readResponse, writeRequest, writeResponse)

	// Delete the keys we just wrote
	//
	deleteResponse, err = ts.store.DeleteTxn(ctx, deleteRequest)
	require.NoError(err, "Failed to delete one or more keys from store - error: %v key: %v", err)
	require.NotNil(deleteResponse)
	require.Equal(0, len(deleteResponse.Records))
	assert.Less(readResponse.Revision, deleteResponse.Revision)

	// and finally, verify none of the keys remain after the delete.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
	assert.Equal(deleteResponse.Revision, readResponse.Revision, "Unexpected value for store revision on read completion")
}

func (ts *storeTestSuite) TestStoreWriteDeleteMultipleTxnPartialRequired() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreWriteDeleteMultipleTxnPartial"
	key := testGenerateKeyFromName(testName)

	writeRequestPartial := testGenerateRequestForWrite(1, key)
	writeRequestComplete := testGenerateRequestForWrite(2, key)

	readRequest := testGenerateRequestForReadWithCondition(len(writeRequestComplete.Records), key, ConditionUnconditional)
	deleteRequest := testGenerateRequestForDeleteWithCondition(len(writeRequestComplete.Records), key, ConditionRequired)

	assert.Equal(1, len(writeRequestPartial.Records))
	assert.Equal(2, len(writeRequestComplete.Records))
	assert.Equal(len(writeRequestComplete.Records), len(readRequest.Records))
	assert.Equal(len(writeRequestComplete.Records), len(deleteRequest.Records))

	// Verify none of the keys exist.
	//
	readResponse, err := ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
	assert.Less(revStoreInitial, readResponse.Revision, "Unexpected value for store revision on read completion")

	// Now write a partial set of keys
	//
	writeResponse, err := ts.store.WriteTxn(ctx, writeRequestPartial)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(readResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

	// Verify we have what we expect by trying to read the complete set of
	// key/value pairs, only some of which should be there.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(len(writeRequestPartial.Records), len(readResponse.Records), "Read returned unexpected number of records")
	assert.Equal(writeResponse.Revision, readResponse.Revision)

	ts.testCompareReadResponseToWrite(readResponse, writeRequestPartial, writeResponse)

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
	deleteResponse, err := ts.store.DeleteTxn(ctx, deleteRequest)
	require.ErrorIs(errors.ErrStoreKeyNotFound(keyComplete), err)
	assert.Nil(deleteResponse)

	// Verify we still have what we expect by trying to read the complete set of
	// key/value pairs, only some of which should be there.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(len(writeRequestPartial.Records), len(readResponse.Records), "Read returned unexpected number of records")
	assert.Equal(writeResponse.Revision, readResponse.Revision)

	ts.testCompareReadResponseToWrite(readResponse, writeRequestPartial, writeResponse)

	// Now write a complete set of keys
	//
	writeResponse, err = ts.store.WriteTxn(ctx, writeRequestComplete)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(readResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

	// Verify we have the complete set of key/value pairs all of which should now be there.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(len(writeRequestComplete.Records), len(readResponse.Records), "Read returned unexpected number of records")
	assert.Equal(writeResponse.Revision, readResponse.Revision)

	ts.testCompareReadResponseToWrite(readResponse, writeRequestComplete, writeResponse)

	// Once again attempt to delete the complete set of keys and this time, the delete should succeed.
	//
	deleteResponse, err = ts.store.DeleteTxn(ctx, deleteRequest)
	require.NoError(err, "Failed to delete one or more keys from store - error: %v key: %v", err)
	require.NotNil(deleteResponse)
	require.Equal(0, len(deleteResponse.Records))
	assert.Less(readResponse.Revision, deleteResponse.Revision)

	// and finally, verify none of the keys remain after the delete.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
	assert.Equal(deleteResponse.Revision, readResponse.Revision, "Unexpected value for store revision on read completion")
}

func (ts *storeTestSuite) TestStoreWriteDeleteMultipleTxnPartialOptional() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreWriteDeleteMultipleTxnPartialOptional"
	key := testGenerateKeyFromName(testName)

	writeRequestPartial := testGenerateRequestForWrite(1, key)
	writeRequestComplete := testGenerateRequestForWrite(2, key)

	readRequest := testGenerateRequestForReadWithCondition(len(writeRequestComplete.Records), key, ConditionUnconditional)
	deleteRequest := testGenerateRequestForDeleteWithCondition(len(writeRequestComplete.Records), key, ConditionUnconditional)

	assert.Equal(1, len(writeRequestPartial.Records))
	assert.Equal(2, len(writeRequestComplete.Records))
	assert.Equal(len(writeRequestComplete.Records), len(readRequest.Records))
	assert.Equal(len(writeRequestComplete.Records), len(deleteRequest.Records))

	// Verify none of the keys exist.
	//
	readResponse, err := ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
	assert.Less(revStoreInitial, readResponse.Revision, "Unexpected value for store revision on read completion")

	// Now write a partial set of keys
	//
	writeResponse, err := ts.store.WriteTxn(ctx, writeRequestPartial)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(readResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

	// Verify we have what we expect by trying to read the complete set of
	// key/value pairs, only some of which should be there.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(len(writeRequestPartial.Records), len(readResponse.Records), "Read returned unexpected number of records")
	assert.Equal(writeResponse.Revision, readResponse.Revision)

	ts.testCompareReadResponseToWrite(readResponse, writeRequestPartial, writeResponse)

	// Attempt to delete the full set of keys. This should succeed and all the
	// keys should have been deleted.
	//
	deleteResponse, err := ts.store.DeleteTxn(ctx, deleteRequest)
	require.NoError(err, "Failed to delete one or more keys from store - error: %v key: %v", err)
	require.NotNil(deleteResponse)
	assert.Equal(0, len(deleteResponse.Records))
	assert.Less(readResponse.Revision, deleteResponse.Revision)

	// and finally, verify none of the keys remain after the delete.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
	assert.Equal(deleteResponse.Revision, readResponse.Revision, "Unexpected value for store revision on read completion")
}

func (ts *storeTestSuite) TestStoreWriteDeleteWithPrefix() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreWriteDeleteWithPrefix"
	key := testGenerateKeyFromName(testName)

	writeRequest := testGenerateRequestForWrite(keySetSize, key)
	readRequest := testGenerateRequestForReadWithCondition(keySetSize, key, ConditionUnconditional)
	deleteRequest := testGenerateRequestForDelete(keySetSize, key)

	assert.Equal(keySetSize, len(writeRequest.Records))
	assert.Equal(keySetSize, len(readRequest.Records))
	assert.Equal(keySetSize, len(deleteRequest.Records))


	// Write the keys to the store
	//
	writeResponse, err := ts.store.WriteTxn(ctx, writeRequest)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(revStoreInitial, writeResponse.Revision, "Unexpected value for store revision on write completion")

	// Verify all the expected keys are present
	//
	readResponse, err := ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(len(writeRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
	assert.Equal(writeResponse.Revision, readResponse.Revision)

	ts.testCompareReadResponseToWrite(readResponse, writeRequest, writeResponse)

	// Now delete the keys by prefix. Note that because of the way the request are
	// built, the supplied "key" argument is an effective prefix for the set of keys.
	//
	deleteResponse, err := ts.store.DeleteWithPrefix(ctx, key)
	require.NoError(err, "Failed to delete the keys from the store - error: %v prefix: %v", err, key)
	require.NotNil(deleteResponse)
	assert.Equal(0, len(deleteResponse.Records))
	assert.Less(writeResponse.Revision, deleteResponse.Revision)

	// and finally, verify none of the keys remain after the delete.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
	assert.Equal(deleteResponse.Revision, readResponse.Revision, "Unexpected value for store revision on read completion")
}

func (ts *storeTestSuite) TestStoreWriteReadDeleteWithoutConnect() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreWriteReadDeleteWithoutConnect"

	writeRequest := testGenerateRequestForWrite(keySetSize, testName)
	readRequest := testGenerateRequestForRead(keySetSize, testName)
	deleteRequest := testGenerateRequestForRead(keySetSize, testName)
	deletePrefix := testGenerateKeyFromNames(testName, "")

	store := NewStore()
	require.NotNil(store, "Failed to get the store as expected")

	response, err := store.WriteTxn(ctx, writeRequest)
	require.ErrorIs(errors.ErrStoreNotConnected("already disconnected"), err, "Unexpected error response - expected: %v got: %v", errors.ErrStoreNotConnected("already disconnected"), err)
	assert.Nil(response)

	response, err = store.ReadTxn(ctx, readRequest)
	require.ErrorIs(errors.ErrStoreNotConnected("already disconnected"), err, "Unexpected error response - expected: %v got: %v", errors.ErrStoreNotConnected("already disconnected"), err)
	assert.Nil(response)

	response, err = store.DeleteTxn(ctx, deleteRequest)
	require.ErrorIs(errors.ErrStoreNotConnected("already disconnected"), err, "Unexpected error response - expected: %v got: %v", errors.ErrStoreNotConnected("already disconnected"), err)
	assert.Nil(response)

	response, err = store.DeleteWithPrefix(ctx, deletePrefix)
	require.ErrorIs(errors.ErrStoreNotConnected("already disconnected"), err, "Unexpected error response - expected: %v got: %v", errors.ErrStoreNotConnected("already disconnected"), err)
	assert.Nil(response)

	store = nil
}

func (ts *storeTestSuite) TestStoreSetWatch() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreSetWatch"
	key := testGenerateKeyFromName(testName)

	writeRequest := testGenerateRequestForWrite(0, key)
	updateRequest := testGenerateRequestForWrite(0, key)
	deleteRequest := testGenerateRequestForDelete(0, key)

	assert.Equal(1, len(writeRequest.Records))
	assert.Equal(1, len(updateRequest.Records))
	assert.Equal(1, len(deleteRequest.Records))

	w, err := ts.store.SetWatch(ctx, key)
	assert.NoError(err, "Failed setting a watch point - error: %v", err)
	require.NotNil(w)


	writeResponse, err := ts.store.WriteTxn(ctx, writeRequest)
	assert.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(revStoreInitial, writeResponse.Revision)

	writeEvent := <-w.Events

	require.NotNil(writeEvent)
	assert.Equal(WatchEventTypeCreate, writeEvent.Type)
	assert.Equal(key, writeEvent.Key)
	assert.Equal(writeResponse.Revision, writeEvent.Revision)

	assert.Equal(RevisionInvalid, writeEvent.OldRev)
	assert.Equal("", writeEvent.OldVal)

	assert.Equal(writeResponse.Revision, writeEvent.NewRev)
	assert.Equal(writeRequest.Records[key].Value, writeEvent.NewVal)


	updateResponse, err := ts.store.WriteTxn(ctx, updateRequest)
	assert.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(updateResponse)
	assert.Equal(0, len(updateResponse.Records))
	assert.Less(writeResponse.Revision, updateResponse.Revision)

	updateEvent := <-w.Events

	require.NotNil(updateEvent)
	assert.Equal(WatchEventTypeUpdate, updateEvent.Type)
	assert.Equal(key, updateEvent.Key)
	assert.Equal(updateResponse.Revision, updateEvent.Revision)

	assert.Equal(writeResponse.Revision, updateEvent.OldRev)
	assert.Equal(writeRequest.Records[key].Value, updateEvent.OldVal)

	assert.Equal(updateResponse.Revision, updateEvent.NewRev)
	assert.Equal(updateRequest.Records[key].Value, updateEvent.NewVal)


	deleteResponse, err := ts.store.DeleteTxn(ctx, deleteRequest)
	assert.NoError(err, "Failed to delete from store - error: %v", err)
	require.NotNil(deleteResponse)
	assert.Equal(0, len(deleteResponse.Records))
	assert.Less(updateResponse.Revision, deleteResponse.Revision)

	deleteEvent := <-w.Events

	require.NotNil(deleteEvent)
	assert.Equal(WatchEventTypeDelete, deleteEvent.Type)
	assert.Equal(key, deleteEvent.Key)
	assert.Equal(deleteResponse.Revision, deleteEvent.Revision)

	assert.Equal(updateResponse.Revision, deleteEvent.OldRev)
	assert.Equal(updateRequest.Records[key].Value, deleteEvent.OldVal)

	assert.Equal(RevisionInvalid, deleteEvent.NewRev)
	assert.Equal("" , deleteEvent.NewVal)

	w.Close(ctx)
}

func (ts *storeTestSuite) TestStoreSetWatchPrefix() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreSetWatchPrefix"
	key := testGenerateKeyFromName(testName)

	writeRequest := testGenerateRequestForWrite(2, key)
	updateRequest := testGenerateRequestForWrite(2, key)
	deleteRequest := testGenerateRequestForDelete(2, key)

	assert.Equal(2, len(writeRequest.Records))
	assert.Equal(2, len(updateRequest.Records))
	assert.Equal(2, len(deleteRequest.Records))

	w, err := ts.store.SetWatchWithPrefix(ctx, key)
	assert.NoError(err, "Failed setting a watch point - error: %v", err)
	require.NotNil(w)


	writeResponse, err := ts.store.WriteTxn(ctx, writeRequest)
	assert.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(revStoreInitial, writeResponse.Revision)

	// At this point we expect two events, one for each of the k/v pairs in the
	// write request. The order of the events is arbitrary.
	//
	for i := 0; i < len(writeRequest.Records); i++ {
		writeEvent := <-w.Events

		require.NotNil(writeEvent)
		assert.Equal(WatchEventTypeCreate, writeEvent.Type)
		assert.Equal(writeResponse.Revision, writeEvent.Revision)

		record, ok := writeRequest.Records[writeEvent.Key]
		require.True(ok)

		assert.Equal(RevisionInvalid, writeEvent.OldRev)
		assert.Equal("", writeEvent.OldVal)

		assert.Equal(writeResponse.Revision, writeEvent.NewRev)
		assert.Equal(record.Value, writeEvent.NewVal)
		}


	updateResponse, err := ts.store.WriteTxn(ctx, updateRequest)
	assert.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(updateResponse)
	assert.Equal(0, len(updateResponse.Records))
	assert.Less(writeResponse.Revision, updateResponse.Revision)

	// Now we expect two update events, one for each of the k/v pairs in the
	// update request. The order of the events is arbitrary.
	//
	for i := 0; i < len(updateRequest.Records); i++ {
		updateEvent := <-w.Events

		require.NotNil(updateEvent)
		assert.Equal(WatchEventTypeUpdate, updateEvent.Type)
		assert.Equal(updateResponse.Revision, updateEvent.Revision)

		record, ok := updateRequest.Records[updateEvent.Key]
		require.True(ok)

		assert.Equal(writeResponse.Revision, updateEvent.OldRev)
		assert.Equal(writeRequest.Records[updateEvent.Key].Value, updateEvent.OldVal)

		assert.Equal(updateResponse.Revision, updateEvent.NewRev)
		assert.Equal(record.Value, updateEvent.NewVal)
	}


	deleteResponse, err := ts.store.DeleteTxn(ctx, deleteRequest)
	assert.NoError(err, "Failed to delete from store - error: %v", err)
	require.NotNil(deleteResponse)
	assert.Equal(0, len(deleteResponse.Records))
	assert.Less(updateResponse.Revision, deleteResponse.Revision)

	// Finally we expect two delete events, one for each of the k/v pairs in the
	// delete request. The order of the events is arbitrary.
	//
	for i := 0; i < len(updateRequest.Records); i++ {
		deleteEvent := <-w.Events

		require.NotNil(deleteEvent)
		assert.Equal(WatchEventTypeDelete, deleteEvent.Type)
		assert.Equal(deleteResponse.Revision, deleteEvent.Revision)

		_, ok := deleteRequest.Records[deleteEvent.Key]
		require.True(ok)

		assert.Equal(updateResponse.Revision, deleteEvent.OldRev)
		assert.Equal(updateRequest.Records[deleteEvent.Key].Value, deleteEvent.OldVal)

		assert.Equal(RevisionInvalid, deleteEvent.NewRev)
		assert.Equal("" , deleteEvent.NewVal)
	}


	w.Close(ctx)
}

func (ts *storeTestSuite) TestStoreGetMemberList() {
	assert  := ts.Assert()
	require := ts.Require()

	response, err := ts.store.GetClusterMembers()
	require.NoError(err, "Failed to fetch member list from store - error: %v", err)
	assert.NotNil(response, "Failed to get a response as expected - error: %v", err)
	assert.GreaterOrEqual(1, len(response.Members), "Failed to get the minimum number of response values")

	for i, node := range response.Members {
		ts.T().Logf("node [%v] Id: %v Name: %v", i, node.ID, node.Name)
		for i, url := range node.ClientURLs {
			ts.T().Logf("  client [%v] URL: %v", i, url)
		}
		for i, url := range node.PeerURLs {
			ts.T().Logf("  peer [%v] URL: %v", i, url)
		}
	}
}

func (ts *storeTestSuite) TestStoreSyncClusterConnections() {
	require := ts.Require()

	err := ts.store.UpdateClusterConnections()
	require.NoError(err, "Failed to update cluster connections - error: %v", err)
}

func (ts *storeTestSuite) TestStoreWriteMultipleTxnCreate() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreWriteMultipleTxnCreate"

	writeRequest := testGenerateRequestForWriteCreate(keySetSize, testName)
	readRequest := testGenerateRequestFromWriteRequest(writeRequest)

	// Verify that none of the keys we care about exist in the store
	//
	readResponse, err := ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	assert.Less(RevisionInvalid, readResponse.Revision, "Unexpected value for store revision given expected failure")
	assert.Equal(0, len(readResponse.Records), "Unexpected numbers of records returned")

	createResponse, err := ts.store.WriteTxn(ctx, writeRequest)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(createResponse)
	assert.Equal(0, len(createResponse.Records))
	assert.Less(revStoreInitial, createResponse.Revision, "Unexpected value for store revision on write(create) completion")

	// The write claimed to succeed, now go fetch the record(s) and verify
	// the revision(s) and value(s) are as expected
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse)
	assert.Equal(createResponse.Revision, readResponse.Revision, "Unexpected value for store revision given no updates")
	assert.Equal(len(writeRequest.Records), len(readResponse.Records), "Unexpected numbers of records returned")

	ts.testCompareReadResponseToWrite(readResponse, writeRequest, createResponse)

	// Try to re-create the same keys. These should fail and the original values and revisions should survive.
	//
	recreateResponse, err := ts.store.WriteTxn(ctx, writeRequest)
	require.ErrorContains(err, errors.ErrStoreAlreadyExists(testName).Error())
	require.Nil(recreateResponse)
	// TODO		assert.Equal(RevisionInvalid, recreateResponse.Records, "Unexpected value for store revision on write(re-create) completion")

	readRecreateResponse, err := ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readRecreateResponse)
	assert.Equal(createResponse.Revision, readRecreateResponse.Revision, "Unexpected value for store revision given no updates")
	assert.Equal(len(writeRequest.Records), len(readRecreateResponse.Records), "Unexpected numbers of records returned")

	ts.testCompareReadResponseToWrite(readResponse, writeRequest, createResponse)
}

func (ts *storeTestSuite) TestStoreWriteMultipleTxnOverwrite() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreWriteMultipleTxnOverwrite"

	writeRequest := testGenerateRequestForWrite(keySetSize, testName)
	readRequest := testGenerateRequestFromWriteRequest(writeRequest)

	// Verify that none of the keys we care about exist in the store
	//
	readResponse, err := ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	assert.Less(RevisionInvalid, readResponse.Revision, "Unexpected value for store revision given expected failure")
	assert.Equal(0, len(readResponse.Records), "Unexpected numbers of records returned")

	createResponse, err := ts.store.WriteTxn(ctx, writeRequest)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(createResponse)
	assert.Equal(0, len(createResponse.Records))
	assert.Less(readResponse.Revision, createResponse.Revision, "Unexpected value for store revision on write completion")

	// The write claimed to succeed, now go fetch the record(s) and verify
	// the revision(s) and value(s) are as expected
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	assert.Equal(createResponse.Revision, readResponse.Revision, "Unexpected value for store revision given no updates")
	assert.Equal(len(writeRequest.Records), len(readResponse.Records), "Unexpected numbers of records returned")

	ts.testCompareReadResponseToWrite(readResponse, writeRequest, createResponse)

	// We verified the write worked, so try an unconditional overwrite. Set the
	// required condition and change the value so we can verify after the update.
	//
	updateRequest := testGenerateRequestFromReadResponse(readResponse)

	updateResponse, err := ts.store.WriteTxn(ctx, updateRequest)
	require.NoError(err, "Failed to write unconditional update to store - error: %v", err)
	require.NotNil(updateResponse)
	assert.Equal(0, len(updateResponse.Records))
	assert.Less(readResponse.Revision, updateResponse.Revision, "Expected new store revision to be greater than the earlier store revision")

	readResponseUpdate, err := ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponseUpdate)
	assert.Equal(updateResponse.Revision, readResponseUpdate.Revision, "Unexpected value for store revision given no updates")
	assert.Equal(len(updateRequest.Records), len(readResponseUpdate.Records), "Unexpected numbers of records returned")

	ts.testCompareReadResponseToWrite(readResponseUpdate, updateRequest, updateResponse)
}

func (ts *storeTestSuite) TestStoreWriteMultipleTxnCompareEqual() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreWriteMultipleTxnCompareEqual"
	key := testGenerateKeyFromName(testName)
	keySetSize := 1

	writeRequest := testGenerateRequestForWrite(keySetSize, key)
	readRequest := testGenerateRequestFromWriteRequest(writeRequest)

	// Verify that none of the keys we care about exist in the store
	//
	readResponse, err := ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	assert.Less(RevisionInvalid, readResponse.Revision, "Unexpected value for store revision given expected failure")
	assert.Equal(0, len(readResponse.Records), "Unexpected numbers of records returned")

	createResponse, err := ts.store.WriteTxn(ctx, writeRequest)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(createResponse)
	require.Equal(0, len(createResponse.Records))
	assert.Less(readResponse.Revision, createResponse.Revision, "Unexpected value for store revision on write(create) completion")

	// The write claimed to succeed, now go fetch the record(s) and verify
	// the revision(s) and value(s) are as expected
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse)
	assert.Equal(createResponse.Revision, readResponse.Revision, "Unexpected value for store revision given no updates")
	assert.Equal(len(writeRequest.Records), len(readResponse.Records), "Unexpected numbers of records returned")

	ts.testCompareReadResponseToWrite(readResponse, writeRequest, createResponse)

	// We verified the write worked, so try a conditional update when the revisions
	// are equal. Set the required condition and change the value so we can verify
	// after the update.
	//
	updateRequest := testGenerateRequestFromReadResponseWithCondition(readResponse, ConditionRevisionEqual)

	updateResponse, err := ts.store.WriteTxn(ctx, updateRequest)
	require.NoError(err, "Failed to write conditional update to store - error: %v", err)
	require.NotNil(updateResponse)
	assert.Equal(0, len(updateResponse.Records))
	assert.Less(readResponse.Revision, updateResponse.Revision, "Expected new store revision to be greater than the earlier store revision")

	// verify the update happened as expected
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse)
	assert.Equal(updateResponse.Revision, readResponse.Revision, "Unexpected value for store revision given no updates")
	assert.Equal(len(updateRequest.Records), len(readResponse.Records), "Unexpected numbers of records returned")

	ts.testCompareReadResponseToWrite(readResponse, updateRequest, updateResponse)
}

func (ts *storeTestSuite) TestStoreListWithPrefix() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreListWithPrefix"
	key := testGenerateKeyFromName(testName)

	writeRequest := testGenerateRequestForWrite(keySetSize, key)

	assert.Equal(keySetSize, len(writeRequest.Records))

	writeResponse, err := ts.store.WriteTxn(ctx, writeRequest)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(revStoreInitial, writeResponse.Revision, "Unexpected value for store revision on write completion")

	// Look for a set of prefixed key/value pairs which we do expect to be present.
	//
	listResponse, err := ts.store.ListWithPrefix(ctx, key)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(listResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(len(writeRequest.Records), len(listResponse.Records), "Failed to get the expected number of response values")

	ts.testCompareReadResponseToWrite(listResponse, writeRequest, writeResponse)

	// Check we got records for each key we asked for
	//
	for k, r := range writeRequest.Records {
		rec, present := listResponse.Records[k]

		assert.True(present, "Missing record for key - %v", k)

		if present {
			assert.Equal(r.Value, rec.Value, "Unexpected value - expected: %q received: %q", r.Value, rec.Value)
		}
	}

	// Check we ONLY got records for the keys we asked for
	//
	for k, r := range listResponse.Records {
		_, present := writeRequest.Records[k]
		assert.True(present, "Extra key: %v record: %v", k, r)
		if present {
			val := writeRequest.Records[k].Value
			assert.Equal(val, r.Value, "key: %v Expected: %q Actual %q", k, val, r.Value)
		}
	}
}

func (ts *storeTestSuite) TestStoreListWithPrefixEmptySet() {
	assert  := ts.Assert()
	require := ts.Require()

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	testName := "TestStoreListWithPrefixEmptySet"
	key := testGenerateKeyFromName(testName)

	writeRequest := testGenerateRequestForWrite(keySetSize, key)
	readRequest := testGenerateRequestForReadWithCondition(keySetSize, key, ConditionUnconditional)

	assert.Equal(keySetSize, len(writeRequest.Records))
	assert.Equal(keySetSize, len(readRequest.Records))

	// Attempt to read from the set of keys before they exist. As this is a "ConditionUnconditional" read
	// request, it should succeed and produce an empty response.
	//
	readResponse, err := ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(0, len(readResponse.Records), "Found %d records when none were expected", len(readResponse.Records))
	assert.Less(revStoreInitial, readResponse.Revision, "Unexpected value for store revision on read completion")

	// Look for a prefix name after verifying the keys are absent.
	//
	// We expect success with a non-nil but empty set.
	//
	listResponse, err := ts.store.ListWithPrefix(ctx, key)
	require.NoError(err, "Unexpected failure attempting to list non-existing key set - error: %v prefixKey: %v", err, key)
	require.NotNil(listResponse, "Failed to get a non-nil response as expected - error: %v prefixKey: %v", err, key)
	assert.Equal(readResponse.Revision, listResponse.Revision)
	assert.Equal(0, len(listResponse.Records), "Got more results than expected")

	if len(listResponse.Records) > 0 {
		for k, r := range listResponse.Records {
			assert.Equal(0, len(listResponse.Records), "Unexpected key/value pair key: %v value: %v", k, r.Value)
		}
	}

	// Now write the keys
	//
	writeResponse, err := ts.store.WriteTxn(ctx, writeRequest)
	require.NoError(err, "Failed to write to store - error: %v", err)
	require.NotNil(writeResponse)
	assert.Equal(0, len(writeResponse.Records))
	assert.Less(readResponse.Revision, writeResponse.Revision, "Unexpected value for store revision on write completion")

	// verify the existence of the keys we just wrote.
	//
	readResponse, err = ts.store.ReadTxn(ctx, readRequest)
	require.NoError(err, "Failed to read from store - error: %v", err)
	require.NotNil(readResponse, "Failed to get a response as expected - error: %v", err)
	assert.Equal(len(readRequest.Records), len(readResponse.Records), "Read returned unexpected number of records")
	assert.Equal(writeResponse.Revision, readResponse.Revision)

	ts.testCompareReadResponseToWrite(readResponse, writeRequest, writeResponse)

	// Now look for a set of prefixed key/value pairs which we now expect and have verified to be present.
	//
	listResponse, err = ts.store.ListWithPrefix(ctx, key)
	require.NoError(err, "Unexpected failure attempting to list existing key set - error: %v prefixKey: %v", err, key)
	require.NotNil(listResponse, "Failed to get a response as expected - error: %v prefixKey: %v", err, key)
	assert.Equal(len(writeRequest.Records), len(listResponse.Records), "Failed to get the expected number of response values")
	assert.Equal(readResponse.Revision, readResponse.Revision)

	ts.testCompareReadResponseToWrite(listResponse, writeRequest, writeResponse)
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(storeTestSuite))
}
