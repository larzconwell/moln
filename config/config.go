// Package config implements routines to read and structure
// server configuration.
package config

import (
	"encoding/json"
	"os"
	"time"
)

// TLS describes the cert/key options for a TLS connection.
type TLS struct {
	Cert string `json:"cert"`
	Key  string `json:"key"`
}

// Config describes generic options for a server.
type Config struct {
	LogDir              string        `json:"logdir"`
	DBAddr              string        `json:"dbaddr"`
	DBNetwork           string        `json:"dbnetwork"`
	DBMaxIdle           int           `json:"dbmaxidle"`
	DBMaxTimeoutStr     string        `json:"dbmaxtimeout"`
	DBMaxTimeout        time.Duration `json:"-"`
	ServerAddr          string        `json:"serveraddr"`
	ServerNetwork       string        `json:"servernetwork"`
	ServerMaxTimeoutStr string        `json:"servermaxtimeout"`
	ServerMaxTimeout    time.Duration `json:"-"`
	TLS                 *TLS          `json:"tls"`
}

// ReadFiles reads the given JSON config files and returns the combined config.
func ReadFiles(files ...string) (*Config, error) {
	var err error
	config := new(Config)

	for _, file := range files {
		fileF, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer fileF.Close()

		decoder := json.NewDecoder(fileF)
		decoder.Decode(config)
	}

	if config.DBMaxTimeoutStr != "" {
		config.DBMaxTimeout, err = time.ParseDuration(config.DBMaxTimeoutStr)
	}
	if config.ServerMaxTimeoutStr != "" {
		config.ServerMaxTimeout, err = time.ParseDuration(config.ServerMaxTimeoutStr)
	}
	return config, err
}
