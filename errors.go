package main

import (
	"errors"
)

var (
	ErrDeviceNameEmpty = errors.New("Device: name cannot be empty")

	ErrUserNameEmpty     = errors.New("User: name cannot be empty")
	ErrUserPasswordEmpty = errors.New("User: password cannot be empty")
)
