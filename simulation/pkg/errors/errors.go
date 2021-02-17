package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrNotInitialized is a new error to indicate initialization failures.
	ErrNotInitialized = errors.New("CloudChamber: initialization failure")

	// ErrWorkloadNotEnabled indicates the specified workload is not enabled
	// for the purposes of deployment or execution.
	ErrWorkloadNotEnabled = errors.New("CloudChamber: workload not enabled")

	// ErrStoreUnableToCreateClient indicates that it is not currently possible
	// to create a client.
	//
	ErrStoreUnableToCreateClient = errors.New("CloudChamber: unable to create a new client")

	// ErrUserUnableToCreate indicates the specified user account cannot be
	// created at this time
	ErrUserUnableToCreate = errors.New("CloudChamber: unable to create a user account at this time")

	// ErrUserAlreadyLoggedIn indicates that the session is currently logged in
	// and a new log in cannot be processed
	ErrUserAlreadyLoggedIn = errors.New("CloudChamber: session already has a logged in user")

	// ErrUserAuthFailed indicates the supplied username and password combination
	// is not valid.
	ErrUserAuthFailed = errors.New("CloudChamber: authentication failed, invalid user name or password")

	// ErrCableStuck indicates that an attempt to control a cable connection
	// has failed because the cable control is faulty.
	ErrCableStuck = errors.New("cable is faulted")

	// ErrNoOperation indicates that the request is superfluous - the element
	// is already in the target state.
	ErrNoOperation = errors.New("repair operation specified the current state, no change occurred")

	// ErrAlreadyStarted indicates that the start machine is already
	// executing, and the start request is in error.
	ErrAlreadyStarted = errors.New("the state machine has already started")

	// ErrInvalidTarget is an error used to indicate that the incoming message had
	// a target element that either was not valid for the message, or an element
	// that could not be found.
	ErrInvalidTarget = errors.New("invalid target specified, request ignored")

	// ErrInvalidMessage indicates that an attempt to process an unexpected
	// message type occurred when receiving a grpc message.
	ErrInvalidMessage = errors.New("invalid message encountered")

	// ErrDelayCanceled indicates to the original waiter that their outstanding
	// stepper delay operation has been canceled.
	ErrDelayCanceled = errors.New("the delay operation was canceled")

	// ErrAlreadyOpen indicates that an attempt was made to open an exporter
	// while it was open and active.
	ErrAlreadyOpen = errors.New("CloudChamber: exporter is already open")

	// ErrOpenAttrsNil indicates that an Open exporter operation was passed no
	// arguments, but argument values were required.
	ErrOpenAttrsNil = errors.New("CloudChamber: Exporter.Open attributes must not be nil")

	// ErrNullItem indicates the supplied item does not exist
	//
	ErrNullItem = errors.New("item not initialized")

	// ErrFunctionNotAvailable indicates the specified object does
	// not have the requested method.
	//
	ErrFunctionNotAvailable = errors.New("function not available")
)

// ErrInventoryChangeTooLate indicates that an attempt to modify an inventory
// element failed because the element had changed since the guard time passed
// in to the modification request.
type ErrInventoryChangeTooLate int64

func (e ErrInventoryChangeTooLate) Error() string {
	return fmt.Sprintf(
		"inventory element has been modified later than the check condition time (tick %d)",
		int64(e))
}

// ErrClientAlreadyInitialized indicates that the attempt to initialize the
// specified internal client library used to mediate access to a microservice
// has failed because it was already initialized.
type ErrClientAlreadyInitialized string

func (eai ErrClientAlreadyInitialized) Error() string {
	return fmt.Sprintf("the %s client has already been initialized", string(eai))
}

// ErrClientNotReady indicates that an attempt to call an internal microservice
// through a client library has failed, as that library has not yet been
// initialized.
type ErrClientNotReady string

func (enr ErrClientNotReady) Error() string {
	return fmt.Sprintf("the %s client is not ready to process requests", string(enr))
}

// ErrTimerNotFound indicates that the specified timer ID was not found when
// attempting to look up an active timer.
type ErrTimerNotFound int

func (etnf ErrTimerNotFound) Error() string {
	return fmt.Sprintf("time ID %d was not found", int(etnf))
}

// ErrStoreNotConnected indicates the store instance does not have a
// currently active client. The Connect() method can be used to establish a client.
//
type ErrStoreNotConnected string

func (esnc ErrStoreNotConnected) Error() string {
	return fmt.Sprintf("CloudChamber: client not currently connected - %s", string(esnc))
}

// ErrStoreConnected indicates the request failed as the store is currently
// connected and the request is not possible in that condition.
//
type ErrStoreConnected string

func (esc ErrStoreConnected) Error() string {
	return fmt.Sprintf("CloudChamber: client currently connected - %s", string(esc))
}

// ErrStoreConnectionFailed indicates that the attempt to establish a connection
// from this client to the store failed
//
type ErrStoreConnectionFailed struct {
	Endpoints []string
	Reason    error
}

func (escf ErrStoreConnectionFailed) Error() string {
	return fmt.Sprintf(
		"CloudChamber: failed to establish connection to store - Endpoints: %q Reason: %q",
		escf.Endpoints, escf.Reason)
}

// ErrStoreIndexNotFound indicates the requested index was not found when the store
// lookup/fetch was attempted.
//
type ErrStoreIndexNotFound string

func (esinf ErrStoreIndexNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: index %q not found", string(esinf))
}

// ErrStoreKeyNotFound indicates the request key was not found when the store
// lookup/fetch was attempted.
//
type ErrStoreKeyNotFound string

