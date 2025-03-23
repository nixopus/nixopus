package validation

import (
	"encoding/json"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/storage"
	"io"
)

type Validator struct {
	storage storage.NotificationRepository
}

func NewValidator(storage storage.NotificationRepository) *Validator {
	return &Validator{
		storage: storage,
	}
}

// ValidateRequest validates different request types
func (v *Validator) ValidateRequest(req interface{}) error {
	switch r := req.(type) {
	case *notification.CreateSMTPConfigRequest:
		return v.validateCreateSMTPConfigRequest(*r)
	case *notification.GetSMTPConfigRequest:
		return v.validateGetSMTPConfigRequest(*r)
	case *notification.UpdateSMTPConfigRequest:
		return v.validateUpdateSMTPConfigRequest(*r)
	case *notification.DeleteSMTPConfigRequest:
		return v.validateDeleteSMTPConfigRequest(*r)
	case *notification.UpdatePreferenceRequest:
		return v.validateUpdatePreferenceRequest(*r)
	default:
		return notification.ErrInvalidRequestType
	}
}

// ParseRequestBody decodes request body into the provided interface
func (v *Validator) ParseRequestBody(req interface{}, body io.ReadCloser, decoded interface{}) error {
	return json.NewDecoder(body).Decode(decoded)
}

// validateCreateSMTPConfigRequest validates create SMTP request fields
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

// validateUpdateSMTPConfigRequest validates update SMTP request fields
func (v *Validator) validateUpdateSMTPConfigRequest(req notification.UpdateSMTPConfigRequest) error {
	if req.ID.String() == "" {
		return notification.ErrMissingID
	}
	return nil
}

// validateDeleteSMTPConfigRequest validates delete SMTP request fields
func (v *Validator) validateDeleteSMTPConfigRequest(req notification.DeleteSMTPConfigRequest) error {
	if req.ID.String() == "" {
		return notification.ErrMissingID
	}
	return nil
}

// validateGetSMTPConfigRequest validates get SMTP request fields
func (v *Validator) validateGetSMTPConfigRequest(req notification.GetSMTPConfigRequest) error {
	if req.ID.String() == "" {
		return notification.ErrMissingID
	}
	return nil
}

// validateUpdatePreferenceRequest validates update preference request fields
func (v *Validator) validateUpdatePreferenceRequest(req notification.UpdatePreferenceRequest) error {
	if req.Type == "" {
		return notification.ErrMissingType
	}
	if req.Category == "" {
		return notification.ErrMissingCategory
	}

	if req.Category != "activity" && req.Category != "security" && req.Category != "update" {
		return notification.ErrInvalidRequestType
	}

	return nil
}
