package main

import (
	"github.com/garyburd/redigo/redis"
)

// UserExists checks if a user exists in the DB
func UserExists(name string) (bool, error) {
	return redis.Bool(DB.Do("exists", "users:"+name))
}

// CreateUser creates a user in the DB
func CreateUser(name, password string) error {
	_, err := DB.Do("hmset", "users:"+name, "name", name, "password", string(password))
	return err
}

// CreateDevice creates a device in the DB
func CreateDevice(user, name, token string) error {
	_, err := DB.Do("hmset", "users:"+user+":device:"+name, "name", name, "token",
		token, "user", user)
	return err
}

// CreateToken creates a token in the DB
func CreateToken(user, device, value string) error {
	_, err := DB.Do("hmset", "tokens:"+value, "device", device, "user", user)
	return err
}

// AddDeviceToUser adds a device to the users devices
func AddDeviceToUser(name, device string) error {
	_, err := DB.Do("sadd", "users:"+name+":devices", device)
	return err
}

// GetUser retrieves a user from the DB
func GetUser(name string) (map[string]string, error) {
	user, err := ToMap(DB.Do("hgetall", "users:"+name))
	if err == redis.ErrNil {
		err = ErrUserNotExist
	}

	return user, err
}

// GetUserDevices gets a users devices from the DB
func GetUserDevices(name string, iterator func(map[string]string) error) ([]map[string]string, error) {
	// Get users device names
	deviceNames, err := redis.Strings(DB.Do("smembers", "users:"+name+":devices"))
	if err != nil {
		return nil, err
	}
	devices := make([]map[string]string, len(deviceNames))

	// Get the users devices
	for i, d := range deviceNames {
		ds := string(d)

		device, err := ToMap(DB.Do("hgetall", "users:"+name+":device:"+ds))
		if err != nil {
			return nil, err
		}

		delete(device, "user")
		devices[i] = device

		if iterator != nil {
			err = iterator(device)
			if err != nil {
				return nil, err
			}
		}
	}

	return devices, nil
}

// GetToken retrieves a token from the DB
func GetToken(value string) (map[string]string, error) {
	token, err := ToMap(DB.Do("hgetall", "tokens:"+value))
	if err == redis.ErrNil {
		err = ErrTokenNotExist
	}

	return token, err
}

// UpdateUser updates the users data
func UpdateUser(name string, attributes ...string) error {
	_, err := DB.Do("hmset", redis.Args{}.Add("users:"+name).AddFlat(attributes))
	return err
}

// DeleteUser deletes the user
func DeleteUser(name string) error {
	_, err := DB.Do("del", "users:"+name)
	return err
}

// DeleteUserDevices deletes the users device set
func DeleteUserDevices(name string) error {
	_, err := DB.Do("del", "users:"+name+":devices")
	return err
}

// DeleteDevice deletes a device from the DB
func DeleteDevice(user, name string) error {
	_, err := DB.Do("del", "users:"+user+":device:"+name)
	return err
}

// DeleteToken deletes a token from the DB
func DeleteToken(value string) error {
	_, err := DB.Do("del", "tokens:"+value)
	return err
}