func (esknf ErrStoreKeyNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: key %q not found", string(esknf))
}

// ErrStoreKeyTypeMismatch indicates the request key was not found when the store
// lookup/fetch was attempted.
//
type ErrStoreKeyTypeMismatch string

func (esktm ErrStoreKeyTypeMismatch) Error() string {
	return fmt.Sprintf("CloudChamber: key %q not the requested type", string(esktm))
}

// ErrStoreNotImplemented indicated the called method does not yet have an
// implementation
//
type ErrStoreNotImplemented string

func (esni ErrStoreNotImplemented) Error() string {
	return fmt.Sprintf("CloudChamber: method %s not currently implemented", string(esni))
}

// ErrStoreKeyReadFailure indicates the read transaction failed.
//
type ErrStoreKeyReadFailure string

func (eskff ErrStoreKeyReadFailure) Error() string {
	return fmt.Sprintf("CloudChamber: transaction failed reading key %q", string(eskff))
}

// ErrStoreKeyWriteFailure indicates the write transaction failed as a result of a pre-condition failure.
//
type ErrStoreKeyWriteFailure string

func (eskwf ErrStoreKeyWriteFailure) Error() string {
	return fmt.Sprintf("CloudChamber: transaction failed writing key %q", string(eskwf))
}

// ErrStoreKeyDeleteFailure indicates the read transaction failed.
//
type ErrStoreKeyDeleteFailure string

func (eskdf ErrStoreKeyDeleteFailure) Error() string {
	return fmt.Sprintf("CloudChamber: transaction failed deleting key %q", string(eskdf))
}

// ErrStoreBadRecordKey indicates the store has found a record with an unrecognized
// format. This generally means the key itself is not properly constructed.
//
type ErrStoreBadRecordKey string

func (esbrk ErrStoreBadRecordKey) Error() string {
	return fmt.Sprintf("CloudChamber: discovered key with an unrecognized format %q", string(esbrk))
}

// ErrStoreBadRecordContent indicates the store has found a record with some content
// that does not match the key. An example might be that the user name used for a key
// does not match the user name field in the record.
//
// There is little consistency checking of this nature in the store itself due to the
// limited knowledge the store component has about the content of records. There
// should be no expectation that the store is taking on the responsibility of any
// consistency checking and any that does occur should be treated as advisory.
//
type ErrStoreBadRecordContent string

func (esbrk ErrStoreBadRecordContent) Error() string {
	return fmt.Sprintf("CloudChamber: discovered record where content does not match key %q", string(esbrk))
}

// ErrStoreBadRecordCount indicates the record count for the operation was not valid.
// This might mean that the store found more, or less, than the number of records expected.
//
type ErrStoreBadRecordCount struct {
	Key      string
	Expected int
	Actual   int
}

func (esbrc ErrStoreBadRecordCount) Error() string {
	return fmt.Sprintf(
		"CloudChamber: unexpected record count for key %q - expected: %d actual %d",
		esbrc.Key, esbrc.Expected, esbrc.Actual)
}

// ErrStoreBadArgCondition indicates the condition argument specified for the update was not valid.
//
type ErrStoreBadArgCondition struct {
	Key       string
	Condition string
}

func (esbac ErrStoreBadArgCondition) Error() string {
	return fmt.Sprintf("CloudChamber: compare operator %q not valid for key %q", esbac.Condition, esbac.Key)
}

// ErrStoreBadArgRevision indicates the supplied revision argument was invalid.
//
type ErrStoreBadArgRevision struct {
	Key       string
	Requested int64
	Actual    int64
}

func (esbar ErrStoreBadArgRevision) Error() string {
	return fmt.Sprintf(
		"CloudChamber: invalid revision argument supplied on update for key %q - requested: %d actual: %d",
		esbar.Key, esbar.Requested, esbar.Actual)
}

// ErrStoreConditionFail indicates the update transaction failed due to a
// failure in the requested condition. This is like a revision mismatch
// but other conditions may apply.
//
type ErrStoreConditionFail struct {
	Key       string
	Requested int64
	Condition string
	Actual    int64
}

func (esucf ErrStoreConditionFail) Error() string {
	return fmt.Sprintf(
		"CloudChamber: condition failure on update for key %q - requested: %d condition: %s actual: %d",
		esucf.Key, esucf.Requested, esucf.Condition, esucf.Actual)
}

// ErrStoreAlreadyExists indicates the key, value pair being created already exists
//
type ErrStoreAlreadyExists string

func (esae ErrStoreAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: condition failure (already exists) on create for key %q", string(esae))
}

// ErrStoreInvalidConfiguration indicates the key, value pair being created already exists
//
type ErrStoreInvalidConfiguration string

func (esic ErrStoreInvalidConfiguration) Error() string {
	return fmt.Sprintf("CloudChamber: invalid store configuration - %s", string(esic))
}

// ErrDuplicateRack indicates duplicates rack names found
type ErrDuplicateRack string

func (edr ErrDuplicateRack) Error() string {
	return fmt.Sprintf("Duplicate rack %q detected", string(edr))
}

// ErrDuplicateBlade indicates duplicates blade indexes found
type ErrDuplicateBlade struct {
	Rack  string
	Blade int64
}

func (edb ErrDuplicateBlade) Error() string {
	return fmt.Sprintf("Duplicate Blade %d in Rack %q detected", edb.Blade, edb.Rack)
}

// ErrDuplicateRegion indicates duplicates region names found
type ErrDuplicateRegion string

func (e ErrDuplicateRegion) Error() string {
	return fmt.Sprintf("Duplicate region %q detected", string(e))
}

