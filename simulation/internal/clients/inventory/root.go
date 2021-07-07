package inventory

import (
	"context"
	"strings"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/namespace"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

// Root is a structure representing the well-known root of the namespace. It
// is used to locate the regions within the namespace represented by the table
// field.
//
// The root object is an in-memory construct only and cannot be persisted to
// the store, or retrieved from it.
//
// TODO: It looks like Root should have an index variant that handles the most
//       primitive store support (e.g. removing teh CRUD functions below).
//       Investigate for the future.
type Root struct {
	Store         *store.Store
	KeyChildIndex string
	Table         namespace.TableName

	revisionInfo

	details *pb.RootDetails
}

func newRoot(store *store.Store, table namespace.TableName) (*Root, error) {

	k, err := namespace.GetKeyForIndexRegions(table)

	if err != nil {
		return nil, err
	}

	r := &Root{
		Store:         store,
		KeyChildIndex: k,
		Table:         table,
	}

	return r, nil
}

// SetDetails is used to attach some attribute information to the object.
//
// For a Root object, the information is not persisted.
//
// The current revision of the region object is reset
//
func (r *Root) SetDetails(details *pb.RootDetails) {
	r.details = details.Clone()
	r.resetRevision()
}

// GetDetails is used to extract the attribute information from the object.
//
// As the Root object is not persisted, the attribute information will either
// be the initialisation value, or whatever was last set using SetDetails()
//
func (r *Root) GetDetails() *pb.RootDetails {
	return r.details.Clone()
}

// Create is not used for a Root object as there is no persistence for this
// object.
//
func (r *Root) Create(_ context.Context, _ ViewType) (int64, error) {
	return store.RevisionInvalid, errors.ErrFunctionNotAvailable
}

// Read is not used for a Root object as there is no persistence for this
// object.
//
func (r *Root) Read(_ context.Context, _ ViewType) (int64, error) {
	return store.RevisionInvalid, errors.ErrFunctionNotAvailable
}

// Update is not used for a Root object as there is no persistence for this
// object.
//
func (r *Root) Update(_ context.Context, _ bool, _ ViewType) (int64, error) {
	return store.RevisionInvalid, errors.ErrFunctionNotAvailable
}

// Delete is not used for a Root object as there is no persistence for this
// object.
//
func (r *Root) Delete(_ context.Context, _ bool, _ ViewType) (int64, error) {
	return store.RevisionInvalid, errors.ErrFunctionNotAvailable
}

// NewChild creates a new region child object within the current
// root using the supplied name. This new object can be used for
// further navigation or for actions involving operations against the
// associated record in the underlying store.
//
// No information is fetched from the underlying store so the attribute
// and revisions fields within the object are not valid.
//
func (r *Root) NewChild(name string) (*Region, error) {

	return newRegion(r.Store, r.Table, name)
}

// ListChildren uses the current object to discover the names of all the
// region child objects in the underlying store for the the current root
// object, The elements of the returned list can be used in subsequent
// NewChild() calls to create new region objects.
//
func (r *Root) ListChildren(ctx context.Context) (int64, []string, error) {

	records, rev, err := r.Store.List(ctx, namespace.KeyRootInventory, r.KeyChildIndex)

	if err = r.mapErrStoreValue(err); err != nil {
		return store.RevisionInvalid, nil, err
	}

	names := make([]string, 0, len(*records))

	for k, v := range *records {

		name := strings.TrimPrefix(k, r.KeyChildIndex)

		if name != namespace.GetNormalizedName(v.Value) {
			return store.RevisionInvalid, nil, errors.ErrIndexKeyValueMismatch{
				Namespace: r.Table.String(),
				Key:       name,
				Value:     v.Value,
			}
		}

		names = append(names, v.Value)
	}

	return rev, names, nil
}

// FetchChildren is used to discover all the child region objects in the
// underlying store for the current root object and to generate a new
// region object for each of those children. It is a convenience wrapper
// around ListChildren() followed by a NewChild() and Read() on each name
// discovered.
//
func (r *Root) FetchChildren(ctx context.Context) (int64, *map[string]Region, error) {

	rev, names, err := r.ListChildren(ctx)
	if err != nil {
		return store.RevisionInvalid, nil, err
	}

	children := make(map[string]Region, len(names))

	for _, v := range names {

		child, err := r.NewChild(v)
		if err != nil {
			return store.RevisionInvalid, nil, err
		}

		_, err = child.Read(ctx, ViewDefinition)
		if err != nil {
			return store.RevisionInvalid, nil, err
		}

		children[v] = *child
	}

	return rev, &children, nil
}

func (r *Root) mapErrStoreValue(err error) error {
	switch err {
	case errors.ErrStoreIndexNotFound(r.KeyChildIndex):
		return errors.ErrIndexNotFound(r.KeyChildIndex)
	}

	return err
}
