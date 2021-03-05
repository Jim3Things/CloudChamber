package inventory

// Inventory reader parses the YAML file and returns Zone. into a pb external zone.

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

const (
	defaultDefinitionFile = "inventory.yaml"
	defaultConfigType     = "yaml"
)

// +++ Intermediate binary format

// ++++++++++
//
// The folowing catXxx structs are used as throwaway data fields to enable
// the Viper package to both read and process a yaml configuration file
// while making use of anchors and aliases to limit the amount of repitition
// within the file.
//
// The values held in this set of structs after unmarshalling the contents of
// the configuration file have no value and should just be discarded.
//
// NOTE: Cannot extend the scheme to include complete regions to be defined
//       with an achor and used as an alias as the file read function returns
//       an error for "excessive aliasing"
//
type catZone struct {
	Name    string
	Details zoneExDetails
	Racks   []rackEx
}

type catRack struct {
	Name    string
	Details rackExDetails
	Pdus    []pduEx
	Tors    []torEx
	Blades  []bladeEx
}
type catPdu struct {
	Details pduExDetails
	Ports   []portEx
}

type catTor struct {
	Details torExDetails
	Ports   []portEx
}

type catBlade struct {
	Details       bladeExDetails
	Capacity      bladeExCapacity
	BootInfo      bladeExBootinfo
	BootOnPowerOn bool
}

type catPort struct {
	Wired bool
	Item  portExTarget
}

// ----------


// This struct is used to load the inventory definition file in a
// temporary format prior to converting to something the rest of
// the system understand and using that to update the persisted
//inventory in the store.
//
type rootEx struct {
	ZoneTypes   []catZone
	RackTypes   []catRack
	PduTypes    []catPdu
	TorTypes    []catTor
	BladeTypes  []catBlade
	Details     rootExDetails
	Regions     []regionEx
}

type rootExDetails struct {
	Name  string
	Notes string
}

type regionEx struct {
	Name    string
	Details regionExDetails
	Zones   []zoneEx	
}

type regionExDetails struct {
	State    string
	Location string
	Notes    string
}

type zoneEx struct {
	Name    string
	Details zoneExDetails
	Racks   []rackEx
}

type zoneExDetails struct {
	Enabled  bool
	State    string
	Location string
	Notes    string
}

type rackEx struct {
	Name    string
	Details rackExDetails
	Pdus    []pduEx
	Tors    []torEx
	Blades  []bladeEx
}

type rackExDetails struct {
	Enabled   bool
	Condition string
	Location  string
	Notes     string
}

type portEx struct {
	Index int64
	Wired bool
	Item  portExTarget
}

type portExTarget struct {
	Type string
	ID   int64
	Port int64
}

type pduEx struct {
	Index   int64
	Details pduExDetails
	Ports   []portEx
}

type pduExDetails struct {
	Enabled  bool
	Condition string	
}

type torEx struct {
	Index   int64
	Details torExDetails
	Ports   []portEx
}

type torExDetails struct {
	Enabled  bool
	Condition string	
}

type bladeEx struct {
	Index         int64
	Details       bladeExDetails
	Capacity      bladeExCapacity
	BootInfo      bladeExBootinfo
	BootOnPowerOn bool
}

type bladeExDetails struct {
	Enabled  bool
	Condition string	
}

type bladeExCapacity struct {
	Arch                   string
	Cores                  int64
	DiskInGb               int64
	MemoryInMb             int64
	NetworkBandwidthInMbps int64
}

type bladeExBootinfo struct {
	Source     string
	Image      string
	Version    string
	Parameters string
}

// --- Intermediate binary format


// ReadInventoryDefinitionFromFileEx imports the inventory from
// an external YAML file and transforms it into the
// internal Cloud chamber binary format.
//
func ReadInventoryDefinitionFromFileEx(ctx context.Context, path string) (*pb.Definition_Root, error) {
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
	xfr := &rootEx{}
	if err := viper.UnmarshalExact(xfr); err != nil {
		return nil, tracing.Error(ctx, "unable to decode into struct, %v", err)
	}

	// Now convert it into its final form
	root, err := toDefinitionRoot(xfr)
	if err != nil {
		return nil, tracing.Error(ctx, err)
	}

	tracing.Info(ctx, "Inventory definition Read: \n%v", root)

	return root, nil
}

var conditionmMap = map[string]pb.Condition{
	"not_in_service": pb.Condition_not_in_service,
    "operational":    pb.Condition_operational,
    "burn_in":        pb.Condition_burn_in,
    "out_for_repair": pb.Condition_out_for_repair,
    "retiring":       pb.Condition_retiring,
    "retired":        pb.Condition_retired,
}

