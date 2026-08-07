package handler

import (
	"encoding/json"
	"io"
)

// decodeJSON decodes the request body into v.
// It returns a descriptive error if the body is not valid JSON.
func decodeJSON(body io.Reader, v any) error {
	return json.NewDecoder(body).Decode(v)
}
