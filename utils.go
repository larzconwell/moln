package main

import (
	"github.com/larzconwell/moln/httpextra"
	"net/http"
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

	return validationErrs, fatal
}

// HandleValidations responds to a request if any failures or validation errors.
func HandleValidations(rw http.ResponseWriter, req *http.Request, errs []string, err error) bool {
	if errs != nil || err != nil {
		msg := make(map[string]interface{})
		status := http.StatusBadRequest
		res := &httpextra.Response{rw, req}

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
