package common

// This module contains the common errors used by the Validate methods for the
// proto files defined in Cloud Chamber.

import (
	"fmt"
)

// ErrMustBeGTE signals that the specified field must be greater than or equal
// to a designated value.
type ErrMustBeGTE struct {
	Field    string
	Actual   int64
	Required int64
}

func (e ErrMustBeGTE) Error() string {
	return fmt.Sprintf(
		"the field %q must be greater than or equal to %d.  It is %d, which is invalid",
		e.Field,
		e.Required,
		e.Actual)
}

// ErrInvalidEnum signals that the specified field does not contain a valid
// enum value.
type ErrInvalidEnum struct {
	Field  string
	Actual int64
}

func (e ErrInvalidEnum) Error() string {
	return fmt.Sprintf(
		"the field %q does not contain a known value.  It is %d, which is invalid",
		e.Field,
		e.Actual)
}

// ErrMinLenMap signals that the specified map field does not contain at least
// the minimum required number of entries.
type ErrMinLenMap struct {
	Field    string
	Actual   int64
	Required int64
}

func (e ErrMinLenMap) Error() string {
	suffix := "s"
	if e.Required == 1 {
		suffix = ""
	}

	return fmt.Sprintf(
		"the field %q must contain at least %d element%s.  It contains %d, which is invalid",
		e.Field,
		e.Required,
		suffix,
		e.Actual)
}
