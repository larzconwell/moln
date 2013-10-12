package main

import (
	"code.google.com/p/go.crypto/bcrypt"
	"encoding/base64"
	"github.com/larzconwell/moln/httpextra"
	"net/http"
	"strings"
)

// Validations validates a set of tests, returning a slice of validation errors if any.
func Validations(tests ...func() (error, error)) ([]string, error) {
	var (
		validationErrs []string
		fatal          error
		err            error
	)

	for _, test := range tests {
		err, fatal = test()
		if fatal != nil {
			break
		}

		if err != nil {
			if validationErrs == nil {
				validationErrs = make([]string, 0)
			}

			validationErrs = append(validationErrs, err.Error())
		}
	}

	if fatal != nil {
		return nil, fatal
	} else {
		return validationErrs, fatal
	}
}

// HandleValidations responds to a request if any failures or validation errors.
func HandleValidations(rw http.ResponseWriter, req *http.Request, errs []string, err error) bool {
	if errs != nil || err != nil {
		msg := make(map[string]interface{})
		status := http.StatusBadRequest
		res := &httpextra.Response{ContentTypes, rw, req}

		if err != nil {
			msg["error"] = err.Error()
			status = http.StatusInternalServerError
		} else {
			msg["errors"] = errs
		}

		res.Send(msg, status)
		return false
	}

	return true
}

// Authenticate authenticates the request handling responses for required auth.
func Authenticate(conn *Conn, rw http.ResponseWriter, req *http.Request) *User {
	authorization := req.Header.Get("Authorization")
	authType := ""
	authValue := ""

	sendErr := func(err string, status int) {
		if status == http.StatusUnauthorized {
			rw.Header().Set("WWW-Authenticate", "Token")
		}

		res := &httpextra.Response{ContentTypes, rw, req}
		res.Send(map[string]string{"error": err}, status)
	}

	if authorization == "" {
		sendErr(http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return nil
	}

	// Split auth type and value
	authSplit := strings.SplitN(authorization, " ", 2)
	authType = strings.ToLower(authSplit[0])

	if len(authSplit) < 2 || authSplit[1] == "" {
		sendErr(ErrNoAuthorizationValue.Error(), http.StatusUnauthorized)
		return nil
	}
	authValue = authSplit[1]

	if authType == "token" {
		user, err := conn.GetUserByToken(authValue)
		if err != nil {
			sendErr(err.Error(), http.StatusInternalServerError)
			return nil
		}

		if user != nil {
			return user
		}
	}

	if authType == "basic" {
		data, err := base64.StdEncoding.DecodeString(authValue)
		if err != nil {
			sendErr(err.Error(), http.StatusInternalServerError)
			return nil
		}
		dataSplit := strings.SplitN(string(data), ":", 2)
		name := dataSplit[0]

		if len(dataSplit) < 2 || dataSplit[1] == "" {
			sendErr(ErrNoAuthorizationPassword.Error(), http.StatusUnauthorized)
			return nil
		}

		user, err := conn.GetUser(name)
		if err != nil {
			sendErr(err.Error(), http.StatusInternalServerError)
			return nil
		}

		if user != nil {
			matches, err := MatchPass(dataSplit[1], user.Password)
			if err != nil {
				sendErr(err.Error(), http.StatusInternalServerError)
				return nil
			}

			if matches {
				return user
			}
		}
	}

	sendErr(http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	return nil
}

// MatchPass checks if a given password matches a given hashed password
func MatchPass(plainPass, hashPass string) (bool, error) {
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
