package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ContentTypes is a map of types and their parse error content
var ContentTypes = map[string]string{
	"application/json":                  "{\"error\": \"{{message}}\"}",
	"application/x-www-form-urlencoded": "",
	"multipart/form-data":               "",
}

// TLS represents the cert and key files for TLS connections
type TLS struct {
	Cert string
	Key  string
}

// Response is a map of key values to send as the requests Content-Type
type Response map[string]interface{}

// Send parses the response and if parsed successfully responds with the given status code,
// otherwise a 500 and an error message is returned
func (res Response) Send(rw http.ResponseWriter, req *http.Request, status int) {
	var (
		contents []byte
		err      error
	)
	reqCt := req.Header.Get("Content-Type")
	rwHeader := rw.Header()

	if reqCt == "" {
		reqCt = "application/json"
	}
	rwHeader.Set("Content-Type", reqCt)

	if status <= 0 {
		status = http.StatusOK
	}

	switch reqCt {
	case "application/json":
		contents, err = json.Marshal(res)
	}

	if err != nil {
		status = http.StatusInternalServerError
		contents = []byte(strings.Replace(ContentTypes[reqCt], "{{message}}", err.Error(), 1))
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
	res.Send(rw, req, http.StatusNotFound)
}

// ContentTypeHandler is a http.HandlerFunc to ensure only valid content types are sent
type ContentTypeHandler struct {
	handler http.Handler
}

func NewContentTypeHandler(handler http.Handler) *ContentTypeHandler {
	return &ContentTypeHandler{handler: handler}
}

// ServeHTTP only calls the handler if the request content type is supported or empty,
// otherwise a 415 is returned with the supported types in the Accept header
func (cth *ContentTypeHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ct := req.Header.Get("Content-Type")
	rwHeader := rw.Header()

	if _, ok := ContentTypes[ct]; ct == "" || ok {
		cth.handler.ServeHTTP(rw, req)
		return
	}

	rwHeader.Set("Content-Type", "application/json")
	for t := range ContentTypes {
		rwHeader.Add("Accept", t)
	}

	rw.WriteHeader(http.StatusUnsupportedMediaType)
	fmt.Fprint(rw, "{\"error\": \""+http.StatusText(http.StatusUnsupportedMediaType)+"\"}")
}

// LogHandler is a http.Handler that logs requests to the writer in Common Log Format
type LogHandler struct {
	writer  io.Writer
	handler http.Handler
}

func NewLogHandler(writer io.Writer, handler http.Handler) *LogHandler {
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
