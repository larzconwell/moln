package main

import (
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/garyburd/redigo/redis"
	"github.com/nu7hatch/gouuid"
	"strings"
	"time"
)

// Database keys.
var (
	UserKey    = "users:{{user}}"
	DevicesKey = "users:{{user}}:devices"
	DeviceKey  = "users:{{user}}:devices:{{device}}"
	TokenKey   = "tokens:{{token}}"
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

// DeviceExists checks if a device exists.
func (db *DBConn) DeviceExists(user, device string) (bool, error) {
	key := strings.Replace(DeviceKey, "{{user}}", user, -1)

	return db.Exists(strings.Replace(key, "{{device}}", device, -1))
}

// GetUser retrieves a user by their name.
func (db *DBConn) GetUser(name string) (*User, error) {
	reply, err := redis.Values(db.Do("hgetall", strings.Replace(UserKey, "{{user}}", name, -1)))
	if err != nil {
		return nil, err
	}

	user := new(User)
	err = redis.ScanStruct(reply, user)
	if err != nil {
		user = nil
	}
	if len(reply) <= 0 {
		user = nil
	}

	return user, err
}

// GetUserByToken retrieves a user by their name.
func (db *DBConn) GetUserByToken(token string) (*User, error) {
	reply, err := redis.Values(db.Do("hgetall", strings.Replace(TokenKey, "{{token}}", token, -1)))
	if err != nil {
		return nil, err
	}

	tok := new(Token)
	err = redis.ScanStruct(reply, tok)
	if err != nil {
		return nil, err
	}
	if len(reply) <= 0 {
		return nil, nil
	}

	return db.GetUser(tok.User)
}

// GetDevices retrieves a users devices.
func (db *DBConn) GetDevices(user string) ([]*Device, error) {
	reply, err := redis.Strings(db.Do("smembers", strings.Replace(DevicesKey, "{{user}}", user, -1)))
	if err != nil {
		return nil, err
	}

	devices := make([]*Device, 0)
	for _, item := range reply {
		device, err := db.GetDevice(user, item)
		if err != nil {
			return nil, err
		}

		devices = append(devices, device)
	}

	return devices, nil
}

// GetDevice retrieves a device.
func (db *DBConn) GetDevice(user, name string) (*Device, error) {
	key := strings.Replace(DeviceKey, "{{user}}", user, -1)

	reply, err := redis.Values(db.Do("hgetall", strings.Replace(key, "{{device}}", name, -1)))
	if err != nil {
		return nil, err
	}

	device := new(Device)
	err = redis.ScanStruct(reply, device)
	if err != nil {
		device = nil
	}
	if len(reply) <= 0 {
		device = nil
	}

	return device, err
}

// DeleteDevices deletes all a users devices
func (db *DBConn) DeleteDevices(name string) error {
	devices, err := db.GetDevices(name)
	if err != nil {
		return err
	}
	user := &User{Name: name}

	for _, device := range devices {
		device.User = user

		err = device.Delete()
		if err != nil {
			return err
		}
	}

	return nil
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

// Delete removes the user data.
func (user *User) Delete() error {
	_, err := DB.Do("del", strings.Replace(UserKey, "{{user}}", user.Name, -1))

	return err
}

// Device represents a single device for a user.
type Device struct {
	Name  string `json:"name" redis:"name"`
	Token string `json:"token" redis:"token"`
	User  *User  `json:"-" redis:"-"`
}

// Validate ensures the data is valid.
func (device *Device) Validate(new bool) ([]string, error) {
	return Validations(func() (error, error) {
		if device.Name == "" {
			return ErrDeviceNameEmpty, nil
		}

		return nil, nil
	}, func() (error, error) {
		if !new || device.Name == "" {
			return nil, nil
		}

		exists, err := DB.DeviceExists(device.User.Name, device.Name)
		if err != nil {
			return nil, err
		}

		if exists {
			return ErrDeviceAlreadyExists, nil
		}

		return nil, nil
	})
}

// Save saves the device data.
func (device *Device) Save(token bool) error {
	if token {
		tok, err := uuid.NewV5(uuid.NamespaceURL, []byte(device.User.Name+device.Name))
		if err != nil {
			return err
		}

		device.Token = tok.String()
	}

	// Add to device set
	key := strings.Replace(DevicesKey, "{{user}}", device.User.Name, -1)
	_, err := DB.Do("sadd", key, device.Name)
	if err != nil {
		return err
	}

	// Add device hash
	key = strings.Replace(DeviceKey, "{{user}}", device.User.Name, -1)
	key = strings.Replace(key, "{{device}}", device.Name, -1)
	_, err = DB.Do("hmset", redis.Args{}.Add(key).AddFlat(device)...)
	if err != nil {
		return err
	}

	// Add token hash
	key = strings.Replace(TokenKey, "{{token}}", device.Token, -1)
	_, err = DB.Do("hmset", redis.Args{}.Add(key).AddFlat(&Token{device.User.Name, device.Name})...)
	return err
}

// Delete removes the device data.
func (device *Device) Delete() error {
	// Remove token hash
	_, err := DB.Do("del", strings.Replace(TokenKey, "{{token}}", device.Token, -1))
	if err != nil {
		return err
	}

	// Remove device hash
	key := strings.Replace(DeviceKey, "{{user}}", device.User.Name, -1)
	_, err = DB.Do("del", strings.Replace(key, "{{device}}", device.Name, -1))
	if err != nil {
		return err
	}

	// Remove from device set
	key = strings.Replace(DevicesKey, "{{user}}", device.User.Name, -1)
	_, err = DB.Do("srem", key, device.Name)
	return err
}

// Token represents a single token for a user and device.
type Token struct {
	User   string `redis:"user"`
	Device string `redis:"device"`
}
