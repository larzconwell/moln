package main

import (
	"net/http"
)

var (
	Routes map[string]*Route = map[string]*Route{}
)

// Route represents a single route, and the details for it
type Route struct {
	Methods []string
	Path    string
	Handler http.HandlerFunc
}
