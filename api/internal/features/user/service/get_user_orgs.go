package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
)

// GetUserOrganizations retrieves the organizations for a given user.
//
// It first checks if the user exists by calling the storage layer's GetUserById method.
// If the user does not exist, it returns ErrUserDoesNotExist.
// If the user exists, it calls the storage layer's GetUserOrganizationsWithRolesAndPermissions method to fetch the organizations.
// If the storage layer returns an error, it returns ErrFailedToGetUserOrganizations.
// If the storage layer succeeds in fetching the organizations, it returns the organizations.
func (u *UserService) GetUserOrganizations(userID string) ([]types.UserOrganizationsResponse, error) {
	orgs, err := u.storage.GetUserOrganizationsWithRolesAndPermissions(userID)
	if err != nil {
		return []types.UserOrganizationsResponse{}, err
	}

	if len(orgs) == 0 {
		return []types.UserOrganizationsResponse{}, nil
	}

	return orgs, nil
}
