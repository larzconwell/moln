// Package httpextra implements routines and types that allow more control
// over incoming requests and outgoing responses.
package httpextra

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// NotFoundHandler is a simple 404 http.Handler.
var NotFoundHandler = http.HandlerFunc(notFoundHandler)

func notFoundHandler(rw http.ResponseWriter, req *http.Request) {
	res := &Response{rw, req}
	res.Send(map[string]string{"error": http.StatusText(http.StatusNotFound)}, http.StatusNotFound)
}

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
	loggedWriter := &ResponseLogger{ResponseWriter: rw}

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

// ContentTypeHandler is a http.Handler that ensures unsupported formats
// have the correct response.
type ContentTypeHandler struct {
	Handler http.Handler
}

func NewContentTypeHandler(handler http.Handler) *ContentTypeHandler {
	return &ContentTypeHandler{handler}
}

// ServeHTTP calls the handler if a requested response content type is supported,
// otherwise a 406 is returned.
func (cth *ContentTypeHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if RequestContentType(req) != nil {
		cth.Handler.ServeHTTP(rw, req)
	} else {
		res := &Response{rw, req}
		res.SendDefault(map[string]string{"error": http.StatusText(http.StatusNotAcceptable)},
			http.StatusNotAcceptable)
	}
}

// SlashHandler is a http.Handler that removes trailing slashes
type SlashHandler struct {
	Handler http.Handler
}

func NewSlashHandler(handler http.Handler) *SlashHandler {
	return &SlashHandler{handler}
}

// ServeHTTP stripes a trailing slash and calls the handler
func (sh *SlashHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		req.URL.Path = strings.TrimRight(req.URL.Path, "/")
		req.RequestURI = req.URL.RequestURI()
	}

	sh.Handler.ServeHTTP(rw, req)
}
