package main

import (
	"net/http"
)

// Route represents the routing details needed to map the path
// corrrectly
type Route struct {
	Name    string
	Path    string
	Methods []string
	Handler http.HandlerFunc
}
