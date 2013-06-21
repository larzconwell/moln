package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	db, err := redis.DialTimeout("tcp", config.RedisAddr, config.MaxTimeout, config.MaxTimeout, config.MaxTimeout)
	if err != nil {
		errorLogger.Fatal(err)
	}
	defer db.Close()

	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi there: %s", r.URL.Path)
	})

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
