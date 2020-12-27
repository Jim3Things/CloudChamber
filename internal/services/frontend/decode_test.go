package frontend

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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