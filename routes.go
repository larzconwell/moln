package main

import (
	"net/http"
)

var Routes map[string]*Route

func init() {
	Routes = make(map[string]*Route)
}

// Route represents a single route, including URL and it's http.HandlerFunc
type Route struct {
	Method  string
	URL     string
	Handler http.HandlerFunc
}

// Add a *Route to a routes map
func AddRoute(URL string, route *Route, routes map[string]*Route) {
	if route.URL == "" {
		route.URL = URL
	}

	routes[URL] = route
}
