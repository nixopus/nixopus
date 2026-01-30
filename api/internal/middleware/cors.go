package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/config"
)

// CorsMiddleware sets the necessary CORS headers for the response. If the request
// method is OPTIONS, it will respond with a 200 status code and return. Otherwise,
// it will call the next handler in the chain.
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/webhook/" {
			next.ServeHTTP(w, r)
			return
		}

		origin := r.Header.Get("Origin")
		allowedOriginConfig := config.AppConfig.CORS.AllowedOrigin
		fmt.Println("allowedOrigin", allowedOriginConfig)

		// Parse comma-separated origins from config
		var allowedOrigins []string
		if allowedOriginConfig != "" {
			// Split by comma and trim whitespace
			configOrigins := strings.Split(allowedOriginConfig, ",")
			for _, o := range configOrigins {
				trimmed := strings.TrimSpace(o)
				if trimmed != "" {
					allowedOrigins = append(allowedOrigins, trimmed)
				}
			}
		}

		// Add localhost origins for development
		allowedOrigins = append(allowedOrigins,
			"http://localhost:3000",
			"http://localhost:7443",
		)

		// Check if the request origin matches any allowed origin
		originAllowed := false
		var matchedOrigin string
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				originAllowed = true
				matchedOrigin = origin
				break
			}
		}

		// Set Access-Control-Allow-Origin header with only ONE value
		// Browsers reject multiple values in this header (must be single origin or *)
		// Use Set() which replaces any existing value to prevent duplicates
		if originAllowed {
			w.Header().Set("Access-Control-Allow-Origin", matchedOrigin)
		}
		// If origin doesn't match any allowed origin, don't set the header
		// This will cause CORS to fail, which is the correct security behavior

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Expose-Headers", "Authorization, X-Organization-Id")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "300")
		headers := []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Organization-Id", "X-Disable-Cache"}
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Header.Get("Upgrade") == "websocket" {
			next.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
