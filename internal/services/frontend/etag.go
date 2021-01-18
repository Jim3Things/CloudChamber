package frontend

import (
	"fmt"
	"strconv"
	"strings"
)

// formatAsEtag converts a number into a string formatted to meet the ETag
// specification.
func formatAsEtag(rev int64) string {
	return fmt.Sprintf("\"%d\"", rev)
}

// parseAsMatchTag converts a string into the revision number that would have
// been provided as a previous ETag, or an error if it is unable to parse it.
func parseAsMatchTag(match string) (int64, error) {
	test := strings.Trim(match, "\"")
	return strconv.ParseInt(test, 10, 64)
}
