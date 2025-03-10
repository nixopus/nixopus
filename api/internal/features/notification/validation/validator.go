package validation

import (
	"encoding/json"
	"io"

	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type Validator struct {
	storage storage.NotificationRepository
}

func NewValidator(storage storage.NotificationRepository) *Validator {
	return &Validator{
		storage: storage,
	}
}

func (v *Validator) ValidateRequest(req interface{}, user shared_types.User) error {
	switch r := req.(type) {
	case *notification.CreateSMTPConfigRequest:
		return v.validateCreateSMTPConfigRequest(*r)
	case *notification.GetSMTPConfigRequest:
		return v.validateGetSMTPConfigRequest(*r)
	case *notification.UpdateSMTPConfigRequest:
		return v.validateUpdateSMTPConfigRequest(*r, user)
	case *notification.DeleteSMTPConfigRequest:
		return v.validateDeleteSMTPConfigRequest(*r, user)
	case *notification.UpdatePreferenceRequest:
		return v.validateUpdatePreferenceRequest(*r)
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

func (v *Validator) validateUpdateSMTPConfigRequest(req notification.UpdateSMTPConfigRequest, user shared_types.User) error {
	if req.ID.String() == "" {
		return notification.ErrMissingID
	}

	smtpConfig, err := v.storage.GetSmtp(req.ID.String())

	if err != nil {
		return err
	}

	if smtpConfig.UserID != user.ID || user.Type != "admin" {
		return notification.ErrPermissionDenied
	}

	return nil
}

func (v *Validator) validateDeleteSMTPConfigRequest(req notification.DeleteSMTPConfigRequest, user shared_types.User) error {
	if req.ID.String() == "" {
		return notification.ErrMissingID
	}

	smtpConfig, err := v.storage.GetSmtp(req.ID.String())
	if err != nil {
		return err
	}

	if smtpConfig == nil {
		return notification.ErrSMTPConfigNotFound
	}

	if smtpConfig.UserID  != user.ID || user.Type != "admin" {
		return notification.ErrPermissionDenied
	}

	return nil
}

func (v *Validator) validateGetSMTPConfigRequest(req notification.GetSMTPConfigRequest) error {
	if req.ID.String() == "" {
		return notification.ErrMissingID
	}
	return nil
}

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
