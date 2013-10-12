package main

import (
	"github.com/larzconwell/moln/httpextra"
	"net/http"
)

func init() {
	getActivities := &Route{"GetActivities", "/activities", []string{"GET"}, GetActivitiesHandler}

	Routes = append(Routes, getActivities)
}

func GetActivitiesHandler(rw http.ResponseWriter, req *http.Request) {
	conn := Pool.Get()
	defer conn.Close()

	user := Authenticate(conn, rw, req)
	if user == nil {
		return
	}
	res := &httpextra.Response{ContentTypes, rw, req}

	activities, err := conn.GetActivities(user.Name)
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	res.Send(activities, http.StatusOK)
}
