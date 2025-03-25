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
	"github.com/raghavyuva/nixopus-api/internal/utils"
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

	return info, nil
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
	// Any authenticated user can create SMTP configs
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
	// Any role (viewer, member, admin) can list domains in their organization
	if role == shared_types.RoleViewer || role == shared_types.RoleMember || role == shared_types.RoleAdmin {
		return nil
	}

	return notification.ErrPermissionDenied
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

	// Only admin or member roles can update
	if role == shared_types.RoleAdmin || role == shared_types.RoleMember {
		return nil
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