var stateMap = map[string]pb.State{
	"out_of_service":  pb.State_out_of_service,
	"in_service":      pb.State_in_service,
	"commissioning":   pb.State_commissioning,
	"assumed_failed":  pb.State_assumed_failed,
	"decommissioning": pb.State_decommissioning,
	"decommissioned":  pb.State_decommissioned,
}

func getConditionFromString(c string) (pb.Condition, bool) {
	cond, ok := conditionmMap[strings.ToLower(c)]

	return cond, ok
}

func getStateFromString(s string) (pb.State, bool) {
	state, ok := stateMap[strings.ToLower(s)]

	return state, ok
}

var methodMap = map[string]pb.BladeBootInfo_Method{
	"local": pb.BladeBootInfo_local,
	"network": pb.BladeBootInfo_network,
}

func getBootMethodFromString(s string) (pb.BladeBootInfo_Method, bool) {
	method, ok := methodMap[strings.ToLower(s)]

	return method, ok
}

var hwTypeMap = map[string]pb.Hardware_HwType{
	"unknown": pb.Hardware_unknown,
	"pdu": 	   pb.Hardware_pdu,
	"tor":     pb.Hardware_tor,
	"blade":   pb.Hardware_blade,
}

func getHwTypeFromString(t string) (pb.Hardware_HwType, bool) {
	hwType, ok := hwTypeMap[strings.ToLower(t)]

	return hwType, ok
}

