package main

import (
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	"net/http"
)

func init() {
	Routes["Users.New"] = &Route{Methods: []string{"POST"}, Path: "/users", Handler: CreateUserHandler}
	Routes["Users.Show"] = &Route{Methods: []string{"GET"}, Path: "/users/{name}", Handler: ShowUserHandler}
}

func CreateUserHandler(rw http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		res := Response{"error": err}
		res.Send(rw, http.StatusBadRequest)
		return
	}
	params := req.PostForm
	name := params.Get("name")
	plainPass := params.Get("password")
	deviceName := params.Get("devicename")
	res := Response{}

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
		exists, err := redis.Bool(DB.Do("exists", "user:"+name))
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

		res.Send(rw, status)
		return
	}

	// Generate password
	password, err := HashPass(plainPass)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, http.StatusInternalServerError)
		return
	}

	// Create user
	_, err = DB.Do("hmset", "user:"+name, "name", name, "password", string(password))
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, http.StatusInternalServerError)
		return
	}

	// Create token for device and token
	token, err := CreateToken(name, deviceName)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, http.StatusInternalServerError)
		return
	}

	// Create device
	_, err = DB.Do("hmset", "user:"+name+":device:"+deviceName, "name", deviceName,
		"token", token, "user", name)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, http.StatusInternalServerError)
		return
	}

	// Create token
	_, err = DB.Do("hmset", "token:"+token, "device", deviceName, "user", name)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, http.StatusInternalServerError)
		return
	}

	// Add device name to user devices
	_, err = DB.Do("sadd", "user:"+name+":devices", deviceName)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, http.StatusInternalServerError)
		return
	}

	res["user"] = map[string]interface{}{
		"name":    name,
		"devices": []map[string]string{map[string]string{"name": deviceName, "token": token}},
	}
	res.Send(rw, 0)
}

func ShowUserHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	res := Response{}

	// Authentication is optional, so we can ignore errors
	authenticated, _ := Authenticate(req, name)

	// Ensure user exists
	exists, err := redis.Bool(DB.Do("exists", "user:"+name))
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, http.StatusInternalServerError)
		return
	}
	if !exists {
		res["error"] = http.StatusText(http.StatusNotFound)
		res.Send(rw, http.StatusNotFound)
		return
	}

	// Get users device names
	deviceNames, err := redis.Strings(DB.Do("smembers", "user:"+name+":devices"))
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, http.StatusInternalServerError)
		return
	}
	devices := make([]map[string]string, len(deviceNames))

	// Get the users devices
	for i, d := range deviceNames {
		device, err := ToMap(DB.Do("hgetall", "user:"+name+":device:"+string(d)))
		if err != nil {
			res["error"] = err.Error()
			res.Send(rw, http.StatusInternalServerError)
			return
		}

		devices[i] = device
	}

	for _, v := range devices {
		delete(v, "user")

		if !authenticated {
			v["token"] = ""
		}
	}

	res["user"] = map[string]interface{}{"name": name, "devices": devices}
	res.Send(rw, 0)
}
