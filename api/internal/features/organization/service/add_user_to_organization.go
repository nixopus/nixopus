package service

import (
	"time"

	"github.com/google/uuid"
	user_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	role_storage "github.com/raghavyuva/nixopus-api/internal/features/role/storage"
)

func (o *OrganizationService) AddUserToOrganization(user types.AddUserToOrganizationRequest, organization shared_types.Organization) error {
	roleId, err := uuid.Parse(user.RoleId)
	if err != nil {
		return types.ErrInvalidRoleID
	}

	existingOrganization, err := o.storage.GetOrganization(user.OrganizationID)
	if err != nil {
		return err
	}
	if existingOrganization.ID == uuid.Nil {
		return types.ErrOrganizationDoesNotExist
	}

	user_storage := user_storage.UserStorage{
		DB:  o.store.DB,
		Ctx: o.Ctx,
	}

	existingUser, err := user_storage.FindUserByID(user.UserID)
	if err != nil {
		return err
	}
	if existingUser.ID == uuid.Nil {
		return types.ErrUserDoesNotExist
	}

	role_storage := role_storage.RoleStorage{
		DB:  o.store.DB,
		Ctx: o.Ctx,
	}
	existingRole, err := role_storage.GetRole(roleId.String())
	if err != nil {
		return err
	}
	if existingRole.ID == uuid.Nil {
		return types.ErrRoleDoesNotExist
	}

	existingUserInOrganization, err := o.storage.FindUserInOrganization(user.UserID, user.OrganizationID)
	if err != nil {
		return err
	}
	if existingUserInOrganization.ID != uuid.Nil {
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
		return types.ErrFailedToAddUserToOrganization
	}

	return nil
}
