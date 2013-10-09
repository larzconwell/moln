package main

import (
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/garyburd/redigo/redis"
	"github.com/nu7hatch/gouuid"
	"strconv"
	"strings"
	"time"
)

// Database keys.
var (
	UserKey       = "users:{{user}}"
	DevicesKey    = "users:{{user}}:devices"
	DeviceKey     = "users:{{user}}:devices:{{device}}"
	ActivitiesKey = "users:{{user}}:activities"
	ActivityKey   = "users:{{user}}:activities:{{activity}}"
	TasksKey      = "users:{{user}}:tasks"
	TaskKey       = "users:{{user}}:tasks:{{task}}"
	TokenKey      = "tokens:{{token}}"
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

// GetActivities retrieves a users activities.
func (db *DBConn) GetActivities(user string) ([]*Activity, error) {
	key := strings.Replace(ActivitiesKey, "{{user}}", user, -1)

	reply, err := redis.Strings(db.Do("lrange", key, 0, -1))
	if err != nil {
		return nil, err
	}

	activities := make([]*Activity, 0)
	for _, item := range reply {
		activity, err := db.GetActivity(user, item)
		if err != nil {
			return nil, err
		}

		activities = append(activities, activity)
	}

	return activities, nil
}

// GetActivity retrieves a activity.
func (db *DBConn) GetActivity(user, time string) (*Activity, error) {
	key := strings.Replace(ActivityKey, "{{user}}", user, -1)

	reply, err := redis.Values(db.Do("hgetall", strings.Replace(key, "{{activity}}", time, -1)))
	if err != nil {
		return nil, err
	}

	activity := new(Activity)
	err = redis.ScanStruct(reply, activity)
	if err != nil {
		activity = nil
	}
	if len(reply) <= 0 {
		activity = nil
	}

	return activity, err
}

// GetTasks retrieves a users tasks.
func (db *DBConn) GetTasks(user string) ([]*Task, error) {
	reply, err := redis.Strings(db.Do("smembers", strings.Replace(TasksKey, "{{user}}", user, -1)))
	if err != nil {
		return nil, err
	}

	tasks := make([]*Task, 0)
	for _, item := range reply {
		task, err := db.GetTask(user, item)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTask retrieves a task.
func (db *DBConn) GetTask(user, id string) (*Task, error) {
	key := strings.Replace(TaskKey, "{{user}}", user, -1)

	reply, err := redis.Values(db.Do("hgetall", strings.Replace(key, "{{task}}", id, -1)))
	if err != nil {
		return nil, err
	}

	task := new(Task)
	err = redis.ScanStruct(reply, task)
	if err != nil {
		task = nil
	}
	if len(reply) <= 0 {
		task = nil
	}

	return task, err
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

// DeleteActivities deletes all a users activities
func (db *DBConn) DeleteActivities(name string) error {
	activities, err := db.GetActivities(name)
	if err != nil {
		return err
	}
	user := &User{Name: name}

	// Delete activity hashes
	for _, activity := range activities {
		activity.User = user

		err = activity.Delete()
		if err != nil {
			return err
		}
	}

	// Delete activity list here, since we can't remove list items individually easily
	_, err = db.Do("del", strings.Replace(ActivitiesKey, "{{user}}", name, -1))
	return err
}

// DeleteTasks deletes all a users tasks
func (db *DBConn) DeleteTasks(name string) error {
	tasks, err := db.GetTasks(name)
	if err != nil {
		return err
	}
	user := &User{Name: name}

	for _, task := range tasks {
		task.User = user

		err = task.Delete()
		if err != nil {
			return err
		}
	}

	return nil
}

/*
  User
*/

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

/*
  Device
*/

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

/*
  Activity
*/

// Activity represents a single activity for a user.
type Activity struct {
	Message string `json:"message" redis:"message"`
	Time    string `json:"time" redis:"time"`
	User    *User  `json:"-" redis:"-"`
}

// Save saves the activity data.
func (activity *Activity) Save() error {
	activity.Time = time.Now().Format(time.RFC3339)

	// Add to activity list
	key := strings.Replace(ActivitiesKey, "{{user}}", activity.User.Name, -1)
	_, err := DB.Do("lpush", key, activity.Time)
	if err != nil {
		return err
	}

	// Add activity hash
	key = strings.Replace(ActivityKey, "{{user}}", activity.User.Name, -1)
	key = strings.Replace(key, "{{activity}}", activity.Time, -1)
	_, err = DB.Do("hmset", redis.Args{}.Add(key).AddFlat(activity)...)
	return err
}

// Delete removes the activity data.
func (activity *Activity) Delete() error {
	// Remove activity hash
	key := strings.Replace(ActivityKey, "{{user}}", activity.User.Name, -1)
	_, err = DB.Do("del", strings.Replace(key, "{{activity}}", activity.Time, -1))
	return err
}

/*
  Task
*/

// Task represents a single task for a user.
type Task struct {
	ID       int    `json:"id" redis:"id"`
	Message  string `json:"message" redis:"message"`
	Complete bool   `json:"complete" redis:"complete"`
	User     *User  `json:"-" redis:"-"`
}

// Delete removes the task data.
func (task *Task) Delete() error {
	id := strconv.Itoa(task.ID)

	// Remove task hash
	key := strings.Replace(TaskKey, "{{user}}", task.User.Name, -1)
	_, err = DB.Do("del", strings.Replace(key, "{{task}}", id, -1))
	if err != nil {
		return err
	}

	// Remove from task set
	key = strings.Replace(TasksKey, "{{user}}", task.User.Name, -1)
	_, err = DB.Do("srem", key, id)
	return err
}

/*
  Token
*/

// Token represents a single token for a user and device.
type Token struct {
	User   string `redis:"user"`
	Device string `redis:"device"`
}
