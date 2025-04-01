package middleware

import (
	"log"
	"net/http"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"status":"error","message":"Internal server error"}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
