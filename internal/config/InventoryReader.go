package config

// Inventory reader parses the YAML file and returns Zone. into a pb external zone.

import (
	"context"
	"fmt"

	"github.com/spf13/viper"

	"github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

const (
	defaultDefinitionFile = "inventory.yaml"
)

// +++ Intermediate binary format

// This mirrors the external zone format except the keys
// are fields of array

type zone struct {
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

// ErrDuplicateRack indicates duplicates rack names found
type ErrDuplicateRack string

func (edr ErrDuplicateRack) Error() string {
	return fmt.Sprintf("Duplicate rack %q detected", string(edr))
}

// ErrDuplicateBlade indicates duplicates blade indexes found
type ErrDuplicateBlade struct {
	rack  string
	blade int64
}

func (edb ErrDuplicateBlade) Error() string {
	return fmt.Sprintf("Duplicate Blade %d in Rack %q detected", edb.blade, edb.rack)
}

// ErrValidationFailure indicates validation failure in blades
type ErrValidationFailure struct {
	rack string
	err  error
}

func (evf ErrValidationFailure) Error() string {
	return fmt.Sprintf("In rack %q: %v", evf.rack, evf.err)
}

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
	if err != nil {
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
func toExternalZone(xfr *zone) (*pb.ExternalZone, error) {
	cfg := &pb.ExternalZone{
		Racks: make(map[string]*pb.ExternalRack),
	}

	for _, r := range xfr.Racks {
		if _, ok := cfg.Racks[r.Name]; ok {
			return nil, ErrDuplicateRack(r.Name)
		}

		cfg.Racks[r.Name] = &pb.ExternalRack{
			Tor:    &pb.ExternalTor{},
			Pdu:    &pb.ExternalPdu{},
			Blades: make(map[int64]*common.BladeCapacity),
		}

		for _, b := range r.Blades {
			if _, ok := cfg.Racks[r.Name].Blades[b.Index]; ok {
				return nil, ErrDuplicateBlade{
					blade: b.Index,
					rack:  r.Name,
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
			return nil, ErrValidationFailure{
				rack: r.Name,
				err:  err,
			}
		}
	}

	return cfg, nil
}

// ReadInventoryDefinitionFromFile imports the inventory from
// an external YAML file and transforms it into the
// internal Cloud chamber binary format.
//
func ReadInventoryDefinitionFromFile(ctx context.Context, path string) (*map[string]*pb.DefinitionZone, error) {
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
	zonemap, err := toDefinitionZone(xfr)
	if err != nil {
		span.AddEvent(ctx, err.Error())
		return nil, err
	}

	span.AddEvent(ctx,
		fmt.Sprintf("Inventory definition Read: \n%v", *zonemap))
	return zonemap, nil
}

// toDefinitionZone converts intermediate values to the final format
// One important differnce is that the intermediate is array based.
// The final format is map based using specific fields in array
// enteries as the map keys
//
func toDefinitionZone(xfr *zone) (*map[string]*pb.DefinitionZone, error) {

	zonemap := make(map[string]*pb.DefinitionZone)

	// Since we only have a single zone at present, there is no loop
	// here. But there will be eventually.
	//
	zone := &pb.DefinitionZone{
		Enabled: true,
		Condition: pb.Definition_operational,
		Location: "DC-PNW-0",
		Notes: "Base zone",
		Racks: make(map[string]*pb.DefinitionRack),
	}

	for _, r := range xfr.Racks {
		if _, ok := zone.Racks[r.Name]; ok {
			return nil, ErrDuplicateRack(r.Name)
		}

		zone.Racks[r.Name] = &pb.DefinitionRack{
			Enabled:   true,
			Condition: pb.Definition_operational,
			Location:  "DC-PNW-0-" + r.Name,
			Notes:     "RackName: " + r.Name,
			Pdus:      make(map[int64]*pb.DefinitionPdu),
			Tors:      make(map[int64]*pb.DefinitionTor),
			Blades:    make(map[int64]*pb.DefinitionBlade),
		}

		// Currently only have one each of Pdu and Tor per-rack.
		//
		// NOTE: At present, the Pdu and Tor are completely
		// synthsized and not actually read from the definition
		// file.
		//
		zone.Racks[r.Name].Pdus[0] = &pb.DefinitionPdu{
			Enabled:   true,
			Powered:   true,
			Condition: pb.Definition_operational,
			Ports:     make(map[int64]*pb.DefinitionPowerPort),
		}

		zone.Racks[r.Name].Tors[0] = &pb.DefinitionTor{
			Enabled:   true,
			Powered:   true,
			Condition: pb.Definition_operational,
			Ports:     make(map[int64]*pb.DefinitionNetworkPort),
		}

		for _, b := range r.Blades {
			if _, ok := zone.Racks[r.Name].Blades[b.Index]; ok {
				return nil, ErrDuplicateBlade{
					blade: b.Index,
					rack:  r.Name,
				}
			}

			zone.Racks[r.Name].Blades[b.Index] = &pb.DefinitionBlade{
				Enabled: true,
				Condition: pb.Definition_operational,
				Capacity: &common.BladeCapacity{
					Cores:                  b.Cores,
					MemoryInMb:             b.MemoryInMb,
					DiskInGb:               b.DiskInGb,
					NetworkBandwidthInMbps: b.NetworkBandwidthInMbps,
					Arch:                   b.Arch,
				},
			}

			zone.Racks[r.Name].Pdus[0].Ports[b.Index] = &pb.DefinitionPowerPort{
				Connected: true,
				Powered:   true,
			}

			zone.Racks[r.Name].Tors[0].Ports[b.Index] = &pb.DefinitionNetworkPort{
				Connected: true,
				Enabled:   true,
			}
		}

		if err := zone.Racks[r.Name].Validate(""); err != nil {
			return nil, ErrValidationFailure{
				rack: r.Name,
				err:  err,
			}
		}
	}

	zonemap["zone1"] = zone

	return &zonemap, nil
}
