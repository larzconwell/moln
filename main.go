package main

import (
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var (
	DB     redis.Conn
	Routes map[string]*Route = make(map[string]*Route, 0)
)

func main() {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	config, err := ReadConfig("config/default.json", "config/"+environment+".json")
	if err != nil {
		log.Fatalln(err)
	}

	errorLogger, errorLogFile, err := NewErrorLogger(filepath.Join(config.LogDir, "errors"))
	if err != nil {
		log.Fatalln(err)
	}
	defer errorLogFile.Close()

	logFile, err := NewLogFile(config.LogDir)
	if err != nil {
		errorLogger.Fatalln(err)
	}
	defer logFile.Close()

	DB, err = redis.DialTimeout(config.RedisNetwork, config.RedisAddr,
		config.MaxTimeout, config.MaxTimeout, config.MaxTimeout)
	if err != nil {
		errorLogger.Fatal(err)
	}
	defer DB.Close()

	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	for name, r := range Routes {
		route := router.NewRoute()
		route.Name(name).Path(r.Path).Methods(r.Methods...)
		route.Handler(http.HandlerFunc(r.Handler))
	}

	server := &http.Server{
		Addr:         config.ServerAddr,
		Handler:      NewSlashHandler(NewLogHandler(logFile, NewContentTypeHandler(router))),
		ReadTimeout:  config.MaxTimeout,
		WriteTimeout: config.MaxTimeout,
	}

	if config.TLS != nil {
		err = server.ListenAndServeTLS(config.TLS.Cert, config.TLS.Key)
	} else {
		err = server.ListenAndServe()
	}
	if err != nil {
		errorLogger.Fatal(err)
	}
}
