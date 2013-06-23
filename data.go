package main

import (
	"code.google.com/p/go.crypto/bcrypt"
	"errors"
	"github.com/garyburd/redigo/redis"
	"github.com/nu7hatch/gouuid"
)

// Create the error slice if nil and append the error
func appendOrCreateErrors(errs []error, err error) []error {
	if errs == nil {
		errs = make([]error, 0)
	}
	errs = append(errs, err)
	return errs
}

// User is a single user with an encrypted password
type User struct {
	Username string `redis:"username"`
	Password []byte `redis:"password"`
}

// Create a new user and hashes the given password
func NewUser(username string, password []byte) (*User, error) {
	user := &User{Username: username}

	if username == "" || len(password) <= 0 {
		return nil, errors.New("Username or password was not given")
	}

	passHash, err := bcrypt.GenerateFromPassword(password, -1)
	if err != nil {
		return nil, err
	}
	user.Password = passHash

	return user, nil
}

// Validate the user data, the slice of errors are the validation errors. Nil if valid
func (user *User) Validate() ([]error, error) {
	var errs []error

	if user.Username == "" {
		errs = appendOrCreateErrors(errs, errors.New("User: Username is empty"))
	}

	if len(user.Password) <= 0 {
		errs = appendOrCreateErrors(errs, errors.New("User: Password is empty"))
	}

	// Check if the user exists
	exists, err := redis.Bool(DB.Do("exists", "user:"+user.Username))
	if err != nil {
		return nil, err
	}

	if exists {
		errs = appendOrCreateErrors(errs, errors.New("User: Username "+user.Username+" exists"))
	}

	return errs, nil
}

// Save the users data
func (user *User) Save() error {
	// Update user fields
	_, err := DB.Do("hmset", redis.Args{}.Add("user:"+user.Username).AddFlat(user)...)
	if err != nil {
		return err
	}

	// Get the users device names
	deviceNames, err := DB.Do("smembers", "user:"+user.Username+":devices")
	if err != nil {
		return err
	}

	for _, n := range deviceNames.([]interface{}) {
		name := string(n.([]byte))

		// Set the devices user
		_, err := DB.Do("hset", "device:"+name, "user", user.Username)
		if err != nil {
			return err
		}

		// Get the devices token
		t, err := DB.Do("hget", "device:"+name, "token")
		if err != nil {
			return err
		}
		token := string(t.([]byte))

		// Set the tokens user
		_, err = DB.Do("hset", "token:"+token, "user", user.Username)
		if err != nil {
			return err
		}
	}

	return nil
}

// Add a new device to the user
func (user *User) AddDevice(name string) (*Device, error) {
	return NewDevice(name, user.Username)
}

// Device is single device with a unique token and a username to link to an owner
type Device struct {
	Name     string `redis:"name"`
	Token    string `redis:"token"`
	Username string `redis:"user"`
}

// Create a new device with a unique token
func NewDevice(name, username string) (*Device, error) {
	device := &Device{Name: name, Username: username}

	if name == "" || username == "" {
		return nil, errors.New("Device name or username was not given")
	}

	id, err := uuid.NewV5(uuid.NamespaceURL, []byte(name+username))
	if err != nil {
		return nil, err
	}
	device.Token = id.String()

	return device, nil
}

// Validate the device data, the slice of errors are the validation errors. Nil if valid
func (device *Device) Validate() ([]error, error) {
	var errs []error

	if device.Name == "" {
		errs = appendOrCreateErrors(errs, errors.New("Device: Name is empty"))
	}

	if device.Username == "" {
		errs = appendOrCreateErrors(errs, errors.New("Device: Username is empty"))
	}

	// Check if the device name is a member of the users devices set
	exists, err := redis.Bool(DB.Do("sismember", "user:"+device.Username+":devices", device.Name))
	if err != nil {
		return nil, err
	}
	if exists {
		errs = appendOrCreateErrors(errs,
			errors.New("Device: Device name "+device.Name+" already exists for "+device.Username))
	}

	// Check if the token exists
	exists, err = redis.Bool(DB.Do("exists", "token:"+device.Token))
	if err != nil {
		return nil, err
	}
	if exists {
		errs = appendOrCreateErrors(errs, errors.New("Device: Token "+device.Name+" exists"))
	}

	return errs, nil
}

// Save the devices data
func (device *Device) Save() error {
	// Update device fields
	_, err := DB.Do("hmset", redis.Args{}.Add("device:"+device.Name).AddFlat(device)...)
	if err != nil {
		return err
	}

	// Set the token fields
	_, err = DB.Do("hmset", "token:"+device.Token, "device", device.Name, "user", device.Username)
	if err != nil {
		return err
	}

	// Add the device to the users devices
	_, err = DB.Do("sadd", "user:"+device.Username+":devices", device.Name)
	return err
}
