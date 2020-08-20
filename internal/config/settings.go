// This module provides support for reading the global service configuration file.
//
// CloudChamber uses a global configuration file that contains the settings for all of the
// support services that make it up.  This is done to avoid replication and drift between,
// e.g., service endpoint definitions as seen by the service and by its callers.

package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/api/global"

	pb "github.com/Jim3Things/CloudChamber/pkg/protos/Stepper"
)

// Default values for the configurable parameters
//
const (
	DefaultGlobalConfigFile     = "cloudchamber.yaml"
	DefaultConfigType           = "yaml"
	ControllerDefaultPort       = 8081
	ControllerDefaultTraceFile  = ".\\controller_trace.txt"
	InventoryDefaultPort        = 8082
	InventoryDefaultTraceFile   = ".\\inventory_trace.txt"
	SimSupportDefaultPort       = 8083
	SimSupportDefaultTraceFile  = ".\\sim_support_trace.txt"
	WebServerDefaultPort        = 8084
	WebServerFEDefaultPort      = 8080
	WebServerDefaultTraceFile   = ".\\web_server_trace.txt"
	DefaultHost                 = ""
	DefaultStepperPolicy        = ""
	DefaultRootFilePath         = "."
	DefaultSystemAccount        = "Admin"
	DefaultSystemPassword       = "SystemPassword"
	StoreDefaultTraceLevel      = 1
	StoreDefaultConnectTimeout  = 5
	StoreDefaultRequestTimeout  = 5
	StoreDefaultEtcdSvcHostname = "localhost"
	StoreDefaultEtcdSvcPort     = 2379

	StoreDefaultTestUseTestNamespace  = false
	StoreDefaultTestUseUniqueInstance = false
	StoreDefaultTestPreCleanStore     = false
)

// GlobalConfig defines the global configuration structure produced from reading
// the config files.  It is structured by the internal services, with each
// internal service having room for the specific settings that it needs.
//
type GlobalConfig struct {
	Controller ControllerType
	Inventory  InventoryType
	SimSupport SimSupportType
	WebServer  WebServerType
	Store      StoreType
}

// Endpoint is a helper type that defines a simple endpoint
type Endpoint struct {
	Hostname string
	Port     int
}

// ControllerType is a helper type describes the controllerd configuration settings
type ControllerType struct {
	// Exposed GRPC endpoint
	EP        Endpoint
	TraceFile string
}

// InventoryType is a helper type that describes the inventoryd configuration settings
type InventoryType struct {
	// Exposed GRPC endpoint
	EP        Endpoint
	TraceFile string
}

// SimSupportType is a helper type that describes the sim_supportd configuration settings
type SimSupportType struct {
	// Exposed GRPC endpoint
	EP        Endpoint
	TraceFile string

	// Name of the initial stepper policy to apply
	StepperPolicy string
}

// GetPolicyType is a function that returns the configured default policy as
// the protobuf-defined enum value.
func (sst SimSupportType) GetPolicyType() pb.StepperPolicy {
	policyName := strings.ToLower(sst.StepperPolicy)
	switch policyName {
	case "manual":
		return pb.StepperPolicy_Manual
	case "automatic":
		return pb.StepperPolicy_Measured
	default:
		return pb.StepperPolicy_Invalid
	}
}

// WebServerType is a helper type that describes the web_server configuration settings
type WebServerType struct {
	// Filesystem path to the static files and scripts
	RootFilePath string

	// Predefined starting account
	SystemAccount string

	// .. and that account's password
	SystemAccountPassword string

	// External http endpoint
	FE Endpoint

	// GPRC endpoint, used for internal notifications
	BE        Endpoint
	TraceFile string
}

// StoreTypeTest describes the specific configured elements for
// a store test
//
type StoreTypeTest struct {
	UseTestNamespace  bool
	UseUniqueInstance bool
	PreCleanStore     bool
}

// StoreType is a structure used to return the configurable elements
// for the Store section of the global configuration file.
//
type StoreType struct {
	ConnectTimeout int
	RequestTimeout int
	TraceLevel     int
	EtcdService    Endpoint
	Test           StoreTypeTest
}

