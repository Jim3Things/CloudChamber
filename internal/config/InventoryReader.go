package config

// Inventory reader parses the YAML file and returns Zone. into a pb external zone.

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/api/global"

	"github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

const(
	defaultDefinitionFile = "inventory.yaml"
)

// +++ Intermediate binary format

// This mirrors the external zone format except the keys
// are fields of array

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

// --- Intermediate binary format

// ReadInventoryDefinition imports the inventory from 
// external YAML file and transforms it into the
// internal Cloud chamber binary format.
func ReadInventoryDefinition(path string) (*pb.ExternalZone, error){

	viper.SetConfigName(defaultDefinitionFile)
	viper.AddConfigPath(path)
	viper.SetConfigType(defaultConfigType)

	tr := global.TraceProvider().Tracer("")

	ctx, span :=tr.Start(
			context.Background(),
			"ReadInventoryDefinition")
	defer span.End()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			err = fmt.Errorf("No inventory definition found at %s/%s (%s)",
					path,
					defaultDefinitionFile,
					defaultConfigType)
		} else {
			err = fmt.Errorf("fatal error reading definition file: %s", err)
		}

		span.AddEvent(ctx, err.Error())
		return nil, err
	} 

	// First we are going to put it into intermediate format 
	xfr := &zone{}
	if err := viper.UnmarshalExact(xfr); err != nil {
		err = fmt.Errorf("unable to decode into struct, %v", err)
		span.AddEvent(ctx, err.Error())
		return nil, err
	}

	// Now convert it into its final form
	cfg, err := toExternalZone(xfr)
	if err != nil{
		span.AddEvent(ctx, err.Error())
		return nil, err
	}
	
	span.AddEvent(ctx,
		fmt.Sprintf("Inventory definition Read: \n%v", cfg))
	return cfg, nil	
}

// toExternalZone converts intermediate values to the final format
// One important differnce is that the intermediate is array based.
// The final format is map based using specific fields in array 
// enteries as the map keys
func toExternalZone(xfr *zone) (*pb.ExternalZone, error){
	cfg := &pb.ExternalZone{ 
		Racks : make(map[string]*pb.ExternalRack),
	}

	for _, r := range xfr.Racks { 
		if _, ok := cfg.Racks [r.Name]; ok { 
			return nil, fmt.Errorf("Duplicate rack %q detected", r.Name)
		}
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

 	return cfg, nil
}

// to check that the unique rack value is returned or 
// unique blade value for a rack is returned
//  before creating a new rack check that rack name is 
// not already in the map
// before creating  a new blade make sure that its index
// is not already in the map
// once we have created a rack call validate on that rack 
// and returns if that gets an error
