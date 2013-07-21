package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

func init() {
	Routes["Users.New"] = &Route{[]string{"POST"}, "/users", CreateUserHandler}
	Routes["Users.Show"] = &Route{[]string{"GET"}, "/users/{name}", ShowUserHandler}
	Routes["Users.Delete"] = &Route{[]string{"DELETE"}, "/users/{name}", DeleteUserHandler}
	Routes["Users.Update"] = &Route{[]string{"PUT"}, "/users/{name}", UpdateUserHandler}
}

func CreateUserHandler(rw http.ResponseWriter, req *http.Request) {
	res := make(Response, 0)
	params, ok := ParseForm(rw, req, res)
	if !ok {
		return
	}
	name := strings.ToLower(params.Get("name"))
	plainPass := params.Get("password")
	deviceName := strings.ToLower(params.Get("devicename"))
	req.Header.Set("Content-Type", "")
	token := ""

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
		_, ok := params["devicename"]
		if ok && deviceName == "" {
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

	err = NewActivityForUser(name, "Created account")
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	_, ok = params["devicename"]
	if ok {
		// Create token for device and token
		token, err = GenerateToken(name, deviceName)
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

		err = NewActivityForUser(name, "Created device "+deviceName+" from "+req.RemoteAddr)
		if err != nil {
			res["error"] = err.Error()
			res.Send(rw, req, http.StatusInternalServerError)
			return
		}
	}

	res["user"] = map[string]interface{}{"name": name}
	devices := []map[string]string{}

	if ok {
		devices = append(devices, map[string]string{"name": deviceName, "token": token})
	}
	res["user"].(map[string]interface{})["devices"] = devices

	res.Send(rw, req, 0)
}

func ShowUserHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := strings.ToLower(vars["name"])
	res := make(Response, 0)

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

	// Authentication is optional, so we can ignore errors
	authenticated, currentUser, _ := Authenticate(req)
	if currentUser != name {
		authenticated = false
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
	name := strings.ToLower(vars["name"])
	res := make(Response, 0)

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

	// Delete the activity
	iterator = func(activity map[string]string) error {
		return DeleteActivity(name, activity["time"])
	}

	_, err = GetUserActivities(name, iterator)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	err = DeleteUserActivities(name)
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
	res := make(Response, 0)
	params, ok := ParseForm(rw, req, res)
	if !ok {
		return
	}
	plainPass := params.Get("password")
	req.Header.Set("Content-Type", "")
	vars := mux.Vars(req)
	name := strings.ToLower(vars["name"])

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

	// Ensure data is valid
	errs, err := Validate(func() (error, error) {
		_, ok := params["password"]
		if ok && plainPass == "" {
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

	args := []string{}

	_, ok = params["password"]
	if ok {
		// Generate password
		password, err := HashPass(plainPass)
		if err != nil {
			res["error"] = err.Error()
			res.Send(rw, req, http.StatusInternalServerError)
			return
		}

		args = append(args, "password", string(password))
	}

	err = UpdateUser(name, args...)
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

	err = NewActivityForUser(name, "Updated account from "+req.RemoteAddr)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	res["user"] = map[string]interface{}{"name": name, "devices": devices}
	res.Send(rw, req, 0)
}
