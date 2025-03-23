package validation

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"path"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// RequestInfo encapsulates parsed request details
type RequestInfo struct {
	Path           string
	Method         string
	ResourceType   string
	Action         string
	DomainID       string
	OrganizationID uuid.UUID
}

// AccessValidator is the main entry point for access validation
func (v *Validator) AccessValidator(w http.ResponseWriter, r *http.Request, user *shared_types.User) error {
	reqInfo, err := parseRequest(r)
	if err != nil {
		return err
	}

	if reqInfo.DomainID == "" && reqInfo.ResourceType == "domain" &&
		(reqInfo.Action == "read" || reqInfo.Action == "update" || reqInfo.Action == "delete") {
		reqInfo.DomainID = v.extractIDFromBody(r)
	}

	switch reqInfo.ResourceType {
	case "domain":
		return v.validateDomainAcess(reqInfo, user)
	default:
		return types.ErrInvalidResource
	}
}

// parseRequest extracts basic information from the HTTP request
func parseRequest(r *http.Request) (*RequestInfo, error) {
	info := &RequestInfo{
		Path:     r.URL.Path,
		Method:   r.Method,
		DomainID: r.URL.Query().Get("id"),
	}

	basePath := path.Base(r.URL.Path)

	switch {
	case path.Dir(r.URL.Path) == "/api/v1" && basePath == "domain":
		info.ResourceType = "domain"
		switch r.Method {
		case http.MethodPost:
			info.Action = "create"
		case http.MethodPut:
			info.Action = "update"
		case http.MethodDelete:
			info.Action = "delete"
		}
	}

	if info.DomainID == "" {
		info.DomainID = extractIDFromPath(r.URL.Path)
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

// validateDomainAcess handles access validation for SMTP endpoints
func (v *Validator) validateDomainAcess(req *RequestInfo, user *shared_types.User) error {
	switch req.Action {
	case "create":
		return v.validateCreateDomainAccess(user)
	case "update":
		return v.validateUpdateDomainAccess(req, user)
	case "delete":
		return v.validateDeleteSMTPAccess(req, user)
	default:
		return types.ErrInvalidRequestType
	}
}

// validateCreateDomainAccess checks if user can create SMTP configs
func (v *Validator) validateCreateDomainAccess(user *shared_types.User) error {
	// Only admin can create SMTP configurations
	if user.Type != shared_types.RoleAdmin {
		return types.ErrAccessDenied
	}
	return nil
}

// validateUpdateDomainAccess checks if user can update an SMTP config
func (v *Validator) validateUpdateDomainAccess(req *RequestInfo, user *shared_types.User) error {
	if req.DomainID == "" {
		return types.ErrMissingID
	}

	domain, err := v.storage.GetDomain(req.DomainID)
	if err != nil {
		return err
	}
	req.OrganizationID = domain.OrganizationID

	// User must be the creator of the SMTP configuration
	if domain.UserID != user.ID {
		return types.ErrPermissionDenied
	}

	return nil
}

// validateDeleteSMTPAccess checks if user can delete an SMTP config
func (v *Validator) validateDeleteSMTPAccess(req *RequestInfo, user *shared_types.User) error {
	if req.DomainID == "" {
		return types.ErrMissingID
	}

	domain, err := v.storage.GetDomain(req.DomainID)
	if err != nil {
		return err
	}
	req.OrganizationID = domain.OrganizationID

	// User must be the creator of the SMTP configuration
	if domain.UserID != user.ID {
		return types.ErrPermissionDenied
	}

	return nil
}