// ErrConfigRegionDuplicate indicates duplicate region names found
//
type ErrConfigRegionDuplicate struct {
	Region string
}

func (e ErrConfigRegionDuplicate) Error() string {
	return fmt.Sprintf(
		"CloudChamber: Configuration detected duplicate Region %s",
		e.Region)
}

// ErrConfigRegionBadState indicates the state for the region was unrecognized
//
type ErrConfigRegionBadState struct {
	Region string
	State  string
}

func (e ErrConfigRegionBadState) Error() string {
	return fmt.Sprintf(
		"CloudChamber: Configuration detected invalid state %q for region %q",
		e.State,
		e.Region)
}

// ErrConfigZoneDuplicate indicates duplicate zone names found
//
type ErrConfigZoneDuplicate struct {
	Region string
	Zone   string
}

func (e ErrConfigZoneDuplicate) Error() string {
	return fmt.Sprintf(
		"CloudChamber: Configuration detected duplicate Zone %q in %s",
		e.Zone,
		regionAddress(e.Region))
}

// ErrConfigZoneBadState indicates the state for the zone was unrecognized
//
type ErrConfigZoneBadState struct {
	Region string
	Zone   string
	State  string
}

func (e ErrConfigZoneBadState) Error() string {
	return fmt.Sprintf(
		"CloudChamber: Configuration detected invalid state %q for %s",
	 	e.State,
		zoneAddress(e.Zone, e.Region))
}

// ErrConfigRackDuplicate indicates duplicate rack names found
//
type ErrConfigRackDuplicate struct {
	Region string
	Zone   string
	Rack   string
}

func (e ErrConfigRackDuplicate) Error() string {
	return fmt.Sprintf(
		"CloudChamber: Configuration detected duplicate Rack %s in %s",
		e.Rack,
		zoneAddress(e.Region, e.Zone))
}

// ErrConfigRackBadCondition indicates the condition for the rack was unrecognized
//
type ErrConfigRackBadCondition struct {
	Region    string
	Zone      string
	Rack      string
	Condition string
}

func (e ErrConfigRackBadCondition) Error() string {
	return fmt.Sprintf(
		"CloudChamber: Configuration detected invalid condition %q for %s",
	 	e.Condition,
		rackAddress(e.Zone, e.Region, e.Rack))
}

// ErrConfigPduDuplicate indicates duplicate pdu indexes found
type ErrConfigPduDuplicate struct {
	Region string
	Zone   string
	Rack   string
	Pdu    int64
}

func (e ErrConfigPduDuplicate) Error() string {
	return fmt.Sprintf(
		"CloudChamber: Configuration detected duplicate Pdu %d in %s",
		e.Pdu,
		pduAddress(e.Region, e.Zone, e.Rack, e.Pdu))
}

// ErrConfigPduBadCondition indicates the condition for the pdu was unrecognized
//
type ErrConfigPduBadCondition struct {
	Region    string
	Zone      string
	Rack      string
	Pdu       int64
	Condition string
}

func (e ErrConfigPduBadCondition) Error() string {
	return fmt.Sprintf(
		"CloudChamber: Configuration detected invalid condition %q for %s",
	 	e.Condition,
		pduAddress(e.Zone, e.Region, e.Rack, e.Pdu))
}

// ErrConfigTorDuplicate indicates duplicate tor indexes found
type ErrConfigTorDuplicate struct {
	Region string
	Zone   string
	Rack   string
	Tor    int64
}

func (e ErrConfigTorDuplicate) Error() string {
	return fmt.Sprintf(
		"CloudChamber: Configuration detected duplicate Tor %d in %s",
		e.Tor,
		torAddress(e.Region, e.Zone, e.Rack, e.Tor))
}

// ErrConfigTorBadCondition indicates the condition for the tor was unrecognized
//
type ErrConfigTorBadCondition struct {
	Region    string
	Zone      string
	Rack      string
	Tor       int64
	Condition string
}

func (e ErrConfigTorBadCondition) Error() string {
	return fmt.Sprintf(
		"CloudChamber: Configuration file detected invalid condition %q for %s",
	 	e.Condition,
		torAddress(e.Region, e.Zone, e.Rack, e.Tor))
}

// ErrConfigBladeDuplicate indicates duplicate tor indexes found
//
type ErrConfigBladeDuplicate struct {
	Region string
	Zone   string
	Rack   string
	Blade  int64
}

func (e ErrConfigBladeDuplicate) Error() string {
	return fmt.Sprintf(
		"CloudChamber: Configuration detected duplicate Blade %d in %s",
		e.Blade,
		bladeAddress(e.Region, e.Zone, e.Rack, e.Blade))
}

// ErrConfigBladeBadCondition indicates the state for the zone was unrecognized
//
type ErrConfigBladeBadCondition struct {
	Region    string
	Zone      string
	Rack      string
	Blade     int64
	Condition string
}

func (e ErrConfigBladeBadCondition) Error() string {
	return fmt.Sprintf(
		"CloudChamber: Configuration file has invalid condition %q for %s",
	 	e.Condition,
		bladeAddress(e.Zone, e.Region, e.Rack, e.Blade))
}

// ErrConfigPowerPortDuplicate indicates duplicate pdu port indexes found
//
type ErrConfigPowerPortDuplicate struct {
	Region string
	Zone   string
	Rack   string
	Pdu    int64
	Port   int64
}

func (e ErrConfigPowerPortDuplicate) Error() string {
	return fmt.Sprintf(
		"CloudChamber: Configuration detected duplicate power port %d in %s",
		e.Port,
		pduAddress(e.Region, e.Zone, e.Rack, e.Pdu))
}

