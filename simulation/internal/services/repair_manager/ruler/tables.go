package ruler

import (
	"strings"

	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

// Key defines the path to a designated field.
type Key struct {
	Table string
	Nodes []string
	Field string
}

// MakeKey produces a key from a path string which uses '/' between table and
// path node elements, and '.' to designate a field.  For instance, the string
// 'table1/path1/path2.field1' would designate table 'table1', a path made up
// of two nodes: 'path1' with 'path2' as its child, and a field called
// 'field1'.
func MakeKey(path string) (*Key, error) {
	pathAndField := strings.Split(path, ".")
	if len(pathAndField) == 1 {
		return nil, errors.ErrMissingFieldName(path)
	}

	if len(pathAndField) > 2 {
		return nil, errors.ErrExtraFieldNames(path)
	}

	tableAndPath := strings.Split(pathAndField[0], "/")

	if len(tableAndPath) == 1 {
		return nil, errors.ErrMissingPath(path)
	}

	return &Key{
		Table: tableAndPath[0],
		Nodes: tableAndPath[1:],
		Field: pathAndField[1],
	}, nil
}

// Tables defines the core interface for a collection of tables.
type Tables interface {
	// GetTable returns the table that matches the Table field in the key
	GetTable(key *Key) (Table, error)
}

// Table defines the core interface for a named table of values
type Table interface {
	// GetValue returns a field value that matches the path and field names
	// in the key
	GetValue(key *Key) (interface{}, error)
}
