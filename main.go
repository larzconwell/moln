package main

import (
	"github.com/larzconwell/moln/config"
	"github.com/larzconwell/moln/loggers"
	"log"
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

	errorLogger, errorLogFile, err := loggers.Error(filepath.Join(conf.LogDir, "errors"))
	if err != nil {
		log.Fatalln(err)
	}
	defer errorLogFile.Close()

	logFile, err := loggers.Access(conf.LogDir)
	if err != nil {
		errorLogger.Fatalln(err)
	}
	defer logFile.Close()
}
