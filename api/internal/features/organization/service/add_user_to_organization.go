package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// AddUserToOrganization adds a user to an organization.
//
// It first checks if the organization and role exist using the IDs from the request.
// If the organization does not exist, it returns ErrOrganizationDoesNotExist.
// If the role does not exist, it returns ErrRoleDoesNotExist.
// It then checks if the user exists using the user ID from the request.
// If the user does not exist, it returns ErrUserDoesNotExist.
// It also checks if the user is already part of the organization using both IDs.
// If the user is already in the organization, it returns ErrUserAlreadyInOrganization.
// If all checks pass, it calls the storage layer's AddUserToOrganization method to add the user to the organization.
// If the addition fails, it returns ErrFailedToAddUserToOrganization.
// Upon successful addition, it returns nil.
func (o *OrganizationService) AddUserToOrganization(request types.AddUserToOrganizationRequest) error {
	o.logger.Log(logger.Info, "adding user to organization", request.UserID)
	roleId, err := uuid.Parse(request.RoleId)
	if err != nil {
		o.logger.Log(logger.Error, types.ErrInvalidRoleID.Error(), err.Error())
		return types.ErrInvalidRoleID
	}

	existingOrganization, err := o.storage.GetOrganization(request.OrganizationID)
	if err != nil {
		o.logger.Log(logger.Error, types.ErrOrganizationDoesNotExist.Error(), err.Error())
		return err
	}

	if existingOrganization.ID == uuid.Nil {
		o.logger.Log(logger.Error, types.ErrOrganizationDoesNotExist.Error(), "")
		return types.ErrOrganizationDoesNotExist
	}

	existingUser, err := o.user_storage.FindUserByID(request.UserID)
	if err != nil {
		o.logger.Log(logger.Error, types.ErrUserDoesNotExist.Error(), err.Error())
		return err
	}

	if existingUser.ID == uuid.Nil {
		o.logger.Log(logger.Error, types.ErrUserDoesNotExist.Error(), "")
		return types.ErrUserDoesNotExist
	}

	existingRole, err := o.role_storage.GetRole(roleId.String())
	if err != nil {
		o.logger.Log(logger.Error, types.ErrRoleDoesNotExist.Error(), err.Error())
		return err
	}
	if existingRole.ID == uuid.Nil {
		o.logger.Log(logger.Error, types.ErrRoleDoesNotExist.Error(), "")
		return types.ErrRoleDoesNotExist
	}

	existingUserInOrganization, err := o.storage.FindUserInOrganization(request.UserID, request.OrganizationID)
	if err != nil {
		o.logger.Log(logger.Error, types.ErrFailedToAddUserToOrganization.Error(), err.Error())
		return err
	}
	if existingUserInOrganization.ID != uuid.Nil {
		o.logger.Log(logger.Error, types.ErrUserAlreadyInOrganization.Error(), "")
		return types.ErrUserAlreadyInOrganization
	}

	organizationUser := shared_types.OrganizationUsers{
		UserID:         existingUser.ID,
		OrganizationID: existingOrganization.ID,
		RoleID:         roleId,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		DeletedAt:      nil,
		ID:             uuid.New(),
	}

	if err := o.storage.AddUserToOrganization(organizationUser); err != nil {
		o.logger.Log(logger.Error, types.ErrFailedToAddUserToOrganization.Error(), err.Error())
		return types.ErrFailedToAddUserToOrganization
	}

	return nil
}
