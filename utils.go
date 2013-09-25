package main

import (
	"github.com/larzconwell/moln/httpextra"
	"net/http"
)

// Validate a set of tests, handling validation errors.
func Validate(rw http.ResponseWriter, req *http.Request, tests ...func() (error, error)) bool {
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

	if validationErrs != nil || fatal != nil {
		var msg = make(map[string]interface{})
		status := http.StatusBadRequest
		res := &httpextra.Response{rw, req}

		if fatal != nil {
			msg["error"] = fatal.Error()
			status = http.StatusInternalServerError
		} else {
			msg["errors"] = validationErrs
		}

		res.Send(msg, status)
		return false
	}

	return true
}
