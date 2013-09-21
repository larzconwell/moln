package handlers

import (
	"fmt"
	"net/http"
)

// NotFound is a simple 404 handler.
var NotFound = http.HandlerFunc(notFound)

func notFound(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(200)
	fmt.Fprint(rw, "404 not found")
}
