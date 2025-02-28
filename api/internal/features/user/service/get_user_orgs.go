package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
)

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
