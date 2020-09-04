package store

import (
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/internal/tracing/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A number of tests use a pre-computed set of keys for the purposes of
// the test. This constant determines the standard size of these sets.
// The value chosen is largely arbitrary. Selecting a value that is too
// large may lead to problems with values to /from the underlying store,
// so be reasonable.
//
const keySetSize = 100

var (
	initialized bool

	configPath *string
)

func commonSetup() {
	setup.Init(exporters.UnitTest)

	configPath = flag.String("config", ".", "path to the configuration file")
	flag.Parse()

	cfg, err := config.ReadGlobalConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to process the global configuration: %v", err)
	}

	Initialize(cfg)
}

func testGenerateRequestForRead(setSize int, setName string) *Request {
	req := &Request{
		Records:    make(map[string]Record, setSize),
		Conditions: make(map[string]Condition, setSize),
	}

	for i := 0; i < setSize; i++ {
		key := fmt.Sprintf("%s/Key%04d", setName, i)
		req.Records[key] = Record{Revision: RevisionInvalid}
		req.Conditions[key] = ConditionUnconditional
	}

	return req
}

func testGenerateRequestForWrite(setSize int, setName string) *Request {
	req := &Request{
		Records:    make(map[string]Record, setSize),
		Conditions: make(map[string]Condition, setSize),
	}

	for i := 0; i < setSize; i++ {
		key := fmt.Sprintf("%s/Key%04d", setName, i)
		val := fmt.Sprintf("%s/Val%04d", setName, i)
		req.Records[key] = Record{Revision: RevisionInvalid, Value: val}
		req.Conditions[key] = ConditionUnconditional
	}

	return req
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

		wRes, ok := writeResponse.Records[k]

		require.Truef(t, ok, "No write response record to match read response record")

		ok = testCompareReadRecordToWriteRecord(&r, &wReq, wRes.Revision)

		assert.Truef(
			t,
			ok,
			"read response does not match write request - key: %s wVal: %s wRev %v rVal %s rRev: %v",
			k,
			wReq.Value,
			wRes.Revision,
			r.Value,
			r.Revision)
	}
}

func testGenerateKeyValueSetOld(setSize int, setName string) []KeyValueArg {

	keyValueSet := make([]KeyValueArg, setSize)

	for i := range keyValueSet {
		keyValueSet[i].key = fmt.Sprintf("%s/Key%04d", setName, i)
		keyValueSet[i].value = fmt.Sprintf("%s/Val%04d", setName, i)
	}

	return keyValueSet
}

func testGenerateKeyValueMapFromKeyValueSetOld(keyValueSet []KeyValueArg) map[string]string {

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

// Build a set of key,value pairs to be created unconditionally in the store.
//
func testGenerateRecordUpdateSetFromKeyValueSetOld(keyValueSet []KeyValueArg, label string, condition Condition) RecordUpdateSet {

	recordUpdateSet := RecordUpdateSet{Label: label, Records: make(map[string]RecordUpdate)}

	for _, kv := range keyValueSet {
		recordUpdateSet.Records[kv.key] =
			RecordUpdate{
				Condition: condition,
				Record: Record{
					Revision: RevisionInvalid,
					Value:    kv.value,
				},
			}
	}

	return recordUpdateSet
}

// TestMain is the Common test startup method.  This is the _only_ Test* function in this
// file.
func TestMain(m *testing.M) {
	commonSetup()
	os.Exit(m.Run())
}
