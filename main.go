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

var (
	DB           *DBConn
	ContentTypes = make(map[string]*httpextra.ContentType)
	Routes       = make([]*Route, 0)
	err          error
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

	DB, err = DBDialTimeout(conf.DBNetwork, conf.DBAddr, conf.MaxTimeout, conf.MaxTimeout, conf.MaxTimeout)
	if err != nil {
		errorLogger.Fatalln(err)
	}
	defer DB.Close()

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
		Addr: conf.ServerAddr,
		Handler: httpextra.NewSlashHandler(httpextra.NewLogHandler(logFile,
			httpextra.NewContentTypeHandler(ContentTypes, router))),
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
