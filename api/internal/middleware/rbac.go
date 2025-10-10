package middleware

import (
	"fmt"
	"net/http"
	"strings"

	appStorage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/recipe/session/claims"
	"github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
	"github.com/supertokens/supertokens-golang/recipe/userroles/userrolesclaims"
	"github.com/supertokens/supertokens-golang/supertokens"
)

// RBACMiddleware validates SuperTokens permission claims for the given resource based on HTTP method.
// It extracts organization ID from header and validates permissions only for organization-specific roles.
func RBACMiddleware(next http.Handler, _ *appStorage.App, resource string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requiredAction := getActionFromMethod(r.Method)
		requiredPermission := resource + ":" + requiredAction

		// Extract organization ID from header
		organizationID := r.Header.Get("X-Organization-Id")
		if organizationID == "" {
			utils.SendErrorResponse(w, "Organization ID is required", http.StatusBadRequest)
			return
		}

		handler := session.VerifySession(&sessmodels.VerifySessionOptions{
			OverrideGlobalClaimValidators: func(globalClaimValidators []claims.SessionClaimValidator, sessionContainer sessmodels.SessionContainer, userContext supertokens.UserContext) ([]claims.SessionClaimValidator, error) {
				//custom validator that checks organization specific permissions
				orgPermissionValidator := claims.SessionClaimValidator{
					ID: "org-permission-validator",
					Validate: func(payload map[string]interface{}, userContext supertokens.UserContext) claims.ClaimValidationResult {
						// Ensure claims are fetched and set in the session
						if err := sessionContainer.FetchAndSetClaim(userrolesclaims.UserRoleClaim); err != nil {
							return claims.ClaimValidationResult{
								IsValid: false,
								Reason:  fmt.Sprintf("failed to fetch user roles: %v", err),
							}
						}

						if err := sessionContainer.FetchAndSetClaim(userrolesclaims.PermissionClaim); err != nil {
							return claims.ClaimValidationResult{
								IsValid: false,
								Reason:  fmt.Sprintf("failed to fetch permissions: %v", err),
							}
						}

						// Get user roles from session
						userRolesResult, err := session.GetClaimValue(sessionContainer.GetHandle(), userrolesclaims.UserRoleClaim)
						if err != nil {
							return claims.ClaimValidationResult{
								IsValid: false,
								Reason:  fmt.Sprintf("failed to get user roles: %v", err),
							}
						}

						// Get permissions claim
						permissionsResult, err := session.GetClaimValue(sessionContainer.GetHandle(), userrolesclaims.PermissionClaim)
						if err != nil {
							return claims.ClaimValidationResult{
								IsValid: false,
								Reason:  fmt.Sprintf("failed to get permissions: %v", err),
							}
						}

						// Extract actual claim values from the result
						var userRolesClaim interface{}
						var permissionsClaim interface{}

						if userRolesResult.OK != nil {
							userRolesClaim = userRolesResult.OK
						} else {
							return claims.ClaimValidationResult{
								IsValid: false,
								Reason:  "user roles claim not found in session",
							}
						}

						if permissionsResult.OK != nil {
							permissionsClaim = permissionsResult.OK
						} else {
							return claims.ClaimValidationResult{
								IsValid: false,
								Reason:  "permissions claim not found in session",
							}
						}

						// Filter roles to only include organization-specific roles
						orgSpecificRoles := filterOrganizationRoles(userRolesClaim, organizationID)

						if len(orgSpecificRoles) == 0 {
							return claims.ClaimValidationResult{
								IsValid: false,
								Reason:  fmt.Sprintf("user has no roles for organization %s", organizationID),
							}
						}

						// Check if user has the required permission for this organization
						hasPermission := checkOrganizationPermission(permissionsClaim, requiredPermission, organizationID)
						if !hasPermission {
							return claims.ClaimValidationResult{
								IsValid: false,
								Reason:  fmt.Sprintf("user lacks permission %s for organization %s", requiredPermission, organizationID),
							}
						}

						return claims.ClaimValidationResult{
							IsValid: true,
						}
					},
				}

				globalClaimValidators = append(globalClaimValidators, orgPermissionValidator)
				return globalClaimValidators, nil
			},
		}, func(w http.ResponseWriter, r *http.Request) { next.ServeHTTP(w, r) })

		handler.ServeHTTP(w, r)
	})
}

// filterOrganizationRoles filters user roles to only include organization-specific roles
func filterOrganizationRoles(userRolesClaim interface{}, organizationID string) []string {
	var orgSpecificRoles []string

	if userRolesClaim == nil {
		return orgSpecificRoles
	}

	wrappedClaim := userRolesClaim.(*struct{ Value interface{} })
	actualRoles := wrappedClaim.Value

	switch roles := actualRoles.(type) {
	case []interface{}:
		for _, role := range roles {
			if roleStr, ok := role.(string); ok {
				if strings.HasPrefix(roleStr, "orgid_"+organizationID+"_") {
					orgSpecificRoles = append(orgSpecificRoles, roleStr)
				}
			}
		}
	case []string:
		for _, role := range roles {
			if strings.HasPrefix(role, "orgid_"+organizationID+"_") {
				orgSpecificRoles = append(orgSpecificRoles, role)
			}
		}
	}

	return orgSpecificRoles
}

// checkOrganizationPermission checks if the user has the required permission for the organization
func checkOrganizationPermission(permissionsClaim interface{}, requiredPermission, organizationID string) bool {
	if permissionsClaim == nil {
		return false
	}

	// Extract permissions from the wrapped claim structure
	wrappedClaim := permissionsClaim.(*struct{ Value interface{} })
	actualPermissions := wrappedClaim.Value

	// Handle different possible claim structures
	switch permissions := actualPermissions.(type) {
	case []interface{}:
		for _, perm := range permissions {
			if permStr, ok := perm.(string); ok {
				if permStr == requiredPermission {
					return true
				}
			}
		}
	case []string:
		for _, perm := range permissions {
			if perm == requiredPermission {
				return true
			}
		}
	}

	return false
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
