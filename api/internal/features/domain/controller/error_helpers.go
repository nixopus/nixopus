package controller

import "github.com/raghavyuva/nixopus-api/internal/features/domain/types"

func isInvalidDomainError(err error) bool {
	switch err {
	case types.ErrInvalidDomainID,
		types.ErrMissingDomainID,
		types.ErrDomainNameInvalid,
		types.ErrDomainNameTooLong,
		types.ErrDomainNameTooShort,
		types.ErrMissingDomainName:
		return true
	default:
		return false
	}
}

func isPermissionError(err error) bool {
	switch err {
	case types.ErrUserDoesNotBelongToOrganization,
		types.ErrPermissionDenied:
		return true
	default:
		return false
	}
}
