package main

import (
	"fmt"
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
	params := req.PostForm
	username := params.Get("username")
	password := []byte(params.Get("password"))
	deviceName := params.Get("devicename")

	user, err := NewUser(username, password)
	if err != nil {
		ErrResponse(rw, http.StatusInternalServerError, err)
		return
	}

	errs, err := user.Validate()
	if err != nil {
		ErrResponse(rw, http.StatusInternalServerError, err)
		return
	}
	if errs != nil {
		MultiErrResponse(rw, http.StatusBadRequest, errs)
		return
	}

	err = user.Save()
	if err != nil {
		ErrResponse(rw, http.StatusInternalServerError, err)
		return
	}

	device, err := user.AddDevice(deviceName)
	if err != nil {
		ErrResponse(rw, http.StatusInternalServerError, err)
		return
	}

	errs, err = device.Validate()
	if err != nil {
		ErrResponse(rw, http.StatusInternalServerError, err)
		return
	}
	if errs != nil {
		MultiErrResponse(rw, http.StatusBadRequest, errs)
		return
	}

	err = device.Save()
	if err != nil {
		ErrResponse(rw, http.StatusInternalServerError, err)
		return
	}

	fmt.Fprintln(rw, user)
	fmt.Fprintln(rw, device)

	// Encode user and return it
}
