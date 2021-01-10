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
	var multipleview = "Defined, Defined, Observed, Actual"
	var expected []bool = []bool {true, true, false, true}
	
	got, err := Decode(keywords, multipleview)
	require.NoError(t, err)
	assert.Equal(t, expected,got)
}

 func TestDecodeBlankSource(t *testing.T){
	var keywords []string = []string  {"Defined", "Actual", "Target", "Observed"}
	var blanksource = ""
	var expected []bool = nil

	got, err := Decode(keywords, blanksource)
	require.Error(t, err)
	assert.Equal(t, expected, got)
}

  func TestDecodemultiplecopiesSource(t *testing.T){
	var source string = "Defined"
	var keyword2 []string = []string {"Defined", "Defined", "Observed", "Defined"}
	var expected []bool = []bool {true, true, false, true}

	got, err := Decode(keyword2, source)
	require.NoError(t, err)
	assert.Equal(t, expected, got)
 }

 
 func TestDecodeemptyarraySource(t *testing.T){
	emptyArray := make ([]string, 0, 4) 
	var somestring = "Defined"
	var expected []bool = nil

	got, err := Decode(emptyArray, somestring)
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

func TestDecodekeywordNil(t *testing.T){
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