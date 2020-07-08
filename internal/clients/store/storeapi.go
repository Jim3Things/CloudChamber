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

		// If we need to do the read to get the revision, we will need an array of the keys
		//
		users := []string{u.Name}

		p := jsonpb.Marshaler{}
		value, err := p.MarshalToString(u)

		if err != nil {
			return err
		}

		// We should probably add a "Create" condition
		//
		recordUpdateSet := RecordUpdateSet{Label: methodName, Records: make(map[string]RecordUpdate)}

		recordUpdateSet.Records[users[0]] =
			RecordUpdate{
				Compare: RevisionCompareUnconditional,
				Record: Record{
					Revision: RevisionUnconditional,
					Value:    value,
				},
			}

		revStore, err := store.WriteMultipleTxn(&recordUpdateSet)

		if err != nil {
			return err
		}

		recordKeySet := RecordKeySet{Label: methodName, Keys: users}

		readResponse, err := store.ReadMultipleTxn(recordKeySet)

		if err != nil {
			return err
		}

		rev = readResponse.Records[users[0]].Revision

		st.Infof(ctx, -1, "Write user record with store revision %v and record revision %v", revStore, rev)

		st.Infof(ctx, -1, "Created User %q with revision %v", users[0], rev)

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

// UserUpdate is a method
//
func (store *Store) UserUpdate() error {
	methodName := tracing.MethodName(1)
	return st.WithSpan(context.Background(), methodName, func(ctx context.Context) (err error) {
		return ErrStoreNotImplemented(methodName)
	})
}

// UserDelete is a method
//
func (store *Store) UserDelete() error {
	methodName := tracing.MethodName(1)
	return st.WithSpan(context.Background(), methodName, func(ctx context.Context) (err error) {
		return ErrStoreNotImplemented(methodName)
	})
}
