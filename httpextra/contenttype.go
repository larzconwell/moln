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
	Default   bool
}

// AddContentType adds a new content type to the list of content types.
func AddContentType(mime, extension, err string, def bool) *ContentType {
	ct := &ContentType{mime, extension, err, def}

	ContentTypes[mime] = ct

	return ct
}

// ContentTypeSupported checks if a given request accepts a supported format.
func ContentTypeSupported(req *http.Request) (supported bool) {
	ext := path.Ext(req.URL.Path)
	if ext == "." {
		ext = ""
	}

	// Check if the extension matches a content type
	if ext != "" {
		for _, ct := range ContentTypes {
			if ct.Extension == ext {
				supported = true
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

			_, ok := ContentTypes[t]
			if ok || t == "*/*" {
				supported = true
				break
			}

			// Check for wildcards
			params = strings.Split(t, "/")
			for ct, _ := range ContentTypes {
				ctSplit := strings.Split(ct, "/")

				if (params[0] == "*" && params[1] == ctSplit[1]) ||
					(params[1] == "*" && params[0] == ctSplit[0]) {
					supported = true
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
			if supported {
				break
			}
		}
	}

	// No extension and no accept headers, so find a default
	for _, ct := range ContentTypes {
		if ct.Default {
			supported = true
			break
		}
	}

	return
}
