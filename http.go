package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// TLS represents the cert and key files for TLS connections
type TLS struct {
	Cert string
	Key  string
}

// Simple 404 handler
func NotFoundHandler(rw http.ResponseWriter, req *http.Request) {
	headers := rw.Header()
	headers.Set("content-type", "application/json; charset=utf-8")

	rw.WriteHeader(http.StatusNotFound)
	fmt.Fprint(rw, "{\"error\": \"Not Found\"}")
}

// LogHandler logs the request/response to the given io.Writer in Common Log Format
type LogHandler struct {
	writer  io.Writer
	handler http.Handler
}

func NewLogHandler(writer io.Writer, handler http.Handler) http.Handler {
	return &LogHandler{
		writer:  writer,
		handler: handler,
	}
}

// ServeHTTP Logs the request and calls the handler
func (lh *LogHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	accessTime := time.Now()

	loggedWriter := &responseLogger{responseWriter: rw}
	lh.handler.ServeHTTP(loggedWriter, req)

	// Get username field if given
	username := "-"
	if req.URL.User != nil {
		name := req.URL.User.Username()
		if name != "" {
			username = name
		}
	}

	fmt.Fprintf(lh.writer, "%s %s - [%s] \"%s %s %s\" %d %d\n",
		req.RemoteAddr,
		username,
		accessTime.Format("02/Jan/2006:15:04:05 -0700"),
		req.Method,
		req.RequestURI,
		req.Proto,
		loggedWriter.status,
		loggedWriter.length)
}

// responseLogger implements http.ResponseWriter, keeps log of status code and content length
type responseLogger struct {
	responseWriter http.ResponseWriter
	status         int
	length         int
}

func (rl *responseLogger) Header() http.Header {
	return rl.responseWriter.Header()
}

// Write calls the responseWriters write, and keeps track of the content length
func (rl *responseLogger) Write(b []byte) (int, error) {
	if rl.status == 0 {
		// The status will be StatusOk if WriteHeader has not been called
		rl.status = http.StatusOK
	}

	size, err := rl.responseWriter.Write(b)
	rl.length += size
	return size, err
}

// WriteHeader sets the status code, and writes the header to the set reponseWriter
func (rl *responseLogger) WriteHeader(status int) {
	rl.responseWriter.WriteHeader(status)
	rl.status = status
}
