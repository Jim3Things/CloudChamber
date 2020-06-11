// This module provides support for reading the global service configuration file.
//
// CloudChamber uses a global configuration file that contains the settings for all of the
// support services that make it up.  This is done to avoid replication and drift between,
// e.g., service endpoint definitions as seen by the service and by its callers.

package config

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/api/global"
)

const (
	DefaultGlobalConfigFile     = "cloudchamber.yaml"
	DefaultConfigType           = "yaml"
	ControllerDefaultPort       = 8081
	InventoryDefaultPort        = 8082
	SimSupportDefaultPort       = 8083
	WebServerDefaultPort        = 8084
	WebServerFEDefaultPort      = 8080
	DefaultHost                 = ""
	DefaultRootFilePath         = "."
	DefaultSystemAccount        = "Admin"
	DefaultSystemPassword       = "SystemPassword"
	StoreDefaultTraceLevel      = 1
	StoreDefaultConnectTimeout  = 5
	StoreDefaultRequestTimeout  = 5
	StoreDefaultEtcdSvcHostname = "localhost"
	StoreDefaultEtcdSvcPort     = 2379

	StoreDefaultTestEmbeddedInstance   = true
	StoreDefaultTestEmbeddedHostname   = "localhost"
	StoreDefaultTestEmbeddedHostAddr   = "127.0.0.1"
	StoreDefaultTestEmbeddedPortClient = 9379
	StoreDefaultTestEmbeddedPortPeer   = 9380
	StoreDefaultTestEmbeddedPath       = "store"
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
	EP Endpoint
}

// InventoryType is a helper type that describes the inventoryd configuration settings
type InventoryType struct {
	// Exposed GRPC endpoint
	EP Endpoint
}

// SimSupportType is a helper type that describes the sim_supportd configuration settings
type SimSupportType struct {
	// Exposed GRPC endpoint
	EP Endpoint
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
	BE Endpoint
}

// StoreTypeTestEmbedded describes the specific configured elements for
// an embedded store InstanceType
//
type StoreTypeTestEmbedded struct {
	EmbeddedInstance   bool
	EmbeddedHostname   string
	EmbeddedHostAddr   string
	EmbeddedPortClient uint16
	EmbeddedPortPeer   uint16
	EmbeddedPath       string
}

// StoreType is a structure used to return the configurable elements
// for the Store section of the global configuration file.
//
type StoreType struct {
	ConnectTimeout int
	RequestTimeout int
	TraceLevel     int
	EtcdService    Endpoint
	Test           StoreTypeTestEmbedded
}

// Create a new global configuration object, preset with defaults
func newGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Controller: ControllerType{
			EP: Endpoint{
				Hostname: DefaultHost,
				Port:     ControllerDefaultPort,
			}},
		Inventory: InventoryType{
			EP: Endpoint{
				Hostname: DefaultHost,
				Port:     InventoryDefaultPort,
			}},
		SimSupport: SimSupportType{
			EP: Endpoint{
				Hostname: DefaultHost,
				Port:     SimSupportDefaultPort,
			}},
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
			}},
		Store: StoreType{
			ConnectTimeout: StoreDefaultConnectTimeout,
			RequestTimeout: StoreDefaultRequestTimeout,
			TraceLevel:     StoreDefaultTraceLevel,
			EtcdService: Endpoint{
				Hostname: StoreDefaultEtcdSvcHostname,
				Port:     StoreDefaultEtcdSvcPort,
			},
			Test: StoreTypeTestEmbedded{
				EmbeddedInstance:   StoreDefaultTestEmbeddedInstance,
				EmbeddedHostname:   StoreDefaultTestEmbeddedHostname,
				EmbeddedHostAddr:   StoreDefaultTestEmbeddedHostAddr,
				EmbeddedPortClient: StoreDefaultTestEmbeddedPortClient,
				EmbeddedPortPeer:   StoreDefaultTestEmbeddedPortPeer,
				EmbeddedPath:       StoreDefaultTestEmbeddedPath,
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
	var storeInstanceConfig string

	switch data.Store.Test.EmbeddedInstance {
	case true:
		storeInstanceConfig = fmt.Sprintf(
			"  Embed: Embedded\n"+
				"    Hostname: %v\n"+
				"    HostAddr: %v\n"+
				"    ClientPort: %v\n"+
				"    PeerPort: %v\n"+
				"    Path: %v\n",
			data.Store.Test.EmbeddedHostname,
			data.Store.Test.EmbeddedHostAddr,
			data.Store.Test.EmbeddedPortClient,
			data.Store.Test.EmbeddedPortPeer,
			data.Store.Test.EmbeddedPath)
	case false:
		storeInstanceConfig = fmt.Sprintf(
			"  Embed: External\n"+
				"    Hostname: %v\n"+
				"    Port: %v\n",
			data.Store.EtcdService.Hostname,
			data.Store.EtcdService.Port)
	}
	return fmt.Sprintf(
		"Controller:\n"+
			"  EP:\n"+
			"    port: %v\n    hostname: %v\n"+
			"Inventory:\n"+
			"  EP:\n"+
			"    port: %v\n    hostname: %v\n"+
			"SimSupport:\n"+
			"  EP:\n"+
			"    port: %v\n    hostname: %v\n"+
			"Webserver:\n"+
			"  FE:\n"+
			"    port: %v\n    hostname: %v\n"+
			"  BE:\n"+
			"    port: %v\n    hostname: %v\n"+
			"  RootFilePath: %s\n"+
			"  SystemAccount: %s\n"+
			"  SystemAccountPassword: %s\n"+
			"Store:"+
			"  ConnectTimeout: %v\n"+
			"  RequestTimeout: %v\n"+
			"  TraceLevel: %v\n"+
			"%s\n",
		data.Controller.EP.Port, data.Controller.EP.Hostname,
		data.Inventory.EP.Port, data.Inventory.EP.Hostname,
		data.SimSupport.EP.Port, data.SimSupport.EP.Hostname,
		data.WebServer.FE.Port, data.WebServer.FE.Hostname,
		data.WebServer.BE.Port, data.WebServer.BE.Hostname,
		data.WebServer.RootFilePath,
		data.WebServer.SystemAccount,
		data.WebServer.SystemAccountPassword,
		data.Store.ConnectTimeout,
		data.Store.RequestTimeout,
		data.Store.TraceLevel,
		storeInstanceConfig)
}
