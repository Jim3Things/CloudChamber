package store

import (
	"context"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"google.golang.org/protobuf/runtime/protoiface"
	"github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

// KeyRoot is used to describe which part of the store namespace
// should be used for the corresponding record access.
//
type KeyRoot int

// The set of available namespace roots used by various record types
//
const (
	KeyRootStoreTest KeyRoot = iota
	KeyRootUsers

	KeyRootInventoryDefinitions
	KeyRootInventoryTargetState
	KeyRootInventoryActualState
	KeyRootInventoryObservedState
	KeyRootInventoryRepairActions

	KeyRootWorkloadDefinitions
	KeyRootWorkloadTargetState
	KeyRootWorkloadActualState
	KeyRootWorkloadObservedState
	KeyRootWorkloadRepairActions
)

const (
	namespaceRootStoreTest = "storetest"
	namespaceRootUsers     = "users"
	namespaceRootInventory = "inventory"
	namespaceRootWorkloads = "workload"

	namespaceRootInventoryDefinition     = namespaceRootInventory + "/" + "definition"
	namespaceRootInventoryTargetState    = namespaceRootInventory + "/" + "target"
	namespaceRootInventoryActualState    = namespaceRootInventory + "/" + "actual"
	namespaceRootInventoryObservedeState = namespaceRootInventory + "/" + "observed"
	namespaceRootInventoryRepairActions  = namespaceRootInventory + "/" + "repair"

	namespaceRootWorkloadDefinition      = namespaceRootWorkloads + "/" + "definition"
	namespaceRootWorkloadTargetState     = namespaceRootWorkloads + "/" + "target"
	namespaceRootWorkloadActualState     = namespaceRootWorkloads + "/" + "actual"
	namespaceRootWorkloadObservedeState  = namespaceRootWorkloads + "/" + "observed"
	namespaceRootWorkloadRepairActions   = namespaceRootWorkloads + "/" + "repair"
)

var namespaceRoots = map[KeyRoot]string{
	KeyRootStoreTest:              namespaceRootStoreTest,
	KeyRootUsers:                  namespaceRootUsers,
	KeyRootInventoryDefinitions:   namespaceRootInventoryDefinition,
	KeyRootInventoryTargetState:   namespaceRootInventoryTargetState,
	KeyRootInventoryActualState:   namespaceRootInventoryActualState,
	KeyRootInventoryObservedState: namespaceRootInventoryObservedeState,
	KeyRootInventoryRepairActions: namespaceRootInventoryRepairActions,
	KeyRootWorkloadDefinitions:    namespaceRootWorkloadDefinition,
	KeyRootWorkloadTargetState:    namespaceRootWorkloadTargetState,
	KeyRootWorkloadActualState:    namespaceRootWorkloadTargetState,
	KeyRootWorkloadObservedState:  namespaceRootWorkloadObservedeState,
	KeyRootWorkloadRepairActions:  namespaceRootWorkloadRepairActions,
}


// Action defines the signature for a function to be invoked when the
// WithAction option is used.
//
type Action func(string) error

// Options is a set of options supplied via zero or more WithXxxx() functions 
//
type Options struct {
	revision int64
	keysOnly bool
	useAsPrefix bool
	action Action
}

func (options *Options) applyOpts(optionsArray []Option) {
	for _, option := range optionsArray {
		option(options)
	}
}

// Option is the signature of the option functions used to select additional
// optional parameters on a base routine call.
// 
type Option func(*Options)

// WithRevision is an option to supply a specific revision that applies to the
// request. For example, to modify a basic Read() request to read a specific
// revision of a record.
//
func WithRevision(rev int64) Option {
	return func(options *Options) {options.revision = rev}
}

// WithPrefix is an option used to indicate the supplied name whoult be used
// as a prefix for the request. This is primarily useful for Read() and Delete()
// calls to indicate the supplied name is the root for a wildcard operation.
//
// Care should be used when applying this option on a Delete() call as a small
// error could easily lead to an entire namespace being inadvertantly deleted.
//
func WithPrefix() Option {
	return func(options *Options) {options.useAsPrefix = true}
}

// WithKeysOnly is an option applying to a Read() request to avoid reading any 
// value(s) associated with the requested set of one or more keys.
//
// This option is primarily useful when attempting to determine which keys are
// present when there is no immediate need to know the associated values. By
// restricting the amount of data being retrieved, this option may lead to an
// increase in performance and/or a reduction in consumed resources.
//
func WithKeysOnly() Option {
	return func(options *Options) {options.keysOnly = true}
}

// WithAction allows a caller to supply an action routine which is invoke
// on each record being processed within the transaction
//
func WithAction(action Action) Option {
	return func(options *Options) {options.action = action}
}



func getNamespaceRootFromKeyRoot(r KeyRoot) string {
	return namespaceRoots[r]
}

func getNamespacePrefixFromKeyRoot(r KeyRoot) string {
	return namespaceRoots[r] + "/"
}

func getKeyFromKeyRootAndName(r KeyRoot, n string) string {
	return namespaceRoots[r] + "/" + GetNormalizedName(n)
}

func getNameFromKeyRootAndKey(r KeyRoot, k string) string {
	n := strings.TrimPrefix(namespaceRoots[r] + "/", k)
	return n
}

// GetKeyFromUsername1 is a utility function to convert a supplied username to
// a store usable key for use when operating with user records.
//
func GetKeyFromUsername1(name string) string {
	return getKeyFromKeyRootAndName(KeyRootUsers, name)
}

// GetNormalizedName is a utility function to prepare a name for use when
// building a key suitable for operating with records in the store
//
func GetNormalizedName(name string) string {
	return strings.ToLower(name)
}

// Request is a struct defining the collection of values needed to make a request
// of the underlying store. Which values need to be set depend on the request.
// For example, setting any "value" for a read request is ignored.
//
type Request struct {
	Reason     string
	Records    map[string]Record
	Conditions map[string]Condition
	Actions    map[string]Action
}


/*
// Operation indicates which operation should be applied to the item in the request.
//
type Operation uint

// The set opf permissible operations on each Item within the set of items in a request
//
const (
	OpRead Operation = iota
	OpUpdate
	OpDelete
)

// Item represents a specific record with 
//
type Item struct {
	Record Record
	Condition Condition
	Operation Operation
	Action Action
}


// Request2 is a struct defining the collection of values needed to make a request
// of the underlying store. Which values need to be set depend on the request.
// For example, setting any "value" for a read request is ignored.
//
type Request2 struct {
	Reason     string
	Items      map[string]Item
}
 */


// Response is a struct defining the set of values returned from a request.
//
type Response struct {
	Revision int64
	Records  map[string]Record
}

// Encode is a default protobuf defined message to JSON encoded string encoder
//
func Encode(m proto.Message) (s string, err error) {
	p := jsonpb.Marshaler{}

	if s, err = p.MarshalToString(m); err != nil {
		return "", err
	}

	return s, nil
}

// Decode is a default JSON encoded string to protobuf defined message decoder
//
func Decode(s string, m protoiface.MessageV1) error {
	if err := jsonpb.Unmarshal(strings.NewReader(s), m); err != nil {
		return err
	}

	return nil
}

// CreateWithEncode is a function to create a single key, value record pair
//
func (store *Store) CreateWithEncode(
	ctx context.Context,
	r KeyRoot,
	n string,
	m protoiface.MessageV1) (revision int64, err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	prefix := getNamespacePrefixFromKeyRoot(r)

	tracing.Info(ctx, "Request to create new %q under prefix %q", n, prefix)

	if err = store.disconnected(ctx); err != nil {
		return RevisionInvalid, err
	}

	v, err := Encode(m)

	if err != nil {
		return RevisionInvalid, err
	}

	request := &Request{
		Records:    make(map[string]Record),
		Conditions: make(map[string]Condition),
	}

	k := getKeyFromKeyRootAndName(r, n)
	request.Records[k] = Record{Revision: RevisionInvalid, Value: v}
	request.Conditions[k] = ConditionCreate

	resp, err := store.WriteTxn(ctx, request)

	// Need to strip the namespace prefix and return something described
	// in terms the caller should recognize
	//
	if err == ErrStoreAlreadyExists(k) {
		return RevisionInvalid, ErrStoreAlreadyExists(n)
	}

	if err != nil {
		return RevisionInvalid, err
	}

	tracing.Info(ctx, "Created record for %q under prefix %q with revision %v", n, prefix, resp.Revision)

	return resp.Revision, nil
}

// Create is a function to create a single key, value record pair
//
func (store *Store) Create(ctx context.Context, r KeyRoot, n string, v string) (revision int64, err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	prefix := getNamespacePrefixFromKeyRoot(r)

	tracing.Info(ctx, "Request to create new %q under prefix %q", n, prefix)

	if err = store.disconnected(ctx); err != nil {
		return RevisionInvalid, err
	}

	request := &Request{
		Records:    make(map[string]Record),
		Conditions: make(map[string]Condition),
	}

	k := getKeyFromKeyRootAndName(r, n)
	request.Records[k] = Record{Revision: RevisionInvalid, Value: v}
	request.Conditions[k] = ConditionCreate

	resp, err := store.WriteTxn(ctx, request)

	// Need to strip the namespace prefix and return something described
	// in terms the caller should recognize
	//
	if err == ErrStoreAlreadyExists(k) {
		return RevisionInvalid, ErrStoreAlreadyExists(n)
	}

	if err != nil {
		return RevisionInvalid, err
	}

	tracing.Info(ctx, "Created record for %q under prefix %q with revision %v", n, prefix, resp.Revision)

	revision = resp.Revision

	return resp.Revision, nil
}

// CreateMultiple is a function to create a set of related key, value pairs within a single operation (txn)
//
func (store *Store) CreateMultiple(ctx context.Context, r KeyRoot, kvs *map[string]string) (revision int64, err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(clients.EnsureTickInContext))
	defer span.End()

	prefix := getNamespacePrefixFromKeyRoot(r)

	tracing.Info(ctx, "Request to create new key set under prefix %q", prefix)

	if err = store.disconnected(ctx); err != nil {
		return RevisionInvalid, err
	}

	request := &Request{
		Records:    make(map[string]Record),
		Conditions: make(map[string]Condition),
	}

	for n, v := range *kvs {
		k := getKeyFromKeyRootAndName(r, n)
		request.Records[k] = Record{Revision: RevisionInvalid, Value: v}
		request.Conditions[k] = ConditionCreate
	}

	resp, err := store.WriteTxn(ctx, request)

	// Need to strip the namespace prefix and return something described
	// in terms the caller should recognize
	//
	if err != nil {
		for k := range request.Records {
			if err == ErrStoreAlreadyExists(k) {
				n := getNameFromKeyRootAndKey(r, k)
				return RevisionInvalid, ErrStoreAlreadyExists(n)
			}
		}
	}

	if err != nil {
		return RevisionInvalid, err
	}

	tracing.Info(ctx, "Created record set under prefix %q with revision %v", prefix, resp.Revision)

	revision = resp.Revision

	return resp.Revision, nil
}

