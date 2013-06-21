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
func (lh *LogHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	startTime := time.Now()

	loggedWriter := &responseLogger{w: w}
	lh.handler.ServeHTTP(loggedWriter, req)

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
		startTime.Format("02/Jan/2006:15:04:05 -0700"),
		req.Method,
		req.RequestURI,
		req.Proto,
		loggedWriter.status,
		loggedWriter.size)
}

type responseLogger struct {
	w      http.ResponseWriter
	status int
	size   int
}

func (rl *responseLogger) Header() http.Header {
	return rl.w.Header()
}

func (rl *responseLogger) Write(b []byte) (int, error) {
	if rl.status == 0 {
		// The status will be StatusOk if WriteHeader has not been called
		rl.status = http.StatusOK
	}

	size, err := rl.w.Write(b)
	rl.size += size
	return size, err
}

func (rl *responseLogger) WriteHeader(s int) {
	rl.w.WriteHeader(s)
	rl.status = s
}
