package frontend

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Input and output for a success case
var views = "Defined, Observed"
var response []string
var expected []bool = []bool {true, false, false, true}

// Input and output for a faiure case
var badView = "AbaddefinedobservedValue" 
var multipleview = "Defined, Defined, Observed, Actual"
var badexpected []bool = nil

func TestDecode(t *testing.T) {

	//var views = []string{"Defined", "Observed"}

	got, err := decode(keywords, views)
	require.NoError(t,err)
	//assert.Equal(t,4,len(got))
	assert.Equal(t,expected, got)
}

func TestDecodeBadView(t *testing.T) {

	//var views = []string{"Defined", "Observed"}

	got, err := decode(keywords, badView)
	require.Error(t, err)
	assert.Equal(t, badexpected, got)
}

func TestDecodeMultipleView(t *testing.T) {
	got, err := decode(keywords, multipleview)
	require.NoError(t, err)
	assert.NotEqual(t, expected,got)
}
