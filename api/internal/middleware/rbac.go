package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/cache"
	betterauth "github.com/nixopus/nixopus/api/internal/features/auth"
	applogger "github.com/nixopus/nixopus/api/internal/features/logger"
	appStorage "github.com/nixopus/nixopus/api/internal/storage"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

// rbacReadResponseBody reads Better Auth HTTP response bodies (swapped in tests).
var rbacReadResponseBody = io.ReadAll

// rbacNewRequestWithContext builds outbound Better Auth membership requests (swapped in tests).
var rbacNewRequestWithContext = http.NewRequestWithContext

// rbacCache is a package-level cache instance for RBAC permissions
var rbacCache *cache.Cache

// InitRBACCache initializes the RBAC cache with a cache instance
func InitRBACCache(c *cache.Cache) {
	rbacCache = c
}

// RBACMiddleware validates permissions for the given resource based on HTTP method.
// It extracts organization ID from header and validates permissions from the database.
// Uses Redis cache to reduce database calls.
func RBACMiddleware(next http.Handler, app *appStorage.App, resource string, l applogger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requiredPermission := buildRequiredPermission(resource, r.Method)
		l.Log(applogger.Debug, "middleware rbac: required permission", fmt.Sprintf("method=%s path=%s permission=%s", r.Method, r.URL.Path, requiredPermission))

		organizationID := extractOrganizationID(w, r, l)
		if organizationID == "" {
			l.Log(applogger.Debug, "middleware rbac: blocked: missing org", fmt.Sprintf("method=%s path=%s", r.Method, r.URL.Path))
			return
		}

		// Get user from context (set by AuthMiddleware)
		userAny := r.Context().Value(types.UserContextKey)
		if userAny == nil {
			l.Log(applogger.Debug, "middleware rbac: blocked: no user in context", fmt.Sprintf("method=%s path=%s", r.Method, r.URL.Path))
			utils.SendErrorResponse(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		user, ok := userAny.(*types.User)
		if !ok {
			l.Log(applogger.Debug, "middleware rbac: blocked: invalid user in context", fmt.Sprintf("method=%s path=%s", r.Method, r.URL.Path))
			utils.SendErrorResponse(w, "Invalid user type in context", http.StatusUnauthorized)
			return
		}

		// Add request to context for Better Auth API calls
		ctx := context.WithValue(r.Context(), "http_request", r)

		// Validate permission
		if !validateUserPermission(ctx, user, organizationID, requiredPermission, app, l) {
			l.Log(applogger.Debug, "middleware rbac: blocked: insufficient permission", fmt.Sprintf("method=%s path=%s user_id=%s permission=%s org_id=%s", r.Method, r.URL.Path, user.ID, requiredPermission, organizationID))
			utils.SendErrorResponse(w, fmt.Sprintf("User lacks permission %s for organization %s", requiredPermission, organizationID), http.StatusForbidden)
			return
		}

		l.Log(applogger.Debug, "middleware rbac: allowed", fmt.Sprintf("method=%s path=%s user_id=%s org_id=%s", r.Method, r.URL.Path, user.ID, organizationID))
		next.ServeHTTP(w, r)
	})
}

// validateUserPermission validates user permissions using database
func validateUserPermission(ctx context.Context, user *types.User, organizationID, requiredPermission string, app *appStorage.App, l applogger.Logger) bool {
	// Try cache first
	if result := validateCachedPermissions(user.ID.String(), organizationID, requiredPermission); result != nil {
		l.Log(applogger.Debug, "middleware rbac: cache hit", fmt.Sprintf("user_id=%s org_id=%s has_perm=%v", user.ID, organizationID, *result))
		return *result
	}

	l.Log(applogger.Debug, "middleware rbac: cache miss", fmt.Sprintf("user_id=%s org_id=%s", user.ID, organizationID))
	// Cache miss: fetch from database
	return validateAndCachePermissions(ctx, user, organizationID, requiredPermission, app, l)
}

// validateCachedPermissions validates permissions using cached data
func validateCachedPermissions(userID, organizationID, requiredPermission string) *bool {
	if rbacCache == nil {
		return nil
	}

	cachedPerms, err := rbacCache.GetRBACPermissions(context.Background(), userID, organizationID)
	if err != nil || cachedPerms == nil {
		return nil
	}

	orgSpecificRoles := filterOrganizationRolesFromStrings(cachedPerms.Roles, organizationID)
	if len(orgSpecificRoles) == 0 {
		result := false
		return &result
	}

	hasPerm := hasPermission(cachedPerms.Permissions, requiredPermission)
	return &hasPerm
}

