package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	user_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
	"github.com/uptrace/bun"
)

// AuthMiddleware is a middleware that checks if the request has a valid
// authorization token. If the token is valid, it adds both the user and
// the authenticated client to the request context.
func AuthMiddleware(next http.Handler, app *storage.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			utils.SendErrorResponse(w, "No authorization token provided", http.StatusUnauthorized)
			return
		}

		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		user, err := verifyToken(token, app.Store.DB, app.Ctx)
		if err != nil {
			log.Printf("Auth error: %v", err)
			utils.SendErrorResponse(w, "Invalid authorization token", http.StatusUnauthorized)
			return
		}

		// Check if 2FA is required but not verified
		claims, err := getTokenClaims(token)
		if err != nil {
			log.Printf("Token claims error: %v", err)
			utils.SendErrorResponse(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		twoFactorEnabled, _ := claims["2fa_enabled"].(bool)
		twoFactorVerified, _ := claims["2fa_verified"].(bool)

		// If 2FA is enabled but not verified, only allow access to 2FA verification endpoints
		if twoFactorEnabled && !twoFactorVerified {
			if !is2FAVerificationEndpoint(r.URL.Path) {
				utils.SendErrorResponse(w, "Two-factor authentication required", http.StatusForbidden)
				return
			}
		}

		log.Printf("User authenticated. ID: %s, Phone: %s", user.ID, user.Email)

		// Skip organization ID check for authentication routes
		if !isAuthEndpoint(r.URL.Path) {
			organizationID := r.Header.Get("X-Organization-Id")
			if organizationID == "" {
				utils.SendErrorResponse(w, "No organization ID provided", http.StatusBadRequest)
				return
			}

			userStorage := user_storage.UserStorage{
				DB:  app.Store.DB,
				Ctx: app.Ctx,
			}
			belongsToOrg, err := userStorage.UserBelongsToOrganization(user.ID.String(), organizationID)
			if err != nil {
				log.Printf("Error checking organization membership: %v", err)
				utils.SendErrorResponse(w, "Error verifying organization membership", http.StatusInternalServerError)
				return
			}

			if !belongsToOrg {
				utils.SendErrorResponse(w, "User does not belong to the specified organization", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), types.OrganizationIDKey, organizationID)
			r = r.WithContext(ctx)
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, types.UserContextKey, user)
		ctx = context.WithValue(ctx, types.AuthTokenKey, token)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func getTokenClaims(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return types.JWTSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

func is2FAVerificationEndpoint(path string) bool {
	// Add paths that should be accessible during 2FA verification
	allowedPaths := []string{
		"/api/v1/auth/2fa-login",
		"/api/v1/auth/verify-2fa",
	}

	for _, allowedPath := range allowedPaths {
		if path == allowedPath {
			return true
		}
	}
	return false
}

func isAuthEndpoint(path string) bool {
	authPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/2fa-login",
		"/api/v1/auth/verify-2fa",
		"/api/v1/auth/refresh-token",
		"/api/v1/auth/logout",
		"/api/v1/auth/setup-2fa",
		"/api/v1/auth/disable-2fa",
		"/api/v1/auth/verify-email",
		"/api/v1/auth/send-verification-email",
		"/api/v1/auth/reset-password",
		"/api/v1/auth/request-password-reset",
		"/api/v1/user",
		"/api/v1/user/",
		"/api/v1/user/organizations",
		"/api/v1/user/name",
	}

	for _, authPath := range authPaths {
		if path == authPath || strings.HasPrefix(path, authPath+"/") {
			return true
		}
	}
	return false
}

// verifyToken validates a JWT token and returns the associated user
func verifyToken(tokenString string, db *bun.DB, ctx context.Context) (*types.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return types.JWTSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return nil, fmt.Errorf("token expired")
			}
		}

		email, ok := claims["email"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid token claims")
		}

		user_storage := user_storage.UserStorage{
			DB:  db,
			Ctx: ctx,
		}
		user, err := user_storage.FindUserByEmail(email)
		if err != nil {
			return nil, err
		}

		return user, nil
	}

	return nil, fmt.Errorf("invalid token")
}
