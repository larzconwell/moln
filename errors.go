package main

import (
	"errors"
)

var (
	ErrDeviceNameEmpty     = errors.New("Device: name cannot be empty")
	ErrDeviceAlreadyExists = errors.New("Device: name already exists")

	ErrUserNameEmpty     = errors.New("User: name cannot be empty")
	ErrUserPasswordEmpty = errors.New("User: password cannot be empty")
	ErrUserAlreadyExists = errors.New("User: name already exists")
)
