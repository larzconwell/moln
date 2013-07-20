package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func init() {
	Routes["Devices.Index"] = &Route{[]string{"GET"}, "/users/{user}/devices", ShowDevicesHandler}
	Routes["Devices.New"] = &Route{[]string{"POST"}, "/users/{user}/devices", CreateDeviceHandler}
	Routes["Devices.Show"] = &Route{[]string{"GET"}, "/users/{user}/devices/{name}", ShowDeviceHandler}
	Routes["Devices.Delete"] = &Route{[]string{"DELETE"}, "/users/{user}/devices/{name}", DeleteDeviceHandler}
}

func ShowDevicesHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	user := vars["user"]
	res := Response{}

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

	// Authentication is optional, so we can ignore errors
	authenticated, currentUser, _ := Authenticate(req)
	if currentUser != user {
		authenticated = false
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

func CreateDeviceHandler(rw http.ResponseWriter, req *http.Request) {
	res := Response{}
	params, ok := ParseForm(rw, req, res)
	if !ok {
		return
	}
	name := params.Get("name")
	req.Header.Set("Content-Type", "")
	vars := mux.Vars(req)
	user := vars["user"]

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

	// Authenticate the request, ensure the authenticated user is the correct user
	authenticated, currentUser, err := Authenticate(req)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}
	if authenticated && currentUser != user {
		res["error"] = ErrUserNotAuthorized.Error()
		res.Send(rw, req, http.StatusForbidden)
		return
	}
	if !authenticated {
		rw.Header().Set("WWW-Authenticate", "Token")
		res["error"] = http.StatusText(http.StatusUnauthorized)
		res.Send(rw, req, http.StatusUnauthorized)
		return
	}

	// Ensure data is valid
	errs, err := Validate(func() (error, error) {
		if name == "" {
			return ErrDeviceNameEmpty, nil
		}

		return nil, nil
	}, func() (error, error) {
		exists, err := DeviceExists(user, name)
		if exists {
			return ErrDeviceAlreadyExists, err
		}

		return nil, err
	})
	if errs != nil || err != nil {
		status := http.StatusBadRequest

		if err != nil {
			res["error"] = err.Error()
			status = http.StatusInternalServerError
		}

		if errs != nil {
			res["errors"] = errs
		}

		res.Send(rw, req, status)
		return
	}

	// Create token for device and token
	token, err := GenerateToken(user, name)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	err = CreateDevice(user, name, token)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	err = CreateToken(user, name, token)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	err = AddDeviceToUser(user, name)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	res["device"] = map[string]string{"name": name, "token": token}
	res.Send(rw, req, 0)
}

func ShowDeviceHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	user := vars["user"]
	name := vars["name"]
	res := Response{}

	// Ensure device exists
	exists, err := DeviceExists(user, name)
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

	// Authentication is optional, so we can ignore errors
	authenticated, currentUser, _ := Authenticate(req)
	if currentUser != user {
		authenticated = false
	}

	device, err := GetDevice(user, name)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	if !authenticated {
		device["token"] = ""
	}

	res["device"] = device
	res.Send(rw, req, 0)
}

func DeleteDeviceHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	user := vars["user"]
	name := vars["name"]
	res := Response{}

	// Get device handling devices not found
	device, err := GetDevice(user, name)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}
	_, ok := device["name"]
	if !ok {
		res["error"] = http.StatusText(http.StatusNotFound)
		res.Send(rw, req, http.StatusNotFound)
		return
	}

	// Authenticate the request, ensure the authenticated user is the correct user
	authenticated, currentUser, err := Authenticate(req)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}
	if authenticated && currentUser != user {
		res["error"] = ErrUserNotAuthorized.Error()
		res.Send(rw, req, http.StatusForbidden)
		return
	}
	if !authenticated {
		rw.Header().Set("WWW-Authenticate", "Token")
		res["error"] = http.StatusText(http.StatusUnauthorized)
		res.Send(rw, req, http.StatusUnauthorized)
		return
	}

	err = DeleteToken(device["token"])
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	err = RemoveDeviceFromUser(user, name)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	err = DeleteDevice(user, name)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	res["device"] = device
	res.Send(rw, req, 0)
}
