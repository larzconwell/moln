package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/larzconwell/moln/config"
	"github.com/larzconwell/moln/httpextra"
	"github.com/larzconwell/moln/loggers"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	env := "development"
	if len(os.Args) > 1 {
		env = os.Args[1]
	}

	conf, err := config.ReadFiles("config/environment.json", "config/"+env+".json")
	if err != nil {
		log.Fatalln(err)
	}

	errorLogger, errorLogFile, err := loggers.Error(filepath.Join(conf.LogDir, "errors.log"))
	if err != nil {
		log.Fatalln(err)
	}
	defer errorLogFile.Close()

	logFile, err := loggers.Access(conf.LogDir)
	if err != nil {
		errorLogger.Fatalln(err)
	}
	defer logFile.Close()

	httpextra.AddContentType("application/json", ".json",
		"{\"error\": \"{{message}}\"}", json.Marshal, true)
	router := mux.NewRouter()
	router.NotFoundHandler = httpextra.NotFoundHandler

	server := &http.Server{
		Addr: conf.ServerAddr,
		Handler: httpextra.NewSlashHandler(httpextra.NewLogHandler(logFile,
			httpextra.NewContentTypeHandler(router))),
		ReadTimeout:  conf.MaxTimeout,
		WriteTimeout: conf.MaxTimeout,
	}

	if conf.TLS != nil {
		err = server.ListenAndServeTLS(conf.TLS.Cert, conf.TLS.Key)
	} else {
		err = server.ListenAndServe()
	}
	if err != nil {
		errorLogger.Fatal(err)
	}
}
