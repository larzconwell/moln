package main

import (
	"github.com/gorilla/mux"
	"github.com/larzconwell/moln/httpextra"
	"net/http"
)

func init() {
	createDevice := &Route{"CreateDevice", "/devices", []string{"POST"}, CreateDeviceHandler}
	getDevice := &Route{"GetDevice", "/devices/{name}", []string{"GET"}, GetDeviceHandler}
	deleteDevice := &Route{"DeleteDevice", "/devices/{name}", []string{"DELETE"}, DeleteDeviceHandler}

	Routes = append(Routes, createDevice, getDevice, deleteDevice)
}

func CreateDeviceHandler(rw http.ResponseWriter, req *http.Request) {
	user := Authenticate(rw, req)
	if user == nil {
		return
	}

	params, ok := httpextra.ParseForm(rw, req)
	if !ok {
		return
	}

	device := &Device{Name: params.Get("name"), User: user}
	errs, err := device.Validate(true)
	ok = HandleValidations(rw, req, errs, err)
	if !ok {
		return
	}
	res := &httpextra.Response{rw, req}

	err = device.Save(true)
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	res.Send(device, http.StatusOK)
}

func GetDeviceHandler(rw http.ResponseWriter, req *http.Request) {
	user := Authenticate(rw, req)
	if user == nil {
		return
	}
	name := mux.Vars(req)["name"]
	res := &httpextra.Response{rw, req}

	device, err := DB.GetDevice(user.Name, name)
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	if device == nil {
		res.Send(map[string]string{"error": http.StatusText(http.StatusNotFound)}, http.StatusNotFound)
		return
	}

	res.Send(device, http.StatusOK)
}

func DeleteDeviceHandler(rw http.ResponseWriter, req *http.Request) {
	user := Authenticate(rw, req)
	if user == nil {
		return
	}
	name := mux.Vars(req)["name"]
	res := &httpextra.Response{rw, req}

	device, err := DB.GetDevice(user.Name, name)
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	if device == nil {
		res.Send(map[string]string{"error": http.StatusText(http.StatusNotFound)}, http.StatusNotFound)
		return
	}
	device.User = user

	err = device.Delete()
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	res.Send(device, http.StatusOK)
}
