package main

import (
	"fmt"
	"net/http"
)

func init() {
	Routes["Users.New"] = &Route{Methods: []string{"POST"}, Path: "/users", Handler: CreateUser}
}

func CreateUser(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprint(rw, req.Method)
}
