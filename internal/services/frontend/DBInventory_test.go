// Unit tests for the web service frontend.
//
// Borrows heavily from the gorilla mux test package.
//
package frontend

import (
	"fmt"
	"log"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
	"github.com/stretchr/testify/assert"
)

// // First DBInventory unit test
func TestListRacks(t *testing.T) {
	unit_test.SetTesting(t)

	request := httptest.NewRequest("GET", "/api/racks/", nil)
	response := doHTTP(request, nil)
	body, err := getBody(response)

	log.Printf("ListRacks: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	// log.Println(string(body))
	s := string(body)
	var splits = strings.Split(s, "\n")
	fmt.Println(splits)
	assert.Equal(t, splits, "[Racks (List) /api/rasks/rack1 /api/racks/rack2 ]", "Error found")

	// func Split(s, )  {

	// }
	// return split(string(s), '\n')
	// Split the string (s) at new line \n.  That will create an array of strings.
	// verify that entry 0 has the racks (List) i.e check with assert.equal
	// verify that entry 1 or entry 2 has /api/racks/rack1
	// verify that entry 1 or entry 2 has /api/racks/rack2
	log.Printf("error: %v", err)
}

// List racks
