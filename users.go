package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func init() {
	Routes["Users.New"] = &Route{[]string{"POST"}, "/users", CreateUserHandler}
	Routes["Users.Show"] = &Route{[]string{"GET"}, "/users/{name}", ShowUserHandler}
	Routes["Users.Delete"] = &Route{[]string{"DELETE"}, "/users/{name}", DeleteUserHandler}
	Routes["Users.Update"] = &Route{[]string{"PUT"}, "/users/{name}", UpdateUserHandler}
}

func CreateUserHandler(rw http.ResponseWriter, req *http.Request) {
	res := Response{}
	params, ok := ParseForm(rw, req, res)
	if !ok {
		return
	}
	name := params.Get("name")
	plainPass := params.Get("password")
	deviceName := params.Get("devicename")
	req.Header.Set("Content-Type", "")

	// Ensure data is valid
	errs, err := Validate(func() (error, error) {
		if name == "" {
			return ErrUserNameEmpty, nil
		}

		return nil, nil
	}, func() (error, error) {
		if plainPass == "" {
			return ErrUserPasswordEmpty, nil
		}

		return nil, nil
	}, func() (error, error) {
		if deviceName == "" {
			return ErrDeviceNameEmpty, nil
		}

		return nil, nil
	}, func() (error, error) {
		exists, err := UserExists(name)
		if exists {
			return ErrUserAlreadyExists, err
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

	// Generate password
	password, err := HashPass(plainPass)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	err = CreateUser(name, string(password))
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	// Create token for device and token
	token, err := GenerateToken(name, deviceName)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	err = CreateDevice(name, deviceName, token)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	err = CreateToken(name, deviceName, token)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	err = AddDeviceToUser(name, deviceName)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	res["user"] = map[string]interface{}{
		"name":    name,
		"devices": []map[string]string{map[string]string{"name": deviceName, "token": token}},
	}
	res.Send(rw, req, 0)
}

func ShowUserHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	res := Response{}

	// Authentication is optional, so we can ignore errors
	authenticated, currentUser, _ := Authenticate(req)
	if currentUser != name {
		authenticated = false
	}

	// Ensure user exists
	exists, err := UserExists(name)
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

	devices, err := GetUserDevices(name, nil)
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

	res["user"] = map[string]interface{}{"name": name, "devices": devices}
	res.Send(rw, req, 0)
}

func DeleteUserHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	res := Response{}

	// Authenticate the request, ensure the authenticated user is the correct user
	authenticated, currentUser, err := Authenticate(req)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}
	if authenticated && currentUser != name {
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

	// Ensure user exists
	exists, err := UserExists(name)
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

	// Delete the device and the devices token
	iterator := func(device map[string]string) error {
		err = DeleteDevice(name, device["name"])
		if err != nil {
			return err
		}

		err = DeleteToken(device["token"])
		return err
	}

	devices, err := GetUserDevices(name, iterator)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	err = DeleteUserDevices(name)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	err = DeleteUser(name)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	res["user"] = map[string]interface{}{"name": name, "devices": devices}
	res.Send(rw, req, 0)
}

func UpdateUserHandler(rw http.ResponseWriter, req *http.Request) {
	res := Response{}
	params, ok := ParseForm(rw, req, res)
	if !ok {
		return
	}
	plainPass := params.Get("password")
	req.Header.Set("Content-Type", "")
	vars := mux.Vars(req)
	name := vars["name"]

	// Authenticate the request, ensure the authenticated user is the correct user
	authenticated, currentUser, err := Authenticate(req)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}
	if authenticated && currentUser != name {
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

	// Ensure user exists
	exists, err := UserExists(name)
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

	// Ensure data is valid
	errs, err := Validate(func() (error, error) {
		if plainPass == "" {
			return ErrUserPasswordEmpty, nil
		}

		return nil, nil
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

	// Generate password
	password, err := HashPass(plainPass)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	err = UpdateUser(name, "password", string(password))
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	devices, err := GetUserDevices(name, nil)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	res["user"] = map[string]interface{}{"name": name, "devices": devices}
	res.Send(rw, req, 0)
}
