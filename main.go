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
	DB     *redis.Conn
	Routes map[string]*Route
)

func init() {
	Routes = make(map[string]*Route)
}

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

	db, err := redis.DialTimeout(config.RedisNetwork, config.RedisAddr,
		config.MaxTimeout, config.MaxTimeout, config.MaxTimeout)
	if err != nil {
		errorLogger.Fatal(err)
	}
	defer db.Close()
	DB = &db

	router := mux.NewRouter()
	router.StrictSlash(true)
	router.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	for name, route := range Routes {
		router.NewRoute().Name(name).Methods(route.Method).Path(route.Path).HandlerFunc(route.Handler)
	}

	server := &http.Server{
		Addr:         config.ServerAddr,
		Handler:      NewLogHandler(logFile, router),
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
