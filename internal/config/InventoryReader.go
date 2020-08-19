// Inventory reader parses the YAML file and returns Zone. into a pb external zone.

package config

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/api/global"

	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

const(
	DefaultDefinitionFile = "inventory.yaml"
)

func ReadInventoryDefinition(path string) (*pb.ExternalZone, error){

	viper.SetConfigName(DefaultDefinitionFile)
	viper.AddConfigPath(path)
	viper.SetConfigType(DefaultConfigType)

	cfg := &pb.ExternalZone{}

	tr := global.TraceProvider().Tracer("")

	ctx, span :=tr.Start(
			context.Background(),
			"ReadInventoryDefinition")
		defer span.End()
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
							// Config file not found; we'll just use the default values
							span.AddEvent(
											ctx,
											fmt.Sprintf(
															"No inventory definitino found at %s/%s (%s), applying defaults.",
															path,
															DefaultDefinitionFile,
															DefaultConfigType))
			} else {
							// Config file was found but another error was produced
							err = fmt.Errorf("fatal error reading definition file: %s", err)
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
			fmt.Sprintf("Inventory definition Read: \n%v", cfg))

return cfg, nil
}


