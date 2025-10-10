package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/validation"
	"github.com/supertokens/supertokens-golang/recipe/passwordless"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *OrganizationsController) customizeInviteLink(inviteLink string, orgID string, email string, role string) string {
	if strings.Contains(inviteLink, "/auth/verify") {
		parts := strings.Split(inviteLink, "/auth/verify")
		if len(parts) == 2 {
			queryAndHash := parts[1]
			if strings.HasPrefix(queryAndHash, "?") {
				queryAndHash = "&" + queryAndHash[1:]
			}
			customInviteLink := fmt.Sprintf("%s/auth/organization-invite?org_id=%s&email=%s&role=%s%s",
				parts[0], orgID, email, role, queryAndHash)
			c.logger.Log(logger.Info, "Customized invite link for organization",
				fmt.Sprintf("Custom InviteLink: %s", customInviteLink))
			return customInviteLink
		}
	}
	return inviteLink
}

func (c *OrganizationsController) SendInvite(f fuego.ContextWithBody[types.InviteSendRequest]) (*shared_types.Response, error) {
	request, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	validator := validation.NewValidator(nil)
	if err := validator.ValidateRequest(&request); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	userContext := &map[string]interface{}{
		"email":           request.Email,
		"organization_id": request.OrganizationID,
		"role":            request.Role,
	}
	inviteLink, err := passwordless.CreateMagicLinkByEmail("public", request.Email, userContext)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	inviteLink = c.customizeInviteLink(inviteLink, request.OrganizationID, request.Email, request.Role)

	return &shared_types.Response{
		Status:  "success",
		Message: "Invite sent successfully",
		Data: map[string]interface{}{
			"email":           request.Email,
			"organization_id": request.OrganizationID,
			"role":            request.Role,
		},
	}, nil
}

func (c *OrganizationsController) ResendInvite(f fuego.ContextWithBody[types.InviteResendRequest]) (*shared_types.Response, error) {
	request, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	validator := validation.NewValidator(nil)
	if err := validator.ValidateRequest(&request); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	userContext := &map[string]interface{}{
		"email":           request.Email,
		"organization_id": request.OrganizationID,
		"role":            request.Role,
	}

	inviteLink, err := passwordless.CreateMagicLinkByEmail("public", request.Email, userContext)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	inviteLink = c.customizeInviteLink(inviteLink, request.OrganizationID, request.Email, request.Role)

	return &shared_types.Response{
		Status:  "success",
		Message: "Invite resent successfully",
		Data: map[string]interface{}{
			"email":           request.Email,
			"organization_id": request.OrganizationID,
			"role":            request.Role,
		},
	}, nil
}
