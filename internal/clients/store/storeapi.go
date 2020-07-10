package store

import (
	"bytes"
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

// UserCreate is a method called by the user management routines to create
// a persistent user record in the store.
//
// No consistency check is performed on the content of the user record itself
//
func (store *Store) UserCreate(ctx context.Context, u *pb.User) (rev int64, err error) {

	key := storeRootUsers + u.Name

	methodName := tracing.MethodName(1)

	err = st.WithSpan(ctx, methodName, func(ctx context.Context) (err error) {

		if err := store.disconnected(ctx); err != nil {
			return err
		}

		p := jsonpb.Marshaler{}
		value, err := p.MarshalToString(u)

		if err != nil {
			return err
		}

		recordSet := RecordUpdateSet{Label: methodName, Records: make(map[string]RecordUpdate)}

		recordSet.Records[key] =
			RecordUpdate{
				Condition: WriteConditionCreate,
				Record: Record{
					Revision: RevisionInvalid,
					Value:    value,
				},
			}

		revStore, err := store.WriteMultipleTxn(ctx, &recordSet)

		if err != nil {
			return err
		}

		rev = revStore

		st.Infof(ctx, -1, "Created user %q with revision %v", u.Name, rev)

		return nil
	})

	return rev, err
}

// UserUpdate is a method used to update the user record for the sepcified user.
//
// No validation is performed on the content of the user record itself
//
func (store *Store) UserUpdate(ctx context.Context, u *pb.User, revision int64) (rev int64, err error) {

	key := storeRootUsers + u.Name

	methodName := tracing.MethodName(1)

	err = st.WithSpan(ctx, methodName, func(ctx context.Context) (err error) {

		if err := store.disconnected(ctx); err != nil {
			return err
		}

		p := jsonpb.Marshaler{}
		value, err := p.MarshalToString(u)

		if err != nil {
			return err
		}

		recordSet := RecordUpdateSet{Label: methodName, Records: make(map[string]RecordUpdate)}

		recordSet.Records[key] =
			RecordUpdate{
				Condition: WriteConditionRevisionEqual,
				Record: Record{
					Revision: revision,
					Value:    value,
				},
			}

		revStore, err := store.WriteMultipleTxn(ctx, &recordSet)

		if err != nil {
			return err
		}

		rev = revStore

		st.Infof(ctx, -1, "Updated user %q with revision %v", u.Name, rev)

		return nil
	})

	return rev, err
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
func (store *Store) UserRead(ctx context.Context, name string) (user *pb.User, rev int64, err error) {

	rev = RevisionInvalid
	key := storeRootUsers + name

	methodName := tracing.MethodName(1)

	err = st.WithSpan(ctx, methodName, func(ctx context.Context) (err error) {

		if err := store.disconnected(ctx); err != nil {
			return err
		}

		// If we need to do the read to get the revision, we will need an array of the keys
		//
		recordKeySet := RecordKeySet{Label: methodName, Keys: []string{key}}

		readResponse, err := store.ReadMultipleTxn(ctx, recordKeySet)

		if err != nil {
			return err
		}

		recordCount := len(readResponse.Records)

		if recordCount != 1 {
			return ErrStoreBadRecordCount(recordCount)
		}

		u := &pb.User{}

		err = jsonpb.Unmarshal(strings.NewReader(readResponse.Records[key].Value), u)

		if err != nil {
			return err
		}

		user = u
		rev = readResponse.Records[key].Revision

		st.Infof(ctx, -1, "Read User %q with revision %v", name, rev)

		return nil
	})

	return user, rev, err
}

// UserRecord represents the revision, user struct pair for a given user
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

		readResponse, err := store.ReadWithPrefix(storeRootUsers)

		if err != nil {
			return err
		}

		rs := &UserRecordSet{StoreRevision: RevisionInvalid, Records: make(map[string]UserRecord, len(readResponse))}

		for _, kv := range readResponse {
			u := &pb.User{}

			err = jsonpb.Unmarshal(bytes.NewBuffer(kv.value), u)

			if err != nil {
				return err
			}

			if !strings.HasPrefix(kv.key, storeRootUsers) {
				return ErrStoreBadRecordKey(kv.key)
			}

			userName := strings.TrimPrefix(kv.key, storeRootUsers)

			if userName != u.Name {
				return ErrStoreBadRecordContent(kv.key)
			}

			rs.Records[userName] = UserRecord{Revision: RevisionInvalid, User: u}
		}

		rs.StoreRevision = RevisionInvalid

		recordSet = rs

		return nil
	})

	return recordSet, err
}