// Create a new global configuration object, preset with defaults
func newGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Controller: ControllerType{
			EP: Endpoint{
				Hostname: DefaultHost,
				Port:     ControllerDefaultPort,
			},
			TraceFile: ControllerDefaultTraceFile,
		},
		Inventory: InventoryType{
			EP: Endpoint{
				Hostname: DefaultHost,
				Port:     InventoryDefaultPort,
			},
			TraceFile: InventoryDefaultTraceFile,
		},
		SimSupport: SimSupportType{
			EP: Endpoint{
				Hostname: DefaultHost,
				Port:     SimSupportDefaultPort,
			},
			TraceFile:     SimSupportDefaultTraceFile,
			StepperPolicy: DefaultStepperPolicy,
		},
		WebServer: WebServerType{
			RootFilePath:          DefaultRootFilePath,
			SystemAccount:         DefaultSystemAccount,
			SystemAccountPassword: DefaultSystemPassword,
			FE: Endpoint{
				Hostname: DefaultHost,
				Port:     WebServerFEDefaultPort,
			},
			BE: Endpoint{
				Hostname: DefaultHost,
				Port:     WebServerDefaultPort,
			},
			TraceFile: WebServerDefaultTraceFile,
		},
		Store: StoreType{
			ConnectTimeout: StoreDefaultConnectTimeout,
			RequestTimeout: StoreDefaultRequestTimeout,
			TraceLevel:     StoreDefaultTraceLevel,
			EtcdService: Endpoint{
				Hostname: StoreDefaultEtcdSvcHostname,
				Port:     StoreDefaultEtcdSvcPort,
			},
			Test: StoreTypeTest{
				UseTestNamespace:  StoreDefaultTestUseTestNamespace,
				UseUniqueInstance: StoreDefaultTestUseUniqueInstance,
				PreCleanStore:     StoreDefaultTestPreCleanStore,
			},
		},
	}
}

// ReadGlobalConfig is a routine to read the configuration file at the specified path.
//
// The configuration file is parsed, and the result returned as a typed object
func ReadGlobalConfig(path string) (*GlobalConfig, error) {
	viper.SetConfigName(DefaultGlobalConfigFile)
	viper.AddConfigPath(path)
	viper.SetConfigType(DefaultConfigType)

	cfg := newGlobalConfig()

	tr := global.TraceProvider().Tracer("")

	ctx, span := tr.Start(
		context.Background(),
		"ReadGlobalConfig")
	defer span.End()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; we'll just use the default values
			span.AddEvent(
				ctx,
				fmt.Sprintf(
					"No config file found at %s/%s (%s), applying defaults.",
					path,
					DefaultGlobalConfigFile,
					DefaultConfigType))
		} else {
			// Config file was found but another error was produced
			err = fmt.Errorf("fatal error reading config file: %s", err)
			span.AddEvent(ctx, err.Error())
			return nil, err
		}
	} else {
		// Fill in the global configuration object from the configuration file
		if err = viper.UnmarshalExact(cfg); err != nil {
			err = fmt.Errorf("unable to decode into struct, %v", err)
			span.AddEvent(ctx, err.Error())
			return nil, err
		}
	}

	span.AddEvent(ctx,
		fmt.Sprintf("Configuration Read: \n%v", ToString(cfg)))

	return cfg, nil
}

// ToString is a function to format the configuration data as a returned string.
func ToString(data *GlobalConfig) string {

	return fmt.Sprintf(
		"Controller:\n"+
			"  EP:\n"+
			"    port: %v\n    hostname: %v\n"+
			"  TraceFile: %s\n"+
			"Inventory:\n"+
			"  EP:\n"+
			"    port: %v\n    hostname: %v\n"+
			"  TraceFile: %s\n"+
			"SimSupport:\n"+
			"  EP:\n"+
			"    port: %v\n    hostname: %v\n"+
			"  TraceFile: %s\n"+
			"  StepperPolicy: %v\n"+
			"Webserver:\n"+
			"  FE:\n"+
			"    port: %v\n    hostname: %v\n"+
			"  BE:\n"+
			"    port: %v\n    hostname: %v\n"+
			"  TraceFile: %s\n"+
			"  RootFilePath: %s\n"+
			"  SystemAccount: %s\n"+
			"  SystemAccountPassword: %s\n"+
			"Store:"+
			"  ConnectTimeout: %v\n"+
			"  RequestTimeout: %v\n"+
			"  TraceLevel: %v\n"+
			"  Test:\n"+
			"    UseTestNamespace: %v\n"+
			"    UseUniqueInstance: %v\n"+
			"    PreCleanStore: %v\n"+
			"",
		data.Controller.EP.Port, data.Controller.EP.Hostname,
		data.Controller.TraceFile,
		data.Inventory.EP.Port, data.Inventory.EP.Hostname,
		data.Inventory.TraceFile,
		data.SimSupport.EP.Port, data.SimSupport.EP.Hostname,
		data.SimSupport.TraceFile,
		data.SimSupport.StepperPolicy,
		data.WebServer.FE.Port, data.WebServer.FE.Hostname,
		data.WebServer.BE.Port, data.WebServer.BE.Hostname,
		data.WebServer.TraceFile,
		data.WebServer.RootFilePath,
		data.WebServer.SystemAccount,
		data.WebServer.SystemAccountPassword,
		data.Store.ConnectTimeout,
		data.Store.RequestTimeout,
		data.Store.TraceLevel,
		data.Store.Test.UseTestNamespace,
		data.Store.Test.UseUniqueInstance,
		data.Store.Test.PreCleanStore)
}
