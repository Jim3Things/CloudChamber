package store

import (
	"context"
	"strings"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
	st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/admin"
	"github.com/golang/protobuf/jsonpb"
)

const (
	storeRootUsers     = "users/"
	storeRootRacks     = "racks/"
	storeRootBlades    = "blades/"
	storeRootWorkloads = "workloads/"
)

func getKeyFromUsername(name string) string {
	return storeRootUsers + name
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

		recordSet.Records[getKeyFromUsername(u.Name)] =
			RecordUpdate{
				Condition: WriteConditionCreate,
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
			condition WriteCondition
		)

		switch {
		case revCond == RevisionInvalid:
			condition = WriteConditionUnconditional

		default:
			condition = WriteConditionRevisionEqual
		}

		key := getKeyFromUsername(u.Name)

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

	key := storeRootUsers + u.Name

	methodName := tracing.MethodName(1)

	err = st.WithSpan(ctx, methodName, func(ctx context.Context) (err error) {

		if err := store.disconnected(ctx); err != nil {
			return err
		}

		recordSet := RecordUpdateSet{Label: methodName, Records: make(map[string]RecordUpdate)}

		recordSet.Records[key] =
			RecordUpdate{
				Condition: WriteConditionRevisionEqual,
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

		key := getKeyFromUsername(name)

		// If we need to do the read to get the revision, we will need an array of the keys
		//
		recordKeySet := RecordKeySet{Label: methodName, Keys: []string{key}}

		if readResponse, err = store.ReadMultipleTxn(ctx, recordKeySet); err != nil {
			return err
		}

		recordCount := len(readResponse.Records)

		switch recordCount {
		default:
			st.Errorf(ctx, -1, "searching for user %q found %v records when expecting just one", name, recordCount)
			return ErrStoreBadRecordCount(name)

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
	methodName := tracing.MethodName(1)

	err = st.WithSpan(ctx, methodName, func(ctx context.Context) (err error) {

		if err := store.disconnected(ctx); err != nil {
			return err
		}

		var readResponse *RecordSet

		if readResponse, err = store.ReadWithPrefix(ctx, storeRootUsers); err != nil {
			return err
		}

		rs := &UserRecordSet{StoreRevision: readResponse.Revision, Records: make(map[string]UserRecord, len(readResponse.Records))}

		for k, r := range readResponse.Records {
			u := &pb.User{}

			err = jsonpb.Unmarshal(strings.NewReader(r.Value), u)

			if err != nil {
				st.Errorf(ctx, -1, "failure unmarshalling record - user: %q rev: %vvalue %q", k, r.Revision, r.Value)
				return err
			}

			if !strings.HasPrefix(k, storeRootUsers) {
				return ErrStoreBadRecordKey(k)
			}

			userName := strings.TrimPrefix(k, storeRootUsers)

			if userName != u.Name {
				return ErrStoreBadRecordContent(k)
			}

			rs.Records[userName] = UserRecord{Revision: r.Revision, User: u}

			st.Infof(ctx, -1, "found record for user %q eith revision %v", u.Name, r.Revision)
		}

		rs.StoreRevision = readResponse.Revision

		st.Infof(ctx, -1, "returned %v records at store revision %v", len(rs.Records), rs.StoreRevision)

		recordSet = rs

		return nil
	})

	return recordSet, err
}
