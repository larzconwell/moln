package main

import (
	"errors"
)

var (
	ErrNoAuthValue    = errors.New("Authentication: authorization header value missing")
	ErrNoAuthPassword = errors.New("Authentication: authorization header password missing")

	ErrDeviceNameEmpty     = errors.New("Device: name cannot be empty")
	ErrDeviceAlreadyExists = errors.New("Device: name already exists")

	ErrUserNameEmpty     = errors.New("User: name cannot be empty")
	ErrUserPasswordEmpty = errors.New("User: password cannot be empty")
	ErrUserAlreadyExists = errors.New("User: name already exists")

	ErrTaskMessageEmpty = errors.New("Task: message cannot be empty")
)
