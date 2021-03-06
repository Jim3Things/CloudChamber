package store

import (
	"context"
	"regexp"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/namespace"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"

	"google.golang.org/protobuf/runtime/protoiface"
)

// Action defines the signature for a function to be invoked when the
// WithAction option is used.
//
type Action func(string) error

// Options is a set of options supplied via zero or more WithXxxx() functions
//
type Options struct {
	revision    int64
	keysOnly    bool
	useAsPrefix bool
	action      Action
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
	return func(options *Options) { options.revision = rev }
}

// WithPrefix is an option used to indicate the supplied name should be used
// as a prefix for the request. This is primarily useful for Read() and Delete()
// calls to indicate the supplied name is the root for a wildcard operation.
//
// Care should be used when applying this option on a Delete() call as a small
// error could easily lead to an entire namespace being inadvertently deleted.
//
func WithPrefix() Option {
	return func(options *Options) { options.useAsPrefix = true }
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
	return func(options *Options) { options.keysOnly = true }
}

// WithAction allows a caller to supply an action routine which is invoke
// on each record being processed within the transaction
//
func WithAction(action Action) Option {
	return func(options *Options) { options.action = action }
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

// // Operation indicates which operation should be applied to the item in the request.
// //
// type Operation uint

// // The set of permissible operations on each Item within the set of items in a request
// //
// const (
// 	OpRead Operation = iota
// 	OpUpdate
// 	OpDelete
// )

// // Item represents a specific record with
// //
// type Item struct {
// 	Record Record
// 	Condition Condition
// 	Operation Operation
// 	Action Action
// }

// // Request2 is a struct defining the collection of values needed to make a request
// // of the underlying store. Which values need to be set depend on the request.
// // For example, setting any "value" for a read request is ignored.
// //
// type Request2 struct {
// 	Reason     string
// 	Items      map[string]Item
// }

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
	r namespace.KeyRoot,
	n string,
	m protoiface.MessageV1) (revision int64, err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	prefix := namespace.GetNamespacePrefixFromKeyRoot(r)

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

	k := namespace.GetKeyFromKeyRootAndName(r, n)
	request.Records[k] = Record{Revision: RevisionInvalid, Value: v}
	request.Conditions[k] = ConditionCreate

	resp, err := store.WriteTxn(ctx, request)

	// Need to strip the namespace prefix and return something described
	// in terms the caller should recognize
	//
	if err == errors.ErrStoreAlreadyExists(k) {
		return RevisionInvalid, errors.ErrStoreAlreadyExists(n)
	}

	if err != nil {
		return RevisionInvalid, err
	}

	tracing.Info(ctx, "Created record for %q under prefix %q with revision %v", n, prefix, resp.Revision)

	return resp.Revision, nil
}

