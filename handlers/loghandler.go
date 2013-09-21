// Package handlers implements routines to handle incoming HTTP
// requests.
package handlers

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// LogHandler is a http.Handler that logs requests in Common Log Format.
type LogHandler struct {
	Writer  io.Writer
	Handler http.Handler
}

func NewLogHandler(writer io.Writer, handler http.Handler) *LogHandler {
	return &LogHandler{writer, handler}
}

// ServeHTTP logs the request and calls the handler.
func (lh *LogHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	accessTime := time.Now()
	loggedWriter := &responseLogger{ResponseWriter: rw}

	lh.Handler.ServeHTTP(loggedWriter, req)

	// Set the username for logging
	username := "-"
	if req.URL.User != nil {
		name := req.URL.User.Username()
		if name != "" {
			username = name
		}
	}

	// Filter out token and password queries
	query := req.URL.Query()
	for k, v := range query {
		replace := ""

		if k == "password" || k == "token" {
			replace = k
		}

		if replace != "" {
			for idx := range v {
				v[idx] = k
			}
		}
	}
	req.URL.RawQuery = query.Encode()

	fmt.Fprintf(lh.Writer, "%s %s - [%s] \"%s %s %s\" %d %d\n",
		req.RemoteAddr,
		username,
		accessTime.Format("02/Jan/2006:15:04:05 -0700"),
		req.Method,
		req.URL.String(),
		req.Proto,
		loggedWriter.Status,
		loggedWriter.Length)
}

// responseLogger is a http.ResponseWriter, it keeps the state of the responses
// status code and content length.
type responseLogger struct {
	ResponseWriter http.ResponseWriter
	Status         int
	Length         int
}

// Header returns the responses headers.
func (rl *responseLogger) Header() http.Header {
	return rl.ResponseWriter.Header()
}

// WriteHeader writes the header, keeping track of the status code.
func (rl *responseLogger) WriteHeader(status int) {
	rl.ResponseWriter.WriteHeader(status)
}

// Write writes the response and keeps track of the content length.
func (rl *responseLogger) Write(b []byte) (int, error) {
	// If no status has been written default to OK
	if rl.Status == 0 {
		rl.Status = http.StatusOK
	}

	size, err := rl.ResponseWriter.Write(b)
	rl.Length += size
	return size, err
}
