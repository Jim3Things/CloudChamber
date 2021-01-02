package frontend

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Input and output for a success case
var keywords []string = []string  {"Defined", "Actual", "Target", "Observed"}
var response []string

func TestDecode(t *testing.T) {

	var views = "Defined, Observed"
	var expected []bool = []bool {true, false, false, true}

	got, err := decode(keywords, views)
	require.NoError(t,err)
	assert.Equal(t,expected, got)
}

func TestDecodeBadView(t *testing.T) {

	var badView = "AbaddefinedobservedValue" 
	var badexpected []bool = nil

	got, err := decode(keywords, badView)
	require.Error(t, err)
	assert.Equal(t, badexpected, got)
}

func TestDecodeMultipleView(t *testing.T) {
	var multipleview = "Defined, Defined, Observed, Actual"
	var expected []bool = []bool {true, false, false, true}
	
	got, err := decode(keywords, multipleview)
	require.NoError(t, err)
	assert.Equal(t, expected,got)
}
 // COMPLETELY BLANK SOURCE	"" WHAT HAPPENS then
func TestDecodeBlankSource(t *testing.T){
	var blanksource = ""
	var expected []bool = []bool {false, false, false, false}

	got, err := decode(keywords, blanksource)
	require.Error(t, err)
	assert.Equal(t, expected, got)
}
 // what happens when the source is NIL
//  func TestDecodeNilSource(t *testing.T){
// 	//var Nilsource = ;	
// 	var expected []bool = []bool {false, false, false, false}

// 	got, err := decode(keywords, Nilsource)
// 	require.Error(t, err)
// 	assert.Equal(t, expected, got)
//  }

 //What happens when the source or keyword is empty array

 func TestDecodeemptyarraySource(t *testing.T){
	var emptyArraysource = make ([]string,0, 4) 
	var expected []bool = []bool {true, false, true, true}

	got, err := decode(keywords, emptyArraysource)
	require.Error(t, err)
	assert.Equal(t, expected, got)
 }
 //What happens when the source or keyword is nil array

 // what happens when the keyword array had multiple copies of an element (SHould be allowed)
//