package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func init() {
	Routes["Activities.Index"] = &Route{[]string{"GET"}, "/users/{user}/activities", ShowActivitiesHandler}
}

func ShowActivitiesHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	user := vars["user"]
	res := Response{}

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

	activities, err := GetUserActivities(user, nil)
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusInternalServerError)
		return
	}

	res["activities"] = activities
	res.Send(rw, req, 0)
}
