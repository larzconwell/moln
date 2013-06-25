package main

import (
	"net/http"
)

func init() {
	Routes["Users.New"] = &Route{Methods: []string{"POST"}, Path: "/users", Handler: CreateUserHandler}
}

func CreateUserHandler(rw http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		ErrResponse(rw, http.StatusInternalServerError, err)
		return
	}
	//params := req.PostForm
	//username := params.Get("username")
	//password := []byte(params.Get("password"))
	//deviceName := params.Get("devicename")
}
