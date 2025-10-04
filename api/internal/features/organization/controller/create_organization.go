package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	"github.com/raghavyuva/nixopus-api/internal/features/supertokens"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/supertokens/supertokens-golang/recipe/userroles"
)

func (c *OrganizationsController) CreateOrganization(f fuego.ContextWithBody[types.CreateOrganizationRequest]) (*shared_types.Response, error) {
	organization, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	c.logger.Log("Creating organization", organization.Name, organization.Description)

	w, r := f.Response(), f.Request()
	if err := c.validator.ValidateRequest(&organization); err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	loggedInUser := utils.GetUser(w, r)
	if loggedInUser == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	createdOrganization, err := c.service.CreateOrganization(&organization)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	c.service.AddUserToOrganization(types.AddUserToOrganizationRequest{
		UserID:         loggedInUser.ID.String(),
		OrganizationID: createdOrganization.ID.String(),
	})

	// Create organization specific roles and assign admin role to the creator
	if err := c.createOrganizationRoles(createdOrganization.ID.String(), loggedInUser.SupertokensUserID); err != nil {
		c.logger.Log(logger.Error, "Failed to create organization roles", err.Error())
		// Don't fail the entire operation for role creation failure
	}

	// c.Notify(notification.NortificationPayloadTypeCreateOrganization, loggedInUser, r, createdOrganization)

	return &shared_types.Response{
		Status:  "success",
		Message: "Organization created successfully",
		Data:    createdOrganization,
	}, nil
}

// createOrganizationRoles creates organization specific roles and assigns admin role to the creator
func (c *OrganizationsController) createOrganizationRoles(organizationID, supertokensUserID string) error {
	// Create organization specific admin role
	adminRoleName := fmt.Sprintf("orgid_%s_admin", organizationID)
	if _, err := userroles.CreateNewRoleOrAddPermissions(adminRoleName, supertokens.GetAdminPermissions(), nil); err != nil {
		return fmt.Errorf("failed to create admin role %s: %w", adminRoleName, err)
	}

	// Create organization specific member role
	memberRoleName := fmt.Sprintf("orgid_%s_member", organizationID)
	if _, err := userroles.CreateNewRoleOrAddPermissions(memberRoleName, supertokens.GetMemberPermissions(), nil); err != nil {
		return fmt.Errorf("failed to create member role %s: %w", memberRoleName, err)
	}

	// Create organization specific viewer role
	viewerRoleName := fmt.Sprintf("orgid_%s_viewer", organizationID)
	if _, err := userroles.CreateNewRoleOrAddPermissions(viewerRoleName, supertokens.GetViewerPermissions(), nil); err != nil {
		return fmt.Errorf("failed to create viewer role %s: %w", viewerRoleName, err)
	}

	// Assign admin role to the organization creator
	if _, err := userroles.AddRoleToUser("public", supertokensUserID, adminRoleName, nil); err != nil {
		return fmt.Errorf("failed to assign admin role %s to user %s: %w", adminRoleName, supertokensUserID, err)
	}

	c.logger.Log(logger.Info, "Successfully created organization roles", organizationID)
	return nil
}
