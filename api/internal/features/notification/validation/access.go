package validation

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"path"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// RequestInfo encapsulates parsed request details
type RequestInfo struct {
	Path           string
	Method         string
	ResourceType   string
	Action         string
	SMTPID         string
	OrganizationID uuid.UUID
}

// AccessValidator is the main entry point for access validation
func (v *Validator) AccessValidator(w http.ResponseWriter, r *http.Request, user *shared_types.User) error {
	reqInfo, err := parseRequest(r)
	if err != nil {
		return err
	}

	if reqInfo.SMTPID == "" && reqInfo.ResourceType == "smtp" &&
		(reqInfo.Action == "read" || reqInfo.Action == "update" || reqInfo.Action == "delete") {
		reqInfo.SMTPID = v.extractIDFromBody(r)
	}

	switch reqInfo.ResourceType {
	case "smtp":
		return v.validateSMTPAccess(reqInfo, user)
	case "preferences":
		return nil
	default:
		return notification.ErrInvalidResource
	}
}

// parseRequest extracts basic information from the HTTP request
func parseRequest(r *http.Request) (*RequestInfo, error) {
	info := &RequestInfo{
		Path:   r.URL.Path,
		Method: r.Method,
		SMTPID: r.URL.Query().Get("id"),
	}

	basePath := path.Base(r.URL.Path)
	switch {
	case path.Dir(r.URL.Path) == "/api/v1/notification" && basePath == "smtp":
		info.ResourceType = "smtp"
		switch r.Method {
		case http.MethodGet:
			info.Action = "read"
		case http.MethodPost:
			info.Action = "create"
		case http.MethodPut:
			info.Action = "update"
		case http.MethodDelete:
			info.Action = "delete"
		}
	case path.Dir(r.URL.Path) == "/api/v1/notification" && basePath == "preferences":
		info.ResourceType = "preferences"
	}

	if info.SMTPID == "" {
		info.SMTPID = extractIDFromPath(r.URL.Path)
	}

	return info, nil
}

// extractIDFromPath gets ID from URL path segments
func extractIDFromPath(urlPath string) string {
	lastSegment := path.Base(urlPath)
	return lastSegment
}

// extractIDFromBody attempts to extract ID from the request body
func (v *Validator) extractIDFromBody(r *http.Request) string {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" && contentType != "application/json; charset=utf-8" {
		return ""
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return ""
	}

	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var requestData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &requestData); err != nil {
		return ""
	}

	if idVal, ok := requestData["id"]; ok {
		switch id := idVal.(type) {
		case string:
			return id
		case map[string]interface{}:
			if strID, ok := id["String"].(string); ok {
				return strID
			}
		}
	}

	return ""
}

// validateSMTPAccess handles access validation for SMTP endpoints
func (v *Validator) validateSMTPAccess(req *RequestInfo, user *shared_types.User) error {
	switch req.Action {
	case "create":
		return v.validateCreateSMTPAccess(user)
	case "read":
		return v.validateReadSMTPAccess(req, user)
	case "update":
		return v.validateUpdateSMTPAccess(req, user)
	case "delete":
		return v.validateDeleteSMTPAccess(req, user)
	default:
		return notification.ErrInvalidRequestType
	}
}

// validateCreateSMTPAccess checks if user can create SMTP configs
func (v *Validator) validateCreateSMTPAccess(user *shared_types.User) error {
	// Only admin can create SMTP configurations
	if user.Type != shared_types.RoleAdmin {
		return notification.ErrAccessDenied
	}
	return nil
}

// validateReadSMTPAccess checks if user can read an SMTP config
func (v *Validator) validateReadSMTPAccess(req *RequestInfo, user *shared_types.User) error {
	if req.SMTPID == "" {
		return notification.ErrMissingID
	}

	smtp, err := v.storage.GetSmtp(req.SMTPID)
	if err != nil {
		return err
	}

	// if user is the creator of the SMTP config, they can read
	if smtp.UserID == user.ID {
		return nil
	}

	req.OrganizationID = smtp.OrganizationID

	// Admin can always read
	if user.Type == shared_types.RoleAdmin {
		return nil
	}

	// Non-admin users must belong to the organization and have read access
	if err := v.checkIfUserBelongsToOrganization(user.Organizations, smtp.OrganizationID); err != nil {
		return err
	}

	return v.checkIfUserHasReadAccess(user.OrganizationUsers, smtp.OrganizationID)
}

// validateUpdateSMTPAccess checks if user can update an SMTP config
func (v *Validator) validateUpdateSMTPAccess(req *RequestInfo, user *shared_types.User) error {
	if req.SMTPID == "" {
		return notification.ErrMissingID
	}

	smtp, err := v.storage.GetSmtp(req.SMTPID)
	if err != nil {
		return err
	}
	req.OrganizationID = smtp.OrganizationID

	// User must be the creator of the SMTP configuration
	if smtp.UserID != user.ID {
		return notification.ErrPermissionDenied
	}

	return nil
}

// validateDeleteSMTPAccess checks if user can delete an SMTP config
func (v *Validator) validateDeleteSMTPAccess(req *RequestInfo, user *shared_types.User) error {
	if req.SMTPID == "" {
		return notification.ErrMissingID
	}

	smtp, err := v.storage.GetSmtp(req.SMTPID)
	if err != nil {
		return err
	}
	req.OrganizationID = smtp.OrganizationID

	// User must be the creator of the SMTP configuration
	if smtp.UserID != user.ID {
		return notification.ErrPermissionDenied
	}

	return nil
}

// checkIfUserBelongsToOrganization verifies if a user belongs to a specific organization
func (v *Validator) checkIfUserBelongsToOrganization(userOrgs []shared_types.Organization, orgID uuid.UUID) error {
	for _, org := range userOrgs {
		if org.ID == orgID {
			return nil
		}
	}
	return notification.ErrUserDoesNotBelongToOrganization
}

// checkIfUserHasReadAccess verifies if a user has read access to a specific organization.
func (v *Validator) checkIfUserHasReadAccess(userOrgs []shared_types.OrganizationUsers, orgID uuid.UUID) error {
	for _, userOrg := range userOrgs {
		if userOrg.OrganizationID != orgID {
			continue
		}

		if userOrg.Role == nil {
			continue
		}

		// If the user has specific permissions, check if they have "read" access on the "organization" resource
		if userOrg.Role.Permissions != nil {
			for _, permission := range userOrg.Role.Permissions {
				if permission.Name == "read" && permission.Resource == "organization" {
					return nil
				}
			}
		}
	}

	return notification.ErrUserDoesNotHavePermissionForTheResource
}
