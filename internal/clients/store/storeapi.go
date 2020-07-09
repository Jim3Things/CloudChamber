package store

import (
	"context"
	"strings"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
	st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/admin"
	"github.com/golang/protobuf/jsonpb"
)

// UserList is a method
//
func (store *Store) UserList() error {
	methodName := tracing.MethodName(1)
	return st.WithSpan(context.Background(), methodName, func(ctx context.Context) (err error) {
		return ErrStoreNotImplemented(methodName)
	})
}

// UserCreate is a method called by the user management
// routines to create a persistent user record in the store.
//
func (store *Store) UserCreate(u *pb.User) (rev int64, err error) {

	methodName := tracing.MethodName(1)

	err = st.WithSpan(context.Background(), methodName, func(ctx context.Context) (err error) {

		if err := store.disconnected(ctx); err != nil {
			return err
		}

		p := jsonpb.Marshaler{}
		value, err := p.MarshalToString(u)

		if err != nil {
			return err
		}

		recordSet := RecordUpdateSet{Label: methodName, Records: make(map[string]RecordUpdate)}

		recordSet.Records[u.Name] =
			RecordUpdate{
				Condition: WriteConditionCreate,
				Record: Record{
					Revision: RevisionInvalid,
					Value:    value,
				},
			}

		revStore, err := store.WriteMultipleTxn(&recordSet)

		if err != nil {
			return err
		}

		rev = revStore

		st.Infof(ctx, -1, "Created user %q with revision %v", u.Name, rev)

		return nil
	})

	return rev, err
}

// UserUpdate is a method
//
func (store *Store) UserUpdate(u *pb.User, revision int64) (rev int64, err error) {

	methodName := tracing.MethodName(1)

	err = st.WithSpan(context.Background(), methodName, func(ctx context.Context) (err error) {

		if err := store.disconnected(ctx); err != nil {
			return err
		}

		p := jsonpb.Marshaler{}
		value, err := p.MarshalToString(u)

		if err != nil {
			return err
		}

		recordSet := RecordUpdateSet{Label: methodName, Records: make(map[string]RecordUpdate)}

		recordSet.Records[u.Name] =
			RecordUpdate{
				Condition: WriteConditionRevisionEqual,
				Record: Record{
					Revision: revision,
					Value:    value,
				},
			}

		revStore, err := store.WriteMultipleTxn(&recordSet)

		if err != nil {
			return err
		}

		rev = revStore

		st.Infof(ctx, -1, "Updated user %q with revision %v", u.Name, rev)

		return nil
	})

	return rev, err
}

// UserRead is a method
//
// What about InvalidRev, NewErrUserNotFound(name)
//
func (store *Store) UserRead(name string) (user *pb.User, rev int64, err error) {

	rev = RevisionInvalid

	methodName := tracing.MethodName(1)

	err = st.WithSpan(context.Background(), methodName, func(ctx context.Context) (err error) {

		if err := store.disconnected(ctx); err != nil {
			return err
		}

		// If we need to do the read to get the revision, we will need an array of the keys
		//
		recordKeySet := RecordKeySet{Label: methodName, Keys: []string{name}}

		readResponse, err := store.ReadMultipleTxn(recordKeySet)

		if err != nil {
			return err
		}

		u := &pb.User{}

		err = jsonpb.Unmarshal(strings.NewReader(readResponse.Records[name].Value), u)

		if err != nil {
			return err
		}

		user = u
		rev = readResponse.Records[name].Revision

		st.Infof(ctx, -1, "Read User %q with revision %v", name, rev)

		return nil
	})

	return user, rev, err
}

// UserDelete is a method
//
func (store *Store) UserDelete() error {
	methodName := tracing.MethodName(1)
	return st.WithSpan(context.Background(), methodName, func(ctx context.Context) (err error) {
		return ErrStoreNotImplemented(methodName)
	})
}
