package validation

import (
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/raghavyuva/nixopus-api/internal/features/extension/storage"
)

type Validator struct {
	validator *validator.Validate
	storage   storage.ExtensionStorageInterface
}

func NewValidator(storage storage.ExtensionStorageInterface) *Validator {
	return &Validator{
		validator: validator.New(),
		storage:   storage,
	}
}

func (v *Validator) ParseRequestBody(r *http.Request, body io.ReadCloser, req interface{}) error {
	return nil
}

func (v *Validator) ValidateRequest(req interface{}) error {
	return v.validator.Struct(req)
}
