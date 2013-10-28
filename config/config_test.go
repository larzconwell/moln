package config

import (
	"testing"
)

func TestRead(t *testing.T) {
	config, err := ReadFiles("environment.json", "development.json")
	if err != nil {
		t.Fatal(err)
	}

	// Test option from environment
	if config.ServerAddr != ":3000" {
		t.Error("ServerAddr option is incorrect")
	}

	// Test option from development
	if config.DBNetwork != "tcp" {
		t.Error("DBNetwork option is incorrect")
	}
}