// toDefinitionRoot converts intermediate values to the final format
// One important difference is that the intermediate is array based.
// The final format is map based using specific fields in array
// entries as the map keys
//
func toDefinitionRoot(xfrRoot *rootEx) (*pb.Definition_Root, error) {

	root := &pb.Definition_Root{
		Details: &pb.RootDetails{
			Name: xfrRoot.Details.Name,
			Notes: xfrRoot.Details.Notes,
		},
		Regions: make(map[string]*pb.Definition_Region, len(xfrRoot.Regions)),
	}

	// For each rack in the supplied configuration, create rack in the
	// zone. Each rack has some details, a set of PDUs, a set of TORs,
	// and a set of blades.
	//
	for _, xfrRegion := range xfrRoot.Regions {
		if _, ok := root.Regions[xfrRegion.Name]; ok {
			return nil, errors.ErrConfigRegionDuplicate{
				Region: xfrRegion.Name,
			}
		}
		var (
			ok         bool
			state      pb.State
			condition  pb.Condition
			hwType     pb.Hardware_HwType
			bootSource pb.BladeBootInfo_Method
		)

		// Create a region and populate it from the supplied configuration
		//
		if state, ok = getStateFromString(xfrRegion.Details.State); !ok {
			return nil, errors.ErrConfigRegionBadState{
				Region: xfrRegion.Name,
				State:  xfrRegion.Details.State,
			}
		}

		region := &pb.Definition_Region{
			Details: &pb.RegionDetails{
				State:    state,
				Location: xfrRegion.Details.Location,
				Notes:    xfrRegion.Details.Notes,
			},
			Zones:   make(map[string]*pb.Definition_Zone, len(xfrRegion.Zones)),
		}
	
		// Iterate over the set of zones in the supplied configuration
		// and add an entry in the internal structs for that zone.
		//
		for _, xfrZone := range xfrRegion.Zones {
			if _, ok = region.Zones[xfrZone.Name]; ok {
				return nil, errors.ErrConfigZoneDuplicate{
					Region: xfrRegion.Name,
					Zone:   xfrZone.Name,
				}
			}

			// Create a zone and populate it from the supplied configuration
			//
			if state, ok = getStateFromString(xfrZone.Details.State); !ok {
				return nil, errors.ErrConfigZoneBadState{
					Region: xfrRegion.Name,
					Zone:   xfrZone.Name,
					State:  xfrZone.Details.State,
				}
			}

			zone := &pb.Definition_Zone{
				Details: &pb.ZoneDetails{
					State:    state,
					Enabled:  xfrZone.Details.Enabled,
					Location: xfrZone.Details.Location,
					Notes:    xfrZone.Details.Notes,
				},
				Racks: make(map[string]*pb.Definition_Rack, len(xfrZone.Racks)),
			}

			// Iterate over the set of racks in the supplied configuration
			// and add an entry in the internal structs for that rack.
			//
			for _, xfrRack := range xfrZone.Racks {
				if _, ok := zone.Racks[xfrRack.Name]; ok {
					return nil, errors.ErrConfigRackDuplicate{
						Region: xfrRegion.Name,
						Zone:   xfrZone.Name,
						Rack:   xfrRack.Name,
					}
				}

				// Create a rack and populate it from the supplied configuration
				//
				if condition, ok = getConditionFromString(xfrRack.Details.Condition); !ok {
					return nil, errors.ErrConfigRackBadCondition{
						Region:    xfrRegion.Name,
						Zone:      xfrZone.Name,
						Rack:      xfrRack.Name,
						Condition: xfrRack.Details.Condition,
					}
				}

				rack := &pb.Definition_Rack{
					Details: &pb.RackDetails{
						Condition: condition,
						Enabled:   xfrRack.Details.Enabled,
						Location:  xfrRack.Details.Location,
						Notes:     xfrRack.Details.Notes,
					},
					Pdus:   make(map[int64]*pb.Definition_Pdu, len(xfrRack.Pdus)),
					Tors:   make(map[int64]*pb.Definition_Tor, len(xfrRack.Tors)),
					Blades: make(map[int64]*pb.Definition_Blade, len(xfrRack.Blades)),
				}

				// A rack contains sub-components for Pdus, Tors and Blades with
				// it being possible to have multiple instances of each.
				//
				// Step 1 - Iterate over the set of pdus in the supplied
				// configuration and add an entry in the internal structs
				// for that pdu.
				//
				for _, xfrPdu := range xfrRack.Pdus {

					if _, ok := rack.Pdus[xfrPdu.Index]; ok {
						return nil, errors.ErrConfigPduDuplicate{
							Region:    xfrRegion.Name,
							Zone:      xfrZone.Name,
							Rack:      xfrRack.Name,
							Pdu:       xfrPdu.Index,
						}
					}

					// Create a pdu and populate it from the supplied configuration
					//
					if condition, ok = getConditionFromString(xfrPdu.Details.Condition); !ok {
						return nil, errors.ErrConfigPduBadCondition{
							Region:    xfrRegion.Name,
							Zone:      xfrZone.Name,
							Rack:      xfrRack.Name,
							Pdu:       xfrPdu.Index,
							Condition: xfrPdu.Details.Condition,
						}
					}

					pdu := &pb.Definition_Pdu{
						Details: &pb.PduDetails{
							Condition: condition,
							Enabled:   xfrPdu.Details.Enabled,
						},
						Ports: make(map[int64]*pb.PowerPort, len(xfrPdu.Ports)),
					}

					// Iterate over the set of power ports in the supplied
					// configuration and add an entry in the internal structs
					// for that port.
					//
					for _, xfrPort := range xfrPdu.Ports {

						if _, ok := pdu.Ports[xfrPort.Index]; ok {
							return nil, errors.ErrConfigPowerPortDuplicate{
								Region: xfrRegion.Name,
								Zone:   xfrZone.Name,
								Rack:   xfrRack.Name,
								Pdu:    xfrPdu.Index,
								Port:	xfrPort.Index,
							}
						}

						// Create a power port and populate it from the supplied
						// configuration
						//
						if hwType, ok = getHwTypeFromString(xfrPort.Item.Type); !ok {
							return nil, errors.ErrConfigPduHwTypeInvalid{
								Region: xfrRegion.Name,
								Zone:   xfrZone.Name,
								Rack:   xfrRack.Name,
								Pdu:    xfrPdu.Index,
								Port:   xfrPort.Index,
								Type:	xfrPort.Item.Type,
							}
						}

						// Create a power port and populate it from the supplied configuration
						//
						pdu.Ports[xfrPort.Index] = &pb.PowerPort{
							Wired: xfrPort.Wired,
							Item: &pb.Hardware{
								Type: hwType,
								Id:   xfrPort.Item.ID,
								Port: xfrPort.Item.Port,
							},
						}
					}

					rack.Pdus[xfrPdu.Index] = pdu
				}

				// Step 2 - Iterate over the set of tors in the supplied
				// configuration and add an entry in the internal structs
				// for that tor.
				//
				for _, xfrTor := range xfrRack.Tors {

					if _, ok := rack.Tors[xfrTor.Index]; ok {
						return nil, errors.ErrConfigTorDuplicate{
							Region:    xfrRegion.Name,
							Zone:      xfrZone.Name,
							Rack:      xfrRack.Name,
							Tor:       xfrTor.Index,
						}
					}

					// Create a pdu and populate it from the supplied configuration
					//
					if condition, ok = getConditionFromString(xfrTor.Details.Condition); !ok {
						return nil, errors.ErrConfigTorBadCondition{
							Region:    xfrRegion.Name,
							Zone:      xfrZone.Name,
							Rack:      xfrRack.Name,
							Tor:       xfrTor.Index,
							Condition: xfrTor.Details.Condition,
						}
					}

					tor := &pb.Definition_Tor{
						Details: &pb.TorDetails{
							Condition: condition,
							Enabled:   xfrTor.Details.Enabled,
						},
						Ports: make(map[int64]*pb.NetworkPort, len(xfrTor.Ports)),
					}

					// Iterate over the set of network ports in the supplied
					// configuration and add an entry in the internal structs
					// for that port.
					//
					for _, xfrPort := range xfrTor.Ports {

						if _, ok := tor.Ports[xfrPort.Index]; ok {
							return nil, errors.ErrConfigNetworkPortDuplicate{
								Region: xfrRegion.Name,
								Zone:   xfrZone.Name,
								Rack:   xfrRack.Name,
								Tor:    xfrTor.Index,
								Port:	xfrPort.Index,
							}
						}

						// Create a power port and populate it from the supplied
						// configuration
						//
						if hwType, ok = getHwTypeFromString(xfrPort.Item.Type); !ok {
							return nil, errors.ErrConfigTorHwTypeInvalid{
								Region: xfrRegion.Name,
								Zone:   xfrZone.Name,
								Rack:   xfrRack.Name,
								Tor:    xfrTor.Index,
								Port:   xfrPort.Index,
								Type:	xfrPort.Item.Type,
							}
						}

						// Create a power port and populate it from the supplied configuration
						//
						tor.Ports[xfrPort.Index] = &pb.NetworkPort{
							Wired: xfrPort.Wired,
							Item: &pb.Hardware{
								Type: hwType,
								Id:   xfrPort.Item.ID,
								Port: xfrPort.Item.Port,
							},
						}
					}

					rack.Tors[xfrTor.Index] = tor
				}

				// Step 3 - Iterate over the set of blades in the supplied
				// configuration and add an entry in the internal structs
				// for that blade.
				//
				for _, xfrBlade := range xfrRack.Blades {

					// If we already have a blade definition at the index
					// for the blade, it MUST be a duplicate, which we do
					// not allow, so fail describing where we found the
					// issue.
					//
					if _, ok := rack.Blades[xfrBlade.Index]; ok {
						return nil, errors.ErrConfigBladeDuplicate{
							Region: xfrRegion.Name,
							Zone:   xfrZone.Name,
							Rack:   xfrRack.Name,
							Blade:  xfrBlade.Index,
						}
					}

					// Create a blade and populate it from the supplied
					// configuration
					//
					if condition, ok = getConditionFromString(xfrBlade.Details.Condition); !ok {
						return nil, errors.ErrConfigBladeBadCondition{
							Region:    xfrRegion.Name,
							Zone:      xfrZone.Name,
							Rack:      xfrRack.Name,
							Blade:     xfrBlade.Index,
							Condition: xfrBlade.Details.Condition,
						}
					}

					if bootSource, ok = getBootMethodFromString(xfrBlade.BootInfo.Source); !ok {
						return nil, errors.ErrConfigBladeBadBootSource{
							Region:     xfrRegion.Name,
							Zone:       xfrZone.Name,
							Rack:       xfrRack.Name,
							Blade:      xfrBlade.Index,
							BootSource: xfrBlade.BootInfo.Source,
						}
					}

					blade := &pb.Definition_Blade{
						Details: &pb.BladeDetails{
							Condition: condition,
							Enabled:   xfrBlade.Details.Enabled,
						},
						Capacity: &pb.BladeCapacity{
							Cores:                  xfrBlade.Capacity.Cores,
							MemoryInMb:             xfrBlade.Capacity.MemoryInMb,
							DiskInGb:               xfrBlade.Capacity.DiskInGb,
							NetworkBandwidthInMbps: xfrBlade.Capacity.NetworkBandwidthInMbps,
							Arch:                   xfrBlade.Capacity.Arch,
						},
						BootInfo: &pb.BladeBootInfo{
							Source: bootSource,
							Image: xfrBlade.BootInfo.Image,
							Version: xfrBlade.BootInfo.Version,
							Parameters: xfrBlade.BootInfo.Parameters,
						},
						BootOnPowerOn: xfrBlade.BootOnPowerOn,
					}

					rack.Blades[xfrBlade.Index] = blade
				}

				zone.Racks[xfrRack.Name] = rack
			}

			region.Zones[xfrZone.Name] = zone
		}

		root.Regions[xfrRegion.Name] = region

		if err := region.Validate(""); err != nil {
			return nil, errors.ErrRegionValidationFailure{
				Region: xfrRegion.Name,
				Err:  err,
			}
		}
	}

	return root, nil
}
