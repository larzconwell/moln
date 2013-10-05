package main

import (
	"github.com/larzconwell/moln/httpextra"
	"net/http"
)

func init() {
	CreateUser := &Route{"CreateUser", "/user", []string{"POST"}, CreateUserHandler}

	Routes = append(Routes, CreateUser)
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