// Create is a function to create a single key, value record pair
//
func (store *Store) Create(ctx context.Context, r namespace.KeyRoot, n string, v string) (revision int64, err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	prefix := namespace.GetNamespacePrefixFromKeyRoot(r)

	tracing.Info(ctx, "Request to create new %q under prefix %q", n, prefix)

	if err = store.disconnected(ctx); err != nil {
		return RevisionInvalid, err
	}

	request := &Request{
		Records:    make(map[string]Record),
		Conditions: make(map[string]Condition),
	}

	k := namespace.GetKeyFromKeyRootAndName(r, n)
	request.Records[k] = Record{Revision: RevisionInvalid, Value: v}
	request.Conditions[k] = ConditionCreate

	resp, err := store.WriteTxn(ctx, request)

	// Need to strip the namespace prefix and return something described
	// in terms the caller should recognize
	//
	if err == errors.ErrStoreAlreadyExists(k) {
		return RevisionInvalid, errors.ErrStoreAlreadyExists(n)
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
func (store *Store) CreateMultiple(ctx context.Context, r namespace.KeyRoot, kvs *map[string]string) (revision int64, err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	prefix := namespace.GetNamespacePrefixFromKeyRoot(r)

	tracing.Info(ctx, "Request to create new key set under prefix %q", prefix)

	if err = store.disconnected(ctx); err != nil {
		return RevisionInvalid, err
	}

	request := &Request{
		Records:    make(map[string]Record),
		Conditions: make(map[string]Condition),
	}

	for n, v := range *kvs {
		k := namespace.GetKeyFromKeyRootAndName(r, n)
		request.Records[k] = Record{Revision: RevisionInvalid, Value: v}
		request.Conditions[k] = ConditionCreate
	}

	resp, err := store.WriteTxn(ctx, request)

	if err != nil {
		// Need to strip the namespace prefix and return something described
		// in terms the caller should recognize
		//
		for k := range request.Records {
			if err == errors.ErrStoreAlreadyExists(k) {
				n := namespace.GetNameFromKeyRootAndKey(r, k)
				return RevisionInvalid, errors.ErrStoreAlreadyExists(n)
			}
		}

		// Nothing more appropriate found so just return what we have.
		//
		return RevisionInvalid, err
	}

	tracing.Info(ctx, "Created record set under prefix %q with revision %d", prefix, resp.Revision)

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
	kr namespace.KeyRoot,
	n string,
	m protoiface.MessageV1) (revision int64, err error) {
	revision = RevisionInvalid

	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	prefix := namespace.GetNamespacePrefixFromKeyRoot(kr)

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

	k := namespace.GetKeyFromKeyRootAndName(kr, n)
	request.Records[k] = Record{Revision: RevisionInvalid}
	request.Conditions[k] = ConditionUnconditional

	if response, err = store.ReadTxn(ctx, request); err != nil {
		return RevisionInvalid, err
	}

	recordCount := len(response.Records)

	switch recordCount {
	default:
		return RevisionInvalid, errors.ErrStoreBadRecordCount{Key: n, Expected: 1, Actual: recordCount}

	case 0:
		return RevisionInvalid, errors.ErrStoreKeyNotFound(n)

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
func (store *Store) Read(ctx context.Context, kr namespace.KeyRoot, n string) (value *string, revision int64, err error) {
	revision = RevisionInvalid
	prefix := namespace.GetNamespacePrefixFromKeyRoot(kr)
	k := namespace.GetKeyFromKeyRootAndName(kr, n)

	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Read value from namespace %s with prefix %s for key %s", n, prefix, k),
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

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

	request.Records[k] = Record{Revision: RevisionInvalid}
	request.Conditions[k] = ConditionUnconditional

	if response, err = store.ReadTxn(ctx, request); err != nil {
		return nil, RevisionInvalid, err
	}

	recordCount := len(response.Records)

	switch recordCount {
	default:
		return nil, RevisionInvalid, errors.ErrStoreBadRecordCount{Key: n, Expected: 1, Actual: recordCount}

	case 0:
		return nil, RevisionInvalid, errors.ErrStoreKeyNotFound(n)

	case 1:
		rev = response.Records[k].Revision
		val = response.Records[k].Value
		tracing.Info(
			ctx,
			tracing.WithReplacement(
				regexp.MustCompile(
					`passwordHash\\\"\:(.*?),\\\"`),
				`passwordHash\":\"...REDACTED...\",\"`),
			"found record with revision %v and value %q",
			rev,
			val)

		revision = rev
		value = &val

		return value, revision, err
	}
}

// Update is a function to conditionally update a value for a single key
//
func (store *Store) Update(ctx context.Context, r namespace.KeyRoot, n string, rev int64, v string) (revision int64, err error) {
	prefix := namespace.GetNamespacePrefixFromKeyRoot(r)
	k := namespace.GetKeyFromKeyRootAndName(r, n)

	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Request to update value in namespace %s with prefix %s for key %s", n, prefix, k),
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

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

	tracing.UpdateSpanName(ctx,
		"Updated value in namespace %s with prefix %s for key %s from revision %d to revision %d",
		n,
		prefix,
		k,
		rev,
		resp.Revision)

	return resp.Revision, nil
}

// UpdateWithEncode is a function to conditionally update a value for a single key
//
func (store *Store) UpdateWithEncode(
	ctx context.Context,
	kr namespace.KeyRoot,
	n string,
	rev int64,
	m protoiface.MessageV1) (revision int64, err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	prefix := namespace.GetNamespacePrefixFromKeyRoot(kr)

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

	k := namespace.GetKeyFromKeyRootAndName(kr, n)
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
func (store *Store) Delete(ctx context.Context, r namespace.KeyRoot, n string, rev int64) (revision int64, err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	prefix := namespace.GetNamespacePrefixFromKeyRoot(r)

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

	k := namespace.GetKeyFromKeyRootAndName(r, n)
	request.Records[k] = Record{Revision: rev}
	request.Conditions[k] = condition

	resp, err := store.DeleteTxn(ctx, request)

	// Need to strip the namespace prefix and return something described
	// in terms the caller should recognize
	//
	if err == errors.ErrStoreKeyNotFound(k) {
		return RevisionInvalid, errors.ErrStoreKeyNotFound(n)
	}

	if err != nil {
		return RevisionInvalid, err
	}

	tracing.Info(ctx, "Deleted record for %q under prefix %q with revision %v resulting in store revision %v", n, prefix, rev, resp.Revision)

	return resp.Revision, nil
}

// DeleteMultiple is a function to delete a set of related key, revision pairs
// within a single operation (txn)
//
func (store *Store) DeleteMultiple(ctx context.Context, r namespace.KeyRoot, kvs *map[string]int64) (int64, error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	prefix := namespace.GetNamespacePrefixFromKeyRoot(r)

	tracing.Info(ctx, "Request to delete key set under prefix %q", prefix)

	if err := store.disconnected(ctx); err != nil {
		return RevisionInvalid, err
	}

	request := &Request{
		Records:    make(map[string]Record, len(*kvs)),
		Conditions: make(map[string]Condition, len(*kvs)),
	}

	for n, rev := range *kvs {
		k := namespace.GetKeyFromKeyRootAndName(r, n)

		condition := ConditionRevisionEqual

		if rev == RevisionInvalid {
			condition = ConditionUnconditional
		}

		request.Records[k] = Record{Revision: rev}
		request.Conditions[k] = condition
	}

	resp, err := store.DeleteTxn(ctx, request)

	if err != nil {
		// Need to strip the namespace prefix and return something described
		// in terms the caller should recognize
		//
		for k := range request.Records {
			if err == errors.ErrStoreKeyNotFound(k) {
				n := namespace.GetNameFromKeyRootAndKey(r, k)
				return RevisionInvalid, errors.ErrStoreKeyNotFound(n)
			}
		}

		// Nothing more appropriate found so just return what we have.
		//
		return RevisionInvalid, err
	}

	tracing.Info(ctx, "Deleted record set under prefix %q with revision %d", prefix, resp.Revision)

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
func (store *Store) List(ctx context.Context, r namespace.KeyRoot, n string) (records *map[string]Record, revision int64, err error) {
	prefix := namespace.GetNamespacePrefixFromKeyRoot(r)
	k := namespace.GetKeyFromKeyRootAndName(r, n)

	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Request to list entries from namespace %s with prefix %s under key %s", n, prefix, k),
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	if err = store.disconnected(ctx); err != nil {
		return nil, RevisionInvalid, err
	}

	response, err := store.ListWithPrefix(ctx, k)

	if err != nil {
		return nil, RevisionInvalid, err
	}

	recs := make(map[string]Record, len(response.Records))

	for k, record := range response.Records {

		if !strings.HasPrefix(k, prefix) {
			return nil, RevisionInvalid, errors.ErrStoreBadRecordKey(k)
		}

		name := strings.TrimPrefix(k, prefix)

		recs[name] = Record{Revision: record.Revision, Value: record.Value}

		tracing.Info(ctx, "found record with key %q for name %q with revision %v", k, name, record.Revision)
	}

	tracing.UpdateSpanName(ctx,
		"Listed %d entries in namespace %s with prefix %s for key %s at store revision %d",
		len(response.Records),
		n,
		prefix,
		k,
		response.Revision)

	return &recs, response.Revision, nil
}

// Watch is a method use to establish a watch point on a portion of
// the namespace identified by the supplied prefix name.
//
func (store *Store) Watch(ctx context.Context, r namespace.KeyRoot, n string) (*Watch, error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	prefix := namespace.GetNamespacePrefixFromKeyRoot(r)

	tracing.UpdateSpanName(ctx, "Request to list keys under prefix %q", prefix)

	k := namespace.GetKeyFromKeyRootAndName(r, n)

	// err = store.SetWatchWithPrefix(ctx, k)
	resp, err := store.SetWatchWithPrefix(ctx, k)

	if err != nil {
		return nil, err
	}

	notifications := make(chan WatchEvent)

	go func() {
		for ev := range resp.Events {
			notifications <- WatchEvent{
				Type:     ev.Type,
				Revision: ev.Revision,
				Key:      namespace.GetNameFromKeyRootAndKey(r, ev.Key),
				NewRev:   ev.NewRev,
				OldRev:   ev.OldRev,
				NewVal:   ev.NewVal,
				OldVal:   ev.OldVal,
			}
		}

		close(notifications)
	}()

	response := &Watch{
		key:    resp.key,
		cancel: resp.cancel,
		Events: notifications,
	}

	return response, nil
}

// Close is used to close the upstream event notification channel.
//
func (w *Watch) Close(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext),
		tracing.WithName("Attempting to close event channel"))
	defer span.End()

	cancel := w.cancel

	if cancel == nil {
		tracing.Debug(ctx, "Second (or subsequent) attempt to close event channel for key %q", w.key)

		return errors.ErrAlreadyClosed{
			Type: "watch event channel",
			Name: w.key,
		}
	}

	w.cancel = nil

	cancel()

	tracing.UpdateSpanName(ctx, "Closed notification channel for key %q", w.key)

	return nil
}
