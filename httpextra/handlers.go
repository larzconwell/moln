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

// NotFoundHandler is a http.Handler responding to 404s.
type NotFoundHandler struct {
	ContentTypes map[string]*ContentType
}

func NewNotFoundHandler(contentTypes map[string]*ContentType) *NotFoundHandler {
	return &NotFoundHandler{contentTypes}
}

// ServeHTTP responds with 404
func (nfh *NotFoundHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	res := &Response{nfh.ContentTypes, rw, req}
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

	// Filter out password queries
	query := req.URL.Query()
	for k, v := range query {
		replace := ""

		if k == "password" {
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
	ContentTypes map[string]*ContentType
	Handler      http.Handler
}

func NewContentTypeHandler(contentTypes map[string]*ContentType, handler http.Handler) *ContentTypeHandler {
	return &ContentTypeHandler{contentTypes, handler}
}

// ServeHTTP calls the handler if a requested response content type is supported,
// otherwise a 406 is returned.
func (cth *ContentTypeHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if RequestContentType(cth.ContentTypes, req) != nil {
		cth.Handler.ServeHTTP(rw, req)
	} else {
		res := &Response{cth.ContentTypes, rw, req}
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
