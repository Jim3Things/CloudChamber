package store

import (
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A number of tests use a pre-computed set of keys for the purposes of
// the test. This constant determines the standard size of these sets.
// The value chosen is largely arbitrary. Selecting a value that is too
// large may lead to problems with values to /from the underlying store,
// so be reasonable.
//
// Limited by etcd configuration option --max-txn-ops which (currently)
// defaults to 128
//
const keySetSize = 10

var (
	initialized bool

	configPath *string

	utf *exporters.Exporter
)

func commonSetup() {
	utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(utf)

	configPath = flag.String("config", "./testdata", "path to the configuration file")
	flag.Parse()

	cfg, err := config.ReadGlobalConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to process the global configuration: %v", err)
	}

	Initialize(cfg)
}

func testGenerateKeyFromNames(prefix string, name string) string {
	return fmt.Sprintf("%s/Key%s", prefix, name)
}

func testGenerateKeyFromName(prefix string) string {
	return fmt.Sprintf("%s/Key", prefix)
}

func testGenerateKeyFromNameAndIndex(name string, index int) string {
	return fmt.Sprintf("%s/Index%04d", name, index)
}

func testGenerateValFromName(name string) string {
	return fmt.Sprintf("%s/Value", name)
}

func testGenerateValFromNameAndIndex(name string, index int) string {
	return fmt.Sprintf("%s/Value%04d", name, index)
}

func testgenerateGenericInitializedRecord(withValue bool, val string) (rec Record) {
	if withValue {
		rec = Record{Revision: RevisionInvalid, Value: val}
	} else {
		rec = Record{Revision: RevisionInvalid}
	}

	return rec
}

func testGenerateGenericRequest(size int, name string, withValue bool, condition Condition) (req *Request) {
	if size == 0 {
		req = &Request{
			Records:    make(map[string]Record, 1),
			Conditions: make(map[string]Condition, 1),
		}

		key := name
		val := testGenerateValFromName(name)

		req.Conditions[key] = condition
		req.Records[key] = testgenerateGenericInitializedRecord(withValue, val)
	} else {
		req = &Request{
			Records:    make(map[string]Record, size),
			Conditions: make(map[string]Condition, size),
		}

		for i := 0; i < size; i++ {
			key := testGenerateKeyFromNameAndIndex(name, i)
			val := testGenerateValFromNameAndIndex(name, i)

			req.Conditions[key] = condition
			req.Records[key] = testgenerateGenericInitializedRecord(withValue, val)
		}
	}

	return req
}

// +++ Helper functions to generate assorted write requests

func testGenerateRequestForWriteInternal(size int, name string, condition Condition) (req *Request) {
	return testGenerateGenericRequest(size, name, true, condition)
}

func testGenerateRequestForWrite(size int, name string) *Request {
	return testGenerateRequestForWriteInternal(size, name, ConditionUnconditional)
}

func testGenerateRequestForWriteCreate(size int, name string) *Request {
	return testGenerateRequestForWriteInternal(size, name, ConditionCreate)
}

func testGenerateRequestForWriteUpdate(size int, name string) *Request {
	return testGenerateRequestForWriteInternal(size, name, ConditionRevisionEqual)
}

func testGenerateRequestForSimpleWrite(name string) *Request {
	return testGenerateRequestForWriteInternal(0, name, ConditionUnconditional)
}

// --- Helper functions to generate assorted write requests

// +++ Helper functions to generate assorted read requests

func testGenerateRequestForReadInternal(size int, name string, condition Condition) (req *Request) {
	return testGenerateGenericRequest(size, name, false, condition)
}

func testGenerateRequestForReadWithCondition(size int, name string, condition Condition) *Request {
	return testGenerateRequestForReadInternal(size, name, condition)
}

func testGenerateRequestForRead(size int, name string) *Request {
	return testGenerateRequestForReadInternal(size, name, ConditionRequired)
}

func testGenerateRequestForSimpleReadWithCondition(name string, condition Condition) *Request {
	return testGenerateRequestForReadInternal(0, name, condition)
}

func testGenerateRequestForSimpleRead(name string) *Request {
	return testGenerateRequestForReadInternal(0, name, ConditionRequired)
}