// ErrConfigNetworkPortDuplicate indicates duplicate tor port indexes found
//
type ErrConfigNetworkPortDuplicate struct {
	Region string
	Zone   string
	Rack   string
	Tor    int64
	Port   int64
}

func (e ErrConfigNetworkPortDuplicate) Error() string {
	return fmt.Sprintf(
		"CloudChamber: Configuration detected duplicate network port %d in %s",
		e.Port,
		torAddress(e.Region, e.Zone, e.Rack, e.Tor))
}

// ErrConfigPduHwTypeInvalid indicates invalid hw type connected to a pdu port
//
type ErrConfigPduHwTypeInvalid struct {
	Region string
	Zone   string
	Rack   string
	Pdu    int64
	Port   int64
	Type   string
}

func (e ErrConfigPduHwTypeInvalid) Error() string {
	return fmt.Sprintf(
		"CloudChamber: Configuration detected invalid hardware type %q at %s",
		e.Type,
		powerPortAddress(e.Region, e.Zone, e.Rack, e.Pdu, e.Port))
}

// ErrConfigTorHwTypeInvalid indicates invalid hw type connected to a tor port
//
type ErrConfigTorHwTypeInvalid struct {
	Region string
	Zone   string
	Rack   string
	Tor    int64
	Port   int64
	Type   string
}

func (e ErrConfigTorHwTypeInvalid) Error() string {
	return fmt.Sprintf(
		"CloudChamber: Configuration detected invalid hardware type %q at %s",
		e.Type,
		networkPortAddress(e.Region, e.Zone, e.Rack, e.Tor, e.Port))
}

// ErrConfigBladeBadBootSource indicates invalid hw type connected to a tor port
//
type ErrConfigBladeBadBootSource struct {
	Region     string
	Zone       string
	Rack       string
	Blade      int64
	BootSource string
}

func (e ErrConfigBladeBadBootSource) Error() string {
	return fmt.Sprintf(
		"CloudChamber: Configuration detected invalid boot source %q at %s",
		e.BootSource,
		bladeAddress(e.Region, e.Zone, e.Rack, e.Blade))
}

// ErrRackValidationFailure indicates validation failure in the attributes associated
// with a rack.
type ErrRackValidationFailure struct {
	Rack string
	Err  error
}

func (evf ErrRackValidationFailure) Error() string {
	return fmt.Sprintf("In rack %q: %v", evf.Rack, evf.Err)
}

// ErrRegionValidationFailure indicates validation failure in the attributes associated
// with a rack.
type ErrRegionValidationFailure struct {
	Region string
	Err    error
}

func (e ErrRegionValidationFailure) Error() string {
	return fmt.Sprintf("In rack %q: %v", e.Region, e.Err)
}

// ErrUserAlreadyExists indicates the attempt to create a new user account
// failed as that user already exists.
//
type ErrUserAlreadyExists string

func (euae ErrUserAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: user %q already exists", string(euae))
}

// ErrUserNotFound indicates the attempt to locate a user account failed as that
// user does not exist.
//
type ErrUserNotFound string

func (eunf ErrUserNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: user %q not found", string(eunf))
}

// ErrUserStaleVersion indicates the attempt to locate a user account failed as that
// user does not exist.
//
type ErrUserStaleVersion string

func (eusv ErrUserStaleVersion) Error() string {
	return fmt.Sprintf("CloudChamber: user %q has a newer version than expected", string(eusv))
}

// ErrUserProtected indicates the attempt to locate a user account failed as that
// user does not exist.
//
type ErrUserProtected string

func (eup ErrUserProtected) Error() string {
	return fmt.Sprintf("CloudChamber: user %q is protected and cannot be deleted", string(eup))
}

// ErrUserBadRecordContent indicates the user record retrieved from the store store
// has some content that does not match the key. An example might be that the user
// name used for a key does not match the user name field in the record.
//
type ErrUserBadRecordContent struct {
	Name  string
	Value string
}

func (eubrc ErrUserBadRecordContent) Error() string {
	return fmt.Sprintf(
		"CloudChamber: discovered record for user %q where the content does not match key %q",
		eubrc.Name, eubrc.Value)
}

// ErrUnableToVerifySystemAccount indicates that the system account has changed
// from what is defined in the configuration
type ErrUnableToVerifySystemAccount struct {
	Name string
	Err  error
}

func (eutvsa ErrUnableToVerifySystemAccount) Error() string {
	return fmt.Sprintf(
		"CloudChamber: unable to verify the standard %q account is using configured password - error %v",
		eutvsa.Name, eutvsa.Err)
}

// ErrRegionAlreadyExists indicates the attempt to create a new region record
// failed as that region already exists.
//
type ErrRegionAlreadyExists struct {
	Region string
}

func (e ErrRegionAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: %s already exists", regionAddress(e.Region))
}

// ErrRegionNotFound indicates the attempt to locate a region record failed as that
// region does not exist.
//
type ErrRegionNotFound struct {
	Region string
}

func (e ErrRegionNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: %s was not found", regionAddress(e.Region))
}

// ErrRegionIndexNotFound indicates the attempt to locate a region index record
// failed as that region index does not exist.
//
type ErrRegionIndexNotFound struct {
	Region string
}

func (e ErrRegionIndexNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: index for %s was not found", regionAddress(e.Region))
}

// ErrRegionChildIndexNotFound indicates the attempt to locate a region child
// index record failed as that region child index does not exist.
//
type ErrRegionChildIndexNotFound struct {
	Region string
}

func (e ErrRegionChildIndexNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: child index for %s was not found", regionAddress(e.Region))
}

