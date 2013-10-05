package main

import (
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/garyburd/redigo/redis"
	"strings"
	"time"
)

// DB keys.
var (
	UserKey = "users:{{user}}"
)

// DBConn wraps redis.Conn that includes methods for data management.
type DBConn struct {
	redis.Conn
}

// DBDialTimeout creates a database connection that has timeouts.
func DBDialTimeout(network, addr string, cTimeout, rTimeout, wTimeout time.Duration) (*DBConn, error) {
	db, err := redis.DialTimeout(network, addr, cTimeout, rTimeout, wTimeout)
	if err != nil {
		return nil, err
	}

	return &DBConn{db}, nil
}

// Exists is a generic check for any key.
func (db *DBConn) Exists(key string) (bool, error) {
	return redis.Bool(db.Do("exists", key))
}

// UserExists checks if a user exists.
func (db *DBConn) UserExists(user string) (bool, error) {
	return db.Exists(strings.Replace(UserKey, "{{user}}", user, -1))
}

// User represents a single users data.
type User struct {
	Name     string `json:"name" redis:"name"`
	Password string `json:"-" redis:"password"`
}

// Validate ensures the data is valid.
func (user *User) Validate(new bool) ([]string, error) {
	return Validations(func() (error, error) {
		if user.Name == "" {
			return ErrUserNameEmpty, nil
		}

		return nil, nil
	}, func() (error, error) {
		if user.Password == "" {
			return ErrUserPasswordEmpty, nil
		}

		return nil, nil
	}, func() (error, error) {
		if !new || user.Name == "" {
			return nil, nil
		}

		exists, err := DB.UserExists(user.Name)
		if err != nil {
			return nil, err
		}

		if exists {
			return ErrUserAlreadyExists, nil
		}

		return nil, nil
	})
}

// Save saves the user data.
func (user *User) Save(hash bool) error {
	if hash {
		pass, err := bcrypt.GenerateFromPassword([]byte(user.Password), -1)
		if err != nil {
			return err
		}

		user.Password = string(pass)
	}

	key := strings.Replace(UserKey, "{{user}}", user.Name, -1)
	_, err := DB.Do("hmset", redis.Args{}.Add(key).AddFlat(user)...)
	return err
}
