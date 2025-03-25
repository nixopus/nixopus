package types

import (
	"errors"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type UserOrganizationsResponse struct {
	Organization shared_types.Organization `json:"organization"`
	Role         shared_types.Role         `json:"role"`
}

type UpdateUserNameRequest struct {
	Name string `json:"name"`
}

var (
	ErrUserDoesNotExist   = errors.New("user does not exist")
	ErrFailedToUpdateUser = errors.New("failed to update user")
)