// ErrRegionStaleVersion indicates the attempt to locate a specific version of a
// region record failed as either that region does not exist, or the specific
// version is no longer present in the store.
//
type ErrRegionStaleVersion struct {
	Region string
}

func (e ErrRegionStaleVersion) Error() string {
	return fmt.Sprintf("CloudChamber: %s has a newer version than expected", regionAddress(e.Region))
}

// ErrZoneAlreadyExists indicates the attempt to create a new zone record
// failed as that zone already exists.
//
type ErrZoneAlreadyExists struct {
	Region string
	Zone   string
}

func (e ErrZoneAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: %s already exists", zoneAddress(e.Zone, e.Region))
}

// ErrZoneNotFound indicates the attempt to locate a zone record failed as that
// zone does not exist.
//
type ErrZoneNotFound struct {
	Region string
	Zone   string
}

func (e ErrZoneNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: %s was not found", zoneAddress(e.Zone, e.Region))
}

// ErrZoneIndexNotFound indicates the attempt to locate a zone index record
// failed as that zone index does not exist.
//
type ErrZoneIndexNotFound struct {
	Region string
	Zone   string
}

func (e ErrZoneIndexNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: index for %s was not found", zoneAddress(e.Zone, e.Region))
}

// ErrZoneChildIndexNotFound indicates the attempt to locate a zone child
// index record failed as that zone child index does not exist.
//
type ErrZoneChildIndexNotFound struct {
	Region string
	Zone   string
}

func (e ErrZoneChildIndexNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: child index for %s was not found", zoneAddress(e.Zone, e.Region))
}

// ErrZoneStaleVersion indicates the attempt to locate a specific version of a
// zone record failed as either that zone does not exist, or the specific
// version is no longer present in the store.
//
type ErrZoneStaleVersion struct {
	Region string
	Zone   string
}

func (e ErrZoneStaleVersion) Error() string {
	return fmt.Sprintf("CloudChamber: %s has a newer version than expected", zoneAddress(e.Zone, e.Region))
}

// ErrRackAlreadyExists indicates the attempt to create a new rack record
// failed as that rack already exists.
//
type ErrRackAlreadyExists struct {
	Region string
	Zone   string
	Rack   string
}

func (e ErrRackAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: %s already exists", rackAddress(e.Region, e.Zone, e.Rack))
}

// ErrRackNotFound indicates the attempt to operate on a rack record failed
// as that record cannot be found.
//
type ErrRackNotFound struct {
	Region string
	Zone   string
	Rack   string
}

func (e ErrRackNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: %s was not found", rackAddress(e.Region, e.Zone, e.Rack))
}

// ErrRackIndexNotFound indicates the attempt to operate on a rack index record
// failed as that record cannot be found.
//
type ErrRackIndexNotFound struct {
	Region string
	Zone   string
	Rack   string
}

func (e ErrRackIndexNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: index for %s was not found", rackAddress(e.Region, e.Zone, e.Rack))
}

// ErrRackPduIndexNotFound indicates the attempt to operate on a rack pdu
// index record failed as that record cannot be found.
//
type ErrRackPduIndexNotFound struct {
	Region string
	Zone   string
	Rack   string
	Pdu    int64
}

func (e ErrRackPduIndexNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: pdu index for %s was not found", pduAddress(e.Region, e.Zone, e.Rack, e.Pdu))
}

// ErrRackTorIndexNotFound indicates the attempt to operate on a rack tor
// index record failed as that record cannot be found.
//
type ErrRackTorIndexNotFound struct {
	Region string
	Zone   string
	Rack   string
	Tor    int64
}

func (e ErrRackTorIndexNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: tor index for %s was not found", torAddress(e.Region, e.Zone, e.Rack, e.Tor))
}

// ErrRackBladeIndexNotFound indicates the attempt to operate on a rack blade
// index record failed as that record cannot be found.
//
type ErrRackBladeIndexNotFound struct {
	Region string
	Zone   string
	Rack   string
	Blade  int64
}

func (e ErrRackBladeIndexNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: blade index for %s was not found", bladeAddress(e.Region, e.Zone, e.Rack, e.Blade))
}

// ErrPduIndexInvalid indicates the attempt to locate a record
// failed as the given index is invalid in some way.
//
type ErrPduIndexInvalid struct {
	Region string
	Zone   string
	Rack   string
	Pdu    string
}

func (e ErrPduIndexInvalid) Error() string {
	return fmt.Sprintf("CloudChamber: %s was not valid", pduAddressName(e.Region, e.Zone, e.Rack, e.Pdu))
}

// ErrPduNotFound indicates the attempt to operate on a pdu record
// failed as that record cannot be found.
//
type ErrPduNotFound struct {
	Region string
	Zone   string
	Rack   string
	Pdu    int64
}

func (e ErrPduNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: %s was not found", pduAddress(e.Region, e.Zone, e.Rack, e.Pdu))
}

// ErrPduIndexNotFound indicates the attempt to locate a record
// failed as the given index is invalid in some way.
//
type ErrPduIndexNotFound struct {
	Region string
	Zone   string
	Rack   string
	Pdu    int64
}

func (e ErrPduIndexNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: index for %s was not found", pduAddress(e.Region, e.Zone, e.Rack, e.Pdu))
}

// ErrPduAlreadyExists indicates the attempt to create a new pdu record
// failed as that pdu already exists.
//
type ErrPduAlreadyExists struct {
	Region string
	Zone   string
	Rack   string
	Pdu    int64
}

func (e ErrPduAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: %s already exists", pduAddress(e.Region, e.Zone, e.Rack, e.Pdu))
}

