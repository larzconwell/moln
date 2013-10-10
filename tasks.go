package main

import (
	"github.com/gorilla/mux"
	"github.com/larzconwell/moln/httpextra"
	"net/http"
)

func init() {
	createTask := &Route{"CreateTask", "/tasks", []string{"POST"}, CreateTaskHandler}
	getTasks := &Route{"GetTasks", "/tasks", []string{"GET"}, GetTasksHandler}
	getTask := &Route{"GetTask", "/tasks/{id}", []string{"GET"}, GetTaskHandler}
	deleteTask := &Route{"DeleteTask", "/tasks/{id}", []string{"DELETE"}, DeleteTaskHandler}

	Routes = append(Routes, createTask, getTasks, getTask, deleteTask)
}

func CreateTaskHandler(rw http.ResponseWriter, req *http.Request) {
	user := Authenticate(rw, req)
	if user == nil {
		return
	}

	params, ok := httpextra.ParseForm(rw, req)
	if !ok {
		return
	}

	task := &Task{Message: params.Get("message"), User: user}
	errs, err := task.Validate()
	ok = HandleValidations(rw, req, errs, err)
	if !ok {
		return
	}
	res := &httpextra.Response{rw, req}

	err = task.Save(true)
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	res.Send(task, http.StatusOK)
}

func GetTasksHandler(rw http.ResponseWriter, req *http.Request) {
	user := Authenticate(rw, req)
	if user == nil {
		return
	}
	res := &httpextra.Response{rw, req}

	tasks, err := DB.GetTasks(user.Name)
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	res.Send(tasks, http.StatusOK)
}

func GetTaskHandler(rw http.ResponseWriter, req *http.Request) {
	user := Authenticate(rw, req)
	if user == nil {
		return
	}
	id := mux.Vars(req)["id"]
	res := &httpextra.Response{rw, req}

	task, err := DB.GetTask(user.Name, id)
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	if task == nil {
		res.Send(map[string]string{"error": http.StatusText(http.StatusNotFound)}, http.StatusNotFound)
		return
	}

	res.Send(task, http.StatusOK)
}

func DeleteTaskHandler(rw http.ResponseWriter, req *http.Request) {
	user := Authenticate(rw, req)
	if user == nil {
		return
	}
	id := mux.Vars(req)["id"]
	res := &httpextra.Response{rw, req}

	task, err := DB.GetTask(user.Name, id)
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	if task == nil {
		res.Send(map[string]string{"error": http.StatusText(http.StatusNotFound)}, http.StatusNotFound)
		return
	}
	task.User = user

	err = task.Delete()
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	res.Send(task, http.StatusOK)
}