// ReadWithDecode is a method to retrieve the user record associated with the
// supplied name, deal with any store related key prefixes, and decode
// the json encoded record into something the caller understands.
//
// NOTE: A future enhancement may occur where the caller supplies an action
//       routine to take care of the decoding of the record itself allowing
//       this layer to only have to deal with the manipulation of the store,
//       the keys used to persist the callers records but not have to worry
//       about the encode/decode formats or indeed the target record itself.
//
func (store *Store) ReadWithDecode(
	ctx context.Context,
	kr KeyRoot,
	n string,
	m protoiface.MessageV1) (revision int64, err error) {
	revision = RevisionInvalid

	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	prefix := getNamespacePrefixFromKeyRoot(kr)

	tracing.Info(ctx, "Request to read and decode %q under prefix %q", n, prefix)

	if err = store.disconnected(ctx); err != nil {
		return RevisionInvalid, err
	}

	var (
		rev      int64
		val      string
		response *Response
	)

	// If we need to do the read to get the revision, we will need an array of the keys
	//
	request := &Request{
		Records:    make(map[string]Record),
		Conditions: make(map[string]Condition),
	}

	k := getKeyFromKeyRootAndName(kr, n)
	request.Records[k] = Record{Revision: RevisionInvalid}
	request.Conditions[k] = ConditionUnconditional

	if response, err = store.ReadTxn(ctx, request); err != nil {
		return RevisionInvalid, err
	}

	recordCount := len(response.Records)

	switch recordCount {
	default:
		return RevisionInvalid, ErrStoreBadRecordCount{n, 1, recordCount}

	case 0:
		return RevisionInvalid, ErrStoreKeyNotFound(n)

	case 1:
		rev = response.Records[k].Revision
		val = response.Records[k].Value

		if err = Decode(val, m); err != nil {
			return RevisionInvalid, err
		}

		tracing.Info(ctx, "found and decoded record for %q under prefix %q with revision %v and value %q", n, prefix, rev, val)

		return rev, nil
	}
}

