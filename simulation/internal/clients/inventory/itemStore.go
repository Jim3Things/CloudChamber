package inventory

import (
	"context"
	"strings"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/namespace"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

// Region, zone and rack are "containers" whereas tor, pdu and blade are "things".
// You can send operations and commands to things, but not containers.
//
// Of the following SetXxx() and GetXxx() methods fetch or set values within the
// fields of an object (cursor) and do NOT perform any operations to the underlying
// store.
//
// the Create(), Read(), Update() and Delete() methods perform the appropriate
// operation to the persisted object based upon the current values of fields
// within the object (cursor) being used and will return an error if needed fields
// are not set.
//
type inventoryItem interface {
	// Create uses the current object to persist the object to the underlying
	// store  and also create any index entries that may be required.
	//
	// This is not valid to call on a root object and doing so will return
	// an error.
	//
	Create(ctx context.Context, view ViewType) (int64, error)

	// Read issues a request to the underlying store to populate all the fields
	// within the cursor object, including any attributes for that object or
	// other information specific to that object.
	//
	// This is not valid to call on a root object and doing so will return
	// an error.
	//
	Read(ctx context.Context, view ViewType) (int64, error)

	// Update will write a record the underlying store using the currently
	// information  in the fields of the object. The update can be either
	// unconditional by setting the unconditional parameter to true, or
	// conditional based on the revision of the object compared to the
	// revision of the associated record in the underlying store.
	//
	// Note the object maintains revision information returned from the
	// store for any actions involving the store, e.g. Create(), Read() or
	// Update().
	//
	Update(ctx context.Context, unconditional bool, view ViewType) (int64, error)

	// Delete is used to remove the persisted copy of the object from the
	// store along with any index information needed to navigate to or
	// through that object. The delete can be either unconditional by
	// setting the unconditional parameter to true, or conditional based
	// on the revision of the object compared to the revision of the
	// associated record in the underlying store.
	//
	// Note the object maintains revision information returned from the
	// store for any actions involving the store, e.g. Create(), Read() or
	// Update().
	//
	Delete(ctx context.Context, unconditional bool, view ViewType) (int64, error)

	// ListChildren uses the current object to discover the names of all the
	// child objects of the current object, The elements of the returned list
	// can be used in subsequent operations to create child objects.
	//
	ListChildren(ctx context.Context) (int64, *[]string, error)

	// Note that it would be ideal to also include the functions that cover
	// creating a child (NewChild), or getting the data for one or more
	// children (FetchChildren), but the lack of generics in the current Go
	// version mean that we'd have to lose type safety.  Consequently, those
	// are omitted until such time as we can include them and retain type
	// safety.
}

// ViewProvider specifies the functions that an instance backed by the store
// must provide to use the itemStore functions.  ViewProvider's functions
// specify how to load stored data into the instance, how to turn the current
// instance's data into a form that can be stored, and how to map generic store
// error codes into ones that are specific to the instance.
//
// TODO: The views will move into their own substructures, and that substructure
//       will implement the viewProvider.  In turn, this will remove the need
//       for Load or Save to pass the ViewType.  Furthermore, we may end up
//       adding a Reset function to this interface that forces the data back
//       to some initial state.
type viewProvider interface {
	// Load converts the string representation that was previously returned
	// from a Save call back into the view's specific content, such that it
	// can be used as part of normal operations.
	Load(view ViewType, value string) error

	// Save returns the view's specific content as a string which can be stored
	// for later retrieval.  If no view data is available, an error is returned.
	Save(view ViewType) (string, error)

	// mapErrStoreValue is a helper function that converts generic store access
	// errors to equivalent view-specific ones, or returns the original error if
	// no mapping for the supplied error is found.
	mapErrStoreValue(err error) error
}

// itemStoreProvider defines the interfaces that an instance using the itemStore
// functions must supply.
type itemStoreProvider interface {
	viewProvider
	keyProvider
	revisionProvider
}

// itemStoreLeaf defines the access functions for an instance that has no
// standard children.
type itemStoreLeaf struct {
	provider itemStoreProvider
}

// Create is used to create a record in the underlying store for the
// object along with the associated index information.
//
// The underlying store record will contain the information currently
// held in the object.
//
// Once the store operation completes successfully, the revision fields
// in the object will be updated to that returned by the store. These can
// either be retrieved by one of the GetRevisionXxx() call or used for
// subsequent conditional operations such as a conditional Update() call.
//
func (is *itemStoreLeaf) Create(ctx context.Context, view ViewType) (int64, error) {
	v, err := is.provider.Save(view)
	if err != nil {
		return store.RevisionInvalid, err
	}

	// Create the child and its index as an atomic pair.
	//
	indexName, err := is.provider.IndexName(view)
	if err != nil {
		return store.RevisionInvalid, err
	}

	keyName, err := is.provider.KeyName(view)
	if err != nil {
		return store.RevisionInvalid, err
	}

	keySet := &map[string]string{
		indexName: is.provider.IndexValue(),
		keyName:   v,
	}

	rev, err := is.provider.DbStore().CreateMultiple(ctx, namespace.KeyRootInventory, keySet)

	if err = is.provider.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	return is.provider.updateRevisionInfo(rev), nil
}

// Read is used to load a record from the underlying store to populate the
// fields in the object and determine the revision values associated with
// that record.
//
// Once the Read() has completed successfully the details and other
// information for the object can be retrieved by any of the GetXxx() methods
// for that object.
//
func (is *itemStoreLeaf) Read(ctx context.Context, view ViewType) (int64, error) {
	if view != ViewDefinition {
		return store.RevisionInvalid, errors.ErrUnimplementedView
	}

	keyName, err := is.provider.KeyName(ViewDefinition)
	if err != nil {
		return store.RevisionInvalid, err
	}

	v, rev, err := is.provider.DbStore().Read(ctx, namespace.KeyRootInventory, keyName)

	if err = is.provider.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	if err = is.provider.Load(view, *v); err != nil {
		return store.RevisionInvalid, err
	}

	return is.provider.updateRevisionInfo(rev), nil
}

// Update is used to persist the information in the fields of the object to
// a record in the underlying store. The update can be either unconditional
// by setting the unconditional parameter to true, or conditional based on
// the revision of the object compared to the revision of the associated
// record in the underlying store.
//
// Once the store operation completes successfully, the revision information
// in the object is updated with that returned from the store.
//
// Update() has no effect on the index information for the object.
//
func (is *itemStoreLeaf) Update(ctx context.Context, unconditional bool, view ViewType) (int64, error) {
	v, err := is.provider.Save(view)
	if err != nil {
		return store.RevisionInvalid, err
	}

	keyName, err := is.provider.KeyName(view)
	if err != nil {
		return store.RevisionInvalid, err
	}

	rev, err := is.provider.DbStore().Update(
		ctx,
		namespace.KeyRootInventory,
		keyName,
		is.provider.GetRevisionForRequest(unconditional),
		v)

	if err = is.provider.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	return is.provider.updateRevisionInfo(rev), nil
}

// Delete is used to remove the persisted copy of the object from the
// store along with any index information needed to navigate to or
// through that object. The delete can be either unconditional by
// setting the unconditional parameter to true, or conditional based
// on the revision of the object compared to the revision of the
// associated record in the underlying store.
//
// Deleting the record from the underlying store has no effect on the
// values held in the fields of the object other than updating the
// revision information using the information returned by the store
// operation.
//
func (is *itemStoreLeaf) Delete(ctx context.Context, unconditional bool, view ViewType) (int64, error) {
	if view != ViewDefinition {
		return store.RevisionInvalid, errors.ErrUnimplementedView
	}

	// Delete the record and its index as an atomic pair.
	//
	indexName, err := is.provider.IndexName(view)
	if err != nil {
		return store.RevisionInvalid, err
	}

	keyName, err := is.provider.KeyName(view)
	if err != nil {
		return store.RevisionInvalid, err
	}

	keySet := &map[string]int64{
		indexName: store.RevisionInvalid,
		keyName:   is.provider.GetRevisionForRequest(unconditional),
	}

	rev, err := is.provider.DbStore().DeleteMultiple(ctx, namespace.KeyRootInventory, keySet)

	if err = is.provider.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, err
	}

	return is.provider.updateRevisionInfo(rev), nil
}

