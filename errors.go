package main

import (
	"errors"
)

var (
	// Authentication errors
	ErrTokenNotExist           = errors.New("Authentication: token does not exist")
	ErrNoAuthorizationValue    = errors.New("Authentication: authorization header value missing")
	ErrNoAuthorizationPassword = errors.New("Authentication: authorization header password missing")

	// Authorization errors
	ErrUserNotAuthorized = errors.New("Authorization: user is not authorized to access this page")

	// User errors
	ErrUserNotExist      = errors.New("User: user does not exist")
	ErrUserNameEmpty     = errors.New("User: name cannot be empty")
	ErrUserPasswordEmpty = errors.New("User: password cannot be empty")
	ErrUserAlreadyExists = errors.New("User: name already exists")

	// Device errors
	ErrDeviceNotExist  = errors.New("Device: device does not exist")
	ErrDeviceNameEmpty = errors.New("Device: name cannot be empty")
)
