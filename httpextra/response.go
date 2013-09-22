package httpextra

import (
	"fmt"
	"net/http"
	"strings"
)

// Response represents a map of key values to send
// to the client.
type Response map[string]interface{}

// Send parses the response and if parsed successfully responsds with the given status
// code, otherwise a 500 is sent with the formats default error message.
func (res *Response) Send(rw http.ResponseWriter, req *http.Request, status int) {
	var (
		contents []byte
		err      error
	)

	ct := RequestContentType(req)
	if ct == nil {
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(rw, "No response content type available")
		return
	}

	rw.Header().Set("Content-Type", ct.Mime)
	contents, err = ct.Marshal(res)
	if err != nil {
		status = http.StatusInternalServerError
		contents = []byte(strings.Replace(ct.Error, "{{message}}", err.Error(), -1))
	}

	rw.WriteHeader(status)
	fmt.Fprint(rw, string(contents))
}

// SendDefault does the same as Send, except uses the default content type for the
// response format.
func (res *Response) SendDefault(rw http.ResponseWriter, req *http.Request, status int) {
	var (
		contents []byte
		err      error
	)

	ct := DefaultContentType()
	if ct == nil {
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(rw, "No response content type available")
		return
	}

	rw.Header().Set("Content-Type", ct.Mime)
	contents, err = ct.Marshal(res)
	if err != nil {
		status = http.StatusInternalServerError
		contents = []byte(strings.Replace(ct.Error, "{{message}}", err.Error(), -1))
	}

	rw.WriteHeader(status)
	fmt.Fprint(rw, string(contents))
}
