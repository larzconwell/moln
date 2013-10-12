package main

import (
	"github.com/gorilla/mux"
	"github.com/larzconwell/moln/httpextra"
	"net/http"
	"strconv"
)

func init() {
	createTask := &Route{"CreateTask", "/tasks", []string{"POST"}, CreateTaskHandler}
	getTasks := &Route{"GetTasks", "/tasks", []string{"GET"}, GetTasksHandler}
	getTask := &Route{"GetTask", "/tasks/{id}", []string{"GET"}, GetTaskHandler}
	updateTask := &Route{"UpdateTask", "/tasks/{id}", []string{"PUT"}, UpdateTaskHandler}
	deleteTask := &Route{"DeleteTask", "/tasks/{id}", []string{"DELETE"}, DeleteTaskHandler}

	Routes = append(Routes, createTask, getTasks, getTask, updateTask, deleteTask)
}

func CreateTaskHandler(rw http.ResponseWriter, req *http.Request) {
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

	task := &Task{Conn: conn, Message: params.Get("message"), User: user}
	errs, err := task.Validate()
	ok = HandleValidations(rw, req, errs, err)
	if !ok {
		return
	}
	res := &httpextra.Response{ContentTypes, rw, req}

	err = task.Save(true)
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	res.Send(task, http.StatusOK)
}

func GetTasksHandler(rw http.ResponseWriter, req *http.Request) {
	conn := Pool.Get()
	defer conn.Close()

	user := Authenticate(conn, rw, req)
	if user == nil {
		return
	}
	res := &httpextra.Response{ContentTypes, rw, req}

	tasks, err := conn.GetTasks(user.Name)
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	res.Send(tasks, http.StatusOK)
}

func GetTaskHandler(rw http.ResponseWriter, req *http.Request) {
	conn := Pool.Get()
	defer conn.Close()

	user := Authenticate(conn, rw, req)
	if user == nil {
		return
	}
	id := mux.Vars(req)["id"]
	res := &httpextra.Response{ContentTypes, rw, req}

	task, err := conn.GetTask(user.Name, id)
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

func UpdateTaskHandler(rw http.ResponseWriter, req *http.Request) {
	params, ok := httpextra.ParseForm(ContentTypes, rw, req)
	if !ok {
		return
	}
	_, messageGiven := params["message"]
	_, completeGiven := params["complete"]
	id := mux.Vars(req)["id"]
	conn := Pool.Get()
	defer conn.Close()

	user := Authenticate(conn, rw, req)
	if user == nil {
		return
	}
	res := &httpextra.Response{ContentTypes, rw, req}

	task, err := conn.GetTask(user.Name, id)
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	if task == nil {
		res.Send(map[string]string{"error": http.StatusText(http.StatusNotFound)}, http.StatusNotFound)
		return
	}
	task.User = user

	if !messageGiven && !completeGiven {
		res.Send(task, http.StatusOK)
		return
	}

	if messageGiven {
		task.Message = params.Get("message")
	}
	if completeGiven {
		complete, err := strconv.ParseBool(params.Get("complete"))
		if err != nil {
			complete = false
		}

		task.Complete = complete
	}
	errs, err := task.Validate()
	ok = HandleValidations(rw, req, errs, err)
	if !ok {
		return
	}

	err = task.Save(false)
	if err != nil {
		res.Send(map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	res.Send(task, http.StatusOK)
}

func DeleteTaskHandler(rw http.ResponseWriter, req *http.Request) {
	conn := Pool.Get()
	defer conn.Close()

	user := Authenticate(conn, rw, req)
	if user == nil {
		return
	}
	id := mux.Vars(req)["id"]
	res := &httpextra.Response{ContentTypes, rw, req}

	task, err := conn.GetTask(user.Name, id)
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
