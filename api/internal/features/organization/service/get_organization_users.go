package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	"github.com/raghavyuva/nixopus-api/internal/features/supertokens"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// GetOrganizationUsersWithRoles fetches the users for a given organization with their roles and permissions from SuperTokens.
//
// It first checks if the organization exists by calling the storage layer's GetOrganization method.
// If the organization does not exist, it returns ErrOrganizationDoesNotExist.
// If the organization exists, it calls the storage layer's GetOrganizationUsers method to fetch the users.
// For each user, it retrieves their roles and permissions from SuperTokens using their SupertokensUserID.
// If the storage layer returns an error, it returns ErrFailedToGetOrganizationUsers.
// If the storage layer succeeds in fetching the users, it returns the users with their roles and permissions.
func (o *OrganizationService) GetOrganizationUsersWithRoles(id string) ([]shared_types.OrganizationUsersWithRoles, error) {
	o.logger.Log(logger.Info, "getting organization users with roles", id)
	existingOrganization, err := o.storage.GetOrganization(id)
	if err != nil && existingOrganization.ID == uuid.Nil {
		o.logger.Log(logger.Error, types.ErrOrganizationDoesNotExist.Error(), "")
		return nil, types.ErrOrganizationDoesNotExist
	}

	users, err := o.storage.GetOrganizationUsers(id)
	if err != nil {
		o.logger.Log(logger.Error, types.ErrFailedToGetOrganizationUsers.Error(), err.Error())
		return nil, types.ErrFailedToGetOrganizationUsers
	}

	// Convert to users with roles and permissions
	var usersWithRoles []shared_types.OrganizationUsersWithRoles
	for _, user := range users {
		userWithRoles := shared_types.OrganizationUsersWithRoles{
			OrganizationUsers: user,
			Roles:             []string{},
			Permissions:       []string{},
		}

		if user.User != nil && user.User.SupertokensUserID != "" {
			roles, permissions, roleErr := supertokens.GetRolesAndPermissionsForUserInOrganization(user.User.SupertokensUserID, id)
			if roleErr != nil {
				o.logger.Log(logger.Warning, "failed to get roles and permissions for user", roleErr.Error())
			} else {
				userWithRoles.Roles = roles
				userWithRoles.Permissions = permissions
			}
		}

		usersWithRoles = append(usersWithRoles, userWithRoles)
	}

	return usersWithRoles, nil
}
