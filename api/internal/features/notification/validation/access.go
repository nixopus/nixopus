package validation

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// AccessValidator is the main entry point for access validation
// Now takes pre-parsed resource type, action and user
func (v *Validator) AccessValidator(resourceType, action string, user *shared_types.User) error {
	switch resourceType {
	case "smtp":
		return v.validateSMTPAccess(action, user)
	case "preferences":
		return nil
	default:
		return notification.ErrInvalidResource
	}
}

// parseResourceAndAction extracts resource type and action from request
// Kept as a helper for HTTP handlers
func ParseResourceAndAction(r *http.Request) (string, string) {
	var resourceType, action string

	path := r.URL.Path
	if path == "/api/v1/notification/smtp" || path == "/api/v1/notification/smtp/" {
		resourceType = "smtp"
	} else if path == "/api/v1/notification/preferences" || path == "/api/v1/notification/preferences/" {
		resourceType = "preferences"
	}

	switch r.Method {
	case http.MethodGet:
		action = "read"
	case http.MethodPost:
		action = "create"
	case http.MethodPut:
		action = "update"
	case http.MethodDelete:
		action = "delete"
	}

	return resourceType, action
}

// validateSMTPAccess handles validation for SMTP endpoints with type safety
func (v *Validator) validateSMTPAccess(action string, user *shared_types.User) error {
	return nil
}

// validateCreateSMTPAccess checks if user can create SMTP configs
// Takes the parsed request directly instead of parsing it again
func (v *Validator) ValidateCreateSMTPAccess(req notification.CreateSMTPConfigRequest, user *shared_types.User) error {
	// Extract organization ID from request
	orgID := req.OrganizationID

	// Check if user belongs to the organization
	err := utils.CheckIfUserBelongsToOrganization(user.Organizations, orgID)
	if err != nil {
		return err
	}

	// Check user's role in the organization
	role, err := utils.GetUserRoleInOrganization(user.OrganizationUsers, orgID)
	if err != nil {
		return err
	}

	// Only admin or member roles can create SMTP configs
	if role == shared_types.RoleAdmin || role == shared_types.RoleMember {
		return nil
	}

	return notification.ErrPermissionDenied
}

// validateReadSMTPAccess checks if user can read an SMTP config
// Takes the parsed request directly
func (v *Validator) ValidateReadSMTPAccess(req notification.GetSMTPConfigRequest, user *shared_types.User) error {
	if req.ID.String() == "" {
		return notification.ErrMissingID
	}

	smtp, err := v.storage.GetSmtp(req.ID.String())
	if err != nil {
		return err
	}

	// If user is the creator of the SMTP config, they can read
	if smtp.UserID == user.ID {
		return nil
	}

	// Check if the user belongs to the domain's organization
	err = utils.CheckIfUserBelongsToOrganization(user.Organizations, smtp.OrganizationID)
	if err != nil {
		return err
	}

	// Check user's role in the organization
	role, err := utils.GetUserRoleInOrganization(user.OrganizationUsers, smtp.OrganizationID)
	if err != nil {
		return err
	}

	// Any role (viewer, member, admin) can view SMTP configs in their organization
	if role == shared_types.RoleViewer || role == shared_types.RoleMember || role == shared_types.RoleAdmin {
		return nil
	}

	return notification.ErrPermissionDenied
}

// validateUpdateSMTPAccess checks if user can update an SMTP config
// Takes the parsed request directly
func (v *Validator) ValidateUpdateSMTPAccess(req notification.UpdateSMTPConfigRequest, user *shared_types.User) error {
	// Validate SMTP ID
	if req.ID.String() == "" {
		return notification.ErrMissingID
	}

	// Get SMTP config
	smtp, err := v.storage.GetSmtp(req.ID.String())
	if err != nil {
		return err
	}

	// If user is the creator, they can update
	if smtp.UserID == user.ID {
		return nil
	}

	// Check if organization IDs match
	if req.OrganizationID != smtp.OrganizationID {
		return notification.ErrAccessDenied
	}

	// Check if user is in the same organization
	err = utils.CheckIfUserBelongsToOrganization(user.Organizations, smtp.OrganizationID)
	if err != nil {
		return err
	}

	// Check user's role in the organization
	role, err := utils.GetUserRoleInOrganization(user.OrganizationUsers, smtp.OrganizationID)
	if err != nil {
		return err
	}

	// Only admin or member roles can update
	if role == shared_types.RoleAdmin || role == shared_types.RoleMember {
		return nil
	}

	return notification.ErrPermissionDenied
}

// validateDeleteSMTPAccess checks if user can delete an SMTP config
// Takes the parsed request directly
func (v *Validator) ValidateDeleteSMTPAccess(req notification.DeleteSMTPConfigRequest, user *shared_types.User) error {
	// Validate SMTP ID
	if req.ID.String() == "" {
		return notification.ErrMissingID
	}

	smtp, err := v.storage.GetSmtp(req.ID.String())
	if err != nil {
		return err
	}

	// Check if user is the creator of the SMTP configuration
	if smtp.UserID == user.ID {
		return nil
	}

	// If not creator, check if user is in the same organization
	err = utils.CheckIfUserBelongsToOrganization(user.Organizations, smtp.OrganizationID)
	if err != nil {
		return err
	}

	// Check user's role in the organization
	role, err := utils.GetUserRoleInOrganization(user.OrganizationUsers, smtp.OrganizationID)
	if err != nil {
		return err
	}

	// Only admin role can delete
	if role == shared_types.RoleAdmin {
		return nil
	}

	return notification.ErrPermissionDenied
}
