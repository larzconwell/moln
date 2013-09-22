package httpextra

import (
	"net/http"
	"path"
	"strings"
)

// ContentTypes is a map of key value items that are used to parse
// a response format.
var ContentTypes = make(map[string]*ContentType)

// ContentType represents a single content type that includes an extension,
// a error message, and if it's the default.
type ContentType struct {
	Mime      string
	Extension string
	Error     string
	Marshal   func(interface{}) ([]byte, error)
	Default   bool
}

// AddContentType adds a new content type to the list of content types.
func AddContentType(mime, extension, err string, marshal func(interface{}) ([]byte, error), def bool) *ContentType {
	ct := &ContentType{mime, extension, err, marshal, def}

	ContentTypes[mime] = ct

	return ct
}

// DefaultContentType returns the content type set to default, or if none are
// set the first one is returned.
func DefaultContentType() (contentType *ContentType) {
	for _, ct := range ContentTypes {
		if ct.Default {
			contentType = ct
			break
		}
	}

	if contentType == nil {
		for _, ct := range ContentTypes {
			contentType = ct
			break
		}
	}

	return
}

// RequestContentType gets an acceptable response format from a request.
func RequestContentType(req *http.Request) (contentType *ContentType) {
	ext := path.Ext(req.URL.Path)
	if ext == "." {
		ext = ""
	}

	// Check if the extension matches a content type
	if ext != "" {
		for _, ct := range ContentTypes {
			if ct.Extension == ext {
				contentType = ct
				break
			}
		}
		return
	}

	// Check through each accept item and handle items with
	// multiple comma seperated values
	var acceptCheck = func(accept string) {
		accepts := strings.Split(accept, ",")
		for _, t := range accepts {
			if t == "" {
				continue
			}

			// Remove any params
			params := strings.Split(t, ";")
			t = params[0]

			ct, ok := ContentTypes[t]
			if ok {
				contentType = ct
				break
			}

			// Check for wildcards
			if t == "*/*" {
				for _, ct := range ContentTypes {
					contentType = ct
					break
				}
				break
			}

			params = strings.Split(t, "/")
			for ct, _ := range ContentTypes {
				ctSplit := strings.Split(ct, "/")

				if (params[0] == "*" && params[1] == ctSplit[1]) ||
					(params[1] == "*" && params[0] == ctSplit[0]) {
					contentType = ContentTypes[ct]
					break
				}
			}
		}
	}

	// Check if any of the accepted responses are supported
	accepts, ok := req.Header[http.CanonicalHeaderKey("accept")]
	if ok {
		for _, accept := range accepts {
			if accept == "" {
				continue
			}

			acceptCheck(accept)
			if contentType != nil {
				break
			}
		}
		return
	}

	// No extension and no accept header, so just get the default
	return DefaultContentType()
}
