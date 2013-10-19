package main

import (
	"code.google.com/p/go.crypto/bcrypt"
	"encoding/base64"
	"github.com/moln/httpextra"
	"net/http"
	"strings"
)

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

// Authenticate authenticates the request handling responses for required auth.
func Authenticate(conn *Conn, rw http.ResponseWriter, req *http.Request) *User {
	authorization := req.Header.Get("Authorization")
	authType := ""

	sendErr := func(err string, status int) {
		res := &httpextra.Response{ContentTypes, rw, req}

		if status == http.StatusUnauthorized {
			if authType == "" {
				authType = "token"
			}
			params := "realm=\"" + req.Host + "\""

			rw.Header().Set("WWW-Authenticate", strings.Title(authType)+" "+params)
		}

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
		sendErr(ErrNoAuthValue.Error(), http.StatusUnauthorized)
		return nil
	}
	authValue := authSplit[1]

	if authType == "token" {
		authValue = strings.SplitN(authValue, " ", 2)[0]

		user, err := tokenAuthenticate(conn, authValue)
		if err != nil {
			sendErr(err.Error(), http.StatusInternalServerError)
			return nil
		}

		if user != nil {
			return user
		}
	}

	if authType == "basic" {
		authValue = strings.SplitN(authValue, " ", 2)[0]
		user, err := basicAuthenticate(conn, authValue)
		if err != nil {
			status := http.StatusInternalServerError
			if err == ErrNoAuthPassword {
				status = http.StatusBadRequest
			}

			sendErr(err.Error(), status)
			return nil
		}

		if user != nil {
			return user
		}
	}

	sendErr(http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	return nil
}

// tokenAuthenticate gets a user from the given token.
func tokenAuthenticate(conn *Conn, token string) (*User, error) {
	return conn.GetUserByToken(token)
}

// basicAuthenticate authenticates according to rfc 2617.
func basicAuthenticate(conn *Conn, userpass string) (*User, error) {
	data, err := base64.StdEncoding.DecodeString(userpass)
	if err != nil {
		return nil, err
	}
	dataSplit := strings.SplitN(string(data), ":", 2)
	name := dataSplit[0]

	if len(dataSplit) < 2 || dataSplit[1] == "" {
		return nil, ErrNoAuthPassword
	}

	user, err := conn.GetUser(name)
	if err != nil {
		return nil, err
	}

	if user != nil {
		matches, err := MatchPass(dataSplit[1], user.Password)
		if err != nil {
			return nil, err
		}

		if matches {
			return user, nil
		} else {
			activity := &Activity{Conn: conn, Message: "Invalid login attempt", User: user}
			err = activity.Save()
			if err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
}
