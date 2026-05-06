package middleware

import (
	"fmt"
	"net/http"

	"github.com/nixopus/nixopus/api/internal/features/logger"
)

// RecoveryMiddleware returns middleware that recovers panics and logs with the app logger.
func RecoveryMiddleware(l logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					data := ""
					if r != nil && r.URL != nil {
						data = fmt.Sprintf("path=%s", r.URL.Path)
					}
					l.LogCtx(r.Context(), logger.Error, fmt.Sprintf("middleware: panic recovered: %v", err), data)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"title":"Internal Server Error","status":500,"detail":"Internal server error"}`))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
