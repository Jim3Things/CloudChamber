package config

// Inventory reader parses the YAML file and returns Zone. into a pb external zone.

import (
	"context"
	"fmt"

	"github.com/spf13/viper"

	"github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/pkg/errors"
	"github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

const (
	defaultDefinitionFile = "inventory.yaml"
)

// +++ Intermediate binary format

// This mirrors the external zone format except the keys
// are fields of array

type xfrZone struct {
	Racks []rack
}

type rack struct {
	Name   string
	Blades []blade
	Tor    tor
	Pdu    pdu
}

type blade struct {
	Index                  int64
	Arch                   string
	Cores                  int64
	DiskInGb               int64
	MemoryInMb             int64
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
func ReadInventoryDefinition(ctx context.Context, path string) (*pb.ExternalZone, error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Read inventory definition from file"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	viper.SetConfigName(defaultDefinitionFile)
	viper.AddConfigPath(path)
	viper.SetConfigType(defaultConfigType)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			err = fmt.Errorf("no inventory definition found at %s/%s (%s)",
				path,
				defaultDefinitionFile,
				defaultConfigType)
		} else {
			err = fmt.Errorf("fatal error reading definition file: %s", err)
		}

		return nil, tracing.Error(ctx, err)
	}

	// First we are going to put it into intermediate format
	xfr := &xfrZone{}
	if err := viper.UnmarshalExact(xfr); err != nil {
		return nil, tracing.Error(ctx, "unable to decode into struct, %v", err)
	}

	// Now convert it into its final form
	cfg, err := toExternalZone(xfr)
	if err != nil {
		return nil, tracing.Error(ctx, err)
	}

	tracing.Info(ctx, "Inventory definition Read: \n%v", cfg)

	return cfg, nil
}

// toExternalZone converts intermediate values to the final format
// One important difference is that the intermediate is array based.
// The final format is map based using specific fields in array
// entries as the map keys
func toExternalZone(xfr *xfrZone) (*pb.ExternalZone, error) {
	cfg := &pb.ExternalZone{
		Racks: make(map[string]*pb.ExternalRack),
	}

	for _, r := range xfr.Racks {
		if _, ok := cfg.Racks[r.Name]; ok {
			return nil, errors.ErrDuplicateRack(r.Name)
		}

		cfg.Racks[r.Name] = &pb.ExternalRack{
			Tor:    &pb.ExternalTor{},
			Pdu:    &pb.ExternalPdu{},
			Blades: make(map[int64]*common.BladeCapacity),
		}

		for _, b := range r.Blades {
			if _, ok := cfg.Racks[r.Name].Blades[b.Index]; ok {
				return nil, errors.ErrDuplicateBlade{
					Blade: b.Index,
					Rack:  r.Name,
				}
			}

			cfg.Racks[r.Name].Blades[b.Index] = &common.BladeCapacity{
				Cores:                  b.Cores,
				MemoryInMb:             b.MemoryInMb,
				DiskInGb:               b.DiskInGb,
				NetworkBandwidthInMbps: b.NetworkBandwidthInMbps,
				Arch:                   b.Arch,
			}
		}

		if err := cfg.Racks[r.Name].Validate(); err != nil {
			return nil, errors.ErrRackValidationFailure{
				Rack: r.Name,
				Err:  err,
			}
		}
	}

	return cfg, nil
}

// ReadInventoryDefinitionFromFile imports the inventory from
// an external YAML file and transforms it into the
// internal Cloud chamber binary format.
//
func ReadInventoryDefinitionFromFile(ctx context.Context, path string) (*pb.Region, error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Read inventory definition from file"),
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	viper.SetConfigName(defaultDefinitionFile)
	viper.AddConfigPath(path)
	viper.SetConfigType(defaultConfigType)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			err = fmt.Errorf("no inventory definition found at %s/%s (%s)",
				path,
				defaultDefinitionFile,
				defaultConfigType)
		} else {
			err = fmt.Errorf("fatal error reading definition file: %s", err)
		}

		return nil, tracing.Error(ctx, err)
	}

	// First we are going to put it into intermediate format
	xfr := &xfrZone{}
	if err := viper.UnmarshalExact(xfr); err != nil {
		return nil, tracing.Error(ctx, "unable to decode into struct, %v", err)
	}

	// Now convert it into its final form
	region, err := toDefinitionRegionInternal(xfr)
	if err != nil {
		return nil, tracing.Error(ctx, err)
	}

	tracing.Info(ctx, "Inventory definition Read: \n%v", region)

	return region, nil
}

