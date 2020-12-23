package ruler

import (
	"fmt"
	"strings"
)

type ValueType int

const (
	ValueInvalid ValueType = iota
	ValueInt64
	ValueInt32
	ValueBool
	ValueString
)

// Leaf holds a data value and its associated data type.  Note that the types
// break down into numeric or string families, thus reducing the number of
// fields needed to store the data.
type Leaf struct {
	vtype  ValueType
	numVal uint64
	strVal string
}

// +++ Constructor functions

// NewLeafString creates a new Leaf entry containing the supplied string value.
func NewLeafString(val string) *Leaf {
	return &Leaf{
		vtype:  ValueString,
		numVal: 0,
		strVal: val,
	}
}

// NewLeafBool creates a new Leaf entry containing the supplied boolean value.
func NewLeafBool(val bool) *Leaf {
	// true is a nonzero value, false is zero
	num := 0
	if val {
		num = 1
	}

	return &Leaf{
		vtype:  ValueBool,
		numVal: uint64(num),
		strVal: "",
	}
}

// NewLeafInt32 creates a new Leaf entry containing the supplied integer value.
func NewLeafInt32(val int32) *Leaf {
	return &Leaf{
		vtype:  ValueInt32,
		numVal: uint64(val),
		strVal: "",
	}
}

// NewLeafInt64 creates a new Leaf entry containing the supplied integer value.
func NewLeafInt64(val int64) *Leaf {
	return &Leaf{
		vtype:  ValueInt64,
		numVal: uint64(val),
		strVal: "",
	}
}

// --- Constructor functions

// Evaluate returns this Leaf element.
func (v *Leaf) Evaluate(*EvalContext) (*Leaf, error) { return v, nil }

// Format returns a descriptive string for this instance.
func (v *Leaf) Format(indent string) string {
	return fmt.Sprintf("%s[Leaf: %s] %q", indent, v.typeName(), v.valueString())
}

// AsName returns a string where the replacement tokens are applied.  In order
// to do so, this function accepts an array of strings that are alternating
// strings of tokens and replacement values.
func (v *Leaf) AsName(replacements []string) (string, error) {
	if v.vtype != ValueString {
		return v.AsString()
	}

	r := strings.NewReplacer(replacements...)

	return r.Replace(v.strVal), nil
}

// AsInt64 returns the int64 value, if it is a numeric value.  int32 and int64
// values are returned as expected, boolean returns 1 for true, 0 for false,
// while string values return an error.
func (v *Leaf) AsInt64() (int64, error) {
	switch v.vtype {
	case ValueBool, ValueInt64, ValueInt32:
		return int64(v.numVal), nil

	default:
		return 0, ErrInvalidType
	}
}

// AsInt32 returns the int632 value, if it is a numeric value.  int32 and int64
// values are returned as expected (including truncation), boolean returns 1
// for true, 0 for false, while string values return an error.
func (v *Leaf) AsInt32() (int32, error) {
	switch v.vtype {
	case ValueBool, ValueInt64, ValueInt32:
		return int32(v.numVal), nil

	default:
		return 0, ErrInvalidType
	}
}

// AsBool returns true if the numeric value is non-zero, false if it is zero,
// and an error if it is a string value.
func (v *Leaf) AsBool() (bool, error) {
	switch v.vtype {
	case ValueBool, ValueInt64, ValueInt32:
		return v.numVal != 0, nil

	default:
		return false, ErrInvalidType
	}
}

// AsString returns either the string value, or the numeric value as a
// formatted string.
func (v *Leaf) AsString() (string, error) {
	switch v.vtype {
	case ValueBool:
		if v.numVal != 0 {
			return "true", nil
		}
		return "false", nil

	case ValueInt32:
		return fmt.Sprintf("%d", int32(v.numVal)), nil

	case ValueInt64:
		return fmt.Sprintf("%d", int64(v.numVal)), nil

	case ValueString:
		return v.strVal, nil

	default:
		return "", ErrInvalidType
	}
}

func (v *Leaf) typeName() string {
	switch v.vtype {
	case ValueBool:
		return "bool"
	case ValueInt32:
		return "int32"
	case ValueInt64:
		return "int64"
	case ValueString:
		return "string"
	default:
		return "Invalid"
	}
}

func (v *Leaf) valueString() string {
	switch v.vtype {
	case ValueBool:
		return fmt.Sprintf("%v", v.numVal != 0)
	case ValueInt32:
		return fmt.Sprintf("%v", int32(v.numVal))
	case ValueInt64:
		return fmt.Sprintf("%v", int64(v.numVal))
	case ValueString:
		return v.strVal
	default:
		return "<<unknown>>"
	}
}