// ErrTorIndexInvalid indicates the attempt to locate a record
// failed as the given index is invalid in some way.
//
type ErrTorIndexInvalid struct {
	Region string
	Zone   string
	Rack   string
	Tor    string
}

func (e ErrTorIndexInvalid) Error() string {
	return fmt.Sprintf("CloudChamber: %s was not valid", torAddressName(e.Region, e.Zone, e.Rack, e.Tor))
}

// ErrTorNotFound indicates the attempt to operate on a tor record
// failed as that record cannot be found.
//
type ErrTorNotFound struct {
	Region string
	Zone   string
	Rack   string
	Tor    int64
}

func (e ErrTorNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: %s was not found", torAddress(e.Region, e.Zone, e.Rack, e.Tor))
}

// ErrTorIndexNotFound indicates the attempt to operate on a tor record
// failed as that record cannot be found.
//
type ErrTorIndexNotFound struct {
	Region string
	Zone   string
	Rack   string
	Tor    int64
}

func (e ErrTorIndexNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: index for %s was not found", torAddress(e.Region, e.Zone, e.Rack, e.Tor))
}

// ErrTorAlreadyExists indicates the attempt to create a new zone record
// failed as that zone already exists.
//
type ErrTorAlreadyExists struct {
	Region string
	Zone   string
	Rack   string
	Tor    int64
}

func (e ErrTorAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: %s already exists", torAddress(e.Region, e.Zone, e.Rack, e.Tor))
}

// ErrBladeIndexInvalid indicates the attempt to locate a record
// failed as the given index is invalid in some way.
//
type ErrBladeIndexInvalid struct {
	Region string
	Zone   string
	Rack   string
	Blade  string
}

func (e ErrBladeIndexInvalid) Error() string {
	return fmt.Sprintf("CloudChamber: %s was not valid", bladeAddressName(e.Region, e.Zone, e.Rack, e.Blade))
}

// ErrBladeNotFound indicates the attempt to operate on a blade record
// failed as that record cannot be found.
//
type ErrBladeNotFound struct {
	Region string
	Zone   string
	Rack   string
	Blade  int64
}

func (e ErrBladeNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: %s was not found", bladeAddress(e.Region, e.Zone, e.Rack, e.Blade))
}

// ErrBladeIndexNotFound indicates the attempt to operate on a blade record
// failed as that record cannot be found.
//
type ErrBladeIndexNotFound struct {
	Region string
	Zone   string
	Rack   string
	Blade  int64
}

func (e ErrBladeIndexNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: index for %s was not found", bladeAddress(e.Region, e.Zone, e.Rack, e.Blade))
}

// ErrBladeAlreadyExists indicates the attempt to create a new blade record
// failed as that blade already exists.
//
type ErrBladeAlreadyExists struct {
	Region string
	Zone   string
	Rack   string
	Blade  int64
}

func (e ErrBladeAlreadyExists) Error() string {
	return fmt.Sprintf("CloudChamber: %s already exists", bladeAddress(e.Region, e.Zone, e.Rack, e.Blade))
}

// ErrPolicyTooLate indicates that the attempt to change the stepper policy
// failed because the policy had changed since the comparison version.  This
// is a protection against racing messages causing odd overwrites.
type ErrPolicyTooLate struct {
	Guard   int64
	Current int64
}

func (e *ErrPolicyTooLate) Error() string {
	return fmt.Sprintf(
		"the SetPolicy operation expects to replace policy version %d, "+
			"but the current policy version is %d",
		e.Guard,
		e.Current)
}

// ErrMustBeEQ signals that the specified field must be equal to a
// designated value.
type ErrMustBeEQ struct {
	Field    string
	Actual   int64
	Required int64
}

func (e ErrMustBeEQ) Error() string {
	return fmt.Sprintf(
		"the field %q must be equal to %d.  It is %d, which is invalid",
		e.Field,
		e.Required,
		e.Actual)
}

// ErrMustBeGTE signals that the specified field must be greater than or equal
// to a designated value.
type ErrMustBeGTE struct {
	Field    string
	Actual   int64
	Required int64
}

func (e ErrMustBeGTE) Error() string {
	return fmt.Sprintf(
		"the field %q must be greater than or equal to %d.  It is %d, which is invalid",
		e.Field,
		e.Required,
		e.Actual)
}

// ErrMustBeLTE signals that the specified field must be less than or equal
// to a designated value.
type ErrMustBeLTE struct {
	Field    string
	Actual   int64
	Required int64
}

func (e ErrMustBeLTE) Error() string {
	return fmt.Sprintf(
		"the field %q must be less than or equal to %d.  It is %d, which is invalid",
		e.Field,
		e.Required,
		e.Actual)
}

// ErrInvalidEnum signals that the specified field does not contain a valid
// enum value.
type ErrInvalidEnum struct {
	Field  string
	Actual int64
}

func (e ErrInvalidEnum) Error() string {
	return fmt.Sprintf(
		"the field %q does not contain a known value.  It is %d, which is invalid",
		e.Field,
		e.Actual)
}

// ErrMinLenMap signals that the specified map field does not contain at least
// the minimum required number of entries.
type ErrMinLenMap struct {
	Field    string
	Actual   int64
	Required int64
}

func (e ErrMinLenMap) Error() string {
	suffix := "s"
	if e.Required == 1 {
		suffix = ""
	}

	return fmt.Sprintf(
		"the field %q must contain at least %d element%s.  It contains %d, which is invalid",
		e.Field,
		e.Required,
		suffix,
		e.Actual)
}

