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
func TestInventoryListRacks(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	request := httptest.NewRequest("GET", "/api/racks/", nil)
	response := doHTTP(request, nil)
	body, err := getBody(response)

	log.Printf("ListRacks: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))

	s := string(body)                   // Converted into a string
	var splits = strings.Split(s, "\n") // Created an array per line
	fmt.Println(splits)                 // just a print statement
	expected := []string{"/api/racks/rack1", "/api/racks/rack2", ""}

	assert.Equal(t, splits[0], "Racks (List)")
	assert.ElementsMatch(t, expected, splits[1:])

	log.Printf("error: %v", err)
}
