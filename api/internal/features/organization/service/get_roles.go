package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// GetRoles retrieves all roles with their permissions from the database.
//
// It calls the role storage's GetRoles method to fetch all roles with their permissions.
// If the storage layer returns an error, it returns ErrFailedToGetRoles.
// If the storage layer succeeds in fetching the roles, it returns the roles.
func (o *OrganizationService) GetRoles() ([]shared_types.Role, error) {
	o.logger.Log(logger.Info, "getting all roles with permissions", "")

	roles, err := o.role_storage.GetRoles()
	if err != nil {
		o.logger.Log(logger.Error, types.ErrFailedToGetRoles.Error(), err.Error())
		return nil, types.ErrFailedToGetRoles
	}

	return roles, nil
}
