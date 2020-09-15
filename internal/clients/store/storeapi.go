package store

import (
	"context"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"google.golang.org/protobuf/runtime/protoiface"

	clients "github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

// KeyRoot is used to describe which part of the store namespace
// should be used for the corresponding record access.
//
type KeyRoot int

// The set of available namespace roots used by various record types
//
const (
	KeyRootUsers KeyRoot = iota
	KeyRootBlades
	KeyRootRacks
	KeyRootWorkloads
	KeyRootStoreTest
)
const (
	namespaceRootUsers     = "users"
	namespaceRootRacks     = "racks"
	namespaceRootBlades    = "blades"
	namespaceRootWorkloads = "workloads"
	namespaceRootStoreTest = "storetest"
)

var namespaceRoots = map[KeyRoot]string{
	KeyRootUsers:     namespaceRootUsers,
	KeyRootRacks:     namespaceRootBlades,
	KeyRootBlades:    namespaceRootBlades,
	KeyRootWorkloads: namespaceRootWorkloads,
	KeyRootStoreTest: namespaceRootStoreTest,
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
}

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
		tracing.WithContextValue(clients.EnsureTickInContext))
	defer span.End()

	prefix := getNamespacePrefixFromKeyRoot(r)

	tracing.Infof(ctx, "Request to create new %q under prefix %q", n, prefix)

	if err = store.disconnected(ctx); err != nil {
		return 0, err
	}

	v, err := Encode(m)

	if err != nil {
		return 0, err
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
		return 0, ErrStoreAlreadyExists(n)
	}

	if err != nil {
		return 0, err
	}

	tracing.Infof(ctx, "Created record for %q under prefix %q with revision %v", n, prefix, resp.Revision)

	return resp.Revision, nil
}

// Create is a function to create a single key, value record pair
//
func (store *Store) Create(ctx context.Context, r KeyRoot, n string, v string) (revision int64, err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(clients.EnsureTickInContext))
	defer span.End()

	prefix := getNamespacePrefixFromKeyRoot(r)

	tracing.Infof(ctx, "Request to create new %q under prefix %q", n, prefix)

	if err = store.disconnected(ctx); err != nil {
		return 0, err
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
		return 0, ErrStoreAlreadyExists(n)
	}

	if err != nil {
		return 0, err
	}

	tracing.Infof(ctx, "Created record for %q under prefix %q with revision %v", n, prefix, resp.Revision)

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
		tracing.WithContextValue(clients.EnsureTickInContext))
	defer span.End()

	prefix := getNamespacePrefixFromKeyRoot(kr)

	tracing.Infof(ctx, "Request to read and decode %q under prefix %q", n, prefix)

	if err = store.disconnected(ctx); err != nil {
		return 0, err
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
		return 0, err
	}

	recordCount := len(response.Records)

	switch recordCount {
	default:
		return 0, ErrStoreBadRecordCount{n, 1, recordCount}

	case 0:
		return 0, ErrStoreKeyNotFound(n)

	case 1:
		rev = response.Records[k].Revision
		val = response.Records[k].Value

		if err = Decode(val, m); err != nil {
			return 0, err
		}

		tracing.Infof(ctx, "found and decoded record for %q under prefix %q with revision %v and value %q", n, prefix, rev, val)

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
		tracing.WithContextValue(clients.EnsureTickInContext))
	defer span.End()

	prefix := getNamespacePrefixFromKeyRoot(kr)

	tracing.Infof(ctx, "Request to read value of %q under prefix %q", n, prefix)

	if err = store.disconnected(ctx); err != nil {
		return nil, 0, err
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
		return nil, 0, err
	}

	recordCount := len(response.Records)

	switch recordCount {
	default:
		return nil, 0, ErrStoreBadRecordCount{n, 1, recordCount}

	case 0:
		return nil, 0, ErrStoreKeyNotFound(n)

	case 1:
		rev = response.Records[k].Revision
		val = response.Records[k].Value
		tracing.Infof(ctx, "found record for %q under prefix %q, with revision %v and value %q", n, prefix, rev, val)

		revision = rev
		value = &val

		return value, revision, err
	}
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
		tracing.WithContextValue(clients.EnsureTickInContext))
	defer span.End()

	prefix := getNamespacePrefixFromKeyRoot(kr)

	tracing.Infof(ctx, "Request to update %q under prefix %q", n, prefix)

	if err = store.disconnected(ctx); err != nil {
		return 0, err
	}

	v, err := Encode(m)
	if err != nil {
		return 0, err
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
		return 0, err
	}

	tracing.Infof(ctx, "Updated record %q under prefix %q from revision %v to revision %v", n, prefix, rev, resp.Revision)

	return resp.Revision, nil
}

// Delete is a function to delete a single key, value record pair
//
func (store *Store) Delete(ctx context.Context, r KeyRoot, n string, rev int64) (revision int64, err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(clients.EnsureTickInContext))
	defer span.End()

	prefix := getNamespacePrefixFromKeyRoot(r)

	tracing.Infof(ctx, "Request to delete %q under prefix %q", n, prefix)

	if err = store.disconnected(ctx); err != nil {
		return 0, err
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
		return 0, ErrStoreKeyNotFound(n)
	}

	if err != nil {
		return 0, err
	}

	tracing.Infof(ctx, "Deleted record for %q under prefix %q with revision %v resulting in store revision %v", n, prefix, rev, resp.Revision)

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
		tracing.WithContextValue(clients.EnsureTickInContext))
	defer span.End()

	prefix := getNamespacePrefixFromKeyRoot(r)

	tracing.Infof(ctx, "Request to list keys under prefix %q", prefix)

	if err = store.disconnected(ctx); err != nil {
		return nil, 0, err
	}

	response, err := store.ListWithPrefix(ctx, prefix)

	if err != nil {
		return nil, 0, err
	}

	recs := make(map[string]Record, len(response.Records))

	for k, record := range response.Records {

		if !strings.HasPrefix(k, prefix) {
			return nil, 0, ErrStoreBadRecordKey(k)
		}

		name := strings.TrimPrefix(k, prefix)

		recs[name] = Record{Revision: record.Revision, Value: record.Value}

		tracing.Infof(ctx, "found record with key %q for name %q with revision %v", k, name, record.Revision)
	}

	tracing.Infof(ctx, "returned %v records at store revision %v", len(response.Records), response.Revision)

	return &recs, response.Revision, nil
}
