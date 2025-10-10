package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	"github.com/raghavyuva/nixopus-api/internal/features/supertokens"
	"github.com/supertokens/supertokens-golang/recipe/userroles"
)

func (o *OrganizationService) UpdateUserRole(request *types.UpdateUserRoleRequest) error {
	o.logger.Log(logger.Info, "updating user role", request.UserID)

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

	if existingUser.SupertokensUserID != "" {
		oldRoleName := fmt.Sprintf("orgid_%s_admin", request.OrganizationID)
		if _, err := userroles.RemoveUserRole("public", existingUser.SupertokensUserID, oldRoleName, nil); err != nil {
			o.logger.Log(logger.Warning, "failed to remove old SuperTokens role", err.Error())
		}

		oldRoleName = fmt.Sprintf("orgid_%s_member", request.OrganizationID)
		if _, err := userroles.RemoveUserRole("public", existingUser.SupertokensUserID, oldRoleName, nil); err != nil {
			o.logger.Log(logger.Warning, "failed to remove old SuperTokens role", err.Error())
		}

		oldRoleName = fmt.Sprintf("orgid_%s_viewer", request.OrganizationID)
		if _, err := userroles.RemoveUserRole("public", existingUser.SupertokensUserID, oldRoleName, nil); err != nil {
			o.logger.Log(logger.Warning, "failed to remove old SuperTokens role", err.Error())
		}

		newRoleName := fmt.Sprintf("orgid_%s_%s", request.OrganizationID, request.Role)
		var permissions []string
		switch request.Role {
		case "admin":
			permissions = supertokens.GetAdminPermissions()
		case "member":
			permissions = supertokens.GetMemberPermissions()
		case "viewer":
			permissions = supertokens.GetViewerPermissions()
		default:
			permissions = supertokens.GetViewerPermissions()
		}

		if _, err := userroles.CreateNewRoleOrAddPermissions(newRoleName, permissions, nil); err != nil {
			o.logger.Log(logger.Error, "failed to create SuperTokens role", err.Error())
			return types.ErrFailedToUpdateUserRole
		}

		if _, err := userroles.AddRoleToUser("public", existingUser.SupertokensUserID, newRoleName, nil); err != nil {
			o.logger.Log(logger.Error, "failed to assign SuperTokens role", err.Error())
			return types.ErrFailedToUpdateUserRole
		}
	}

	if err := o.cache.InvalidateOrgMembership(o.Ctx, request.UserID, request.OrganizationID); err != nil {
		o.logger.Log(logger.Error, "failed to invalidate organization membership cache", err.Error())
	}

	o.logger.Log(logger.Info, "user role updated successfully", request.UserID)
	return nil
}
