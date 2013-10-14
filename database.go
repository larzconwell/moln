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

// connect creates a redis.Conn for pool connections.
func connect() (redis.Conn, error) {
	return redis.DialTimeout(Config.DBNetwork, Config.DBAddr, Config.DBMaxTimeout,
		Config.DBMaxTimeout, Config.DBMaxTimeout)
}

// ping is used to check if the connection is responding.
func ping(conn redis.Conn, t time.Time) error {
	_, err := conn.Do("ping")
	return err
}

// DBPool is a wrapped redis.Pool that gets a Conn instead of redis.Conn.
type DBPool struct {
	Pool *redis.Pool
}

// NewDBPool creates a new redis pool for requests.
func NewDBPool() *DBPool {
	pool := redis.NewPool(connect, Config.DBMaxIdle)
	pool.IdleTimeout = Config.DBMaxTimeout
	pool.TestOnBorrow = ping

	return &DBPool{pool}
}

// Get gets a connection and wraps it a Conn.
func (pool *DBPool) Get() *Conn {
	return &Conn{pool.Pool.Get()}
}

// Close delegates to the redis.Pool.Close.
func (pool *DBPool) Close() error {
	return pool.Pool.Close()
}

// Conn wraps redis.Conn adding methods for data management.
type Conn struct {
	redis.Conn
}

// exists is a generic check for any key.
func (conn *Conn) exists(key string) (bool, error) {
	return redis.Bool(conn.Do("exists", key))
}

// UserExists checks if a user exists.
func (conn *Conn) UserExists(user string) (bool, error) {
	return conn.exists(strings.Replace(UserKey, "{{user}}", user, -1))
}

// DeviceExists checks if a device exists.
func (conn *Conn) DeviceExists(user, device string) (bool, error) {
	key := strings.Replace(DeviceKey, "{{user}}", user, -1)

	return conn.exists(strings.Replace(key, "{{device}}", device, -1))
}

// GetUser retrieves a user by their name.
func (conn *Conn) GetUser(name string) (*User, error) {
	reply, err := redis.Values(conn.Do("hgetall", strings.Replace(UserKey, "{{user}}", name, -1)))
	if err != nil {
		return nil, err
	}

	user := &User{Conn: conn}
	err = redis.ScanStruct(reply, user)
	if err != nil {
		user = nil
	}
	if len(reply) <= 0 {
		user = nil
	}

	return user, err
}

// GetUserByToken retrieves a user by a token.
func (conn *Conn) GetUserByToken(token string) (*User, error) {
	reply, err := redis.Values(conn.Do("hgetall", strings.Replace(TokenKey, "{{token}}", token, -1)))
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

	return conn.GetUser(tok.User)
}

// GetDevices retrieves a users devices.
func (conn *Conn) GetDevices(user string) ([]*Device, error) {
	reply, err := redis.Strings(conn.Do("smembers", strings.Replace(DevicesKey, "{{user}}", user, -1)))
	if err != nil {
		return nil, err
	}

	devices := make([]*Device, 0)
	for _, item := range reply {
		device, err := conn.GetDevice(user, item)
		if err != nil {
			return nil, err
		}

		devices = append(devices, device)
	}

	return devices, nil
}

