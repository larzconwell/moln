package main

import (
	"github.com/garyburd/redigo/redis"
)

// UserExists checks if a user exists in the DB
func UserExists(user string) (bool, error) {
	return redis.Bool(DB.Do("exists", "users:"+user))
}

// DeviceExists checks if a device exists in the DB
func DeviceExists(user, device string) (bool, error) {
	return redis.Bool(DB.Do("exists", "users:"+user+":devices:"+device))
}

// CreateUser creates a user in the DB
func CreateUser(user, password string) error {
	_, err := DB.Do("hmset", "users:"+user, "name", user, "password", password)
	return err
}

// CreateDevice creates a device in the DB
func CreateDevice(user, device, token string) error {
	_, err := DB.Do("hmset", "users:"+user+":devices:"+device, "name", device, "token",
		token, "user", user)
	return err
}

// CreateToken creates a token in the DB
func CreateToken(user, device, token string) error {
	_, err := DB.Do("hmset", "tokens:"+token, "device", device, "user", user)
	return err
}

// GetUser retrieves a user from the DB
func GetUser(user string) (map[string]string, error) {
	u, err := ToMap(DB.Do("hgetall", "users:"+user))
	if err == redis.ErrNil {
		err = ErrUserNotExist
	}

	return u, err
}

// GetUserDevices gets a users devices from the DB
func GetUserDevices(user string, iterator func(map[string]string) error) ([]map[string]string, error) {
	// Get users device names
	deviceNames, err := redis.Strings(DB.Do("smembers", "users:"+user+":devices"))
	if err != nil {
		return nil, err
	}
	devices := make([]map[string]string, len(deviceNames))

	// Get the users devices
	for i, d := range deviceNames {
		ds := string(d)

		device, err := ToMap(DB.Do("hgetall", "users:"+user+":devices:"+ds))
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

// GetDevice retrieves a device from the DB
func GetDevice(user, device string) (map[string]string, error) {
	d, err := ToMap(DB.Do("hgetall", "users:"+user+":devices:"+device))
	if err == redis.ErrNil {
		err = ErrDeviceNotExist
	}

	delete(d, "user")
	return d, err
}

// GetToken retrieves a token from the DB
func GetToken(token string) (map[string]string, error) {
	t, err := ToMap(DB.Do("hgetall", "tokens:"+token))
	if err == redis.ErrNil {
		err = ErrTokenNotExist
	}

	return t, err
}

// UpdateUser updates the users data
func UpdateUser(user string, attributes ...string) error {
	var err error

	if len(attributes) > 0 {
		_, err = DB.Do("hmset", redis.Args{}.Add("users:"+user).AddFlat(attributes)...)
	}
	return err
}

// DeleteUser deletes the user
func DeleteUser(user string) error {
	_, err := DB.Do("del", "users:"+user)
	return err
}

// DeleteUserDevices deletes the users device set
func DeleteUserDevices(user string) error {
	_, err := DB.Do("del", "users:"+user+":devices")
	return err
}

// DeleteDevice deletes a device from the DB
func DeleteDevice(user, device string) error {
	_, err := DB.Do("del", "users:"+user+":devices:"+device)
	return err
}

// DeleteToken deletes a token from the DB
func DeleteToken(token string) error {
	_, err := DB.Do("del", "tokens:"+token)
	return err
}

// AddDeviceToUser adds a device to the users devices
func AddDeviceToUser(user, device string) error {
	_, err := DB.Do("sadd", "users:"+user+":devices", device)
	return err
}

// RemoveDeviceFromUser removes a device from the users devices
func RemoveDeviceFromUser(user, device string) error {
	_, err := DB.Do("srem", "users:"+user+":devices", device)
	return err
}
