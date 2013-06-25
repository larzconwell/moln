package main

import (
	"errors"
	"github.com/garyburd/redigo/redis"
	"github.com/nu7hatch/gouuid"
)

// Create the error slice if nil and append the error
func createAndAppendErrors(errs []error, err ...error) []error {
	if errs == nil {
		errs = make([]error, 0)
	}
	return append(errs, err...)
}

// Data implements access to update specific sections of the database and provide validation
// for the data being set
type Data interface {
	// Validate should validate the data to be set, a nil error slice means
	// no failed validations
	Validate() ([]error, error)
	// Save should save the data set to the database
	Save() error
	ValidateAndSave() ([]error, error)
	// Delete should remove the data from the database
	Delete() error
}

// User implements Data allowing updates to a user object
type User struct {
	Name     string   `redis:"name" json:"name"`
	Password string   `redis:"password" json:"-"`
	Devices  []string `redis:"-" json:"devices"`
}

func FindUser(name string) (*User, error) { return nil, nil }

/*
 * Device
 */

// Device implements Data allowing updates to a users device
type Device struct {
	Name  string `redis:"name" json:"name"`
	Token string `redis:"token" json:"token"`
	User  string `redis:"user" json:"-"`
	new   bool
	token *Token
	user  *User
}

// NewDevice creates a new device with a unique token
func NewDevice(name, user string) (*Device, error) {
	device := &Device{Name: name, User: user, new: true}
	token, err := NewToken(name, user)
	if err != nil {
		return nil, err
	}

	device.SetToken(token)
	return device, err
}

// FindDevice finds the device in the DB
func FindDevice(name, user string) (*Device, error) {
	reply, err := DB.Do("hgetall", "user:"+user+":device:"+name)
	if err != nil {
		return nil, err
	}

	device := &Device{}
	err = redis.ScanStruct(reply.([]interface{}), device)
	return device, err
}

// DeleteDevice deletes the device in the DB
func DeleteDevice(name, user string) (*Device, error) {
	device, err := FindDevice(name, user)
	if err != nil {
		return nil, err
	}

	return device, device.Delete()
}

// DeviceExists checks if the device exsts in the DB
func DeviceExists(name, user string) (bool, error) {
	return redis.Bool(DB.Do("exists", "user:"+user+":device:"+name))
}

func (device *Device) String() string {
	return device.Name
}

// SetToken sets the underlying token
func (device *Device) SetToken(token *Token) {
	device.Token = token.Value
	device.token = token
}

// SetUser sets the underlying user
func (device *Device) SetUser(user *User) {
	device.User = user.Name
	device.user = user

	if device.token != nil {
		device.token.SetUser(user)
	}
}

// FindToken finds the devices token, using the set token if exists
func (device *Device) FindToken() (*Token, error) {
	var err error
	token := device.token

	if token == nil {
		token, err = FindToken(device.Token)
	}

	if token != nil {
		device.SetToken(token)
		token.SetDevice(device)

		if device.user != nil {
			token.SetUser(device.user)
		}
	}

	return token, err
}

// FindUser finds the devices user, using the set user if exists
func (device *Device) FindUser() (*User, error) {
	var err error
	user := device.user

	if user == nil {
		user, err = FindUser(device.User)
	}

	if user != nil {
		device.SetUser(user)
		// TODO: How to add a users device?
	}

	return user, err
}

// Validate ensures the device name is not taken for the user and validates the devices token
func (device *Device) Validate() ([]error, error) {
	var (
		errs []error
		err  error
	)

	// Only check the devices name uniqueness if it's a new item
	if device.new {
		exists, err := redis.Bool(DB.Do("exists", "user:"+device.User+":device:"+device.Name))
		if err != nil {
			return errs, err
		}
		if exists {
			errs = createAndAppendErrors(errs, errors.New("Device: name already exists"))
		}
	}

	token := device.token
	if token == nil {
		token, err = device.FindToken()
		if err != nil {
			return errs, err
		}
	}

	// Validate token appending errors
	es, err := token.Validate()
	if err != nil {
		return errs, err
	}
	if es != nil {
		errs = createAndAppendErrors(errs, es...)
	}

	return errs, err
}

// Save writes the device and it's token to the DB
func (device *Device) Save() error {
	var err error

	token := device.token
	if token == nil {
		token, err = device.FindToken()
		if err != nil {
			return err
		}
	}

	if token != nil {
		err = token.Save()
		if err != nil {
			return err
		}
	}

	_, err = DB.Do("hmset", redis.Args{}.Add("user:"+device.User+":device:"+device.Name).AddFlat(device)...)
	if err == nil {
		// Unset new since saving causes the item to not be new anymore
		device.new = false
	}

	return err
}