// GetDevice retrieves a device.
func (conn *Conn) GetDevice(user, name string) (*Device, error) {
	key := strings.Replace(DeviceKey, "{{user}}", user, -1)

	reply, err := redis.Values(conn.Do("hgetall", strings.Replace(key, "{{device}}", name, -1)))
	if err != nil {
		return nil, err
	}

	device := &Device{Conn: conn}
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
func (conn *Conn) GetActivities(user string) ([]*Activity, error) {
	key := strings.Replace(ActivitiesKey, "{{user}}", user, -1)

	reply, err := redis.Strings(conn.Do("lrange", key, 0, -1))
	if err != nil {
		return nil, err
	}

	activities := make([]*Activity, 0)
	for _, item := range reply {
		activity, err := conn.GetActivity(user, item)
		if err != nil {
			return nil, err
		}

		activities = append(activities, activity)
	}

	return activities, nil
}

// GetActivity retrieves a activity.
func (conn *Conn) GetActivity(user, time string) (*Activity, error) {
	key := strings.Replace(ActivityKey, "{{user}}", user, -1)

	reply, err := redis.Values(conn.Do("hgetall", strings.Replace(key, "{{activity}}", time, -1)))
	if err != nil {
		return nil, err
	}

	activity := &Activity{Conn: conn}
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
func (conn *Conn) GetTasks(user string) ([]*Task, error) {
	reply, err := redis.Strings(conn.Do("smembers", strings.Replace(TasksKey, "{{user}}", user, -1)))
	if err != nil {
		return nil, err
	}

	tasks := make([]*Task, 0)
	for _, item := range reply {
		task, err := conn.GetTask(user, item)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTask retrieves a task.
func (conn *Conn) GetTask(user, id string) (*Task, error) {
	key := strings.Replace(TaskKey, "{{user}}", user, -1)

	reply, err := redis.Values(conn.Do("hgetall", strings.Replace(key, "{{task}}", id, -1)))
	if err != nil {
		return nil, err
	}

	task := &Task{Conn: conn}
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
func (conn *Conn) DeleteDevices(name string) error {
	devices, err := conn.GetDevices(name)
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
func (conn *Conn) DeleteActivities(name string) error {
	activities, err := conn.GetActivities(name)
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
	_, err = conn.Do("del", strings.Replace(ActivitiesKey, "{{user}}", name, -1))
	return err
}

// DeleteTasks deletes all a users tasks
func (conn *Conn) DeleteTasks(name string) error {
	tasks, err := conn.GetTasks(name)
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

// User represents a single users hash data.
type User struct {
	*Conn    `json:"-" redis:"-"`
	Name     string `json:"name" redis:"name"`
	Password string `json:"-" redis:"password"`
}

// Validate ensures the data is valid, if new it'll check if exists.
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

		exists, err := user.UserExists(user.Name)
		if err != nil {
			return nil, err
		}

		if exists {
			return ErrUserAlreadyExists, nil
		}

		return nil, nil
	})
}

// Save saves the user data, hashing the password if needed.
func (user *User) Save(genHash bool) error {
	if genHash {
		pass, err := bcrypt.GenerateFromPassword([]byte(user.Password), -1)
		if err != nil {
			return err
		}

		user.Password = string(pass)
	}

	key := strings.Replace(UserKey, "{{user}}", user.Name, -1)
	_, err := user.Do("hmset", redis.Args{}.Add(key).AddFlat(user)...)
	return err
}

// Delete removes the user data.
func (user *User) Delete() error {
	_, err := user.Do("del", strings.Replace(UserKey, "{{user}}", user.Name, -1))

	return err
}

/*
  Device
*/

// Device represents a single device hash for a user.
type Device struct {
	*Conn `json:"-" redis:"-"`
	Name  string `json:"name" redis:"name"`
	Token string `json:"token" redis:"token"`
	User  *User  `json:"-" redis:"-"`
}

// Validate ensures the data is valid, if new it'll check if it exists.
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

		exists, err := device.DeviceExists(device.User.Name, device.Name)
		if err != nil {
			return nil, err
		}

		if exists {
			return ErrDeviceAlreadyExists, nil
		}

		return nil, nil
	})
}

// Save saves the device data, generating a token if needed.
func (device *Device) Save(genToken bool) error {
	if genToken {
		now := time.Now().String()
		tok, err := uuid.NewV5(uuid.NamespaceURL, []byte(now+device.User.Name+device.Name))
		if err != nil {
			return err
		}

		device.Token = tok.String()
	}

	// Add to device set
	key := strings.Replace(DevicesKey, "{{user}}", device.User.Name, -1)
	_, err := device.Do("sadd", key, device.Name)
	if err != nil {
		return err
	}

	// Add device hash
	key = strings.Replace(DeviceKey, "{{user}}", device.User.Name, -1)
	key = strings.Replace(key, "{{device}}", device.Name, -1)
	_, err = device.Do("hmset", redis.Args{}.Add(key).AddFlat(device)...)
	if err != nil {
		return err
	}

	// Add token hash
	key = strings.Replace(TokenKey, "{{token}}", device.Token, -1)
	_, err = device.Do("hmset", redis.Args{}.Add(key).AddFlat(&Token{device.User.Name, device.Name})...)
	return err
}

// Delete removes the device data.
func (device *Device) Delete() error {
	// Remove token hash
	_, err := device.Do("del", strings.Replace(TokenKey, "{{token}}", device.Token, -1))
	if err != nil {
		return err
	}

	// Remove device hash
	key := strings.Replace(DeviceKey, "{{user}}", device.User.Name, -1)
	_, err = device.Do("del", strings.Replace(key, "{{device}}", device.Name, -1))
	if err != nil {
		return err
	}

	// Remove from device set
	key = strings.Replace(DevicesKey, "{{user}}", device.User.Name, -1)
	_, err = device.Do("srem", key, device.Name)
	return err
}

/*
  Activity
*/

// Activity represents a single activity hash for a user.
type Activity struct {
	*Conn   `json:"-" redis:"-"`
	Message string `json:"message" redis:"message"`
	Time    string `json:"time" redis:"time"`
	User    *User  `json:"-" redis:"-"`
}

// Save saves the activity data.
func (activity *Activity) Save() error {
	activity.Time = time.Now().Format(time.RFC3339)

	// Add to activity list
	key := strings.Replace(ActivitiesKey, "{{user}}", activity.User.Name, -1)
	_, err := activity.Do("lpush", key, activity.Time)
	if err != nil {
		return err
	}

	// Add activity hash
	key = strings.Replace(ActivityKey, "{{user}}", activity.User.Name, -1)
	key = strings.Replace(key, "{{activity}}", activity.Time, -1)
	_, err = activity.Do("hmset", redis.Args{}.Add(key).AddFlat(activity)...)
	return err
}

// Delete removes the activity data.
func (activity *Activity) Delete() error {
	// Unfortunately there's no easy way to remove the list item...so only use this if you plan
	// to remove the list item yourself or deleting all the activities.

	key := strings.Replace(ActivityKey, "{{user}}", activity.User.Name, -1)
	_, err = activity.Do("del", strings.Replace(key, "{{activity}}", activity.Time, -1))
	return err
}

/*
  Task
*/

// Task represents a single task hash for a user.
type Task struct {
	*Conn    `json:"-" redis:"-"`
	ID       int    `json:"id" redis:"id"`
	Message  string `json:"message" redis:"message"`
	Category string `json:"category" redis:"category"`
	Complete bool   `json:"complete" redis:"complete"`
	User     *User  `json:"-" redis:"-"`
}

// Validate ensures the data is valid.
func (task *Task) Validate() ([]string, error) {
	return Validations(func() (error, error) {
		if task.Message == "" {
			return ErrTaskMessageEmpty, nil
		}

		return nil, nil
	})
}

// Save saves the task data, generating an id if needed.
func (task *Task) Save(genID bool) error {
	if genID {
		tasks, err := task.GetTasks(task.User.Name)
		if err != nil {
			return err
		}
		largest := 0

		for _, task := range tasks {
			if task.ID > largest {
				largest = task.ID
			}
		}

		task.ID = largest + 1
	}
	id := strconv.Itoa(task.ID)

	// Add to tasks set
	key := strings.Replace(TasksKey, "{{user}}", task.User.Name, -1)
	_, err := task.Do("sadd", key, id)
	if err != nil {
		return err
	}

	// Add task hash
	key = strings.Replace(TaskKey, "{{user}}", task.User.Name, -1)
	key = strings.Replace(key, "{{task}}", id, -1)
	_, err = task.Do("hmset", redis.Args{}.Add(key).AddFlat(task)...)
	return err
}

// Delete removes the task data.
func (task *Task) Delete() error {
	id := strconv.Itoa(task.ID)

	// Remove task hash
	key := strings.Replace(TaskKey, "{{user}}", task.User.Name, -1)
	_, err = task.Do("del", strings.Replace(key, "{{task}}", id, -1))
	if err != nil {
		return err
	}

	// Remove from task set
	key = strings.Replace(TasksKey, "{{user}}", task.User.Name, -1)
	_, err = task.Do("srem", key, id)
	return err
}

/*
  Token
*/

// Token represents a single token hash for a user and device.
type Token struct {
	User   string `redis:"user"`
	Device string `redis:"device"`
}
