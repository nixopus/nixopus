package validation

import (
	"encoding/json"
	"io"

	"github.com/raghavyuva/nixopus-api/internal/features/notification"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateRequest(req interface{}) error {
	switch r := req.(type) {
	case notification.CreateSMTPConfigRequest:
		return v.validateCreateSMTPConfigRequest(r)
	case notification.GetSMTPConfigRequest:
		return v.validateGetSMTPConfigRequest(r)
	case notification.UpdateSMTPConfigRequest:
		return v.validateUpdateSMTPConfigRequest(r)
	case notification.DeleteSMTPConfigRequest:
		return v.validateDeleteSMTPConfigRequest(r)
	default:
		return notification.ErrInvalidRequestType
	}
}

func (v *Validator) ParseRequestBody(req interface{}, body io.ReadCloser, decoded interface{}) error {
	return json.NewDecoder(body).Decode(decoded)
}

func (v *Validator) validateCreateSMTPConfigRequest(req notification.CreateSMTPConfigRequest) error {
	if req.Host == "" {
		return notification.ErrMissingHost
	}
	if req.Port == 0 {
		return notification.ErrMissingPort
	}
	if req.Username == "" {
		return notification.ErrMissingUsername
	}
	if req.Password == "" {
		return notification.ErrMissingPassword
	}
	return nil
}

func (v *Validator) validateUpdateSMTPConfigRequest(req notification.UpdateSMTPConfigRequest) error {
	if req.ID.String() == "" {
		return notification.ErrMissingID
	}
	return nil
}

func (v *Validator) validateDeleteSMTPConfigRequest(req notification.DeleteSMTPConfigRequest) error {
	if req.ID.String() == "" {
		return notification.ErrMissingID
	}
	return nil
}

func (v *Validator) validateGetSMTPConfigRequest(req notification.GetSMTPConfigRequest) error {
	if req.ID.String() == "" {
		return notification.ErrMissingID
	}
	return nil
}