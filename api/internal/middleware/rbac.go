package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/cache"
	appStorage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// rbacCache is a package-level cache instance for RBAC permissions
var rbacCache *cache.Cache

// InitRBACCache initializes the RBAC cache with a cache instance
func InitRBACCache(c *cache.Cache) {
	rbacCache = c
}

// RBACMiddleware validates permissions for the given resource based on HTTP method.
// It extracts organization ID from header and validates permissions from the database.
// Uses Redis cache to reduce database calls.
func RBACMiddleware(next http.Handler, app *appStorage.App, resource string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requiredPermission := buildRequiredPermission(resource, r.Method)
		organizationID := extractOrganizationID(w, r)
		if organizationID == "" {
			return
		}

		// Get user from context (set by AuthMiddleware)
		userAny := r.Context().Value(types.UserContextKey)
		if userAny == nil {
			utils.SendErrorResponse(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		user, ok := userAny.(*types.User)
		if !ok {
			utils.SendErrorResponse(w, "Invalid user type in context", http.StatusUnauthorized)
			return
		}

		// Add request to context for Better Auth API calls
		ctx := context.WithValue(r.Context(), "http_request", r)

		// Validate permission
		if !validateUserPermission(ctx, user, organizationID, requiredPermission, app) {
			utils.SendErrorResponse(w, fmt.Sprintf("User lacks permission %s for organization %s", requiredPermission, organizationID), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// validateUserPermission validates user permissions using database
func validateUserPermission(ctx context.Context, user *types.User, organizationID, requiredPermission string, app *appStorage.App) bool {
	// Try cache first
	if result := validateCachedPermissions(user.ID.String(), organizationID, requiredPermission); result != nil {
		return *result
	}

	// Cache miss: fetch from database
	return validateAndCachePermissions(ctx, user, organizationID, requiredPermission, app)
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
func validateAndCachePermissions(ctx context.Context, user *types.User, organizationID, requiredPermission string, app *appStorage.App) bool {
	// Get the HTTP request from context to forward cookies to Better Auth
	req := ctx.Value("http_request")
	var httpReq *http.Request
	if req != nil {
		httpReq, _ = req.(*http.Request)
	}

	// Get user's role from Better Auth organization membership
	member, err := getBetterAuthOrganizationMember(ctx, httpReq, user.ID.String(), organizationID)
	if err != nil || member == nil {
		// If we can't verify membership, deny access
		return false
	}

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
func getBetterAuthOrganizationMember(ctx context.Context, originalReq *http.Request, userID, organizationID string) (*BetterAuthMember, error) {
	betterAuthURL := os.Getenv("BETTER_AUTH_URL")
	if betterAuthURL == "" {
		betterAuthURL = os.Getenv("OCTOAGENT_URL")
	}
	if betterAuthURL == "" {
		betterAuthURL = "http://localhost:9090"
	}

	betterAuthAPI := betterAuthURL + "/api/auth"
	url := fmt.Sprintf("%s/organization/list-members?organizationId=%s", betterAuthAPI, organizationID)

	// Create request with timeout
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Forward cookies from original request if available (for Better Auth authentication)
	if originalReq != nil {
		for _, cookie := range originalReq.Cookies() {
			req.AddCookie(cookie)
		}
	}

	// Make the request with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch organization members: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Better Auth API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var members []BetterAuthMember

	// Try to unmarshal as array first (direct response)
	if err := json.Unmarshal(body, &members); err != nil {
		// If that fails, try as object with data/members field
		var responseObj map[string]interface{}
		if err2 := json.Unmarshal(body, &responseObj); err2 != nil {
			log.Printf("DEBUG getBetterAuthOrganizationMember: Failed to parse response as array or object. Response body: %s", string(body))
			return nil, fmt.Errorf("failed to parse response: %w (also tried object format: %v)", err, err2)
		}

		// Check for common response wrapper fields
		if data, ok := responseObj["data"]; ok {
			// Convert to []BetterAuthMember
			dataBytes, _ := json.Marshal(data)
			if err := json.Unmarshal(dataBytes, &members); err != nil {
				log.Printf("DEBUG getBetterAuthOrganizationMember: Failed to parse data field. Data: %s", string(dataBytes))
				return nil, fmt.Errorf("failed to parse data array: %w", err)
			}
		} else if membersData, ok := responseObj["members"]; ok {
			membersBytes, _ := json.Marshal(membersData)
			if err := json.Unmarshal(membersBytes, &members); err != nil {
				log.Printf("DEBUG getBetterAuthOrganizationMember: Failed to parse members field. Members: %s", string(membersBytes))
				return nil, fmt.Errorf("failed to parse members array: %w", err)
			}
		} else {
			// Try to unmarshal the whole object as a single member (if it's a single member response)
			var singleMember BetterAuthMember
			if err := json.Unmarshal(body, &singleMember); err == nil && singleMember.UserID != "" {
				members = []BetterAuthMember{singleMember}
			} else {
				log.Printf("DEBUG getBetterAuthOrganizationMember: Response is not array or single member. Response: %s", string(body))
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

// getPermissionsForRole returns permissions for a given role
// TODO: This should be fetched from database or a configuration
// TEMPORARY: All roles have all permissions - will be fixed later
func getPermissionsForRole(role string) []string {
	// All permissions for all roles (temporary fix)
	allPermissions := []string{
		"user:create", "user:read", "user:update", "user:delete",
		"organization:create", "organization:read", "organization:update", "organization:delete",
		"role:create", "role:read", "role:update", "role:delete",
		"permission:create", "permission:read", "permission:update", "permission:delete",
		"domain:create", "domain:read", "domain:update", "domain:delete",
		"github-connector:create", "github-connector:read", "github-connector:update", "github-connector:delete",
		"notification:create", "notification:read", "notification:update", "notification:delete",
		"file-manager:create", "file-manager:read", "file-manager:update", "file-manager:delete",
		"deploy:create", "deploy:read", "deploy:update", "deploy:delete",
		"container:create", "container:read", "container:update", "container:delete",
		"audit:create", "audit:read", "audit:update", "audit:delete",
		"terminal:create", "terminal:read", "terminal:update", "terminal:delete",
		"feature_flags:read", "feature_flags:update",
		"dashboard:read", "extension:read", "extension:create", "extension:update", "extension:delete",
		"healthcheck:create", "healthcheck:read", "healthcheck:update", "healthcheck:delete",
	}

	// Return all permissions for any role (temporary)
	return allPermissions
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

// extractOrganizationID extracts and validates organization ID from request header
func extractOrganizationID(w http.ResponseWriter, r *http.Request) string {
	organizationID := r.Header.Get("X-Organization-Id")
	if organizationID == "" {
		utils.SendErrorResponse(w, "Organization ID is required", http.StatusBadRequest)
		return ""
	}

	// Validate UUID format
	if _, err := uuid.Parse(organizationID); err != nil {
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
		if strings.HasPrefix(role, prefix) || role == "admin" || role == "member" || role == "viewer" {
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
