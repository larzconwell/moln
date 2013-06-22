package main

import (
	"fmt"
	"net/http"
)

func init() {
	Routes["Users.New"] = &Route{Method: "POST", Path: "/users", Handler: CreateUser}
}

func CreateUser(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprint(rw, req.Method)
}
