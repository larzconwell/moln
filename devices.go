package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func init() {
	Routes["Devices.index"] = &Route{[]string{"GET"}, "/users/{user}/devices", ShowDevicesHandler}
}

func ShowDevicesHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	user := vars["user"]
	res := Response{}

	// Authentication is optional, so we can ignore errors
	authenticated, currentUser, _ := Authenticate(req)
	if currentUser != user {
		authenticated = false
	}

	// Ensure user exists
	exists, err := UserExists(user)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}
	if !exists {
		res["error"] = http.StatusText(http.StatusNotFound)
		res.Send(rw, req, http.StatusNotFound)
		return
	}

	devices, err := GetUserDevices(user, nil)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	for _, v := range devices {
		if !authenticated {
			v["token"] = ""
		}
	}

	res["devices"] = devices
	res.Send(rw, req, 0)
}
