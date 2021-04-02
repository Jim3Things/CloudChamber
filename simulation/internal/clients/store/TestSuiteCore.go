package store

import (
	"context"
	"flag"
	"sync"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/internal/config"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing/exporters"
)

var (
	lock         sync.Mutex
	initialized  bool = false
)

type testSuiteCore struct {
	suite.Suite

	utf   *exporters.Exporter
	cfg   *config.GlobalConfig
}

func runOnce() (cfg *config.GlobalConfig, err error) {
	lock.Lock()
	defer lock.Unlock()

	if !initialized {
		configPath := flag.String("config", "./testdata", "path to the configuration file")

		flag.Parse()

		cfg, err = config.ReadGlobalConfig(*configPath)
		if err != nil {
			return nil, err
		}


		Initialize(context.Background(), cfg)

		initialized = true
	}

	return cfg, nil
}

func (ts *testSuiteCore) SetupSuite() {
	require := ts.Require()

	cfg, err := runOnce()
	require.NoError(err, "failed to process the global configuration")

	ts.utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(ts.utf)

	ts.cfg = cfg
}

func (ts *testSuiteCore) SetupTest() {
	require := ts.Require()

	require.NoError(ts.utf.Open(ts.T()))
}

func (ts *testSuiteCore) TearDownTest() {
	ts.utf.Close()
}
