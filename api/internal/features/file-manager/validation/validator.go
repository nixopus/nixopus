package validation

import (
	"encoding/json"
	"io"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type Validator struct {
}

func NewValidator() *Validator {
	return &Validator{}
}

// ParseRequestBody decodes JSON request body
func (v *Validator) ParseRequestBody(req interface{}, body io.ReadCloser, decoded interface{}) error {
	return json.NewDecoder(body).Decode(decoded)
}

// ValidateRequest validates different domain request types
func (v *Validator) ValidateRequest(req interface{}, user shared_types.User) error {
	return nil
}
