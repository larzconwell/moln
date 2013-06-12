package main

import (
	"github.com/garyburd/redigo/redis"
	"log"
	"os"
)

func main() {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	config, err := ReadConfig("config/default.json", "config/"+environment+".json")
	if err != nil {
		log.Fatal(err)
	}

	logger, logFile, err := NewLogger(config.LogDir)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	// Create the base redis connection
	dbConn, err := redis.DialTimeout("tcp", config.RedisAddr, config.MaxTimeout, config.MaxTimeout, config.MaxTimeout)
	if err != nil {
		logger.Fatal(err)
	}
	defer dbConn.Close()

	// Use the file logger with the redis connection
	db := redis.NewLoggingConn(dbConn, logger, "")
	defer db.Close()

	// define routes
	// start server

	logger.Println("server")
}
