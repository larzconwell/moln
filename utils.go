package main

import (
	"code.google.com/p/go.crypto/bcrypt"
	"encoding/base64"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/nu7hatch/gouuid"
	"net/http"
	"net/url"
	"strings"
	"time"
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

// GenerateToken creates a unique token from a user name and device name
func GenerateToken(user, device string) (string, error) {
	token, err := uuid.NewV5(uuid.NamespaceURL, []byte(strings.ToLower(user+device)))
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
		token, err := GetToken(token)
		if err == nil {
			authenticated = true
		}

		return authenticated, token["user"], err
	}

	// Check for basic authorization based authentication
	if authType == "basic" {
		data, err := base64.StdEncoding.DecodeString(authValue)
		if err != nil {
			return authenticated, "", err
		}
		dataSplit := strings.SplitN(string(data), ":", 2)
		name := strings.ToLower(dataSplit[0])

		if len(dataSplit) < 2 || dataSplit[1] == "" {
			return authenticated, "", ErrNoAuthorizationPassword
		}

		user, err := GetUser(name)
		if err != nil || user["password"] == "" {
			return authenticated, "", err
		}
		matches, err := MatchPass(user["password"], dataSplit[1])
		if err != nil {
			return authenticated, "", err
		}
		if matches {
			authenticated = true
		}

		return authenticated, name, err
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

// ParseForm parse the request form handling errors
func ParseForm(rw http.ResponseWriter, req *http.Request, res Response) (url.Values, bool) {
	err := req.ParseForm()
	if err != nil {
		res["error"] = err.Error()
		res.Send(rw, req, http.StatusBadRequest)
		return nil, false
	}
	return req.PostForm, true
}

// NewActivityForUser creates a new activity for the user
func NewActivityForUser(user, message string) error {
	t := time.Now().Format(time.RFC3339Nano)

	err := CreateActivity(user, t, message)
	if err != nil {
		return err
	}

	err = AddActivityToUser(user, t)
	return err
}
