package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/nixopus/nixopus/api/internal/types"
)

const RequestIDHeader = "X-Request-ID"

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = generateRequestID()
		}
		w.Header().Set(RequestIDHeader, requestID)
		ctx := context.WithValue(r.Context(), types.RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
