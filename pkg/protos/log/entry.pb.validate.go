// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: github.com/Jim3Things/CloudChamber/pkg/protos/log/entry.proto

package log

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang/protobuf/ptypes"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = ptypes.DynamicAny{}
)

// define the regex for a UUID once up-front
var _entry_uuidPattern = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

// Validate checks the field values on Module with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *Module) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Impact

	// no validation rules for Name

	return nil
}

// ModuleValidationError is the validation error returned by Module.Validate if
// the designated constraints aren't met.
type ModuleValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ModuleValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ModuleValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ModuleValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ModuleValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ModuleValidationError) ErrorName() string { return "ModuleValidationError" }

// Error satisfies the builtin error interface
func (e ModuleValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sModule.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ModuleValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ModuleValidationError{}

// Validate checks the field values on Event with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *Event) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Tick

	// no validation rules for Reason

	// no validation rules for Text

	for idx, item := range m.GetImpacted() {
		_, _ = idx, item

		if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return EventValidationError{
					field:  fmt.Sprintf("Impacted[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	return nil
}

// EventValidationError is the validation error returned by Event.Validate if
// the designated constraints aren't met.
type EventValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e EventValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e EventValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e EventValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e EventValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e EventValidationError) ErrorName() string { return "EventValidationError" }

// Error satisfies the builtin error interface
func (e EventValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sEvent.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = EventValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = EventValidationError{}

// Validate checks the field values on Entry with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *Entry) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Name

	// no validation rules for SpanID

	// no validation rules for ParentID

	// no validation rules for Status

	for idx, item := range m.GetEvent() {
		_, _ = idx, item

		if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return EntryValidationError{
					field:  fmt.Sprintf("Event[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	return nil
}

// EntryValidationError is the validation error returned by Entry.Validate if
// the designated constraints aren't met.
type EntryValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e EntryValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e EntryValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e EntryValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e EntryValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e EntryValidationError) ErrorName() string { return "EntryValidationError" }

// Error satisfies the builtin error interface
func (e EntryValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sEntry.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = EntryValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = EntryValidationError{}