// Read is a method to retrieve the user record associated with the
// supplied name, deal with any store related key prefixes, and decode
// the json encoded record into something the caller understands.
//
// NOTE: A future enhancement may occur where the caller supplies an action
//       routine to take care of the decoding of the record itself allowing
//       this layer to only have to deal with the manipulation of the store,
//       the keys used to persist the callers records but not have to worry
//       about the encode/decode formats or indeed the target record itself.
//
func (store *Store) Read(ctx context.Context, kr KeyRoot, n string) (value *string, revision int64, err error) {
	revision = RevisionInvalid

	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	prefix := getNamespacePrefixFromKeyRoot(kr)

	tracing.Info(ctx, "Request to read value of %q under prefix %q", n, prefix)

	if err = store.disconnected(ctx); err != nil {
		return nil, RevisionInvalid, err
	}

	var (
		rev      int64
		val      string
		response *Response
	)

	// If we need to do the read to get the revision, we will need an array of the keys
	//
	request := &Request{
		Records:    make(map[string]Record),
		Conditions: make(map[string]Condition),
	}

	k := getKeyFromKeyRootAndName(kr, n)
	request.Records[k] = Record{Revision: RevisionInvalid}
	request.Conditions[k] = ConditionUnconditional

	if response, err = store.ReadTxn(ctx, request); err != nil {
		return nil, RevisionInvalid, err
	}

	recordCount := len(response.Records)

	switch recordCount {
	default:
		return nil, RevisionInvalid, ErrStoreBadRecordCount{n, 1, recordCount}

	case 0:
		return nil, RevisionInvalid, ErrStoreKeyNotFound(n)

	case 1:
		rev = response.Records[k].Revision
		val = response.Records[k].Value
		tracing.Info(ctx, "found record for %q under prefix %q, with revision %v and value %q", n, prefix, rev, val)

		revision = rev
		value = &val

		return value, revision, err
	}
}

