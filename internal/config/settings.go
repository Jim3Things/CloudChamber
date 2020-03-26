package config

import (
    "fmt"

    "github.com/spf13/viper"
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

type Endpoint struct {
    Hostname string
    Port int
}

type ControllerType struct {
    Hostname string
    Port int
}

type InventoryType struct {
    Hostname string
    Port int
}

type SimSupportType struct {
    Hostname string
    Port int
}

type WebServerType struct {
    RootFilePath string
    SystemAccount string
    FE Endpoint
    BE Endpoint
}

type GlobalConfig struct {
    Controller ControllerType
    Inventory InventoryType
    SimSupport SimSupportType
    WebServer WebServerType
}

func newGlobalConfig() *GlobalConfig {
    return &GlobalConfig{
        Controller: ControllerType{
            Hostname: DefaultHost,
            Port: ControllerDefaultPort,
        },
        Inventory: InventoryType{
            Hostname: DefaultHost,
            Port: InventoryDefaultPort,
        },
        SimSupport: SimSupportType{
            Hostname: DefaultHost,
            Port:     SimSupportDefaultPort,
        },
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
            },
        },
    }
}

func ToString(data *GlobalConfig) string {
    return fmt.Sprintf(
        "Controller:\n" +
            "-- port: %v\n-- hostname: %v\n" +
            "Inventory:\n" +
            "-- port: %v\n-- hostname: %v\n" +
            "SimSupport:\n" +
            "-- port: %v\n-- hostname: %v\n" +
            "Webserver:\n" +
            "  FE:\n" +
            "  -- port: %v\n  -- hostname: %v\n" +
            "  BE:\n" +
            "  -- port: %v\n  -- hostname: %v\n" +
            "  RootFilePath: %s\n" +
            "  SystemAccount: %s\n",
        data.Controller.Port, data.Controller.Hostname,
        data.Inventory.Port, data.Inventory.Hostname,
        data.SimSupport.Port, data.SimSupport.Hostname,
        data.WebServer.FE.Port, data.WebServer.FE.Hostname,
        data.WebServer.BE.Port, data.WebServer.BE.Hostname,
        data.WebServer.RootFilePath,
        data.WebServer.SystemAccount)
}

func ReadGlobalConfig(path string) (*GlobalConfig, error) {
    viper.SetConfigName(DefaultGlobalConfigFile)
    viper.AddConfigPath(path)
    viper.SetConfigType(DefaultConfigType)

    cfg := newGlobalConfig()

    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            // Config file not found; ignore error if desired
            fmt.Printf(
                "No config file found at %s/%s (%s), applying defaults.\n",
                path,
                DefaultGlobalConfigFile,
                DefaultConfigType)
        } else {
            // Config file was found but another error was produced
            return nil, fmt.Errorf("Fatal error config file: %s \n", err)
        }
    } else {
        // Fill in the global configuration object from the configuration file
        if err = viper.UnmarshalExact(cfg); err != nil {
            return nil, fmt.Errorf("unable to decode into struct, %v", err)
        }
    }

    return cfg, nil
}
