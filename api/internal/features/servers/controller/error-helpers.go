package controller

import "github.com/raghavyuva/nixopus-api/internal/features/servers/types"

func isInvalidServerError(err error) bool {
	switch err {
	case types.ErrInvalidServerID,
		types.ErrMissingServerID,
		types.ErrServerNameInvalid,
		types.ErrServerNameTooLong,
		types.ErrServerNameTooShort,
		types.ErrMissingServerName,
		types.ErrMissingHost,
		types.ErrInvalidHost,
		types.ErrMissingPort,
		types.ErrInvalidPort,
		types.ErrMissingUsername,
		types.ErrMissingSSHAuth,
		types.ErrBothSSHAuthProvided,
		types.ErrInvalidSSHPrivateKeyPath,
		types.ErrSSHConnectionFailed,
		types.ErrSSHAuthenticationFailed,
		types.ErrMissingStatus,
		types.ErrInvalidStatus:
		return true
	default:
		return false
	}
}

func isPermissionError(err error) bool {
	switch err {
	case types.ErrUserDoesNotBelongToOrganization,
		types.ErrPermissionDenied,
		types.ErrAccessDenied,
		types.ErrUserDoesNotHavePermissionForTheResource:
		return true
	default:
		return false
	}
}