// Update is a function to conditionally update a value for a single key
//
func (store *Store) Update(ctx context.Context, r KeyRoot, n string, rev int64,	v string) (revision int64, err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(clients.EnsureTickInContext))
	defer span.End()

	prefix := getNamespacePrefixFromKeyRoot(r)
	k      := getKeyFromKeyRootAndName(r, n)

	tracing.Info(ctx, "Request to update %q under prefix %q", n, prefix)

	if err = store.disconnected(ctx); err != nil {
		return RevisionInvalid, err
	}

	var condition Condition

	switch {
	case rev == RevisionInvalid:
		condition = ConditionUnconditional

	default:
		condition = ConditionRevisionEqual
	}

	request := &Request{
		Records:    make(map[string]Record),
		Conditions: make(map[string]Condition)}

	request.Records[k] = Record{Revision: rev, Value: v}
	request.Conditions[k] = condition

	resp, err := store.WriteTxn(ctx, request)

	if err != nil {
		return RevisionInvalid, err
	}

	tracing.Info(ctx,
		"Updated record %q under prefix %q from revision %v to revision %v",
		n, prefix, rev, resp.Revision)

	return resp.Revision, nil
}

// UpdateWithEncode is a function to conditionally update a value for a single key
//
func (store *Store) UpdateWithEncode(
	ctx context.Context,
	kr KeyRoot,
	n string,
	rev int64,
	m protoiface.MessageV1) (revision int64, err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	prefix := getNamespacePrefixFromKeyRoot(kr)

	tracing.Info(ctx, "Request to update %q under prefix %q", n, prefix)

	if err = store.disconnected(ctx); err != nil {
		return RevisionInvalid, err
	}

	v, err := Encode(m)
	if err != nil {
		return RevisionInvalid, err
	}

	var condition Condition

	switch {
	case rev == RevisionInvalid:
		condition = ConditionUnconditional

	default:
		condition = ConditionRevisionEqual
	}

	request := &Request{
		Records:    make(map[string]Record),
		Conditions: make(map[string]Condition)}

	k := getKeyFromKeyRootAndName(kr, n)
	request.Records[k] = Record{Revision: rev, Value: v}
	request.Conditions[k] = condition

	resp, err := store.WriteTxn(ctx, request)

	if err != nil {
		return RevisionInvalid, err
	}

	tracing.Info(ctx,
		"Updated record %q under prefix %q from revision %v to revision %v",
		n, prefix, rev, resp.Revision)

	return resp.Revision, nil
}

