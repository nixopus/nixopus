package validation

import (
	"strings"

	"github.com/nixopus/nixopus/api/internal/features/domain/storage"
	"github.com/nixopus/nixopus/api/internal/features/domain/types"
)

type Validator struct {
	storage storage.DomainStorageInterface
}

func NewValidator(storage storage.DomainStorageInterface) *Validator {
	return &Validator{
		storage: storage,
	}
}

func (v *Validator) ValidateName(name string) error {
	if name == "" {
		return types.ErrMissingDomainName
	}

	if len(name) < 3 {
		return types.ErrDomainNameTooShort
	}

	if len(name) > 255 {
		return types.ErrDomainNameTooLong
	}

	if !strings.Contains(name, ".") {
		return types.ErrDomainNameInvalid
	}

	tld := strings.Split(name, ".")[1]
	if len(tld) < 2 || len(tld) > 63 {
		return types.ErrDomainNameInvalid
	}

	return nil
}
