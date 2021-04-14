package inventory

import (
	"context"
	"flag"
	"sync"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/simulation/internal/config"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing/exporters"
)

var (
	lock         sync.Mutex
	globalCfg    *config.GlobalConfig
	initialized  bool = false
)

type testSuiteCore struct {
	suite.Suite

	utf   *exporters.Exporter
	cfg   *config.GlobalConfig
	store *store.Store
}

func runOnce() (*config.GlobalConfig, error) {
	lock.Lock()
	defer lock.Unlock()

	if !initialized {
		configPath := flag.String("config", "./testdata", "path to the configuration file")

		flag.Parse()

		cfg, err := config.ReadGlobalConfig(*configPath)
		if err != nil {
			return nil, err
		}

		globalCfg = cfg
		initialized = true
	}

	return globalCfg, nil
}

func (ts *testSuiteCore) SetupSuite() {
	require := ts.Require()

	cfg, err := runOnce()
	require.NoError(err, "failed to process the global configuration")

	ts.utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(ts.utf)

	store.Initialize(context.Background(), cfg)

	ts.cfg = cfg
	ts.store = store.NewStore()
}

func (ts *testSuiteCore) SetupTest() {
	require := ts.Require()

	require.NoError(ts.utf.Open(ts.T()))
	require.NoError(ts.store.Connect())
}

func (ts *testSuiteCore) TearDownTest() {
	ts.store.Disconnect()
	ts.utf.Close()
}
