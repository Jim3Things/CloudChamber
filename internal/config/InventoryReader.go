// Inventory reader parses the YAML file and returns Zone. into a pb external zone.

package config

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/api/global"

	"github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

const(
	DefaultDefinitionFile = "inventory.yaml"
)

type zone struct {
	Racks []rack
}

type rack struct {
	Name string
	Blades []blade
	Tor tor
	Pdu pdu
}

type blade struct {
	Index int64
	Arch string
	Cores int64
	DiskInGb int64
	MemoryInMb int64
	NetworkBandwidthInMbps int64
}

type tor struct {

}

type pdu struct {

}

func ReadInventoryDefinition(path string) (*pb.ExternalZone, error){

	viper.SetConfigName(DefaultDefinitionFile)
	viper.AddConfigPath(path)
	viper.SetConfigType(DefaultConfigType)

	
	tr := global.TraceProvider().Tracer("")

	ctx, span :=tr.Start(
			context.Background(),
			"ReadInventoryDefinition")
	defer span.End()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; we'll just use the default values
			err = fmt.Errorf("No inventory definition found at %s/%s (%s)",
					path,
					DefaultDefinitionFile,
					DefaultConfigType)
			span.AddEvent(ctx, err.Error())
			return nil, err
		} else {
			// Config file was found but another error was produced
			err = fmt.Errorf("fatal error reading definition file: %s", err)
			span.AddEvent(ctx, err.Error())
			return nil, err
		}
	} 

	// Fill in the global configuration object from the configuration file
	xfr := &zone{}
	if err := viper.UnmarshalExact(xfr); err != nil {
		err = fmt.Errorf("unable to decode into struct, %v", err)
		span.AddEvent(ctx, err.Error())
		return nil, err
	}
	// function to translate xfr into an external zone
	cfg := ToExternalZone(xfr)
	span.AddEvent(ctx,
		fmt.Sprintf("Inventory definition Read: \n%v", cfg))
	return cfg, nil	
}

func ToExternalZone (xfr *zone) *pb.ExternalZone{
	cfg := &pb.ExternalZone{
		Racks : make(map[string]*pb.ExternalRack),
	}
	for _, r := range xfr.Racks {
		cfg.Racks [r.Name] = &pb.ExternalRack{
			Tor: &pb.ExternalTor{},
			Pdu: &pb.ExternalPdu{},
			Blades: make(map[int64]*common.BladeCapacity),
		}
		for _, b := range r.Blades{
			cfg.Racks [r.Name].Blades[b.Index] = &common.BladeCapacity{
				Cores: b.Cores,
				MemoryInMb: b.MemoryInMb,
				DiskInGb: b.DiskInGb,
				NetworkBandwidthInMbps: b.NetworkBandwidthInMbps,
				Arch: b.Arch,
			}
		}
	}
 return cfg
}
