package service_deprecated

import (
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// Deprecated: Use SupertokensGetRoles instead
//
// It calls the role storage's GetRoles method to fetch all roles with their permissions.
// If the storage layer returns an error, it returns ErrFailedToGetRoles.
// If the storage layer succeeds in fetching the roles, it returns the roles.
func (o *OrganizationService) GetRoles() ([]shared_types.Role, error) {
	return []shared_types.Role{}, nil
}
