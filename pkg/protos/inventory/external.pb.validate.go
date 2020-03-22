// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: github.com/Jim3Things/CloudChamber/pkg/protos/inventory/external.proto

package inventory

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
var _external_uuidPattern = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

// Validate checks the field values on External with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *External) Validate() error {
	if m == nil {
		return nil
	}

	return nil
}

// ExternalValidationError is the validation error returned by
// External.Validate if the designated constraints aren't met.
type ExternalValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ExternalValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ExternalValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ExternalValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ExternalValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ExternalValidationError) ErrorName() string { return "ExternalValidationError" }

// Error satisfies the builtin error interface
func (e ExternalValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sExternal.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ExternalValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ExternalValidationError{}

// Validate checks the field values on ExternalPdu with the rules defined in
// the proto definition for this message. If any rules are violated, an error
// is returned.
func (m *ExternalPdu) Validate() error {
	if m == nil {
		return nil
	}

	return nil
}

// ExternalPduValidationError is the validation error returned by
// ExternalPdu.Validate if the designated constraints aren't met.
type ExternalPduValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ExternalPduValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ExternalPduValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ExternalPduValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ExternalPduValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ExternalPduValidationError) ErrorName() string { return "ExternalPduValidationError" }

// Error satisfies the builtin error interface
func (e ExternalPduValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sExternalPdu.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ExternalPduValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ExternalPduValidationError{}

// Validate checks the field values on ExternalTor with the rules defined in
// the proto definition for this message. If any rules are violated, an error
// is returned.
func (m *ExternalTor) Validate() error {
	if m == nil {
		return nil
	}

	return nil
}

// ExternalTorValidationError is the validation error returned by
// ExternalTor.Validate if the designated constraints aren't met.
type ExternalTorValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ExternalTorValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ExternalTorValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ExternalTorValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ExternalTorValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ExternalTorValidationError) ErrorName() string { return "ExternalTorValidationError" }

// Error satisfies the builtin error interface
func (e ExternalTorValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sExternalTor.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ExternalTorValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ExternalTorValidationError{}

// Validate checks the field values on ExternalRack with the rules defined in
// the proto definition for this message. If any rules are violated, an error
// is returned.
func (m *ExternalRack) Validate() error {
	if m == nil {
		return nil
	}

	if v, ok := interface{}(m.GetPdu()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ExternalRackValidationError{
				field:  "Pdu",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if v, ok := interface{}(m.GetTor()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ExternalRackValidationError{
				field:  "Tor",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(m.GetBlades()) < 1 {
		return ExternalRackValidationError{
			field:  "Blades",
			reason: "value must contain at least 1 pair(s)",
		}
	}

	for key, val := range m.GetBlades() {
		_ = val

		// no validation rules for Blades[key]

		if v, ok := interface{}(val).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return ExternalRackValidationError{
					field:  fmt.Sprintf("Blades[%v]", key),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	return nil
}

// ExternalRackValidationError is the validation error returned by
// ExternalRack.Validate if the designated constraints aren't met.
type ExternalRackValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ExternalRackValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ExternalRackValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ExternalRackValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ExternalRackValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ExternalRackValidationError) ErrorName() string { return "ExternalRackValidationError" }

// Error satisfies the builtin error interface
func (e ExternalRackValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sExternalRack.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ExternalRackValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ExternalRackValidationError{}

// Validate checks the field values on ExternalZone with the rules defined in
// the proto definition for this message. If any rules are violated, an error
// is returned.
func (m *ExternalZone) Validate() error {
	if m == nil {
		return nil
	}

	if len(m.GetRacks()) < 1 {
		return ExternalZoneValidationError{
			field:  "Racks",
			reason: "value must contain at least 1 pair(s)",
		}
	}

	for key, val := range m.GetRacks() {
		_ = val

		// no validation rules for Racks[key]

		if v, ok := interface{}(val).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return ExternalZoneValidationError{
					field:  fmt.Sprintf("Racks[%v]", key),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	return nil
}

// ExternalZoneValidationError is the validation error returned by
// ExternalZone.Validate if the designated constraints aren't met.
type ExternalZoneValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ExternalZoneValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ExternalZoneValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ExternalZoneValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ExternalZoneValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ExternalZoneValidationError) ErrorName() string { return "ExternalZoneValidationError" }

// Error satisfies the builtin error interface
func (e ExternalZoneValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sExternalZone.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ExternalZoneValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ExternalZoneValidationError{}