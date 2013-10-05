package main

import (
	"github.com/larzconwell/moln/httpextra"
	"net/http"
)

func init() {
	createUser := &Route{"CreateUser", "/user", []string{"POST"}, CreateUserHandler}
	getUser := &Route{"GetUser", "/user", []string{"GET"}, GetUserHandler}
	updateUser := &Route{"UpdateUser", "/user", []string{"PUT"}, UpdateUserHandler}

	Routes = append(Routes, createUser, getUser, updateUser)
}

func CreateUserHandler(rw http.ResponseWriter, req *http.Request) {
	params, ok := httpextra.ParseForm(rw, req)
	if !ok {
		return
	}
	_, deviceGiven := params["device"]
	res := &httpextra.Response{rw, req}

	user := &User{params.Get("name"), params.Get("password")}
	errs, err := user.Validate(true)
	ok = HandleValidations(rw, req, errs, err)
	if !ok {
		return
	}

	device := &Device{Name: params.Get("device"), User: user}
	if deviceGiven {
		// Set new to false, since we don't need to check for existance
		errs, err := device.Validate(false)
		ok = HandleValidations(rw, req, errs, err)
		if !ok {
			return
		}
	}

	err = user.Save(true)
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	if deviceGiven {
		err = device.Save(true)
		if err != nil {
			res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}

		res.Send(map[string]interface{}{"user": user, "device": device}, http.StatusOK)
	} else {
		res.Send(user, http.StatusOK)
	}
}

func GetUserHandler(rw http.ResponseWriter, req *http.Request) {
	user := Authenticate(rw, req)
	if user == nil {
		return
	}
	res := &httpextra.Response{rw, req}

	devices, err := DB.GetDevices(user.Name)
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	res.Send(map[string]interface{}{"user": user, "devices": devices}, http.StatusOK)
}

func UpdateUserHandler(rw http.ResponseWriter, req *http.Request) {
	params, ok := httpextra.ParseForm(rw, req)
	if !ok {
		return
	}
	_, passwordGiven := params["password"]

	user := Authenticate(rw, req)
	if user == nil {
		return
	}
	res := &httpextra.Response{rw, req}

	if !passwordGiven {
		res.Send(user, http.StatusOK)
		return
	}

	user.Password = params.Get("password")
	errs, err := user.Validate(false)
	ok = HandleValidations(rw, req, errs, err)
	if !ok {
		return
	}

	err = user.Save(true)
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	res.Send(user, http.StatusOK)
}
