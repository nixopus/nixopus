package validation

import (
	"encoding/json"
	"io"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}


func (v *Validator) ParseRequestBody(req interface{}, body io.ReadCloser, decoded interface{}) error {
	return json.NewDecoder(body).Decode(decoded)
}