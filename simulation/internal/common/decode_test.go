package common

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type decodeTestSuite struct {
	suite.Suite
}

func (ts *decodeTestSuite) keywords() []string {
	return []string{"Defined", "Actual", "Target", "Observed"}
}

func (ts *decodeTestSuite) doSuccess(
	keywords []string,
	view string,
	results []bool) {
	require := ts.Require()

	got, err := Decode(keywords, view)
	require.NoError(err)
	require.Equal(results, got)
}

func (ts *decodeTestSuite) doFail(
	keywords []string,
	view string,
	errString string) {
	require := ts.Require()

	got, err := Decode(keywords, view)
	require.EqualError(err, errString)
	require.Nil(got)
}

func (ts *decodeTestSuite) TestDecode() {
	ts.doSuccess(
		ts.keywords(),
		"Defined, Observed",
		[]bool{true, false, false, true})
}

func (ts *decodeTestSuite) TestDecodeBadView() {
	ts.doFail(
		ts.keywords(),
		"AbaddefinedobservedValue",
		`The source "AbaddefinedobservedValue" was not found in keywords [Defined Actual Target Observed].`)
}

func (ts *decodeTestSuite) TestDecodeMultipleView() {
	ts.doSuccess(
		ts.keywords(),
		"Defined, Defined, Observed, Actual",
		[]bool{true, true, false, true})
}

func (ts *decodeTestSuite) TestDecodeBlankSource() {
	ts.doFail(
		ts.keywords(),
		"",
		`The source "" was not found in keywords [Defined Actual Target Observed].`)
}

func (ts *decodeTestSuite) TestDecodeMultipleCopiesSource() {
	ts.doSuccess(
		[]string{"Defined", "Defined", "Observed", "Defined"},
		"Defined",
		[]bool{true, true, false, true})
}

func (ts *decodeTestSuite) TestDecodeEmptyArraySource() {
	ts.doFail(
		make([]string, 0, 4),
		"Defined",
		`The source "Defined" was not found in keywords [].`)
}

func (ts *decodeTestSuite) TestDecodeNilKeyword() {
	ts.doFail(
		nil,
		"Defined",
		`The source "Defined" was not found in keywords [].`)
}

func (ts *decodeTestSuite) TestDecodeKeywordNil() {
	ts.doFail(
		nil,
		"Defined, Observed",
		`The source "Defined, Observed" was not found in keywords [].`)
}

func (ts *decodeTestSuite) TestDecodeBestView() {
	ts.doSuccess(
		ts.keywords(),
		"Defined, Actual, Target, Observed",
		[]bool{true, true, true, true})
}

func TestDecodeSuite(t *testing.T) {
	suite.Run(t, new(decodeTestSuite))
}