// ErrMaxLenMap signals that the specified map field does not contain at least
// the minimum required number of entries.
type ErrMaxLenMap struct {
	Field  string
	Actual int64
	Limit  int64
}

func (e ErrMaxLenMap) Error() string {
	suffix := "s"
	if e.Limit == 1 {
		suffix = ""
	}

	return fmt.Sprintf(
		"the field %q must contain no more then %d element%s.  It contains %d, which is invalid",
		e.Field,
		e.Limit,
		suffix,
		e.Actual)
}

// ErrInvalidID signals that the specified tracing ID does not contain a valid
// value.
type ErrInvalidID struct {
	Field string
	Type  string
	ID    string
}

func (e ErrInvalidID) Error() string {
	return fmt.Sprintf("the field %q must be a valid %s ID.  It contains %q, which is invalid",
		e.Field,
		e.Type,
		e.ID)
}

// ErrIDMustBeEmpty signals that the specified ID contains a value when it must
// not (due to consistency rules involving other fields).
type ErrIDMustBeEmpty struct {
	Field  string
	Actual string
}

func (e ErrIDMustBeEmpty) Error() string {
	return fmt.Sprintf("the field %q must be empty.  It contains %q, which is invalid",
		e.Field,
		e.Actual)
}

// ErrMinLenString signals that the specified string does not meet a minimum length criteria.
//
type ErrMinLenString struct {
	Field    string
	Actual   int64
	Required int64
}

func (e ErrMinLenString) Error() string {
	suffix := "s"
	if e.Required == 1 {
		suffix = ""
	}

	return fmt.Sprintf(
		"the field %q must contain at least %d character%s.  It contains %d, which is invalid",
		e.Field,
		e.Required,
		suffix,
		e.Actual)
}

// ErrItemMustBeEmpty signals that the specified Item's port contains a value when it must
// not (due to consistency rules involving other fields).
//
type ErrItemMustBeEmpty struct {
	Field  string
	Item   string
	Port   int64
	Actual string
}

func (e ErrItemMustBeEmpty) Error() string {
	return fmt.Sprintf("the field %q for %q port %q must be empty.  It contains %q, which is invalid",
		e.Field,
		e.Item,
		e.Port,
		e.Actual)
}

// ErrItemMissingValue signals that the specified tracing ID does not contain any value.
//
type ErrItemMissingValue struct {
	Field string
	Item  string
	Port  int64
}

func (e ErrItemMissingValue) Error() string {
	return fmt.Sprintf("the field %q for %q port %q must have a value.",
		e.Field,
		e.Item,
		e.Port)
}

// ErrInvalidItemSelf signals that the specified item has wired a port to itself.
//
type ErrInvalidItemSelf struct {
	Field  string
	Item   string
	Port   int64
	Actual string
}

func (e ErrInvalidItemSelf) Error() string {
	return fmt.Sprintf(
		"the field %q for %q port %q must be a valid type.  It contains %q, which connects to itself",
		e.Field,
		e.Item,
		e.Port,
		e.Actual)
}

// ErrUnexpectedMessage is the standard error when an incoming request arrives in
// a state that is not expecting it.
type ErrUnexpectedMessage struct {
	Msg   string
	State string
}

func (um *ErrUnexpectedMessage) Error() string {
	return fmt.Sprintf("unexpected message %q while in state %q", um.Msg, um.State)
}

// ErrInvalidOpenAttrsType indicates that the argument passed to the
// Exporter.Open call used a type that was not expected or supported.
type ErrInvalidOpenAttrsType struct {
	Expected string
	Actual   string
}

func (e ErrInvalidOpenAttrsType) Error() string {
	return fmt.Sprintf(
		"CloudChamber: exporter open expected type %q, received type %q",
		e.Expected,
		e.Actual)
}

// ErrInvalidArgLen indicates that the number of child nodes for a rule
// operation is incorrect.
type ErrInvalidArgLen struct {
	Op       string
	Required string
	Actual   int
}

func (e ErrInvalidArgLen) Error() string {
	return fmt.Sprintf("operation %s expects %s but received %d", e.Op, e.Required, e.Actual)
}

// ErrMissingFieldName indicates that the path to child nodes is missing an
// expected path element.
type ErrMissingFieldName string

func (e ErrMissingFieldName) Error() string {
	return fmt.Sprintf(
		"key must have a table name, one or more path elements, and one field name.  "+
			"no field name was found in %q.", string(e))
}

// ErrExtraFieldNames indicates that an unexpected field name appears in the
// path to a child element.
type ErrExtraFieldNames string

func (e ErrExtraFieldNames) Error() string {
	return fmt.Sprintf(
		"key must have a table name, one or more path elements, and one field name.  "+
			"multiple possible field were names found in %q.",
		string(e))
}

// ErrMissingPath indicates that a required path value is missing from the path
// supplied for a child node.  This is typically the path element denoting a
// collection, such a 'racks' for the collection of rack elements in a zone.
type ErrMissingPath string

func (e ErrMissingPath) Error() string {
	return fmt.Sprintf(
		"key must have a table name, one or more path elements, and one field name.  "+
			"No path elements were found in %q.",
		string(e))
}

// ErrInvalidType indicates that a rule's leaf value type is invalid for the
// requested extraction operation.  For instance, a string is not a valid type
// when extraction as a boolean value is requested.
type ErrInvalidType int

func (e ErrInvalidType) Error() string {
	return fmt.Sprintf("unexpected value type %d encountered", e)
}

// ErrInvalidRuleOp indicates that an intermediate node for a rule contains an
// execution operation that is unexpected, and cannot be processed.
type ErrInvalidRuleOp int

