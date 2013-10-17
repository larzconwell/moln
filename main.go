package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/moln/server/config"
	"github.com/moln/server/httpextra"
	"github.com/moln/loggers"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var (
	Pool         *DBPool
	Config       *config.Config
	ContentTypes = make(map[string]*httpextra.ContentType)
	Routes       = make([]*Route, 0)
	err          error
)

func main() {
	env := "development"
	if len(os.Args) > 1 {
		env = os.Args[1]
	}

	Config, err = config.ReadFiles("config/environment.json", "config/"+env+".json")
	if err != nil {
		log.Fatalln(err)
	}

	errorLogger, errorLogFile, err := loggers.Error(filepath.Join(Config.LogDir, "errors.log"))
	if err != nil {
		log.Fatalln(err)
	}
	defer errorLogFile.Close()

	logFile, err := loggers.Access(Config.LogDir)
	if err != nil {
		errorLogger.Fatalln(err)
	}
	defer logFile.Close()

	Pool = NewDBPool()
	defer Pool.Close()

	ContentTypes["application/json"] = &httpextra.ContentType{"application/json", ".json",
		"{\"error\": \"{{message}}\"}", json.Marshal, true}
	router := mux.NewRouter()
	router.NotFoundHandler = httpextra.NewNotFoundHandler(ContentTypes)

	for _, r := range Routes {
		route := router.NewRoute()
		route.Name(r.Name).Path(r.Path + "{ext:(\\.[a-z]+)?}").Methods(r.Methods...)
		route.HandlerFunc(r.Handler)
	}

	server := &http.Server{
		Addr: Config.ServerAddr,
		Handler: httpextra.NewSlashHandler(httpextra.NewLogHandler(logFile,
			httpextra.NewContentTypeHandler(ContentTypes, router))),
		ReadTimeout:  Config.ServerMaxTimeout,
		WriteTimeout: Config.ServerMaxTimeout,
	}

	if Config.TLS != nil {
		err = server.ListenAndServeTLS(Config.TLS.Cert, Config.TLS.Key)
	} else {
		err = server.ListenAndServe()
	}
	if err != nil {
		errorLogger.Fatal(err)
	}
}
