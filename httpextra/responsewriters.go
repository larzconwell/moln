package httpextra

import (
	"net/http"
)

// ResponseLogger is a http.ResponseWriter, it keeps the state of the responses
// status code and content length.
type ResponseLogger struct {
	ResponseWriter http.ResponseWriter
	Status         int
	Length         int
}

// Header returns the responses headers.
func (rl *ResponseLogger) Header() http.Header {
	return rl.ResponseWriter.Header()
}

// WriteHeader writes the header, keeping track of the status code.
func (rl *ResponseLogger) WriteHeader(status int) {
	rl.ResponseWriter.WriteHeader(status)
	rl.Status = status
}

// Write writes the response and keeps track of the content length.
func (rl *ResponseLogger) Write(b []byte) (int, error) {
	// If no status has been written default to OK
	if rl.Status == 0 {
		rl.Status = http.StatusOK
	}

	size, err := rl.ResponseWriter.Write(b)
	rl.Length += size
	return size, err
}