// toDefinitionRegionInternal converts intermediate values to the final format
// One important difference is that the intermediate is array based.
// The final format is map based using specific fields in array
// entries as the map keys
//
func toDefinitionRegionInternal(xfr *xfrZone) (*pb.Region, error) {

	region := &pb.Region{Zones: make(map[string]*pb.DefinitionZoneInternal)}

	// Since we only have a single zone at present, there is no loop
	// here. But there will be eventually.
	//
	zone := &pb.DefinitionZoneInternal{
		Details: &pb.DefinitionZone{
			Enabled:   true,
			Condition: pb.DefinitionZone_operational,
			Location:  "DC-PNW-0",
			Notes:     "Base zone",
		},
		Racks: make(map[string]*pb.DefinitionRackInternal),
	}

	// For each rack in the supplied configuration, create rack in the
	// zone. Each rack has some details, a set of PDUs, a set of TORs,
	// and a set of blades.
	//
	for _, r := range xfr.Racks {
		if _, ok := zone.Racks[r.Name]; ok {
			return nil, errors.ErrDuplicateRack(r.Name)
		}

		zone.Racks[r.Name] = &pb.DefinitionRackInternal{
			Details: &pb.DefinitionRack{
				Enabled:   true,
				Condition: pb.Definition_operational,
				Location:  "DC-PNW-0-" + r.Name,
				Notes:     "RackName: " + r.Name,
			},
			Pdus:   make(map[int64]*pb.DefinitionPdu),
			Tors:   make(map[int64]*pb.DefinitionTor),
			Blades: make(map[int64]*pb.DefinitionBlade),
		}

		// Currently only have one each of Pdu and Tor per-rack.
		//
		// NOTE: At present, the Pdu and Tor are completely
		// synthesized and not actually read from the definition
		// file.
		//
		zone.Racks[r.Name].Pdus[0] = &pb.DefinitionPdu{
			Enabled:   true,
			Condition: pb.Definition_operational,
			Ports:     make(map[int64]*pb.DefinitionPowerPort),
		}

		zone.Racks[r.Name].Tors[0] = &pb.DefinitionTor{
			Enabled:   true,
			Condition: pb.Definition_operational,
			Ports:     make(map[int64]*pb.DefinitionNetworkPort),
		}

		// We do support more than a single blade for each rack
		// so iterate over each of the blades in the supplied
		// configuration.
		//
		for _, b := range r.Blades {

			// If we already have a blade definition at the index
			// for the blade, it MUST be a duplicate, which we do
			// not allow, so fail describing where we found the
			// issue.
			//
			if _, ok := zone.Racks[r.Name].Blades[b.Index]; ok {
				return nil, errors.ErrDuplicateBlade{
					Blade: b.Index,
					Rack:  r.Name,
				}
			}

			// Add a blade definition based upon the supplied
			// configuration.and add in the fields which do not
			// (currently) have an existence in the configuration
			// file.
			//
			zone.Racks[r.Name].Blades[b.Index] = &pb.DefinitionBlade{
				Enabled:   true,
				Condition: pb.Definition_operational,
				Capacity: &common.BladeCapacity{
					Cores:                  b.Cores,
					MemoryInMb:             b.MemoryInMb,
					DiskInGb:               b.DiskInGb,
					NetworkBandwidthInMbps: b.NetworkBandwidthInMbps,
					Arch:                   b.Arch,
				},
			}

			// For the given blade index, add a matching connection
			// in the PDU to allow power for the blade to be
			// connected and controlled.
			//
			zone.Racks[r.Name].Pdus[0].Ports[b.Index] = &pb.DefinitionPowerPort{
				Wired: true,
				Item: &pb.DefinitionItem{
					Type: pb.DefinitionItem_blade,
					Id:   b.Index,
					Port: 0,
				},
			}

			// For the given blade index, add a matching connection
			// in the TOR to allow a network for the blade to be
			// connected and controlled.
			//
			zone.Racks[r.Name].Tors[0].Ports[b.Index] = &pb.DefinitionNetworkPort{
				Wired: true,
				Item: &pb.DefinitionItem{
					Type: pb.DefinitionItem_blade,
					Id:   b.Index,
					Port: 0,
				},
			}
		}

		if err := zone.Racks[r.Name].Validate(""); err != nil {
			return nil, errors.ErrRackValidationFailure{
				Rack: r.Name,
				Err:  err,
			}
		}
	}

	region.Zones["zone1"] = zone

	return region, nil
}
