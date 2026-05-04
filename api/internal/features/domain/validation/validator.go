package validation

import (
	"fmt"
	"strings"

	"github.com/nixopus/nixopus/api/internal/features/domain/storage"
	"github.com/nixopus/nixopus/api/internal/features/domain/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

type Validator struct {
	storage storage.DomainStorageInterface
	Logger  *logger.Logger // optional; nil disables validation logs
}

func NewValidator(storage storage.DomainStorageInterface) *Validator {
	return NewValidatorWithLogger(storage, nil)
}

// NewValidatorWithLogger is like NewValidator but attaches a logger for Debug detail on rule failures.
func NewValidatorWithLogger(storage storage.DomainStorageInterface, log *logger.Logger) *Validator {
	return &Validator{
		storage: storage,
		Logger:  log,
	}
}

func (v *Validator) log(sev logger.Severity, msg, data string) {
	if v == nil || v.Logger == nil {
		return
	}
	v.Logger.Log(sev, msg, data)
}

func (v *Validator) ValidateName(name string) error {
	v.log(logger.Debug, "validation: ValidateName", fmt.Sprintf("name=%q", name))

	if name == "" {
		v.log(logger.Debug, "validation: ValidateName rejected", "reason=empty")
		return types.ErrMissingDomainName
	}

	if len(name) < 3 {
		v.log(logger.Debug, "validation: ValidateName rejected", fmt.Sprintf("reason=too_short name=%q len=%d", name, len(name)))
		return types.ErrDomainNameTooShort
	}

	if len(name) > 255 {
		v.log(logger.Debug, "validation: ValidateName rejected", fmt.Sprintf("reason=too_long name=%q len=%d", name, len(name)))
		return types.ErrDomainNameTooLong
	}

	if !strings.Contains(name, ".") {
		v.log(logger.Debug, "validation: ValidateName rejected", fmt.Sprintf("reason=no_dot name=%q", name))
		return types.ErrDomainNameInvalid
	}

	tld := strings.Split(name, ".")[1]
	if len(tld) < 2 || len(tld) > 63 {
		v.log(logger.Debug, "validation: ValidateName rejected", fmt.Sprintf("reason=invalid_label name=%q label_len=%d", name, len(tld)))
		return types.ErrDomainNameInvalid
	}

	v.log(logger.Debug, "validation: ValidateName ok", fmt.Sprintf("name=%q", name))
	return nil
}
