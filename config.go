package main

import (
	"encoding/json"
	"os"
	"time"
)

// Config describes the configuration for the server
type Config struct {
	LogDir        string
	RedisAddr     string
	ServerAddr    string
	MaxTimeoutStr string
	MaxTimeout    time.Duration
	TLS           *TLS
}

// Read the given JSON configuration files, returning the final config
func ReadConfig(files ...string) (*Config, error) {
	var err error
	config := &Config{}

	for _, file := range files {
		fileF, err := os.OpenFile(file, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return config, err
		}
		defer fileF.Close()

		decoder := json.NewDecoder(fileF)
		decoder.Decode(config)
	}

	config.MaxTimeout, err = time.ParseDuration(config.MaxTimeoutStr)
	return config, err
}
