package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/cache"
	appStorage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/recipe/session/claims"
	"github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
	"github.com/supertokens/supertokens-golang/recipe/userroles/userrolesclaims"
	"github.com/supertokens/supertokens-golang/supertokens"
)

// rbacCache is a package-level cache instance for RBAC permissions
var rbacCache *cache.Cache

// InitRBACCache initializes the RBAC cache with a cache instance
func InitRBACCache(c *cache.Cache) {
	rbacCache = c
}

// RBACMiddleware validates SuperTokens permission claims for the given resource based on HTTP method.
// It extracts organization ID from header and validates permissions only for organization-specific roles.
// Uses Redis cache to reduce SuperTokens API calls.
func RBACMiddleware(next http.Handler, _ *appStorage.App, resource string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requiredPermission := buildRequiredPermission(resource, r.Method)
		organizationID := extractOrganizationID(w, r)
		if organizationID == "" {
			return
		}

		validator := createPermissionValidator(organizationID, requiredPermission)
		handler := session.VerifySession(&sessmodels.VerifySessionOptions{
			OverrideGlobalClaimValidators: validator,
		}, func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})

		handler.ServeHTTP(w, r)
	})
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
	return organizationID
}

// createPermissionValidator creates a claim validator for organization permissions
func createPermissionValidator(organizationID, requiredPermission string) func([]claims.SessionClaimValidator, sessmodels.SessionContainer, supertokens.UserContext) ([]claims.SessionClaimValidator, error) {
	return func(globalClaimValidators []claims.SessionClaimValidator, sessionContainer sessmodels.SessionContainer, userContext supertokens.UserContext) ([]claims.SessionClaimValidator, error) {
		validator := claims.SessionClaimValidator{
			ID: "org-permission-validator",
			Validate: func(payload map[string]interface{}, userContext supertokens.UserContext) claims.ClaimValidationResult {
				return validateUserPermission(sessionContainer, organizationID, requiredPermission)
			},
		}

		globalClaimValidators = append(globalClaimValidators, validator)
		return globalClaimValidators, nil
	}
}

// validateUserPermission validates user permissions, checking cache first, then SuperTokens
func validateUserPermission(sessionContainer sessmodels.SessionContainer, organizationID, requiredPermission string) claims.ClaimValidationResult {
	userID := sessionContainer.GetUserID()

	// Try cache first
	if result := validateCachedPermissions(userID, organizationID, requiredPermission); result != nil {
		return *result
	}

	// Cache miss: fetch from SuperTokens
	return validateAndCachePermissions(sessionContainer, userID, organizationID, requiredPermission)
}

// validateCachedPermissions validates permissions using cached data
func validateCachedPermissions(userID, organizationID, requiredPermission string) *claims.ClaimValidationResult {
	if rbacCache == nil {
		return nil
	}

	cachedPerms, err := rbacCache.GetRBACPermissions(context.Background(), userID, organizationID)
	if err != nil || cachedPerms == nil {
		return nil
	}

	orgSpecificRoles := filterOrganizationRolesFromStrings(cachedPerms.Roles, organizationID)
	if len(orgSpecificRoles) == 0 {
		return &claims.ClaimValidationResult{
			IsValid: false,
			Reason:  fmt.Sprintf("user has no roles for organization %s", organizationID),
		}
	}

	if !hasPermission(cachedPerms.Permissions, requiredPermission) {
		return &claims.ClaimValidationResult{
			IsValid: false,
			Reason:  fmt.Sprintf("user lacks permission %s for organization %s", requiredPermission, organizationID),
		}
	}

	return &claims.ClaimValidationResult{IsValid: true}
}

// validateAndCachePermissions fetches permissions from SuperTokens, caches them, and validates
func validateAndCachePermissions(sessionContainer sessmodels.SessionContainer, userID, organizationID, requiredPermission string) claims.ClaimValidationResult {
	userRolesClaim, permissionsClaim, err := fetchClaims(sessionContainer)
	if err != nil {
		return claims.ClaimValidationResult{
			IsValid: false,
			Reason:  err.Error(),
		}
	}

	// Extract and cache permissions
	roles := extractRolesAsStrings(userRolesClaim)
	permissions := extractPermissionsAsStrings(permissionsClaim)
	cachePermissions(userID, organizationID, roles, permissions)

	// Validate organization roles
	orgSpecificRoles := filterOrganizationRoles(userRolesClaim, organizationID)
	if len(orgSpecificRoles) == 0 {
		return claims.ClaimValidationResult{
			IsValid: false,
			Reason:  fmt.Sprintf("user has no roles for organization %s", organizationID),
		}
	}

	// Validate permission
	if !checkOrganizationPermission(permissionsClaim, requiredPermission) {
		return claims.ClaimValidationResult{
			IsValid: false,
			Reason:  fmt.Sprintf("user lacks permission %s for organization %s", requiredPermission, organizationID),
		}
	}

	return claims.ClaimValidationResult{IsValid: true}
}