func (e ErrInvalidRuleOp) Error() string {
	return fmt.Sprintf("unexpected operation %d encountered", e)
}

// ErrSessionNotFound indicates that no session with the supplied ID was found.
type ErrSessionNotFound int64

func (e ErrSessionNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: session %d not found", int64(e))
}

// ErrTableNameInvalid indicates the supplied region name is not one of the valid options.
//
type ErrTableNameInvalid struct {
	Name            string
	ActualTable     string
	DefinitionTable string
	ObservedTable   string
	TargetTable     string
}

func (e ErrTableNameInvalid) Error() string {
	return fmt.Sprintf(
		"CloudChamber: table name %q is not one of the valid options [%q, %q, %q, %q]",
		e.Name,
		e.ActualTable,
		e.DefinitionTable,
		e.ObservedTable,
		e.TargetTable,
	)
}

// ErrTableNameMissing indicates the supplied region name is absent or otherwise not properly specified.
//
type ErrTableNameMissing string

func (etnm ErrTableNameMissing) Error() string {
	return fmt.Sprintf("CloudChamber: table name %q is missing or not properly specified", string(etnm))
}

// ErrRegionNameMissing indicates the supplied region name is absent or otherwise not properly specified.
//
type ErrRegionNameMissing string

func (ernm ErrRegionNameMissing) Error() string {
	return fmt.Sprintf("CloudChamber: region name %q is missing or not properly specified", string(ernm))
}

// ErrZoneNameMissing indicates the supplied zone name is absent or otherwise not properly specified.
//
type ErrZoneNameMissing string

func (eznm ErrZoneNameMissing) Error() string {
	return fmt.Sprintf("CloudChamber: zone name %q is missing or not properly specified", string(eznm))
}

// ErrRackNameMissing indicates the supplied rack name is absent or otherwise not properly specified.
//
type ErrRackNameMissing string

func (ernm ErrRackNameMissing) Error() string {
	return fmt.Sprintf("CloudChamber: zone name %q is missing or not properly specified", string(ernm))
}

// ErrBladeIDInvalid indicates the supplied bladeID was out of range, either < less than 0 or greater than maxBladeID
//
type ErrBladeIDInvalid struct {
	Value int64
	Limit int64
}

func (ebii ErrBladeIDInvalid) Error() string {
	return fmt.Sprintf("CloudChamber: bladeID %d is out of range (0 to %d)", ebii.Value, ebii.Limit)
}

// ErrPduIDInvalid indicates the supplied pduID was out of range, either < less than 0 or greater than maxPduID
//
type ErrPduIDInvalid struct {
	Value int64
	Limit int64
}

func (epii ErrPduIDInvalid) Error() string {
	return fmt.Sprintf("CloudChamber: pduID %d is out of range (0 to %d)", epii.Value, epii.Limit)
}

// ErrTorIDInvalid indicates the supplied torID was out of range, either < less than 0 or greater than maxTorID
//
type ErrTorIDInvalid struct {
	Value int64
	Limit int64
}

func (etii ErrTorIDInvalid) Error() string {
	return fmt.Sprintf("CloudChamber: torID %d is out of range (0 to %d)", etii.Value, etii.Limit)
}

// ErrRootNotFound indicates the attempt to operate on the specified namespace table
// failed as that part of the namespace cannot be found.
//
type ErrRootNotFound struct {
	namespace string
}

func (ernf ErrRootNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: unable to find the root of the %q namespace", ernf.namespace)
}

// ErrIndexNotFound indicates the requested index was not found when the store
// lookup/fetch was attempted.
//
type ErrIndexNotFound string

func (einf ErrIndexNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: index %q not found", string(einf))
}

// ErrRevisionNotAvailable indicates the requested detail for the item have not
// yet been establiehed.
//
type ErrRevisionNotAvailable string

func (erna ErrRevisionNotAvailable) Error() string {
	return fmt.Sprintf("CloudChamber: %q revision not available", string(erna))
}

// ErrDetailsNotAvailable indicates the requested detail for the item have not
// yet been establiehed.
//
type ErrDetailsNotAvailable string

func (edna ErrDetailsNotAvailable) Error() string {
	return fmt.Sprintf("CloudChamber: %q details not available", string(edna))
}

// ErrPortsNotAvailable indicates the requested detail for the item have not
// yet been establiehed.
//
type ErrPortsNotAvailable string

func (epna ErrPortsNotAvailable) Error() string {
	return fmt.Sprintf("CloudChamber: %q ports not available", string(epna))
}

// ErrCapacityNotAvailable indicates the requested capacity information for
// the item has not yet been establiehed.
//
type ErrCapacityNotAvailable string

func (ecna ErrCapacityNotAvailable) Error() string {
	return fmt.Sprintf("CloudChamber: %q capacity not available", string(ecna))
}

// ErrBootInfoNotAvailable indicates the requested boot information for the item
// have not yet been establiehed.
//
type ErrBootInfoNotAvailable string

func (ebina ErrBootInfoNotAvailable) Error() string {
	return fmt.Sprintf("CloudChamber: %q boot information not available", string(ebina))
}

// ErrIndexKeyValueMismatch indicates the requested boot information for the item
// have not yet been establiehed.
//
type ErrIndexKeyValueMismatch struct {
	Namespace string
	Key       string
	Value     string
}

func (ekvm ErrIndexKeyValueMismatch) Error() string {
	return fmt.Sprintf("CloudChamber: mismatch in index key %q for returned value %q in the %q namespace", ekvm.Key, ekvm.Value, ekvm.Namespace)
}
