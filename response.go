package main

import (
	"fmt"
	"net/http"
)

// Send an JSON error response with the given status and/or message
func ErrResponse(rw http.ResponseWriter, status int, message string) {
	if message == "" {
		message = http.StatusText(status)
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(status)
	fmt.Fprint(rw, "{\"error\": \""+message+"\"}")
}
