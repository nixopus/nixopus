package middleware

import "net/http"

// SecurityHeadersMiddleware adds defensive HTTP headers to every response.
// HSTS is intentionally omitted — Caddy handles TLS termination and sets
// Strict-Transport-Security closer to the edge where it belongs.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Permitted-Cross-Domain-Policies", "none")
		next.ServeHTTP(w, r)
	})
}
