package middleware

import (
	"fmt"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/types"
)

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