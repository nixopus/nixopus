package middleware

import (
	"net/http"

	appStorage "github.com/raghavyuva/nixopus-api/internal/storage"

	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/recipe/session/claims"
	"github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
	"github.com/supertokens/supertokens-golang/recipe/userroles/userrolesclaims"
	"github.com/supertokens/supertokens-golang/supertokens"
)

// RBACMiddleware validates SuperTokens permission claims for the given resource based on HTTP method.
// It appends a PermissionClaim validator dynamically via VerifySession.
func RBACMiddleware(next http.Handler, _ *appStorage.App, resource string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requiredAction := getActionFromMethod(r.Method)
		requiredPermission := resource + ":" + requiredAction

		handler := session.VerifySession(&sessmodels.VerifySessionOptions{
			OverrideGlobalClaimValidators: func(globalClaimValidators []claims.SessionClaimValidator, sessionContainer sessmodels.SessionContainer, userContext supertokens.UserContext) ([]claims.SessionClaimValidator, error) {
				globalClaimValidators = append(globalClaimValidators, userrolesclaims.PermissionClaimValidators.Includes(requiredPermission, nil, nil))
				return globalClaimValidators, nil
			},
		}, func(w http.ResponseWriter, r *http.Request) { next.ServeHTTP(w, r) })

		handler.ServeHTTP(w, r)
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