// Delete is a function to delete a single key, value record pair
//
func (store *Store) Delete(ctx context.Context, r KeyRoot, n string, rev int64) (revision int64, err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	prefix := getNamespacePrefixFromKeyRoot(r)

	tracing.Info(ctx, "Request to delete %q under prefix %q", n, prefix)

	if err = store.disconnected(ctx); err != nil {
		return RevisionInvalid, err
	}

	var condition Condition

	switch {
	case rev == RevisionInvalid:
		condition = ConditionUnconditional

	default:
		condition = ConditionRevisionEqual
	}

	request := &Request{
		Records:    make(map[string]Record),
		Conditions: make(map[string]Condition),
	}

	k := getKeyFromKeyRootAndName(r, n)
	request.Records[k] = Record{Revision: rev}
	request.Conditions[k] = condition

	resp, err := store.DeleteTxn(ctx, request)

	// Need to strip the namespace prefix and return something described
	// in terms the caller should recognize
	//
	if err == ErrStoreKeyNotFound(k) {
		return RevisionInvalid, ErrStoreKeyNotFound(n)
	}

	if err != nil {
		return RevisionInvalid, err
	}

	tracing.Info(ctx, "Deleted record for %q under prefix %q with revision %v resulting in store revision %v", n, prefix, rev, resp.Revision)

	return resp.Revision, nil
}

// List is a method to return all the user records using a single call.
//
// NOTE: The returned set of records may exist but contain no records.
//
// NOTE: This should only be used at present if the number of user records
//       is limited as there is a limit to the number of records that can
//       be fetched from the store in a single shot. Eventually this will
//       be updated to use an "interrupted" enum style call to allow for
//		 an essentially infinite number of records.
//
func (store *Store) List(ctx context.Context, r KeyRoot) (records *map[string]Record, revision int64, err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	prefix := getNamespacePrefixFromKeyRoot(r)

	tracing.Info(ctx, "Request to list keys under prefix %q", prefix)

	if err = store.disconnected(ctx); err != nil {
		return nil, RevisionInvalid, err
	}

	response, err := store.ListWithPrefix(ctx, prefix)

	if err != nil {
		return nil, RevisionInvalid, err
	}

	recs := make(map[string]Record, len(response.Records))

	for k, record := range response.Records {

		if !strings.HasPrefix(k, prefix) {
			return nil, RevisionInvalid, ErrStoreBadRecordKey(k)
		}

		name := strings.TrimPrefix(k, prefix)

		recs[name] = Record{Revision: record.Revision, Value: record.Value}

		tracing.Info(ctx, "found record with key %q for name %q with revision %v", k, name, record.Revision)
	}

	tracing.Info(ctx, "returned %v records at store revision %v", len(response.Records), response.Revision)

	return &recs, response.Revision, nil
}
