package store

import (
	"context"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
	st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/admin"

	"google.golang.org/protobuf/runtime/protoiface"
)

// KeyRoot is used to describe which part of the store namespace
// should be used for the corresponding record access.
//
type KeyRoot int

// The set of avalable namespace roots used by vcarious record types
//
const (
	KeyRootUsers KeyRoot = iota
	KeyRootBlades
	KeyRootRacks
	KeyRootWorkloads
)
const (
	namespaceRootUsers     = "users/"
	namespaceRootRacks     = "racks/"
	namespaceRootBlades    = "blades/"
	namespaceRootWorkloads = "workloads/"
)

var namespaceRoots = map[KeyRoot]string{
	KeyRootUsers:     namespaceRootUsers,
	KeyRootRacks:     namespaceRootBlades,
	KeyRootBlades:    namespaceRootBlades,
	KeyRootWorkloads: namespaceRootWorkloads,
}

func getNamespaceRootFromKeyRoot(r KeyRoot) string {
	return namespaceRoots[r]
}

func getKeyFromKeyRootAndName(r KeyRoot, n string) string {
	return namespaceRoots[r] + GetNormalizedName(n)
}

// GetKeyFromUsername is a utility function to convert a supplied username to
// a store usable key for use when operating with user records.
//
func GetKeyFromUsername(name string) string {
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

// CreateNew is a function to create a single key, value record pair
//
func (store *Store) CreateNew(ctx context.Context, r KeyRoot, n string, m protoiface.MessageV1) (revision int64, err error) {

	err = st.WithSpan(ctx, tracing.MethodName(1), func(ctx context.Context) (err error) {

		if err = store.disconnected(ctx); err != nil {
			return err
		}

		v, err := Encode(m)

		if err != nil {
			return err
		}

		request := &Request{Records: make(map[string]Record), Conditions: make(map[string]Condition)}

		k := getKeyFromKeyRootAndName(r, n)
		request.Records[k] = Record{Revision: RevisionInvalid, Value: v}
		request.Conditions[k] = ConditionCreate

		resp, err := store.WriteTxn(ctx, request)

		// Need to strip the namespace prefix and return something described
		// in terms the caller should recognize
		//
		if err == ErrStoreRecordExists(k) {
			return ErrStoreRecordExists(n)
		}

		if err != nil {
			return err
		}

		st.Infof(ctx, -1, "Created record for %q with revision %v", n, resp.Revision)

		revision = resp.Revision

		return nil
	})

	return revision, err
}

// CreateNewValue is a function to create a single key, value record pair
//
func (store *Store) CreateNewValue(ctx context.Context, r KeyRoot, n string, v string) (revision int64, err error) {

	err = st.WithSpan(ctx, tracing.MethodName(1), func(ctx context.Context) (err error) {

		if err = store.disconnected(ctx); err != nil {
			return err
		}

		request := &Request{Records: make(map[string]Record), Conditions: make(map[string]Condition)}

		k := getKeyFromKeyRootAndName(r, n)
		request.Records[k] = Record{Revision: RevisionInvalid, Value: v}
		request.Conditions[k] = ConditionCreate

		resp, err := store.WriteTxn(ctx, request)

		// Need to strip the namespace prefix and return something described
		// in terms the caller should recognize
		//
		if err == ErrStoreRecordExists(k) {
			return ErrStoreRecordExists(n)
		}

		if err != nil {
			return err
		}

		st.Infof(ctx, -1, "Created record for %q with revision %v", n, resp.Revision)

		revision = resp.Revision

		return nil
	})

	return revision, err
}

// ReadNew is a method to retrieve the user record associated with the
// supplied name, deal with any store related key prefixes, and decore
// the json encoded record into something the caller understands.
//
// NOTE: A future enhancement may occur where the caller supplies an action
//       routine to take care of the decoding of the record itself allowing
//       this layer to only have to deal with the manipulation of the store,
//       the keys used to persist the callers records but not have to worry
//       about the encode/decode formats or indeed the target record itself.
//
func (store *Store) ReadNew(ctx context.Context, r KeyRoot, n string, m protoiface.MessageV1) (revision int64, err error) {

	revision = RevisionInvalid

	err = st.WithSpan(ctx, tracing.MethodName(1), func(ctx context.Context) (err error) {

		if err := store.disconnected(ctx); err != nil {
			return err
		}

		var (
			rev      int64
			val      string
			response *Response
		)

		key := GetKeyFromUsername(n)

		// If we need to do the read to get the revision, we will need an array of the keys
		//
		request := &Request{Records: make(map[string]Record), Conditions: make(map[string]Condition)}

		k := getKeyFromKeyRootAndName(r, n)
		request.Records[k] = Record{Revision: RevisionInvalid}
		request.Conditions[k] = ConditionUnconditional

		if response, err = store.ReadTxn(ctx, request); err != nil {
			return err
		}

		recordCount := len(response.Records)

		switch recordCount {
		default:
			return ErrStoreBadRecordCount{n, 1, recordCount}

		case 0:
			return ErrStoreKeyNotFound(n)

		case 1:
			rev = response.Records[key].Revision
			val = response.Records[key].Value
			st.Infof(ctx, -1, "found record for %q with revision %v and value %q", n, rev, val)

			err = Decode(val, m)

			st.Infof(ctx, -1, "Read record for %q with revision %v", n, rev)

			revision = rev

			return nil
		}
	})

	return revision, err
}

// ReadNewValue is a method to retrieve the user record associated with the
// supplied name, deal with any store related key prefixes, and decore
// the json encoded record into something the caller understands.
//
// NOTE: A future enhancement may occur where the caller supplies an action
//       routine to take care of the decoding of the record itself allowing
//       this layer to only have to deal with the manipulation of the store,
//       the keys used to persist the callers records but not have to worry
//       about the encode/decode formats or indeed the target record itself.
//
func (store *Store) ReadNewValue(ctx context.Context, r KeyRoot, n string) (value *string, revision int64, err error) {

	revision = RevisionInvalid

	err = st.WithSpan(ctx, tracing.MethodName(1), func(ctx context.Context) (err error) {

		if err := store.disconnected(ctx); err != nil {
			return err
		}

		var (
			rev      int64
			val      string
			response *Response
		)

		key := GetKeyFromUsername(n)

		// If we need to do the read to get the revision, we will need an array of the keys
		//
		request := &Request{Records: make(map[string]Record), Conditions: make(map[string]Condition)}

		k := getKeyFromKeyRootAndName(r, n)
		request.Records[k] = Record{Revision: RevisionInvalid}
		request.Conditions[k] = ConditionUnconditional

		if response, err = store.ReadTxn(ctx, request); err != nil {
			return err
		}

		recordCount := len(response.Records)

		switch recordCount {
		default:
			return ErrStoreBadRecordCount{n, 1, recordCount}

		case 0:
			return ErrStoreKeyNotFound(n)

		case 1:
			rev = response.Records[key].Revision
			val = response.Records[key].Value
			st.Infof(ctx, -1, "found record for %q with revision %v and value %q", n, rev, val)

			st.Infof(ctx, -1, "Read record for %q with revision %v", n, rev)

			revision = rev
			value = &val

			return nil
		}
	})

	return value, revision, err
}

// UpdateNew is a function to conditionally update a value for a single key
//
func (store *Store) UpdateNew(ctx context.Context, r KeyRoot, n string, rev int64, m protoiface.MessageV1) (revision int64, err error) {

	err = st.WithSpan(ctx, tracing.MethodName(1), func(ctx context.Context) (err error) {

		if err = store.disconnected(ctx); err != nil {
			return err
		}

		v, err := Encode(m)

		if err != nil {
			return err
		}

		var condition Condition

		switch {
		case rev == RevisionInvalid:
			condition = ConditionUnconditional

		default:
			condition = ConditionRevisionEqual
		}

		request := &Request{Records: make(map[string]Record), Conditions: make(map[string]Condition)}

		k := getKeyFromKeyRootAndName(r, n)
		request.Records[k] = Record{Revision: rev, Value: v}
		request.Conditions[k] = condition

		resp, err := store.WriteTxn(ctx, request)

		if err != nil {
			return err
		}

		st.Infof(ctx, -1, "Updated record for %q from revision %v to revision %v", n, rev, resp.Revision)

		revision = resp.Revision

		return nil
	})

	return revision, err
}

// DeleteNew is a function to create a single key, value record pair
//
func (store *Store) DeleteNew(ctx context.Context, r KeyRoot, n string, rev int64) (revision int64, err error) {

	err = st.WithSpan(ctx, tracing.MethodName(1), func(ctx context.Context) (err error) {

		if err = store.disconnected(ctx); err != nil {
			return err
		}

		var condition Condition

		switch {
		case rev == RevisionInvalid:
			condition = ConditionUnconditional

		default:
			condition = ConditionRevisionEqual
		}

		request := &Request{Records: make(map[string]Record), Conditions: make(map[string]Condition)}

		k := getKeyFromKeyRootAndName(r, n)
		request.Records[k] = Record{Revision: rev}
		request.Conditions[k] = condition

		resp, err := store.DeleteTxn(ctx, request)

		// Need to strip the namespace prefix and return something described
		// in terms the caller should recognize
		//
		if err == ErrStoreKeyNotFound(k) {
			return ErrStoreKeyNotFound(n)
		}

		if err != nil {
			return err
		}

		st.Infof(ctx, -1, "Deleted record for %q with revision %v at store revision %v", n, rev, resp.Revision)

		revision = resp.Revision

		return nil
	})

	return revision, err
}

// UserRecord represents the revision/user struct pair for a given user
//
type UserRecord struct {
	Revision int64
	User     *pb.User
}

// UserRecordSet is a collection of UserRecord structs used to describe
// a set of users all with a common store revision.
//
type UserRecordSet struct {
	StoreRevision int64
	Records       map[string]UserRecord
}

// UserCreate is a method called by the user management routines to create
// a persistent user record in the store.
//
// No consistency check is performed on the content of the user record itself
//
func (store *Store) UserCreate(ctx context.Context, u *pb.User) (revision int64, err error) {

	methodName := tracing.MethodName(1)

	err = st.WithSpan(ctx, methodName, func(ctx context.Context) (err error) {

		if err := store.disconnected(ctx); err != nil {
			return err
		}

		var (
			rev int64
			val string
		)

		p := jsonpb.Marshaler{}

		if val, err = p.MarshalToString(u); err != nil {
			return err
		}

		recordSet := RecordUpdateSet{Label: methodName, Records: make(map[string]RecordUpdate)}

		recordSet.Records[GetKeyFromUsername(u.Name)] =
			RecordUpdate{
				Condition: ConditionCreate,
				Record: Record{
					Revision: RevisionInvalid,
					Value:    val,
				},
			}

		if rev, err = store.WriteMultipleTxn(ctx, &recordSet); err != nil {
			return err
		}

		st.Infof(ctx, -1, "Created user %q with revision %v", u.Name, rev)

		revision = rev

		return nil
	})

	return revision, err
}

// UserUpdate is a method used to update the user record for the sepcified user.
//
// No validation is performed on the content of the user record itself
//
func (store *Store) UserUpdate(ctx context.Context, u *pb.User, revCond int64) (revision int64, err error) {

	methodName := tracing.MethodName(1)

	err = st.WithSpan(ctx, methodName, func(ctx context.Context) (err error) {

		if err := store.disconnected(ctx); err != nil {
			return err
		}

		var (
			rev       int64
			val       string
			condition Condition
		)

		switch {
		case revCond == RevisionInvalid:
			condition = ConditionUnconditional

		default:
			condition = ConditionRevisionEqual
		}

		key := GetKeyFromUsername(u.Name)

		p := jsonpb.Marshaler{}

		if val, err = p.MarshalToString(u); err != nil {
			return err
		}

		recordSet := RecordUpdateSet{Label: methodName, Records: make(map[string]RecordUpdate)}

		recordSet.Records[key] =
			RecordUpdate{
				Condition: condition,
				Record: Record{
					Revision: revCond,
					Value:    val,
				},
			}

		if rev, err = store.WriteMultipleTxn(ctx, &recordSet); err != nil {
			return err
		}

		st.Infof(ctx, -1, "Updated user %q using condition revision %v to revision %v using condition %v", u.Name, revCond, rev, condition)

		revision = rev

		return nil
	})

	return revision, err
}

// UserDelete is a method used to delete the user record for the specified user
//
func (store *Store) UserDelete(ctx context.Context, u *pb.User, revision int64) (rev int64, err error) {

	key := GetKeyFromUsername(u.Name)

	methodName := tracing.MethodName(1)

	err = st.WithSpan(ctx, methodName, func(ctx context.Context) (err error) {

		if err := store.disconnected(ctx); err != nil {
			return err
		}

		recordSet := RecordUpdateSet{Label: methodName, Records: make(map[string]RecordUpdate)}

		recordSet.Records[key] =
			RecordUpdate{
				Condition: ConditionRevisionEqual,
				Record: Record{
					Revision: revision,
					Value:    "",
				},
			}

		revStore, err := store.DeleteMultipleTxn(ctx, &recordSet)

		if err != nil {
			return err
		}

		rev = revStore

		st.Infof(ctx, -1, "Deleted user %q with revision %v", u.Name, rev)

		return nil
	})

	return rev, err
}

// UserRead is a method to retrieve the user record associated with the
// supplied name, deal with any store related key prefixes, and decore
// the json encoded record into something the caller understands.
//
// NOTE: A future enhancement may occur where the caller supplies an action
//       routine to take care of the decoding of the record itself allowing
//       this layer to only have to deal with the manipulation of the store,
//       the keys used to persist the callers records but not have to worry
//       about the encode/decode formats or indeed the target record itself.
//
func (store *Store) UserRead(ctx context.Context, name string) (user *pb.User, revision int64, err error) {

	revision = RevisionInvalid

	methodName := tracing.MethodName(1)

	err = st.WithSpan(ctx, methodName, func(ctx context.Context) (err error) {

		if err := store.disconnected(ctx); err != nil {
			return err
		}

		var (
			rev          int64
			val          string
			readResponse *RecordSet
		)

		key := GetKeyFromUsername(name)

		// If we need to do the read to get the revision, we will need an array of the keys
		//
		recordKeySet := RecordKeySet{Label: methodName, Keys: []string{key}}

		if readResponse, err = store.ReadMultipleTxn(ctx, recordKeySet); err != nil {
			return err
		}

		recordCount := len(readResponse.Records)

		switch recordCount {
		default:
			return ErrStoreBadRecordCount{name, 1, recordCount}

		case 0:
			return ErrStoreKeyNotFound(name)

		case 1:
			rev = readResponse.Records[key].Revision
			val = readResponse.Records[key].Value
			st.Infof(ctx, -1, "found record for user %q with revision %v and value %q", name, rev, val)

			u := &pb.User{}
			if err = jsonpb.Unmarshal(strings.NewReader(val), u); err != nil {
				return err
			}

			st.Infof(ctx, -1, "Read User %q with revision %v", name, rev)

			user = u
			revision = rev

			return nil
		}
	})

	return user, revision, err
}

// UserList is a method to return all the user records using a single call.
//
// NOTE: The returned UserRecordSet may exist but contain no records.
//
// NOTE: This should only be used at present if the number of user records
//       is limited as there is a limit to the number of records that can
//       be fetched from the store in a single shot. Eventually this will
//       be updated to user an "interupted" enum style call to allow for
//		 an essentially infinite number of records.
//
func (store *Store) UserList(ctx context.Context) (recordSet *UserRecordSet, err error) {

	err = st.WithSpan(ctx, tracing.MethodName(1), func(ctx context.Context) (err error) {

		if err := store.disconnected(ctx); err != nil {
			return err
		}

		var readResponse *RecordSet

		prefix := getNamespaceRootFromKeyRoot(KeyRootUsers)

		if readResponse, err = store.ReadWithPrefix(ctx, prefix); err != nil {
			return err
		}

		rs := &UserRecordSet{StoreRevision: readResponse.Revision, Records: make(map[string]UserRecord, len(readResponse.Records))}

		for k, r := range readResponse.Records {
			u := &pb.User{}

			err = jsonpb.Unmarshal(strings.NewReader(r.Value), u)

			if err != nil {
				st.Errorf(ctx, -1, "failure unmarshalling record - user: %q rev: %v value %q", k, r.Revision, r.Value)
				return err
			}

			if !strings.HasPrefix(k, prefix) {
				return ErrStoreBadRecordKey(k)
			}

			userName := strings.TrimPrefix(k, prefix)

			if userName != GetNormalizedName(u.Name) {
				return ErrStoreBadRecordContent(k)
			}

			rs.Records[userName] = UserRecord{Revision: r.Revision, User: u}

			st.Infof(ctx, -1, "found record for user %q (%q) with revision %v", userName, u.Name, r.Revision)
		}

		rs.StoreRevision = readResponse.Revision

		st.Infof(ctx, -1, "returned %v records at store revision %v", len(rs.Records), rs.StoreRevision)

		recordSet = rs

		return nil
	})

	return recordSet, err
}
