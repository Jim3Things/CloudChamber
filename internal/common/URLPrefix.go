package common

import (
	"net/http"
	"strings"
)

// URLPrefix extracts the URL used by the request, and returns it after
// ensuring that it has a trailing '/'
func URLPrefix(r *http.Request) string {
	b := r.URL.String()
	if !strings.HasSuffix(b, "/") {
		b += "/"
	}
	return b
}

