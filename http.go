package main

import (
	"encoding/json"
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

// Response is a map of key values to send as a JSON response
type Response map[string]interface{}

// Send parses the response and if parsed successfully responsed with give status code otherwise
// 500, and a parse error message
func (res Response) Send(rw http.ResponseWriter, status int) {
	rw.Header().Set("Content-Type", "application/json")

	if status <= 0 {
		status = http.StatusOK
	}

	contents, err := json.Marshal(res)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(rw, "{\"error\": \""+err.Error()+"\"}")
		return
	}

	rw.WriteHeader(status)
	fmt.Fprint(rw, string(contents))
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

/*
 * Handlers
 */

// NotFoundHandler is a simple 404 http.HandlerFunc
func NotFoundHandler(rw http.ResponseWriter, req *http.Request) {
	res := Response{"error": http.StatusText(http.StatusNotFound)}
	res.Send(rw, http.StatusNotFound)
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
	res := Response{"error": http.StatusText(http.StatusMethodNotAllowed)}
	res.Send(rw, http.StatusMethodNotAllowed)
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

	// Filter out token and password queries
	query := req.URL.Query()
	for k, v := range query {
		replace := ""

		if k == "password" || k == "token" {
			replace = k
		}

		if replace != "" {
			for idx := range v {
				v[idx] = "[" + k + "]"
			}
		}
	}
	req.URL.RawQuery = query.Encode()

	fmt.Fprintf(lh.writer, "%s %s - [%s] \"%s %s %s\" %d %d\n",
		req.RemoteAddr,
		username,
		accessTime.Format("02/Jan/2006:15:04:05 -0700"),
		req.Method,
		req.URL.String(),
		req.Proto,
		loggedWriter.status,
		loggedWriter.length)
}
