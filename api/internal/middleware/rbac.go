package middleware

import (
	appStorage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
	"net/http"
)

// RBACMiddleware is a middleware that checks if the user has the required permissions
// to access a specific resource in the organization.
func RBACMiddleware(next http.Handler, app *appStorage.App, resource string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(types.UserContextKey).(*types.User)
		if !ok {
			utils.SendErrorResponse(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		orgID, ok := r.Context().Value(types.OrganizationIDKey).(string)
		if !ok {
			utils.SendErrorResponse(w, "Organization ID not found in context", http.StatusBadRequest)
			return
		}

		var userRole *types.Role
		for _, orgUser := range user.OrganizationUsers {
			if orgUser.OrganizationID.String() == orgID {
				userRole = orgUser.Role
				break
			}
		}

		if userRole == nil {
			utils.SendErrorResponse(w, "User does not have a role in this organization", http.StatusForbidden)
			return
		}
		requiredAction := getActionFromMethod(r.Method)

		hasPermission := false
		for _, rp := range userRole.Permissions {
			if rp.Resource == resource &&
				rp.Name == requiredAction {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			utils.SendErrorResponse(w, "User does not have permission to "+requiredAction+" "+resource, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
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
