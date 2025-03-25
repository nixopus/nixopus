package validation

import (
	"net/http"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// RequestInfo encapsulates parsed request details
type RequestInfo struct {
	Path           string
	Method         string
	ResourceType   string
	Action         string
	DomainID       string
	OrganizationID string
}

// AccessValidator is the main entry point for access validation
func (v *Validator) AccessValidator(w http.ResponseWriter, r *http.Request, user *shared_types.User, req interface{}) error {
	reqInfo, err := parseRequest(r)
	if err != nil {
		return err
	}

	if reqInfo.DomainID == "" && reqInfo.ResourceType == "domain" &&
		(reqInfo.Action == "update" || reqInfo.Action == "delete") {
		reqInfo.DomainID = v.extractIDFromBody(req)
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
		Path:   r.URL.Path,
		Method: r.Method,
	}

	pathDir := path.Dir(r.URL.Path)
	basePath := path.Base(r.URL.Path)

	if pathDir == "/api/v1" && basePath == "domain" {
		info.ResourceType = "domain"
		switch r.Method {
		case http.MethodPost:
			info.Action = "create"
		case http.MethodPut:
			info.Action = "update"
		case http.MethodDelete:
			info.Action = "delete"
		case http.MethodGet:
			info.Action = "read"
		}
	}

	if pathDir == "/api/v1" && basePath == "domains" {
		info.ResourceType = "domain"
		if r.Method == http.MethodGet {
			info.Action = "list"
			info.OrganizationID = r.URL.Query().Get("id")
		}
	}

	if strings.HasPrefix(r.URL.Path, "/api/v1/domain/") && r.Method == http.MethodGet {
		info.ResourceType = "domain"
		info.Action = "read"
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 4 {
			info.DomainID = parts[4]
		}
	}

	return info, nil
}

// extractIDFromBody attempts to extract ID from the request body
func (v *Validator) extractIDFromBody(req interface{}) string {
	if req == nil {
		return ""
	}
	switch r := req.(type) {
	case *types.UpdateDomainRequest:
		return r.ID
	case *types.DeleteDomainRequest:
		return r.ID
	default:
		return ""
	}
}

// validateDomainAcess handles access validation for domain endpoints
func (v *Validator) validateDomainAcess(req *RequestInfo, user *shared_types.User) error {
	switch req.Action {
	case "create":
		return v.validateCreateDomainAccess(user)
	case "read":
		return v.validateReadDomainAccess(req, user)
	case "list":
		return v.validateListDomainsAccess(req, user)
	case "update":
		return v.validateUpdateDomainAccess(req, user)
	case "delete":
		return v.validateDeleteDomainAccess(req, user)
	default:
		return types.ErrInvalidRequestType
	}
}

// validateCreateDomainAccess checks if user can create domain configs
func (v *Validator) validateCreateDomainAccess(user *shared_types.User) error {
	// Any authenticated user can create domains
	return nil
}

// validateReadDomainAccess checks if user can read a specific domain
func (v *Validator) validateReadDomainAccess(req *RequestInfo, user *shared_types.User) error {
	if req.DomainID == "" {
		return types.ErrMissingID
	}

	domain, err := v.storage.GetDomain(req.DomainID)
	if err != nil {
		return err
	}
	req.OrganizationID = domain.OrganizationID.String()

	// User can read domain if they created it
	if domain.UserID == user.ID {
		return nil
	}

	// Check if the user belongs to the domain's organization
	err = utils.CheckIfUserBelongsToOrganization(user.Organizations, domain.OrganizationID)
	if err != nil {
		return err
	}

	// Check user's role in the organization
	role, err := utils.GetUserRoleInOrganization(user.OrganizationUsers, domain.OrganizationID)
	if err != nil {
		return err
	}
	// Any role (viewer, member, admin) can list domains in their organization
	if role == shared_types.RoleViewer || role == shared_types.RoleMember || role == shared_types.RoleAdmin {
		return nil
	}

	return types.ErrPermissionDenied
}

// validateListDomainsAccess checks if user can list domains
func (v *Validator) validateListDomainsAccess(req *RequestInfo, user *shared_types.User) error {
	// Check if the user belongs to the domain's organization
	err := utils.CheckIfUserBelongsToOrganization(user.Organizations, uuid.MustParse(req.OrganizationID))
	if err != nil {
		return err
	}

	// Check user's role in the organization
	role, err := utils.GetUserRoleInOrganization(user.OrganizationUsers, uuid.MustParse(req.OrganizationID))
	if err != nil {
		return err
	}

	// Any role (viewer, member, admin) can list domains in their organization
	if role == shared_types.RoleViewer || role == shared_types.RoleMember || role == shared_types.RoleAdmin {
		return nil
	}

	return types.ErrPermissionDenied
}

// validateUpdateDomainAccess checks if user can update a domain
func (v *Validator) validateUpdateDomainAccess(req *RequestInfo, user *shared_types.User) error {
	if req.DomainID == "" {
		return types.ErrMissingID
	}

	domain, err := v.storage.GetDomain(req.DomainID)
	if err != nil {
		return err
	}
	req.OrganizationID = domain.OrganizationID.String()

	// User can always update domains they created
	if domain.UserID == user.ID {
		return nil
	}

	// If not creator, check if user is in the same organization
	err = utils.CheckIfUserBelongsToOrganization(user.Organizations, domain.OrganizationID)
	if err != nil {
		return err
	}

	// Check user's role in the organization
	role, err := utils.GetUserRoleInOrganization(user.OrganizationUsers, domain.OrganizationID)
	if err != nil {
		return err
	}

	// Only admin or member roles can update
	if role == shared_types.RoleAdmin || role == shared_types.RoleMember {
		return nil
	}

	return types.ErrPermissionDenied
}

// validateDeleteDomainAccess checks if user can delete a domain
func (v *Validator) validateDeleteDomainAccess(req *RequestInfo, user *shared_types.User) error {
	if req.DomainID == "" {
		return types.ErrMissingID
	}

	domain, err := v.storage.GetDomain(req.DomainID)
	if err != nil {
		return err
	}
	req.OrganizationID = domain.OrganizationID.String()

	// User can always delete domains they created
	if domain.UserID == user.ID {
		return nil
	}

	// If not creator, check if user is in the same organization
	err = utils.CheckIfUserBelongsToOrganization(user.Organizations, domain.OrganizationID)
	if err != nil {
		return err
	}

	// Check user's role in the organization
	role, err :=utils.GetUserRoleInOrganization(user.OrganizationUsers, domain.OrganizationID)
	if err != nil {
		return err
	}

	// Only admin role can delete
	if role == shared_types.RoleAdmin {
		return nil
	}

	return types.ErrPermissionDenied
}
