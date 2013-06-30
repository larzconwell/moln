package main

import (
	"code.google.com/p/go.crypto/bcrypt"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/nu7hatch/gouuid"
	"net/http"
	"strings"
)

var (
	// Authentication errors
	ErrTokenNotExist           = errors.New("Authentication: token does not exist")
	ErrNoAuthorizationValue    = errors.New("Authentication: authorization header value missing")
	ErrNoAuthorizationPassword = errors.New("Authentication: authorization header password missing")

	// Authorization errors
	ErrUserNotAuthorized = errors.New("Authorization: user is not authorized to access this page")

	// User errors
  ErrUserNotExist = errors.New("User: user does not exist")

	// Validation errors
	ErrUserNameEmpty     = errors.New("User: name cannot be empty")
	ErrUserPasswordEmpty = errors.New("User: password cannot be empty")
	ErrDeviceNameEmpty   = errors.New("Device: name cannot be empty")
	ErrUserAlreadyExists = errors.New("User: name already exists")
)

// createAndAppendErrorMsgs creates the given list if nil then appends the items to it
func createAndAppendErrorMsgs(list []string, items ...string) []string {
	if list == nil {
		list = make([]string, 0)
	}

	return append(list, items...)
}

// Validate tests each of the given functions, appending the first error if not nil, if the
// second error is not nil, validate returns with the error
func Validate(tests ...func() (error, error)) ([]string, error) {
	var errs []string

	for _, test := range tests {
		err, fatal := test()
		if fatal != nil {
			return nil, fatal
		}

		if err != nil {
			errs = createAndAppendErrorMsgs(errs, err.Error())
		}
	}

	return errs, nil
}

// HashPass creates a hashed password from the given plain text password
func HashPass(plainPass string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(plainPass), -1)
}

// MatchPass checks if a given password matches a given hashed password
func MatchPass(hashPass, plainPass string) (bool, error) {
	match := false

	err := bcrypt.CompareHashAndPassword([]byte(hashPass), []byte(plainPass))
	if err == nil {
		match = true
	}
	if err == bcrypt.ErrMismatchedHashAndPassword {
		err = nil
	}

	return match, err
}

// CreateToken creates a unique token from a user name and device name
func CreateToken(user, device string) (string, error) {
	token, err := uuid.NewV5(uuid.NamespaceURL, []byte(user+device))
	if err != nil {
		return "", err
	}

	return token.String(), nil
}

// Authenticate authenticates the request returning the authenticated users name
func Authenticate(req *http.Request) (bool, string, error) {
	authenticated := false

	// Token from URL/Body query
	token := req.FormValue("token")
	authorization := req.Header.Get("Authorization")
	authType := ""
	authValue := ""

	// Get auth type from authorization
	if authorization != "" {
		authSplit := strings.SplitN(authorization, " ", 2)
		authType = strings.ToLower(authSplit[0])

		// If no token we can assume their using auth based, so check to
		// ensure a value was given
		if token == "" {
			if len(authSplit) < 2 || authSplit[1] == "" {
				return authenticated, "", ErrNoAuthorizationValue
			}

			authValue = authSplit[1]
		}
	}

	// If no token value then get it from authValue
	if token == "" && authType == "token" {
		token = authValue
	}

	// Check for token based authentication
	if token != "" {
		user, err := redis.String(DB.Do("hget", "token:"+token, "user"))
		if err == redis.ErrNil {
			err = ErrTokenNotExist
		}
		if err == nil {
			authenticated = true
		}

		return authenticated, user, err
	}

	// Check for basic authorization based authentication
	if authType == "basic" {
		data, err := base64.StdEncoding.DecodeString(authValue)
		if err != nil {
			return authenticated, "", err
		}
		dataSplit := strings.SplitN(string(data), ":", 2)
		user := dataSplit[0]

		if len(dataSplit) < 2 || dataSplit[1] == "" {
			return authenticated, "", ErrNoAuthorizationPassword
		}

		pass, err := redis.String(DB.Do("hget", "users:"+user, "password"))
		if err == redis.ErrNil {
			err = ErrUserNotExist
		}
		matches, err := MatchPass(pass, dataSplit[1])
		if err != nil {
			return authenticated, "", err
		}
		if matches {
			authenticated = true
		}

		return authenticated, user, err
	}

	return authenticated, "", nil
}

// ToMap converts a redis multi-bulk reply to a map with key value strings
func ToMap(reply interface{}, err error) (map[string]string, error) {
	if err != nil {
		return nil, err
	}

	switch reply := reply.(type) {
	case []interface{}:
		result := make(map[string]string)
		key := ""

		for i, v := range reply {
			if v == nil {
				continue
			}

			v, ok := v.([]byte)
			if !ok {
				return nil, fmt.Errorf("ToMap: unexpected element type %T", v)
			}
			str := string(v)

			if i%2 == 0 {
				// Key name
				result[str] = ""
				key = str
			} else {
				// Key value
				result[key] = str
			}
		}

		return result, nil
	case nil:
		return nil, redis.ErrNil
	case redis.Error:
		return nil, reply
	}

	return nil, fmt.Errorf("ToMap: unexpected type %T", reply)
}
