package config

import (
	"testing"
)

func TestRead(t *testing.T) {
	config, err := Read("environment.json", "development.json")
	if err != nil {
		t.Fatal(err)
	}

	// Test option from environment
	if config.LogDir != "logs" {
		t.Error("LogDir option is incorrect")
	}

	// Test option from development
	if config.RedisNetwork != "tcp" {
		t.Error("RedisNetwork option is incorrect")
	}
}
