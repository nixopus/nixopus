package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/types"
)

func ServerIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverID := r.URL.Query().Get("server_id")
		if serverID == "" {
			serverID = r.Header.Get("X-Server-Id")
		}
		if serverID != "" {
			if _, err := uuid.Parse(serverID); err == nil {
				ctx := context.WithValue(r.Context(), types.ServerIDKey, serverID)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}
