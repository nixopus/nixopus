package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

// AddUserToOrganization adds a user to an organization.
// This function is mostly used by the admin whenever he creates a new organization
// he will be added as the first user to the organization.
func (o *OrganizationService) AddUserToOrganization(request types.AddUserToOrganizationRequest, tx ...bun.Tx) error {
	o.logger.Log(logger.Info, "adding user to organization", request.UserID)

	var storage storage.OrganizationRepository = o.storage
	if len(tx) > 0 {
		storage = o.storage.WithTx(tx[0])
	}

	existingOrganization, err := storage.GetOrganization(request.OrganizationID)
	if err != nil {
		o.logger.Log(logger.Error, types.ErrOrganizationDoesNotExist.Error(), err.Error())
		return types.ErrOrganizationDoesNotExist
	}

	if existingOrganization == nil || existingOrganization.ID == uuid.Nil {
		o.logger.Log(logger.Error, types.ErrOrganizationDoesNotExist.Error(), "")
		return types.ErrOrganizationDoesNotExist
	}

	existingUser, err := o.user_storage.FindUserByID(request.UserID)
	if err != nil {
		o.logger.Log(logger.Error, types.ErrUserDoesNotExist.Error(), err.Error())
		return types.ErrUserDoesNotExist
	}

	if existingUser == nil || existingUser.ID == uuid.Nil {
		o.logger.Log(logger.Error, types.ErrUserDoesNotExist.Error(), "")
		return types.ErrUserDoesNotExist
	}

	existingUserInOrganization, err := storage.FindUserInOrganization(request.UserID, request.OrganizationID)
	if err != nil {
		o.logger.Log(logger.Error, "failed to check user organization membership", err.Error())
		return types.ErrInternalServer
	}

	if existingUserInOrganization != nil && existingUserInOrganization.ID != uuid.Nil {
		o.logger.Log(logger.Error, types.ErrUserAlreadyInOrganization.Error(), "")
		return types.ErrUserAlreadyInOrganization
	}

	organizationUser := shared_types.OrganizationUsers{
		UserID:         existingUser.ID,
		OrganizationID: existingOrganization.ID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		DeletedAt:      nil,
		ID:             uuid.New(),
	}

	if err := storage.AddUserToOrganization(organizationUser); err != nil {
		o.logger.Log(logger.Error, types.ErrFailedToAddUserToOrganization.Error(), err.Error())
		return types.ErrFailedToAddUserToOrganization
	}

	if err := o.cache.InvalidateOrgMembership(o.Ctx, request.UserID, request.OrganizationID); err != nil {
		o.logger.Log(logger.Error, "failed to invalidate organization membership cache", err.Error())
	}

	o.logger.Log(logger.Info, "user added to organization successfully", request.UserID)
	return nil
}
