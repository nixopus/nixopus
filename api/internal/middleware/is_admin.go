package middleware

import (
	"fmt"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/types"
)

// IsAdmin is a middleware that ensures the user is an admin.
//
// This middleware retrieves the user from the request context and checks if the user
// has an admin type. If the user is not found in the context or is not an admin, an
// error response is sent to the client. If the user is an admin, the request proceeds
// to the next handler.
//
// Parameters:
//   - next: the next http.Handler to be called if the user is an admin.
//
// Returns:
//   - http.Handler: a handler that checks the user's admin status before proceeding.
func IsAdmin(next http.Handler) http.Handler {
	fmt.Println("IsAdmin middleware")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(types.UserContextKey).(*types.User)
		if !ok {
			http.Error(w, "User not found in context", http.StatusInternalServerError)
			return
		}

		if user.Type != "admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		fmt.Printf("Logged in as admin: %s \n", user.Email)

		next.ServeHTTP(w, r)
	})
}