// validateAndCachePermissions fetches permissions from Better Auth, caches them, and validates
func validateAndCachePermissions(ctx context.Context, user *types.User, organizationID, requiredPermission string, app *appStorage.App, l applogger.Logger) bool {
	// Get the HTTP request from context to forward cookies to Better Auth
	req := ctx.Value("http_request")
	var httpReq *http.Request
	if req != nil {
		httpReq, _ = req.(*http.Request)
	}

	// Get user's role from Better Auth organization membership
	member, err := getBetterAuthOrganizationMember(ctx, httpReq, user.ID.String(), organizationID, l)
	if err != nil || member == nil {
		l.Log(applogger.Debug, "middleware rbac: list-members failed", fmt.Sprintf("user_id=%s org_id=%s error=%v", user.ID, organizationID, err))
		// If we can't verify membership, deny access
		return false
	}

	l.Log(applogger.Debug, "middleware rbac: member resolved", fmt.Sprintf("user_id=%s org_id=%s role=%v", user.ID, organizationID, member.Role))

	// Extract role from Better Auth member data
	// Better Auth can return role as string or array
	var role string
	if member.Role != nil {
		if roleStr, ok := member.Role.(string); ok {
			role = roleStr
		} else if roleArr, ok := member.Role.([]interface{}); ok && len(roleArr) > 0 {
			if roleStr, ok := roleArr[0].(string); ok {
				role = roleStr
			}
		}
	}

	// Default to "member" if no role found
	if role == "" {
		role = "member"
	}

	roles := []string{role}
	permissions := getPermissionsForRole(role)

	l.Log(applogger.Debug, "middleware rbac: role resolved", fmt.Sprintf("role=%s permission_count=%d has_required=%v", role, len(permissions), hasPermission(permissions, requiredPermission)))

	// Cache permissions
	cachePermissions(user.ID.String(), organizationID, roles, permissions)

	// Validate permission
	return hasPermission(permissions, requiredPermission)
}

