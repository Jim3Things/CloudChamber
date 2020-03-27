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
    DefaultGlobalConfigFile = "cloudchamber.yaml"
    DefaultConfigType = "yaml"
    ControllerDefaultPort = 8081
    InventoryDefaultPort = 8082
    SimSupportDefaultPort = 8083
    WebServerDefaultPort = 8084
    WebServerFEDefaultPort = 8080
    DefaultHost = ""
    DefaultRootFilePath = "."
    DefaultSystemAccount = "Admin"
)

// Define the global configuration structure produced from reading the config
// files.  It is structured by the internal services, with each internal service
// having room for the specific settings that it needs.
//
type GlobalConfig struct {
    Controller ControllerType
    Inventory InventoryType
    SimSupport SimSupportType
    WebServer WebServerType
}

// Helper type that defines a simple endpoint
type Endpoint struct {
    Hostname string
    Port int
}

// controllerd configuration settings
type ControllerType struct {
    // Exposed GRPC endpoint
    EP Endpoint
}

// inventoryd configuration settings
type InventoryType struct {
    // Exposed GRPC endpoint
    EP Endpoint
}

// sim_supportd configuration settings
type SimSupportType struct {
    // Exposed GRPC endpoint
    EP Endpoint
}

// web_server configuration settings
type WebServerType struct {
    // Filesystem path to the static files and scripts
    RootFilePath string

    // Predefined starting account
    SystemAccount string

    // External http endpoint
    FE Endpoint

    // GPRC endpoint, used for internal notifications
    BE Endpoint
}

// Create a new global configuration object, preset with defaults
func newGlobalConfig() *GlobalConfig {
    return &GlobalConfig{
        Controller: ControllerType{
            EP: Endpoint{
                Hostname: DefaultHost,
                Port: ControllerDefaultPort,
            }},
        Inventory: InventoryType{
            EP: Endpoint{
                Hostname: DefaultHost,
                Port: InventoryDefaultPort,
            }},
        SimSupport: SimSupportType{
            EP: Endpoint{
                Hostname: DefaultHost,
                Port: SimSupportDefaultPort,
            }},
        WebServer: WebServerType{
            RootFilePath: DefaultRootFilePath,
            SystemAccount: DefaultSystemAccount,
            FE: Endpoint{
                Hostname: DefaultHost,
                Port:     WebServerFEDefaultPort,
            },
            BE: Endpoint{
                Hostname: DefaultHost,
                Port:     WebServerDefaultPort,
            }},
    }
}

// Read the configuration file at the specified path.
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
                    "No config file found at %s/%s (%s), applying defaults.\n",
                    path,
                    DefaultGlobalConfigFile,
                    DefaultConfigType))
        } else {
            // Config file was found but another error was produced
            err = fmt.Errorf("Fatal error reading config file: %s \n", err)
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

// Format the configuration data.
func ToString(data *GlobalConfig) string {
    return fmt.Sprintf(
        "Controller:\n" +
            "  EP:\n" +
            "    port: %v\n    hostname: %v\n" +
            "Inventory:\n" +
            "  EP:\n" +
            "    port: %v\n    hostname: %v\n" +
            "SimSupport:\n" +
            "  EP:\n" +
            "    port: %v\n    hostname: %v\n" +
            "Webserver:\n" +
            "  FE:\n" +
            "    port: %v\n    hostname: %v\n" +
            "  BE:\n" +
            "    port: %v\n    hostname: %v\n" +
            "  RootFilePath: %s\n" +
            "  SystemAccount: %s\n",
        data.Controller.EP.Port, data.Controller.EP.Hostname,
        data.Inventory.EP.Port, data.Inventory.EP.Hostname,
        data.SimSupport.EP.Port, data.SimSupport.EP.Hostname,
        data.WebServer.FE.Port, data.WebServer.FE.Hostname,
        data.WebServer.BE.Port, data.WebServer.BE.Hostname,
        data.WebServer.RootFilePath,
        data.WebServer.SystemAccount)
}
