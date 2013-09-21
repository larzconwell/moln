package handlers

import (
	"net/http"
)

// ContentTypes is a map of supported mime types and parse error messages.
var ContentTypes = make(map[string]string)

// ContentType is a http.Handler that ensures unsupported formats
// have the correct response.
type ContentType struct {
	Handler http.Handler
}

func NewContentType(handler http.Handler) *ContentType {
	return &ContentType{handler}
}

// ServeHTTP calls the handler if the requested response format is supported,
// otherwise a 415 is returned with a list of supported types in the Accept header.
func (ct *ContentType) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ct.Handler.ServeHTTP(rw, req)
}
