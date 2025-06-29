package middleware

import (
	"fmt"
	"net/http"
	"os"
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
		allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
		fmt.Println("allowedOrigin", allowedOrigin)

		allowedOrigins := []string{
			allowedOrigin,
			"http://localhost:3000",
			"http://localhost:7443",
		}

		originAllowed := false
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				originAllowed = true
				break
			}
		}

		if originAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if origin != "" {
			if len(allowedOrigins) > 0 {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigins[0])
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, Sec-WebSocket-Extensions, Sec-WebSocket-Key, Sec-WebSocket-Version, X-Organization-Id")
		w.Header().Set("Access-Control-Expose-Headers", "Authorization, X-Organization-Id")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "300")

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
