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
	userName := params.Get("name")
	plainPass := params.Get("password")
	deviceName := params.Get("device")

	ok = Validate(rw, req, func() (error, error) {
		if userName == "" {
			return ErrUserNameEmpty, nil
		}

		return nil, nil
	}, func() (error, error) {
		if plainPass == "" {
			return ErrUserPasswordEmpty, nil
		}

		return nil, nil
  }, func() (error, error) {
    _, ok = params["device"]
    if ok && deviceName == "" {
      return ErrDeviceNameEmpty, nil
    }

    return nil, nil
  })
	if !ok {
		return
	}

	res := &httpextra.Response{rw, req}
	res.Send(params, 200)
}
