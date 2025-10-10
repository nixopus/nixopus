package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	"github.com/supertokens/supertokens-golang/recipe/userroles"
)

func (o *OrganizationService) RemoveUserFromOrganization(request *types.RemoveUserFromOrganizationRequest) error {
	o.logger.Log(logger.Info, "removing user from organization", request.UserID)

	existingOrganization, err := o.storage.GetOrganization(request.OrganizationID)
	if err != nil || existingOrganization.ID == uuid.Nil {
		o.logger.Log(logger.Error, types.ErrOrganizationDoesNotExist.Error(), "")
		return types.ErrOrganizationDoesNotExist
	}

	existingUser, err := o.user_storage.FindUserByID(request.UserID)
	if err != nil || existingUser.ID == uuid.Nil {
		o.logger.Log(logger.Error, types.ErrUserDoesNotExist.Error(), "")
		return types.ErrUserDoesNotExist
	}

	existingUserInOrganization, err := o.storage.FindUserInOrganization(request.UserID, request.OrganizationID)
	if err != nil || existingUserInOrganization.ID == uuid.Nil {
		o.logger.Log(logger.Error, types.ErrUserNotInOrganization.Error(), "")
		return types.ErrUserNotInOrganization
	}

	if err := o.storage.RemoveUserFromOrganization(request.UserID, request.OrganizationID); err != nil {
		o.logger.Log(logger.Error, types.ErrFailedToRemoveUserFromOrganization.Error(), err.Error())
		return types.ErrFailedToRemoveUserFromOrganization
	}

	if existingUser.SupertokensUserID != "" {
		roleName := fmt.Sprintf("orgid_%s_admin", request.OrganizationID)
		if _, err := userroles.RemoveUserRole("public", existingUser.SupertokensUserID, roleName, nil); err != nil {
			o.logger.Log(logger.Warning, "failed to remove SuperTokens role", err.Error())
		}

		roleName = fmt.Sprintf("orgid_%s_member", request.OrganizationID)
		if _, err := userroles.RemoveUserRole("public", existingUser.SupertokensUserID, roleName, nil); err != nil {
			o.logger.Log(logger.Warning, "failed to remove SuperTokens role", err.Error())
		}

		roleName = fmt.Sprintf("orgid_%s_viewer", request.OrganizationID)
		if _, err := userroles.RemoveUserRole("public", existingUser.SupertokensUserID, roleName, nil); err != nil {
			o.logger.Log(logger.Warning, "failed to remove SuperTokens role", err.Error())
		}
	}

	if err := o.cache.InvalidateOrgMembership(o.Ctx, request.UserID, request.OrganizationID); err != nil {
		o.logger.Log(logger.Error, "failed to invalidate organization membership cache", err.Error())
	}

	o.logger.Log(logger.Info, "user removed from organization successfully", request.UserID)
	return nil
}
