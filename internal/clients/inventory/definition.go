package inventory

import (
	"context"

	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

type definitionBase interface {
	SetAddress(addr Address) error
	GetKey() (*string, error)
	Create(ctx context.Context) error
	Read(ctx context.Context) (*interface{}, error)
	Update(ctx context.Context) error
	Delete(ctx context.Context) error
}

type definitionRegion interface {
	definitionBase
	SetName(ctx context.Context, region string) error
	ListZones(ctx context.Context) (*map[string]*interface{}, error)
}

type definitionZone interface {
	definitionBase
	SetName(ctx context.Context, region string, zone string) error
	ListRacks(ctx context.Context) (*map[string]*interface{}, error)
}

type definitionRack interface {
	definitionBase
	SetName(ctx context.Context, region string, zone string, rack string) error
	ListPdus(ctx context.Context)   (*map[int64]*interface{}, error)
	ListTors(ctx context.Context)   (*map[int64]*interface{}, error)
	ListBlades(ctx context.Context) (*map[int64]*interface{}, error)
}

type definitionPdu interface {
	definitionBase
	SetName(ctx context.Context, region string, zone string, rack string, pdu int64) error
}

type definitionTor interface {
	definitionBase
	SetName(ctx context.Context, region string, zone string, rack string, tor int64) error
}

type definitionBlade interface {
	definitionBase
	SetName(ctx context.Context, region string, zone string, rack string, blade int64) error
}


type bladeRecord struct {
	revision int64
	record *pb.StoreRecordDefinitionBlade
}

type pduRecord struct {
	revision int64
	record *pb.StoreRecordDefinitionPdu
}

type torRecord struct {
	revision int64
	record *pb.StoreRecordDefinitionTor
}

type rackRecord struct {
	revision int64
	record *pb.StoreRecordDefinitionRack
}

type zoneRecord struct {
	revision int64
	record *pb.StoreRecordDefinitionZone
}

type regionRecord struct {
	revision int64
	record *pb.StoreRecordDefinitionRegion
}

// Blade is a
//
type Blade struct {
	region string
	zone   string
	rack   string
	id     int64
	key    string

	record *bladeRecord
}

// Pdu is a
//
type Pdu struct {
	region string
	zone   string
	rack   string
	id     int64
	key    string

	record *pduRecord
}

// Tor is a
//
type Tor struct {
	region string
	zone   string
	rack   string
	id     int64
	key    string

	record torRecord
}

// Rack is a
//
type Rack struct {

	region string
	zone   string
	rack   string
	key    string

	record *rackRecord	
}

// Zone is a
//
type Zone struct {

	region string
	zone   string
	key    string

	record *zoneRecord	
}

// Region is a
//
type Region struct {

	region string
	zone   string
	rack   string
	key    string

	record *regionRecord	
}

// SetAddress is
//
func (b *Blade) SetAddress(addr Address) error {

	return nil
}

// GetKey is
//
func (b *Blade) GetKey() (*string, error) {
	return nil, nil
}

// Create is
//
func (b *Blade) Create(ctx context.Context) error {

	return nil
}

// Read is
//
func (b *Blade) Read(ctx context.Context) (*pb.DefinitionBlade, error) {

	return nil, nil
}

// Update is
//
func (b *Blade) Update(ctx context.Context) error {

	return nil
}

// Delete is
//
func (b *Blade) Delete(ctx context.Context) error {

	return nil
}


// SetName is a 
//
func (b *Blade) SetName(ctx context.Context, region string, zone string, rack string, blade int64) error {

	key, err := GetKeyForBlade(DefinitionTable, region, zone, rack, blade)

	if nil != err {
		return err
	}

	b.region = region
	b.zone   = zone
	b.rack   = rack
	b.id     = blade

	b.key    = key
	b.record = nil

	return nil
}

