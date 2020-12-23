package ruler

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type LeafTestSuite struct {
	suite.Suite
}

func (ts *LeafTestSuite) TestEvaluate() {
	require := ts.Require()

	val := NewLeafBool(true)
	v, err := val.Evaluate(nil)
	require.NoError(err)
	require.Same(val, v)

	val = NewLeafInt32(1)
	v, err = val.Evaluate(nil)
	require.NoError(err)
	require.Same(val, v)

	val = NewLeafInt64(2)
	v, err = val.Evaluate(nil)
	require.NoError(err)
	require.Same(val, v)

	val = NewLeafString("test")
	v, err = val.Evaluate(nil)
	require.NoError(err)
	require.Same(val, v)
}

func (ts *LeafTestSuite) TestFormat() {
	assert := ts.Assert()

	indent := "    "

	boolVal := NewLeafBool(true)
	assert.Equal("    [Leaf: bool] \"true\"", boolVal.Format(indent))

	int32Val := NewLeafInt32(23)
	assert.Equal("    [Leaf: int32] \"23\"", int32Val.Format(indent))

	int64Val := NewLeafInt64(230000000)
	assert.Equal("    [Leaf: int64] \"230000000\"", int64Val.Format(indent))

	stringVal := NewLeafString("this is a test")
	assert.Equal("    [Leaf: string] \"this is a test\"", stringVal.Format(indent))
}

func (ts *LeafTestSuite) TestAsName() {
	assert := ts.Assert()
	require := ts.Require()

	replacements := []string{
		"%id%", "2",
		"%rack%", "rack1",
		"%workload%", "fred",
	}

	pattern := "/racks/%rack%/blades/%id%"

	stringVal := NewLeafString(pattern)

	s, err := stringVal.AsString()
	require.NoError(err)
	assert.Equal(pattern, s)

	path, err := stringVal.AsName(replacements)
	require.NoError(err)
	assert.Equal("/racks/rack1/blades/2", path)

	stringVal = NewLeafString("/workload/rack1")
	path, err = stringVal.AsName(replacements)
	require.NoError(err)
	assert.Equal("/workload/rack1", path)
}

func (ts *LeafTestSuite) testBoolVal(l *Leaf, expected bool, expErr error) {
	assert := ts.Assert()
	require := ts.Require()

	b, err := l.AsBool()
	require.Equal(expErr, err)
	assert.Equal(expected, b)
}

func (ts *LeafTestSuite) TestAsBool() {
	ts.testBoolVal(NewLeafBool(true), true, nil)
	ts.testBoolVal(NewLeafBool(false), false, nil)

	ts.testBoolVal(NewLeafInt32(1), true, nil)
	ts.testBoolVal(NewLeafInt32(2), true, nil)
	ts.testBoolVal(NewLeafInt32(0), false, nil)

	ts.testBoolVal(NewLeafInt64(1), true, nil)
	ts.testBoolVal(NewLeafInt64(2), true, nil)
	ts.testBoolVal(NewLeafInt64(0), false, nil)

	ts.testBoolVal(NewLeafString("test"), false, ErrInvalidType)
}

func (ts *LeafTestSuite) testInt32Val(l *Leaf, expected int32, expErr error) {
	assert := ts.Assert()
	require := ts.Require()

	b, err := l.AsInt32()
	require.Equal(expErr, err)
	assert.Equal(expected, b)
}

func (ts *LeafTestSuite) TestAsInt32() {
	ts.testInt32Val(NewLeafBool(true), 1, nil)
	ts.testInt32Val(NewLeafBool(false), 0, nil)

	ts.testInt32Val(NewLeafInt32(1), 1, nil)
	ts.testInt32Val(NewLeafInt32(-1), -1, nil)
	ts.testInt32Val(NewLeafInt32(2), 2, nil)

	ts.testInt32Val(NewLeafInt64(1), 1, nil)
	ts.testInt32Val(NewLeafInt64(-1), -1, nil)
	ts.testInt32Val(NewLeafInt64(2), 2, nil)

	ts.testInt32Val(NewLeafString("test"), 0, ErrInvalidType)
}

func (ts *LeafTestSuite) testInt64Val(l *Leaf, expected int64, expErr error) {
	assert := ts.Assert()
	require := ts.Require()

	b, err := l.AsInt64()
	require.Equal(expErr, err)
	assert.Equal(expected, b)
}

func (ts *LeafTestSuite) TestAsInt64() {
	ts.testInt64Val(NewLeafBool(true), 1, nil)
	ts.testInt64Val(NewLeafBool(false), 0, nil)

	ts.testInt64Val(NewLeafInt32(1), 1, nil)
	ts.testInt64Val(NewLeafInt32(-1), -1, nil)
	ts.testInt64Val(NewLeafInt32(2), 2, nil)

	ts.testInt64Val(NewLeafInt64(1), 1, nil)
	ts.testInt64Val(NewLeafInt64(-1), -1, nil)
	ts.testInt64Val(NewLeafInt64(2), 2, nil)

	ts.testInt64Val(NewLeafString("test"), 0, ErrInvalidType)
}

func (ts *LeafTestSuite) testStringVal(l *Leaf, expected string, expErr error) {
	assert := ts.Assert()
	require := ts.Require()

	b, err := l.AsString()
	require.Equal(expErr, err)
	assert.Equal(expected, b)
}

func (ts *LeafTestSuite) TestAsString() {
	ts.testStringVal(NewLeafBool(true), "true", nil)
	ts.testStringVal(NewLeafBool(false), "false", nil)

	ts.testStringVal(NewLeafInt32(3), "3", nil)
	ts.testStringVal(NewLeafInt32(-3), "-3", nil)

	ts.testStringVal(NewLeafInt64(3), "3", nil)
	ts.testStringVal(NewLeafInt64(-3), "-3", nil)

	ts.testStringVal(NewLeafString("test"), "test", nil)
	ts.testStringVal(NewLeafString("%test%"), "%test%", nil)
}

func (ts *LeafTestSuite) testNameVal(l *Leaf, expected string, expErr error) {
	assert := ts.Assert()
	require := ts.Require()

	b, err := l.AsName([]string{"%test%", "fred", "%true%", "bogus"})
	require.Equal(expErr, err)
	assert.Equal(expected, b)
}

func (ts *LeafTestSuite) TestAsName2() {
	ts.testNameVal(NewLeafBool(true), "true", nil)
	ts.testNameVal(NewLeafBool(false), "false", nil)

	ts.testNameVal(NewLeafInt32(3), "3", nil)
	ts.testNameVal(NewLeafInt32(-3), "-3", nil)

	ts.testNameVal(NewLeafInt64(3), "3", nil)
	ts.testNameVal(NewLeafInt64(-3), "-3", nil)

	ts.testNameVal(NewLeafString("test"), "test", nil)
	ts.testNameVal(NewLeafString("%test%"), "fred", nil)
}

func TestLeafTestSuite(t *testing.T) {
	suite.Run(t, new(LeafTestSuite))
}
