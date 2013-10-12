package httpextra

import (
	"net/http"
	"net/url"
)

// ParseForm parses the request form, handling errors.
func ParseForm(contentTypes map[string]*ContentType, rw http.ResponseWriter, req *http.Request) (url.Values, bool) {
	err := req.ParseForm()
	if err != nil {
		res := Response{contentTypes, rw, req}
		res.Send(map[string]string{"error": err.Error()}, http.StatusBadRequest)

		return nil, false
	}

	return req.PostForm, true
}
