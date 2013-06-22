package main

import (
	"net/http"
)

// Route represents a single route, and the details for it
type Route struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}
