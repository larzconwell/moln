package main

import (
	"github.com/gorilla/mux"
	"github.com/moln/httpextra"
	"net/http"
)

func init() {
	createDevice := &Route{"CreateDevice", "/devices", []string{"POST"}, CreateDeviceHandler}
	getDevices := &Route{"GetDevices", "/devices", []string{"GET"}, GetDevicesHandler}
	getDevice := &Route{"GetDevice", "/devices/{name}", []string{"GET"}, GetDeviceHandler}
	deleteDevice := &Route{"DeleteDevice", "/devices/{name}", []string{"DELETE"}, DeleteDeviceHandler}

	Routes = append(Routes, createDevice, getDevices, getDevice, deleteDevice)
}

func CreateDeviceHandler(rw http.ResponseWriter, req *http.Request) {
	params, ok := httpextra.ParseForm(ContentTypes, rw, req)
	if !ok {
		return
	}
	conn := Pool.Get()
	defer conn.Close()

	user := Authenticate(conn, rw, req)
	if user == nil {
		return
	}

	device := &Device{Conn: conn, Name: params.Get("name"), User: user}
	errs, err := device.Validate(true)
	ok = HandleValidations(rw, req, errs, err)
	if !ok {
		return
	}
	res := &httpextra.Response{ContentTypes, rw, req}

	err = device.Save(true)
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	activity := &Activity{Conn: conn, Message: "Created device " + device.Name, User: user}
	err = activity.Save()
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	res.Send(device, http.StatusOK)
}

func GetDevicesHandler(rw http.ResponseWriter, req *http.Request) {
	conn := Pool.Get()
	defer conn.Close()

	user := Authenticate(conn, rw, req)
	if user == nil {
		return
	}
	res := &httpextra.Response{ContentTypes, rw, req}

	devices, err := conn.GetDevices(user.Name)
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	res.Send(devices, http.StatusOK)
}

func GetDeviceHandler(rw http.ResponseWriter, req *http.Request) {
	conn := Pool.Get()
	defer conn.Close()

	user := Authenticate(conn, rw, req)
	if user == nil {
		return
	}
	name := mux.Vars(req)["name"]
	res := &httpextra.Response{ContentTypes, rw, req}

	device, err := conn.GetDevice(user.Name, name)
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
	conn := Pool.Get()
	defer conn.Close()

	user := Authenticate(conn, rw, req)
	if user == nil {
		return
	}
	name := mux.Vars(req)["name"]
	res := &httpextra.Response{ContentTypes, rw, req}

	device, err := conn.GetDevice(user.Name, name)
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

	activity := &Activity{Conn: conn, Message: "Deleted device " + device.Name, User: user}
	err = activity.Save()
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	res.Send(device, http.StatusOK)
}
