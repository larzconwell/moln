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
	res := &httpextra.Response{rw, req}

	user := &User{params.Get("name"), params.Get("password")}
	errs, err := user.Validate(true)
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
