// This module contains the common errors used by the Validate methods for the
// proto files defined in Cloud Chamber
package common

import (
    "fmt"
)

// The field must be greater than or equal to a designated value
type ErrMustBeGTE struct {
    Field string
    Actual int64
    Required int64
}

func (e ErrMustBeGTE) Error() string {
    return fmt.Sprintf(
        "the field %q must be greater than or equal to %d.  It is %d, which is invalid",
        e.Field,
        e.Required,
        e.Actual)
}

// The field must contain a recognized enum value
type ErrInvalidEnum struct {
    Field string
    Actual int64
}

func (e ErrInvalidEnum) Error() string {
    return fmt.Sprintf(
        "the field %q does not contain a known value.  It is %d, which is invalid",
        e.Field,
        e.Actual)
}

// The supplied map must contain at least a designated number of entries
type ErrMinLenMap struct {
    Field string
    Actual int64
    Required int64
}

func (e ErrMinLenMap) Error() string {
    suffix := "s"
    if e.Required == 1 { suffix = "" }

    return fmt.Sprintf(
        "the field %q must contain at least %d element%s.  It contains %d, which is invalid",
        e.Field,
        e.Required,
        suffix,
        e.Actual)
}