package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Err is an single error message
type Err struct {
	Message string `json:"error"`
}

// MultiErr has multiple error messages
type MultiErr struct {
	Messages []string `json:"errors"`
}

// Send an JSON error response with the given status and error message
func ErrResponse(rw http.ResponseWriter, status int, err error) {
	if err == nil {
		err = errors.New(http.StatusText(status))
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(status)

	encoder := json.NewEncoder(rw)
	e := encoder.Encode(&Err{err.Error()})
	if e != nil {
		fmt.Fprint(rw, "{\"error\": \""+e.Error()+"\"}")
	}
}

// Send a JSON error response with multiple errors
func MultiErrResponse(rw http.ResponseWriter, status int, errs []error) {
	errMessages := make([]string, len(errs))

	for idx, err := range errs {
		errMessages[idx] = err.Error()
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(status)

	encoder := json.NewEncoder(rw)
	e := encoder.Encode(&MultiErr{errMessages})
	if e != nil {
		fmt.Fprint(rw, "{\"error\": \""+e.Error()+"\"}")
	}
}
