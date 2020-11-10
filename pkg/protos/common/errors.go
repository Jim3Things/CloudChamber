package common

// This module contains the common errors used by the Validate methods for the
// proto files defined in Cloud Chamber.

import (
	"fmt"
)

// ErrMustBeEQ signals that the specified field must be equal to a
// designated value.
type ErrMustBeEQ struct {
	Field    string
	Actual   int64
	Required int64
}

func (e ErrMustBeEQ) Error() string {
	return fmt.Sprintf(
		"the field %q must be equal to %d.  It is %d, which is invalid",
		e.Field,
		e.Required,
		e.Actual)
}

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

// ErrMaxLenMap signals that the specified map field does not contain at least
// the minimum required number of entries.
type ErrMaxLenMap struct {
	Field  string
	Actual int64
	Limit  int64
}

func (e ErrMaxLenMap) Error() string {
	suffix := "s"
	if e.Limit == 1 {
		suffix = ""
	}

	return fmt.Sprintf(
		"the field %q must contain no more then %d element%s.  It contains %d, which is invalid",
		e.Field,
		e.Limit,
		suffix,
		e.Actual)
}

// ErrInvalidID signals that the specified tracing ID does not contain a valid
// value.
type ErrInvalidID struct {
	Field string
	Type  string
	ID    string
}

func (e ErrInvalidID) Error() string {
	return fmt.Sprintf("the field %q must be a valid %s ID.  It contains %q, which is invalid",
		e.Field,
		e.Type,
		e.ID)
}

// ErrIDMustBeEmpty signals that the specified ID contains a value when it must
// not (due to consistency rules involving other fields).
type ErrIDMustBeEmpty struct {
	Field string
	Actual string
}

func (e ErrIDMustBeEmpty) Error() string {
	return fmt.Sprintf("the field %q must be emtpy.  It contains %q, which is invalid",
		e.Field,
		e.Actual)
}

// ErrIDMustHaveValue signals that the specified tracing ID does not contain any value.
//
type ErrIDMustHaveValue struct {
	Field string
	Actual string
}

func (e ErrIDMustHaveValue) Error() string {
	return fmt.Sprintf("the field %q must have a value.  It contains %q, which is invalid",
		e.Field,
		e.Actual)
}

// ErrMinLenString signals that the specified string does not meet a minimum lenth criteria.
//
type ErrMinLenString struct {
	Field    string
	Actual   int64
	Required int64
}

func (e ErrMinLenString) Error() string {
	suffix := "s"
	if e.Required == 1 {
		suffix = ""
	}

	return fmt.Sprintf(
		"the field %q must contain at least %d character%s.  It contains %d, which is invalid",
		e.Field,
		e.Required,
		suffix,
		e.Actual)
}

// ErrItemMustBeEmpty signals that the specified Item's port contains a value when it must
// not (due to consistency rules involving other fields).
//
type ErrItemMustBeEmpty struct {
	Field string
	Item string
	Port int64
	Actual string
}

func (e ErrItemMustBeEmpty) Error() string {
	return fmt.Sprintf("the field %q for %q port %q must be emtpy.  It contains %q, which is invalid",
		e.Field,
		e.Item,
		e.Port,
		e.Actual)
}

// ErrItemMissingValue signals that the specified tracing ID does not contain any value.
//
type ErrItemMissingValue struct {
	Field string
	Item string
	Port int64
}

func (e ErrItemMissingValue) Error() string {
	return fmt.Sprintf("the field %q for %q port %q must have a value.",
		e.Field,
		e.Item,
		e.Port)
}

// ErrInvalidItemSelf signals that the specified item has wired a port to itsel.
//
type ErrInvalidItemSelf struct {
	Field string
	Item string
	Port int64
	Actual string
}

func (e ErrInvalidItemSelf) Error() string {
	return fmt.Sprintf(
		"the field %q for %q port %q must be a valid type.  It contains %q, which connects to itself",
		e.Field,
		e.Item,
		e.Port,
		e.Actual)
}

