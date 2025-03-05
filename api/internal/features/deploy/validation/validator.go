package validation

import (
	"encoding/json"
	"io"
)

type Validator struct {
}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ParseRequestBody(req interface{}, body io.ReadCloser, decoded interface{}) error {
	return json.NewDecoder(body).Decode(decoded)
}

func (v *Validator) ValidateRequest(req interface{}) error {
	// switch r := req.(type) {
	// case *types.CreateDeployRequest:
	// 	return v.validateCreateDeployRequest(*r)
	// case *types.UpdateDeployRequest:
	// 	return v.validateUpdateDeployRequest(*r)
	// case *types.DeleteDeployRequest:
	// 	return v.validateDeleteDeployRequest(*r)
	// default:
	// 	return types.ErrInvalidRequestType
	// }
	return nil
}
