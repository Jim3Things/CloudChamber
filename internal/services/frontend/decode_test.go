package frontend

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecode(t *testing.T) {
	var keywords []string = []string  {"Defined", "Actual", "Target", "Observed"}
	var views = "Defined, Observed"
	var expected []bool = []bool {true, false, false, true}

	got, err := Decode(keywords, views)
	require.NoError(t,err)
	assert.Equal(t,expected, got)
}

func TestDecodeBadView(t *testing.T) {
	var keywords []string = []string  {"Defined", "Actual", "Target", "Observed"}
	var badView = "AbaddefinedobservedValue" 
	var badexpected []bool = nil

	got, err := Decode(keywords, badView)

	require.Error(t, err)
	var s = err.Error()
	fmt.Printf(s)
	assert.Equal(t, badexpected, got)
}

func TestDecodeMultipleView(t *testing.T) {
	var keywords []string = []string  {"Defined", "Actual", "Target", "Observed"}
	var multipleView = "Defined, Defined, Observed, Actual"
	var expected []bool = []bool {true, true, false, true}
	
	got, err := Decode(keywords, multipleView)
	require.NoError(t, err)
	assert.Equal(t, expected,got)
}

 func TestDecodeBlankSource(t *testing.T){
	var keywords []string = []string  {"Defined", "Actual", "Target", "Observed"}
	var blankSource = ""
	var expected []bool = nil

	got, err := Decode(keywords, blankSource)
	require.Error(t, err)
	assert.Equal(t, expected, got)
}

  func TestDecodeMultipleCopiesSource(t *testing.T){
	var source string = "Defined"
	var keyword2 []string = []string {"Defined", "Defined", "Observed", "Defined"}
	var expected []bool = []bool {true, true, false, true}

	got, err := Decode(keyword2, source)
	require.NoError(t, err)
	assert.Equal(t, expected, got)
 }

 
 func TestDecodeEmptyArraySource(t *testing.T){
	emptyArray := make ([]string, 0, 4) 
	var someString = "Defined"
	var expected []bool = nil

	got, err := Decode(emptyArray, someString)
	require.Error(t, err)
	assert.Equal(t, expected, got)
 }

  func TestDecodeNilKeyword(t *testing.T){
	var source string = "Defined"
	var keyword2 []string = nil
	var expected []bool = nil

	got, err := Decode(keyword2, source)
	require.Error(t, err)
	assert.Equal(t, expected, got)
 }

func TestDecodeKeywordNil(t *testing.T){
	var keywords []string = nil
	var views = "Defined, Observed"
	var expected []bool = nil

	got, err := Decode(keywords, views)
	require.Error(t, err)
	assert.Equal(t, expected, got)
}

func TestDecodeBestView(t *testing.T) {
	var keywords []string = []string  {"Defined", "Actual", "Target", "Observed"}
	var views = "Defined, Actual, Target, Observed"
	var expected []bool = []bool {true, true, true, true}

	got, err := Decode(keywords, views)
	require.NoError(t,err)
	assert.Equal(t,expected, got)
}