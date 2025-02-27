package service

import (
	"time"

	"github.com/google/uuid"
	user_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	role_storage "github.com/raghavyuva/nixopus-api/internal/features/role/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

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

	user_storage := user_storage.UserStorage{
		DB:  o.storage.DB,
		Ctx: o.storage.Ctx,
	}
	existingUser, err := user_storage.FindUserByID(request.UserID)
	if err != nil {
		o.logger.Log(logger.Error, types.ErrUserDoesNotExist.Error(), err.Error())
		return err
	}

	if existingUser.ID == uuid.Nil {
		o.logger.Log(logger.Error, types.ErrUserDoesNotExist.Error(), "")
		return types.ErrUserDoesNotExist
	}

	role_storage := role_storage.RoleStorage{
		DB:  o.storage.DB,
		Ctx: o.storage.Ctx,
	}
	existingRole, err := role_storage.GetRole(roleId.String())
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