// BetterAuthMember represents a member from Better Auth organization API
type BetterAuthMember struct {
	ID             string      `json:"id"`
	UserID         string      `json:"userId"`
	OrganizationID string      `json:"organizationId"`
	Role           interface{} `json:"role"` // Can be string or array
	CreatedAt      string      `json:"createdAt"`
	UpdatedAt      string      `json:"updatedAt"`
	User           struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"user"`
}

// getBetterAuthOrganizationMember fetches organization membership from Better Auth API
func getBetterAuthOrganizationMember(ctx context.Context, originalReq *http.Request, userID, organizationID string, l applogger.Logger) (*BetterAuthMember, error) {
	betterAuthURL := os.Getenv("AUTH_SERVICE_URL")
	if betterAuthURL == "" {
		betterAuthURL = "http://localhost:9090"
	}

	betterAuthAPI := betterAuthURL + "/api/auth"
	url := fmt.Sprintf("%s/organization/list-members?organizationId=%s", betterAuthAPI, organizationID)

	// Create request with timeout
	req, err := rbacNewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if originalReq != nil {
		if authHeader := originalReq.Header.Get("Authorization"); authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}
		if apiKey := originalReq.Header.Get("x-api-key"); apiKey != "" {
			req.Header.Set("x-api-key", apiKey)
		}
		for _, cookie := range originalReq.Cookies() {
			req.AddCookie(cookie)
		}
	}

	resp, err := betterauth.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch organization members: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := rbacReadResponseBody(resp.Body)
		return nil, fmt.Errorf("Better Auth API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := rbacReadResponseBody(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var members []BetterAuthMember

	// Try to unmarshal as array first (direct response)
	if err := json.Unmarshal(body, &members); err != nil {
		// If that fails, try as object with data/members field
		var responseObj map[string]interface{}
		if err2 := json.Unmarshal(body, &responseObj); err2 != nil {
			l.Log(applogger.Debug, "middleware rbac: list-members parse failed", fmt.Sprintf("expected_array_or_object body=%s", string(body)))
			return nil, fmt.Errorf("failed to parse response: %w (also tried object format: %v)", err, err2)
		}

		// Check for common response wrapper fields
		if data, ok := responseObj["data"]; ok {
			// Convert to []BetterAuthMember
			dataBytes, _ := json.Marshal(data)
			if err := json.Unmarshal(dataBytes, &members); err != nil {
				l.Log(applogger.Debug, "middleware rbac: list-members data field parse failed", fmt.Sprintf("data=%s", string(dataBytes)))
				return nil, fmt.Errorf("failed to parse data array: %w", err)
			}
		} else if membersData, ok := responseObj["members"]; ok {
			membersBytes, _ := json.Marshal(membersData)
			if err := json.Unmarshal(membersBytes, &members); err != nil {
				l.Log(applogger.Debug, "middleware rbac: list-members members field parse failed", fmt.Sprintf("members=%s", string(membersBytes)))
				return nil, fmt.Errorf("failed to parse members array: %w", err)
			}
		} else {
			// Try to unmarshal the whole object as a single member (if it's a single member response)
			var singleMember BetterAuthMember
			if err := json.Unmarshal(body, &singleMember); err == nil && singleMember.UserID != "" {
				members = []BetterAuthMember{singleMember}
			} else {
				l.Log(applogger.Debug, "middleware rbac: list-members unexpected shape", fmt.Sprintf("body=%s", string(body)))
				return nil, fmt.Errorf("response does not contain array or single member: %s", string(body))
			}
		}
	}

	// Find the current user in the members list
	for i := range members {
		if members[i].UserID == userID || members[i].User.ID == userID {
			return &members[i], nil
		}
	}

	// User not found in organization
	return nil, fmt.Errorf("user %s is not a member of organization %s", userID, organizationID)
}

// rolePermissions defines permissions per organization role.
// owner/admin: full access; member: create/read/update/delete on most resources; viewer: read-only.
var rolePermissions = map[string][]string{
	"owner": {
		"user:create", "user:read", "user:update", "user:delete",
		"organization:create", "organization:read", "organization:update", "organization:delete",
		"role:create", "role:read", "role:update", "role:delete",
		"permission:create", "permission:read", "permission:update", "permission:delete",
		"domain:create", "domain:read", "domain:update", "domain:delete",
		"github-connector:create", "github-connector:read", "github-connector:update", "github-connector:delete",
		"notification:create", "notification:read", "notification:update", "notification:delete",
		"deploy:create", "deploy:read", "deploy:update", "deploy:delete",
		"container:create", "container:read", "container:update", "container:delete",
		"audit:create", "audit:read", "audit:update", "audit:delete",
		"terminal:create", "terminal:read", "terminal:update", "terminal:delete",
		"feature_flags:read", "feature_flags:update",
		"dashboard:read", "extension:read", "extension:create", "extension:update", "extension:delete",
		"healthcheck:create", "healthcheck:read", "healthcheck:update", "healthcheck:delete",
		"server:create", "server:read", "server:update", "server:delete",
		"trail:create", "trail:read", "trail:update", "trail:delete",
		"execute:create", "execute:read", "execute:update", "execute:delete",
		"machine:create", "machine:read", "machine:update", "machine:delete",
		"mcp:create", "mcp:read", "mcp:update", "mcp:delete",
	},
	"admin": {
		"user:create", "user:read", "user:update", "user:delete",
		"organization:create", "organization:read", "organization:update", "organization:delete",
		"role:create", "role:read", "role:update", "role:delete",
		"permission:create", "permission:read", "permission:update", "permission:delete",
		"domain:create", "domain:read", "domain:update", "domain:delete",
		"github-connector:create", "github-connector:read", "github-connector:update", "github-connector:delete",
		"notification:create", "notification:read", "notification:update", "notification:delete",
		"deploy:create", "deploy:read", "deploy:update", "deploy:delete",
		"container:create", "container:read", "container:update", "container:delete",
		"audit:create", "audit:read", "audit:update", "audit:delete",
		"terminal:create", "terminal:read", "terminal:update", "terminal:delete",
		"feature_flags:read", "feature_flags:update",
		"dashboard:read", "extension:read", "extension:create", "extension:update", "extension:delete",
		"healthcheck:create", "healthcheck:read", "healthcheck:update", "healthcheck:delete",
		"server:create", "server:read", "server:update", "server:delete",
		"trail:create", "trail:read", "trail:update", "trail:delete",
		"execute:create", "execute:read", "execute:update", "execute:delete",
		"machine:create", "machine:read", "machine:update", "machine:delete",
		"mcp:create", "mcp:read", "mcp:update", "mcp:delete",
	},
	"member": {
		"user:read",
		"organization:read",
		"role:read", "permission:read",
		"domain:create", "domain:read", "domain:update", "domain:delete",
		"github-connector:create", "github-connector:read", "github-connector:update", "github-connector:delete",
		"notification:create", "notification:read", "notification:update", "notification:delete",
		"deploy:create", "deploy:read", "deploy:update", "deploy:delete",
		"container:create", "container:read", "container:update", "container:delete",
		"audit:read",
		"terminal:create", "terminal:read", "terminal:update", "terminal:delete",
		"feature_flags:read",
		"dashboard:read", "extension:read", "extension:create", "extension:update", "extension:delete",
		"healthcheck:create", "healthcheck:read", "healthcheck:update", "healthcheck:delete",
		"server:create", "server:read", "server:update", "server:delete",
		"trail:create", "trail:read", "trail:update", "trail:delete",
		"execute:create", "execute:read", "execute:update", "execute:delete",
		"machine:create", "machine:read", "machine:update",
		"mcp:read", "mcp:update",
	},
	"viewer": {
		"user:read", "organization:read", "role:read", "permission:read",
		"domain:read", "github-connector:read", "notification:read",
		"deploy:read", "container:read", "audit:read", "terminal:read",
		"feature_flags:read", "dashboard:read", "extension:read",
		"healthcheck:read", "server:read", "trail:read", "execute:read",
		"machine:read", "mcp:read",
	},
}

// getPermissionsForRole returns permissions for a given organization role.
// Unknown roles default to viewer (read-only) for safety.
func getPermissionsForRole(role string) []string {
	role = strings.ToLower(strings.TrimSpace(role))
	if perms, ok := rolePermissions[role]; ok {
		return perms
	}
	// Custom roles (e.g. orgid_xxx_custom) or unknown: default to viewer
	return rolePermissions["viewer"]
}

// cacheRBACPermissionsFromMember extracts role from member and caches permissions.
// Called by Auth middleware when it fetches member data, so RBAC avoids duplicate API call.
func cacheRBACPermissionsFromMember(userID, organizationID string, member *BetterAuthMember) {
	var role string
	if member.Role != nil {
		if roleStr, ok := member.Role.(string); ok {
			role = roleStr
		} else if roleArr, ok := member.Role.([]interface{}); ok && len(roleArr) > 0 {
			if roleStr, ok := roleArr[0].(string); ok {
				role = roleStr
			}
		}
	}
	if role == "" {
		role = "member"
	}
	roles := []string{role}
	permissions := getPermissionsForRole(role)
	cachePermissions(userID, organizationID, roles, permissions)
}

// cachePermissions caches user permissions for future requests
func cachePermissions(userID, organizationID string, roles, permissions []string) {
	if rbacCache == nil {
		return
	}

	cachedPerms := &cache.CachedRBACPermissions{
		Roles:       roles,
		Permissions: permissions,
	}
	_ = rbacCache.SetRBACPermissions(context.Background(), userID, organizationID, cachedPerms)
}

// buildRequiredPermission constructs the required permission string from resource and HTTP method
func buildRequiredPermission(resource, method string) string {
	action := getActionFromMethod(method)
	return resource + ":" + action
}

// extractOrganizationID extracts and validates organization ID from request header or auth context.
// Falls back to OrganizationIDKey from AuthMiddleware (Better Auth session) when header is missing.
func extractOrganizationID(w http.ResponseWriter, r *http.Request, l applogger.Logger) string {
	organizationID := r.Header.Get("X-Organization-Id")
	if organizationID == "" {
		// Fallback: use org from auth context (set by AuthMiddleware from Better Auth session)
		if orgAny := r.Context().Value(types.OrganizationIDKey); orgAny != nil {
			if orgStr, ok := orgAny.(string); ok && orgStr != "" {
				organizationID = orgStr
				l.Log(applogger.Debug, "middleware rbac: using org from context", fmt.Sprintf("method=%s path=%s", r.Method, r.URL.Path))
			}
		}
	}
	if organizationID == "" {
		l.Log(applogger.Debug, "middleware rbac: missing X-Organization-Id and context org", fmt.Sprintf("method=%s path=%s", r.Method, r.URL.Path))
		utils.SendErrorResponse(w, "Organization ID is required", http.StatusBadRequest)
		return ""
	}

	// Validate UUID format
	if _, err := uuid.Parse(organizationID); err != nil {
		l.Log(applogger.Debug, "middleware rbac: invalid X-Organization-Id", fmt.Sprintf("org_id=%s method=%s path=%s", organizationID, r.Method, r.URL.Path))
		utils.SendErrorResponse(w, "Invalid organization ID format", http.StatusBadRequest)
		return ""
	}

	return organizationID
}

// filterOrganizationRolesFromStrings filters roles to only include organization-specific roles
func filterOrganizationRolesFromStrings(roles []string, organizationID string) []string {
	prefix := buildOrgRolePrefix(organizationID)
	var orgSpecificRoles []string

	for _, role := range roles {
		if strings.HasPrefix(role, prefix) || role == "owner" || role == "admin" || role == "member" || role == "viewer" {
			orgSpecificRoles = append(orgSpecificRoles, role)
		}
	}

	return orgSpecificRoles
}

// hasPermission checks if the required permission exists in the permissions list
func hasPermission(permissions []string, requiredPermission string) bool {
	for _, perm := range permissions {
		if perm == requiredPermission {
			return true
		}
	}
	return false
}

// buildOrgRolePrefix builds the organization role prefix
func buildOrgRolePrefix(organizationID string) string {
	return "orgid_" + organizationID + "_"
}

// getActionFromMethod maps HTTP methods to permission actions
func getActionFromMethod(method string) string {
	switch method {
	case http.MethodGet:
		return "read"
	case http.MethodPost:
		return "create"
	case http.MethodPut, http.MethodPatch:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return "read"
	}
}