// ListChildren is a stub function for rack objects as there are no generic
// child objects.
func (is *itemStoreLeaf) ListChildren(_ context.Context) (int64, *[]string, error) {
	return store.RevisionInvalid, nil, errors.ErrFunctionNotAvailable
}

// itemStore extends itemStoreLeaf with support for the listing of standard
// child objects.
type itemStore struct {
	itemStoreLeaf
}

// ListChildren uses the current object to discover the names of all the
// child objects in the underlying store for the the current itemStore instance.
// The elements of the returned list can be used in subsequent calls to create
// new children or retrieve their data from the store.
//
func (is *itemStore) ListChildren(ctx context.Context) (int64, []string, error) {
	childIndexName, err := is.provider.ChildIndexName(ViewDefinition)
	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	records, rev, err := is.provider.DbStore().List(ctx, namespace.KeyRootInventory, childIndexName)

	if err = is.provider.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, nil, err
	}

	names := make([]string, 0, len(*records))

	for k, v := range *records {

		name := strings.TrimPrefix(k, childIndexName)

		if name != namespace.GetNormalizedName(v.Value) {
			return store.RevisionInvalid, nil, errors.ErrIndexKeyValueMismatch{
				Namespace: is.provider.TableName().String(),
				Key:       name,
				Value:     v.Value}
		}

		names = append(names, v.Value)
	}

	return rev, names, nil
}