// ValidateAndSave validates and saves the data if valid
func (device *Device) ValidateAndSave() ([]error, error) {
	errs, err := device.Validate()
	if errs == nil && err == nil {
		err = device.Save()
	}

	return errs, err
}

// Delete the device and it's token from the DB
func (device *Device) Delete() error {
	var err error

	token := device.token
	if token == nil {
		token, err = device.FindToken()
		if err != nil {
			return err
		}
	}

	if token != nil {
		err = token.Delete()
		if err != nil {
			return err
		}
	}

	_, err = DB.Do("del", "user:"+device.User+":device:"+device.Name)
	if err == nil {
		// Set new to true, so validation will check uniqueness
		device.new = true
	}

	return err
}

/*
 * Token
 */

// Token implements Data allowing manipulation of the token linking to users and devices
type Token struct {
	Value  string `redis:"-"`
	Device string `redis:"device"`
	User   string `redis:"user"`
	new    bool
	device *Device
	user   *User
}

// NewToken creates a new unique token
func NewToken(device, user string) (*Token, error) {
	token := &Token{Device: device, User: user, new: true}
	err := token.UpdateValue()

	return token, err
}

// FindToken finds the token in the DB
func FindToken(t string) (*Token, error) {
	reply, err := DB.Do("hgetall", "token:"+t)
	if err != nil {
		return nil, err
	}

	token := &Token{Value: t}
	err = redis.ScanStruct(reply.([]interface{}), token)
	return token, err
}

// DeleteToken deletes the token in the DB
func DeleteToken(t string) (*Token, error) {
	token, err := FindToken(t)
	if err != nil {
		return nil, err
	}

	return token, token.Delete()
}

// TokenExists checks if the token exists in the DB
func TokenExists(t string) (bool, error) {
	return redis.Bool(DB.Do("exists", "token:"+t))
}

func (token *Token) String() string {
	return token.Value
}

// UpdateToken creates a uuid for the tokens device and user
func (token *Token) UpdateValue() error {
	id, err := uuid.NewV5(uuid.NamespaceURL, []byte(token.Device+token.User))
	if id != nil {
		token.Value = id.String()
	}

	return err
}

// SetDevice sets the underlying device
func (token *Token) SetDevice(device *Device) {
	token.Device = device.Name
	token.device = device
}

// SetUser sets the underlying user
func (token *Token) SetUser(user *User) {
	token.User = user.Name
	token.user = user
}

// FindDevice finds the tokens device, returning the set device if exists
func (token *Token) FindDevice() (*Device, error) {
	var err error
	device := token.device

	if device == nil {
		device, err = FindDevice(token.Device, token.User)
	}

	if device != nil {
		token.SetDevice(device)
		device.SetToken(token)
	}

	return device, err
}

// FindUser finds the tokens user, returning the set user if exists
func (token *Token) FindUser() (*User, error) {
	var err error
	user := token.user

	if user == nil {
		user, err = FindUser(token.User)
	}

	if user != nil {
		token.SetUser(user)
	}

	return user, err
}

// Validate ensures the tokens uniqueness
func (token *Token) Validate() ([]error, error) {
	var errs []error

	// Only check if the token exists if it's new
	if token.new {
		exists, err := redis.Bool(DB.Do("exists", "token:"+token.Value))
		if err != nil {
			return errs, err
		}
		if exists {
			errs = createAndAppendErrors(errs, errors.New("Token: token already exists"))
		}
	}

	return errs, nil
}

// Save writes the token to the DB
func (token *Token) Save() error {
	_, err := DB.Do("hmset", redis.Args{}.Add("token:"+token.Value).AddFlat(token)...)
	if err == nil {
		// Unset new since saving causes the item to not be new anymore
		token.new = false
	}

	return err
}

// ValidateAndSave validates and saves if data is valid
func (token *Token) ValidateAndSave() ([]error, error) {
	errs, err := token.Validate()
	if errs == nil && err == nil {
		err = token.Save()
	}

	return errs, err
}

// Delete removes the token from the database
func (token *Token) Delete() error {
	_, err := DB.Do("del", "token:"+token.Value)
	if err == nil {
		// Set new to true, so validation will check uniqueness
		token.new = true
	}

	return err
}
