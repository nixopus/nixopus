package service

import (
	"github.com/google/uuid"
	user_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
)

// RemoveUserFromOrganization removes a user from an organization.
//
// It first checks if the organization exists using the organization ID from the request.
// If the organization does not exist, it returns ErrOrganizationDoesNotExist.
// It then checks if the user exists using the user ID from the request.
// If the user does not exist, it returns ErrUserDoesNotExist.
// It also checks if the user is part of the organization using both IDs.
// If the user is not part of the organization, it returns ErrUserNotInOrganization.
// If all checks pass, it calls the storage layer's RemoveUserFromOrganization method
// to remove the user from the organization.
// If the removal fails, it returns ErrFailedToRemoveUserFromOrganization.
// Upon successful removal, it returns nil.
func (o *OrganizationService) RemoveUserFromOrganization(user *types.RemoveUserFromOrganizationRequest) error {
	existingOrganization, err := o.storage.GetOrganization(user.OrganizationID)
	if err == nil && existingOrganization.ID == uuid.Nil {
		return types.ErrOrganizationDoesNotExist
	}
	user_storage := user_storage.UserStorage{
		DB:  o.store.DB,
		Ctx: o.Ctx,
	}
	existingUser, err := user_storage.FindUserByID(user.UserID)
	if err == nil && existingUser.ID == uuid.Nil {
		return types.ErrUserDoesNotExist
	}

	existingUserInOrganization, err := o.storage.FindUserInOrganization(user.UserID, user.OrganizationID)
	if err == nil && existingUserInOrganization.ID == uuid.Nil {
		return types.ErrUserNotInOrganization
	}

	if err := o.storage.RemoveUserFromOrganization(user.UserID, user.OrganizationID); err != nil {
		return types.ErrFailedToRemoveUserFromOrganization
	}

	return nil
}
