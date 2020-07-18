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

// Build a set of key,value pairs to be created unconditionally in the store.
//
func testGenerateRecordUpdateSetFromKeyValueSet(keyValueSet []KeyValueArg, label string, condition Condition) RecordUpdateSet {

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
