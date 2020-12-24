package inventory

// Inventory reader parses the YAML file and returns Zone. into a pb external zone.

import (
	"context"
	"fmt"

	"github.com/spf13/viper"

	"github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

const (
	defaultDefinitionFile = "inventory.yaml"
	defaultConfigType     = "yaml"
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
			Blades: make(map[int64]*pb.BladeCapacity),
		}

		for _, b := range r.Blades {
			if _, ok := cfg.Racks[r.Name].Blades[b.Index]; ok {
				return nil, errors.ErrDuplicateBlade{
					Blade: b.Index,
					Rack:  r.Name,
				}
			}

			cfg.Racks[r.Name].Blades[b.Index] = &pb.BladeCapacity{
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
func ReadInventoryDefinitionFromFile(ctx context.Context, path string) (*pb.DefinitionRegion, error) {
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
// One important differnce is that the intermediate is array based.
// The final format is map based using specific fields in array
// entries as the map keys
//
func toDefinitionRegionInternal(xfr *zone) (*pb.DefinitionRegion, error) {

	region := &pb.DefinitionRegion{
		Details: &pb.RegionDetails{},
		Zones: make(map[string]*pb.DefinitionZone)}

	// Since we only have a single zone at present, there is no loop
	// here. But there will be eventually.
	//
	zone := &pb.DefinitionZone{
		Details: &pb.ZoneDetails{
			Enabled:   true,
			State:     pb.State_in_service,
			Location:  "DC-PNW-0",
			Notes:     "Base zone",
		},
		Racks: make(map[string]*pb.DefinitionRack),
	}

	// For each rack in the supplied configuration, create rack in the
	// zone. Each rack has some details, a set of PDUs, a set of TORs,
	// and a set of blades.
	//
	for _, r := range xfr.Racks {
		if _, ok := zone.Racks[r.Name]; ok {
			return nil, errors.ErrDuplicateRack(r.Name)
		}

		rack := &pb.DefinitionRack{
			Details: &pb.RackDetails{
				Enabled:   true,
				Condition: pb.Condition_operational,
				Location:  "DC-PNW-0-" + r.Name,
				Notes:     "RackName: " + r.Name,
			},
			Pdus:   make(map[int64]*pb.DefinitionPdu),
			Tors:   make(map[int64]*pb.DefinitionTor),
			Blades:    make(map[int64]*pb.DefinitionBlade),
		}

		// Currently only have one each of Pdu and Tor per-rack.
		//
		// NOTE: At present, the Pdu and Tor are completely
		// synthesized and not actually read from the definition
		// file.
		//
		rack.Pdus[0] = &pb.DefinitionPdu{
			Details: &pb.PduDetails{
				Enabled:   true,
				Condition: pb.Condition_operational,
			},
			Ports:     make(map[int64]*pb.PowerPort),
		}

		rack.Tors[0] = &pb.DefinitionTor{
			Details: &pb.TorDetails{
				Enabled:   true,
				Condition: pb.Condition_operational,
			},
			Ports:     make(map[int64]*pb.NetworkPort),
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
			if _, ok := rack.Blades[b.Index]; ok {
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
			rack.Blades[b.Index] = &pb.DefinitionBlade{
				Details: &pb.BladeDetails{
				Enabled: true,
					Condition: pb.Condition_operational,
				},
				Capacity: &pb.BladeCapacity{
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
			rack.Pdus[0].Ports[b.Index] = &pb.PowerPort{
				Wired: true,
				Item:  &pb.Hardware{
					Type: pb.Hardware_blade,
					Id:   b.Index,
					Port: 0,
				},
			}

			// For the given blade index, add a matching connection
			// in the TOR to allow a network for the blade to be
			// connected and controlled.
			//
			rack.Tors[0].Ports[b.Index] = &pb.NetworkPort{
				Wired: true,
				Item:  &pb.Hardware{
					Type: pb.Hardware_blade,
					Id:   b.Index,
					Port: 0,
				},
			}
		}

		if err := rack.Validate(""); err != nil {
			return nil, errors.ErrRackValidationFailure{
				Rack: r.Name,
				Err:  err,
		}

		zone.Racks[r.Name] = rack
	}

	region.Zones[DefaultZone] = zone

	return region, nil
}
