package service_deprecated

import (
// "github.com/google/uuid"
// "github.com/raghavyuva/nixopus-api/internal/features/logger"
// "github.com/raghavyuva/nixopus-api/internal/features/organization/types"
)

// Deprecated: Use SupertokensUpdateUserRole instead
//
// It first checks if the organization exists using the organization ID.
// If the organization does not exist, it returns ErrOrganizationDoesNotExist.
// It then checks if the user exists using the user ID.
// If the user does not exist, it returns ErrUserDoesNotExist.
// It also checks if the user is part of the organization using both IDs.
// If the user is not part of the organization, it returns ErrUserNotInOrganization.
// If all checks pass, it calls the storage layer's UpdateUserRole method
// to update the user's role in the organization.
// If the update fails, it returns ErrFailedToUpdateUserRole.
// Upon successful update, it returns nil.
func (o *OrganizationService) UpdateUserRole(userID, organizationID, role string) error {
	// o.logger.Log(logger.Info, "updating user role", userID)
	// existingOrganization, err := o.storage.GetOrganization(organizationID)
	// if err == nil && existingOrganization.ID == uuid.Nil {
	// 	o.logger.Log(logger.Error, types.ErrOrganizationDoesNotExist.Error(), "")
	// 	return types.ErrOrganizationDoesNotExist
	// }

	// existingUser, err := o.user_storage.FindUserByID(userID)
	// if err == nil && existingUser.ID == uuid.Nil {
	// 	o.logger.Log(logger.Error, types.ErrUserDoesNotExist.Error(), "")
	// 	return types.ErrUserDoesNotExist
	// }

	// existingUserInOrganization, err := o.storage.FindUserInOrganization(userID, organizationID)
	// if err == nil && existingUserInOrganization.ID == uuid.Nil {
	// 	o.logger.Log(logger.Error, types.ErrUserNotInOrganization.Error(), "")
	// 	return types.ErrUserNotInOrganization
	// }

	// existingRole, err := o.role_storage.GetRoleByName(role)
	// if err == nil && existingRole.ID == uuid.Nil {
	// 	o.logger.Log(logger.Error, types.ErrRoleDoesNotExist.Error(), "")
	// 	return types.ErrRoleDoesNotExist
	// }

	// if err := o.storage.UpdateUserRole(userID, organizationID, existingRole.ID); err != nil {
	// 	o.logger.Log(logger.Error, types.ErrFailedToUpdateUserRole.Error(), err.Error())
	// 	return types.ErrFailedToUpdateUserRole
	// }

	// // Invalidate cache for organization membership
	// if err := o.cache.InvalidateOrgMembership(o.Ctx, userID, organizationID); err != nil {
	// 	o.logger.Log(logger.Error, "failed to invalidate organization membership cache", err.Error())
	// }

	return nil
}