// --- Helper functions to generate assorted read requests

// +++ Helper functions to generate assorted delete requests

func testGenerateRequestForDeleteInternal(size int, name string, condition Condition) *Request {
	return testGenerateGenericRequest(size, name, false, condition)
}

func testGenerateRequestForDeleteWithCondition(size int, name string, condition Condition) *Request {
	return testGenerateRequestForDeleteInternal(size, name, condition)
}

func testGenerateRequestForDelete(size int, name string) *Request {
	return testGenerateRequestForDeleteInternal(size, name, ConditionUnconditional)
}

func testGenerateRequestForSimpleDeleteWithCondition(name string, condition Condition) *Request {
	return testGenerateRequestForDeleteInternal(0, name, condition)
}

func testGenerateRequestForSimpleDelete(name string) *Request {
	return testGenerateRequestForDeleteInternal(0, name, ConditionRequired)
}

// --- Helper functions to generate assorted delete requests

// +++ Helper functions to generate assorted requests for chained operations

func testGenerateRequestFromWriteRequest(request *Request) *Request {
	setSize := len(request.Records)
	req := &Request{
		Records:    make(map[string]Record, setSize),
		Conditions: make(map[string]Condition, setSize),
	}

	for k := range request.Records {
		req.Records[k] = Record{Revision: RevisionInvalid}
		req.Conditions[k] = ConditionUnconditional
	}

	return req
}

func testGenerateRequestFromReadResponseWithCondition(response *Response, condition Condition) *Request {
	size := len(response.Records)
	req := &Request{
		Records:    make(map[string]Record, size),
		Conditions: make(map[string]Condition, size),
	}

	for k, r := range response.Records {
		val := r.Value + "Update"
		rev := r.Revision

		if condition == ConditionUnconditional {
			rev = RevisionInvalid
		}

		req.Records[k] = Record{Revision: rev, Value: val}
		req.Conditions[k] = condition
	}

	return req
}

func testGenerateRequestFromReadResponse(response *Response) *Request {
	return testGenerateRequestFromReadResponseWithCondition(response, ConditionUnconditional)
}

func testCompareReadRecordToWriteRecord(rRec *Record, wRec *Record, wRev int64) bool {

	if rRec.Value != wRec.Value {
		return false
	}

	if rRec.Revision != wRev {
		return false
	}

	return true
}

func testCompareReadResponseToWrite(
	t *testing.T,
	readResponse *Response,
	writeRequest *Request,
	writeResponse *Response,
) {

	// Fist check that we have the same number of records in the
	// response as in the request
	//
	assert.Equalf(
		t,
		len(writeRequest.Records),
		len(readResponse.Records),
		"record count mismatch",
	)

	// Now, check that we have a matching read response value and
	// revision for each record in the write request/response pair
	//
	for k, w := range writeRequest.Records {

		r, ok := readResponse.Records[k]

		require.Truef(t, ok, "No read response record to match request record for key: %s val:, %s", k, w.Value)

		ok = testCompareReadRecordToWriteRecord(&r, &w, writeResponse.Revision)

		assert.Truef(
			t,
			ok,
			"read response does not match write request - key: %s wVal: %s wRev %v rVal %s rRev: %v",
			k,
			w.Value,
			writeResponse.Revision,
			r.Value,
			r.Revision,
		)
	}

	// Finally, do we have a write request/response pair for each
	// record in the read response
	//
	for k, r := range readResponse.Records {

		wReq, ok := writeRequest.Records[k]

		require.Truef(t, ok, "No write request record to match read response record")

		ok = testCompareReadRecordToWriteRecord(&r, &wReq, writeResponse.Revision)

		assert.Truef(
			t,
			ok,
			"write request does not match read response - key: %s wVal: %s wRev %v rVal %s rRev: %v",
			k,
			wReq.Value,
			writeResponse.Revision,
			r.Value,
			r.Revision)
	}
}

// TestMain is the Common test startup method.  This is the _only_ Test* function in this
// file.
func TestMain(m *testing.M) {
	commonSetup()
	os.Exit(m.Run())
}
