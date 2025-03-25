package utils

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

// checkIfUserBelongsToOrganization verifies if a user belongs to a specific organization
func CheckIfUserBelongsToOrganization(userOrgs []types.Organization, orgID uuid.UUID) error {
	for _, org := range userOrgs {
		if org.ID == orgID {
			return nil
		}
	}
	return types.ErrUserDoesNotBelongToOrganization
}

// getUserRoleInOrganization determines the user's role in an organization
func GetUserRoleInOrganization(userOrgs []types.OrganizationUsers, orgID uuid.UUID) (string, error) {
	for _, userOrg := range userOrgs {
		if userOrg.OrganizationID == orgID {
			if userOrg.Role == nil {
				return "", types.ErrNoRoleAssigned
			}

			return userOrg.Role.Name, nil
		}
	}

	return "", types.ErrUserDoesNotBelongToOrganization
}
