package httpextra

import (
	"fmt"
	"net/http"
	"strings"
)

// Response represents a structure that marshals content
// to send to the client.
type Response struct {
	RW  http.ResponseWriter
	Req *http.Request
}

// Send parses the response and if parsed successfully responds with the given status
// code, otherwise a 500 is sent with the formats default error message.
func (res *Response) Send(data interface{}, status int) {
	res.send(data, status, RequestContentType(res.Req))
}

// SendDefault does the same as Send, except uses the default content type for the
// response format.
func (res *Response) SendDefault(data interface{}, status int) {
	res.send(data, status, DefaultContentType())
}

// send does the actual marshal and response
func (res *Response) send(data interface{}, status int, ct *ContentType) {
	var (
		contents []byte
		err      error
	)

	if ct == nil {
		res.RW.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(res.RW, "No response content type available")
		return
	}

	res.RW.Header().Set("Content-Type", ct.Mime)
	contents, err = ct.Marshal(data)
	if err != nil {
		status = http.StatusInternalServerError
		contents = []byte(strings.Replace(ct.Error, "{{message}}", err.Error(), -1))
	}

	res.RW.WriteHeader(status)
	fmt.Fprint(res.RW, string(contents))
}
