package main

// Returns a JSON string to send as an error message
func ErrResponse(err string) string {
	return "{\"error\": \"" + err + "\"}"
}
