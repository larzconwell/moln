package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// TLS represents the cert and key files for TLS connections
type TLS struct {
	Cert string
	Key  string
}

// NotFoundHandler is a simple 404 http.HandlerFunc
func NotFoundHandler(rw http.ResponseWriter, req *http.Request) {
	ErrResponse(rw, http.StatusNotFound, nil)
}

// MethodHandler is a http.Handler to manage allowed methods to a resource
type MethodHandler struct {
	Allowed []string
	handler http.Handler
}

func NewMethodHandler(allowed []string, handler http.Handler) http.Handler {
	return &MethodHandler{
		Allowed: allowed,
		handler: handler,
	}
}

// ServeHTTP Only calls the handler if the request method is allowed, otherwise a 405
// response is sent
func (mh *MethodHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for _, method := range mh.Allowed {
		if req.Method == method {
			mh.handler.ServeHTTP(rw, req)
			return
		}
	}

	rw.Header().Set("Allow", strings.Join(mh.Allowed, ", "))
	ErrResponse(rw, http.StatusMethodNotAllowed, nil)
}

// LogHandler is a http.Handler that logs requests to the writer in Common Log Format
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

// responseLogger is a http.ResponseWriter it keeps status code and content length
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