// fetchClaims fetches and validates user roles and permissions claims from SuperTokens
func fetchClaims(sessionContainer sessmodels.SessionContainer) (interface{}, interface{}, error) {
	if err := sessionContainer.FetchAndSetClaim(userrolesclaims.UserRoleClaim); err != nil {
		return nil, nil, fmt.Errorf("failed to fetch user roles: %v", err)
	}

	if err := sessionContainer.FetchAndSetClaim(userrolesclaims.PermissionClaim); err != nil {
		return nil, nil, fmt.Errorf("failed to fetch permissions: %v", err)
	}

	userRolesResult, err := session.GetClaimValue(sessionContainer.GetHandle(), userrolesclaims.UserRoleClaim)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user roles: %v", err)
	}

	permissionsResult, err := session.GetClaimValue(sessionContainer.GetHandle(), userrolesclaims.PermissionClaim)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get permissions: %v", err)
	}

	if userRolesResult.OK == nil {
		return nil, nil, fmt.Errorf("user roles claim not found in session")
	}

	if permissionsResult.OK == nil {
		return nil, nil, fmt.Errorf("permissions claim not found in session")
	}

	return userRolesResult.OK, permissionsResult.OK, nil
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

// extractRolesAsStrings extracts roles from the claim as a string slice
func extractRolesAsStrings(userRolesClaim interface{}) []string {
	return extractStringSlice(userRolesClaim)
}

// extractPermissionsAsStrings extracts permissions from the claim as a string slice
func extractPermissionsAsStrings(permissionsClaim interface{}) []string {
	return extractStringSlice(permissionsClaim)
}

// extractStringSlice extracts a string slice from a wrapped claim value
func extractStringSlice(claim interface{}) []string {
	if claim == nil {
		return []string{}
	}

	switch v := claim.(type) {
	case []string:
		return v
	case []interface{}:
		return convertInterfaceSliceToStringSlice(v)
	}

	actualValue := extractValueFromWrapper(claim)
	if actualValue == nil {
		return []string{}
	}

	switch v := actualValue.(type) {
	case []string:
		return v
	case []interface{}:
		return convertInterfaceSliceToStringSlice(v)
	default:
		return []string{}
	}
}

// extractValueFromWrapper safely extracts the Value field from various wrapper types
func extractValueFromWrapper(claim interface{}) interface{} {
	if wrappedClaim, ok := claim.(*struct{ Value interface{} }); ok {
		return wrappedClaim.Value
	}

	if wrappedClaim, ok := claim.(struct{ Value interface{} }); ok {
		return wrappedClaim.Value
	}

	if m, ok := claim.(map[string]interface{}); ok {
		if val, exists := m["Value"]; exists {
			return val
		}
		if val, exists := m["value"]; exists {
			return val
		}
	}

	return nil
}

// convertInterfaceSliceToStringSlice converts []interface{} to []string
func convertInterfaceSliceToStringSlice(v []interface{}) []string {
	var result []string
	for _, item := range v {
		if str, ok := item.(string); ok {
			result = append(result, str)
		}
	}
	return result
}

// filterOrganizationRolesFromStrings filters roles to only include organization-specific roles
func filterOrganizationRolesFromStrings(roles []string, organizationID string) []string {
	prefix := buildOrgRolePrefix(organizationID)
	var orgSpecificRoles []string

	for _, role := range roles {
		if strings.HasPrefix(role, prefix) {
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

// filterOrganizationRoles filters user roles to only include organization-specific roles
func filterOrganizationRoles(userRolesClaim interface{}, organizationID string) []string {
	roles := extractRolesAsStrings(userRolesClaim)
	return filterOrganizationRolesFromStrings(roles, organizationID)
}

// checkOrganizationPermission checks if the user has the required permission for the organization
func checkOrganizationPermission(permissionsClaim interface{}, requiredPermission string) bool {
	permissions := extractPermissionsAsStrings(permissionsClaim)
	return hasPermission(permissions, requiredPermission)
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
