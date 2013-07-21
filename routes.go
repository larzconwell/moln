package main

import (
	"net/http"
)

// Route represents a single route, and the details for it
type Route struct {
	Methods []string
	Path    string
	Handler http.HandlerFunc
}
