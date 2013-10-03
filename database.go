package main

import (
	"github.com/garyburd/redigo/redis"
	"time"
)

// DBConn redis.Conn and includes methods for data management
type DBConn struct {
	redis.Conn
}

// DBDialTimeout creates a database connection that has timeouts
func DBDialTimeout(network, addr string, cTimeout, rTimeout, wTimeout time.Duration) (*DBConn, error) {
	db, err := redis.DialTimeout(network, addr, cTimeout, rTimeout, wTimeout)
	if err != nil {
		return nil, err
	}

	return &DBConn{db}, nil
}
