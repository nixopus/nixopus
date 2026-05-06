package middleware

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/cache"
	betterauth "github.com/nixopus/nixopus/api/internal/features/auth"
	applogger "github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/storage"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

// defaultVerifySessionFn is the production Better Auth session verifier.
func defaultVerifySessionFn(r *http.Request) (*betterauth.SessionResponse, error) {
	return betterauth.VerifySession(r)
}

// verifySessionFn calls Better Auth session verification (swapped in tests).
var verifySessionFn = defaultVerifySessionFn

// jsonMarshalSessionResponse marshals session payloads for Redis (swapped in tests).
var jsonMarshalSessionResponse = json.Marshal

// authNewRequestForAPIKey builds the fallback API-key session probe (swapped in tests).
var authNewRequestForAPIKey = http.NewRequest

// sessionCacheKey computes a SHA-256 hash of the auth-relevant headers to use as a
// Redis cache key. Requests with identical auth credentials get the same key.
func sessionCacheKey(r *http.Request) string {
	h := sha256.New()
	h.Write([]byte(r.Header.Get("Cookie")))
	h.Write([]byte{0})
	h.Write([]byte(r.Header.Get("Authorization")))
	h.Write([]byte{0})
	h.Write([]byte(r.Header.Get("x-api-key")))
	return hex.EncodeToString(h.Sum(nil))
}

// verifySessionCached attempts to resolve a session from Redis cache first,
// falling back to the auth service on cache miss. Caches successful results.
func verifySessionCached(r *http.Request, c *cache.Cache) (*betterauth.SessionResponse, error) {
	if c == nil {
		return verifySessionWithFallback(r)
	}

	cacheKey := sessionCacheKey(r)
	ctx := r.Context()

	if data, err := c.GetSession(ctx, cacheKey); err == nil && data != nil {
		var cached betterauth.SessionResponse
		if err := json.Unmarshal(data, &cached); err == nil && cached.User.ID != "" {
			return &cached, nil
		}
	}

	resp, err := verifySessionWithFallback(r)
	if err != nil {
		return nil, err
	}

	// Only cache sessions that have an activeOrganizationId. New users have a race
	// condition where the session is created before their default org is set up in
	// user.create.after, so activeOrganizationId is null in the initial session.
	// Caching that null would lock new users out for the full SessionCacheTTL.
	if resp.Session.ActiveOrganizationID != nil && *resp.Session.ActiveOrganizationID != "" {
		if data, err := jsonMarshalSessionResponse(resp); err == nil {
			_ = c.SetSession(ctx, cacheKey, data)
		}
	}

	return resp, nil
}

// verifySessionWithFallback tries cookie/bearer auth first, then falls back to x-api-key.
func verifySessionWithFallback(r *http.Request) (*betterauth.SessionResponse, error) {
	sessionResp, err := verifySessionFn(r)
	if err != nil {
		apiKeyHeader := r.Header.Get("x-api-key")
		if apiKeyHeader == "" {
			return nil, err
		}
		apiKeyReq, reqErr := authNewRequestForAPIKey("GET", "", nil)
		if reqErr != nil {
			return nil, fmt.Errorf("failed to create API key request: %w", reqErr)
		}
		apiKeyReq.Header.Set("x-api-key", apiKeyHeader)
		if origin := r.Header.Get("Origin"); origin != "" {
			apiKeyReq.Header.Set("Origin", origin)
		}
		if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
			apiKeyReq.Header.Set("X-Forwarded-Proto", proto)
		}
		sessionResp, err = verifySessionFn(apiKeyReq)
		if err != nil {
			return nil, err
		}
	}
	return sessionResp, nil
}

// AuthMiddleware is a middleware that checks if the request has a valid
// Better Auth session. If the session is valid, it adds both the user and
// the authenticated client to the request context.
func AuthMiddleware(next http.Handler, app *storage.App, c *cache.Cache, l applogger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if handled := tryM2MJWTAuth(ctx, w, r, next, app, l); handled {
			return
		}

		disableCache := r.Header.Get("X-Disable-Cache")
		sessionCache := c
		if disableCache == "true" {
			sessionCache = nil
		}

		sessionResp, err := verifySessionCached(r, sessionCache)
		if err != nil {
			l.LogCtx(r.Context(), applogger.Error, fmt.Sprintf("middleware auth: verify session failed: %v", err), fmt.Sprintf("path=%s", r.URL.Path))
			utils.SendErrorResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		betterAuthUserID := sessionResp.User.ID

		userIDUUID, err := uuid.Parse(betterAuthUserID)
		if err != nil {
			l.LogCtx(r.Context(), applogger.Error, "middleware auth: invalid Better Auth user id format", fmt.Sprintf("user_id=%s path=%s", betterAuthUserID, r.URL.Path))
			utils.SendErrorResponse(w, "Invalid user ID", http.StatusUnauthorized)
			return
		}

		var user *types.User
		if sessionCache != nil {
			if cached, _ := sessionCache.GetUserByID(ctx, betterAuthUserID); cached != nil {
				user = cached
			}
		}

		if user == nil {
			var dbUser types.User
			err = app.Store.DB.NewSelect().
				Model(&dbUser).
				Where("id = ?", userIDUUID).
				Scan(ctx)

			if err != nil {
				if err == sql.ErrNoRows {
					l.LogCtx(r.Context(), applogger.Error, "middleware auth: user not found in database", fmt.Sprintf("user_id=%s path=%s", betterAuthUserID, r.URL.Path))
					utils.SendErrorResponse(w, "User not found", http.StatusUnauthorized)
					return
				}
				l.LogCtx(r.Context(), applogger.Error, fmt.Sprintf("middleware auth: user lookup failed: %v", err), fmt.Sprintf("user_id=%s path=%s", betterAuthUserID, r.URL.Path))
				utils.SendErrorResponse(w, "Failed to fetch user", http.StatusInternalServerError)
				return
			}

			dbUser.ComputeCompatibilityFields()
			user = &dbUser

			if sessionCache != nil {
				_ = sessionCache.SetUserByID(ctx, betterAuthUserID, user)
			}
		}

		ctx = context.WithValue(ctx, types.UserContextKey, user)

		if !isAuthEndpoint(r.URL.Path) {
			organizationID, err := resolveAndVerifyOrganization(ctx, r, c, app, betterAuthUserID, sessionResp, l)
			if err != nil {
				l.LogCtx(r.Context(), applogger.Error, fmt.Sprintf("middleware auth: resolve organization failed: %v", err), fmt.Sprintf("path=%s user_id=%s", r.URL.Path, betterAuthUserID))
				statusCode := http.StatusBadRequest
				if strings.Contains(err.Error(), "does not belong") {
					statusCode = http.StatusForbidden
				}
				utils.SendErrorResponse(w, err.Error(), statusCode)
				return
			}

			ctx = context.WithValue(ctx, types.OrganizationIDKey, organizationID)
		}

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func tryM2MJWTAuth(ctx context.Context, w http.ResponseWriter, r *http.Request, next http.Handler, app *storage.App, l applogger.Logger) bool {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return false
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if !isJWT(token) {
		return false
	}

	orgID, err := validateM2MJWT(ctx, token, r.Header.Get("X-Organization-Id"))
	if err != nil {
		l.LogCtx(r.Context(), applogger.Info, "middleware auth: M2M JWT validation failed, using session auth", fmt.Sprintf("path=%s error=%v", r.URL.Path, err))
		return false
	}

	userIDStr := r.Header.Get("X-User-Id")
	if userIDStr == "" {
		l.LogCtx(r.Context(), applogger.Error, "middleware auth: M2M request missing X-User-Id", fmt.Sprintf("path=%s", r.URL.Path))
		utils.SendErrorResponse(w, "M2M requests require X-User-Id header", http.StatusUnauthorized)
		return true
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		l.LogCtx(r.Context(), applogger.Error, "middleware auth: invalid X-User-Id", fmt.Sprintf("user_id=%q path=%s", userIDStr, r.URL.Path))
		utils.SendErrorResponse(w, "Invalid X-User-Id format", http.StatusUnauthorized)
		return true
	}

	var user types.User
	err = app.Store.DB.NewSelect().Model(&user).Where("id = ?", userUUID).Scan(ctx)
	if err != nil {
		l.LogCtx(r.Context(), applogger.Error, fmt.Sprintf("middleware auth: M2M user lookup failed: %v", err), fmt.Sprintf("user_id=%s path=%s", userIDStr, r.URL.Path))
		utils.SendErrorResponse(w, "User not found", http.StatusUnauthorized)
		return true
	}
	user.ComputeCompatibilityFields()

	ctx = context.WithValue(ctx, types.UserContextKey, &user)
	ctx = context.WithValue(ctx, types.OrganizationIDKey, orgID)
	r = r.WithContext(ctx)
	next.ServeHTTP(w, r)
	return true
}

func isAuthEndpoint(path string) bool {
	authPaths := []string{
		"/api/v1/auth/bootstrap",
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
		// Better Auth endpoints
		"/api/auth",
	}

	for _, authPath := range authPaths {
		if path == authPath || strings.HasPrefix(path, authPath+"/") {
			return true
		}
	}
	return false
}

// extractOrgIDFromSession extracts organization ID from session response or X-Organization-Id header.
func extractOrgIDFromSession(sessionResp *betterauth.SessionResponse, r *http.Request) string {
	if sessionResp != nil && sessionResp.Session.ActiveOrganizationID != nil && *sessionResp.Session.ActiveOrganizationID != "" {
		return *sessionResp.Session.ActiveOrganizationID
	}
	return r.Header.Get("X-Organization-Id")
}

// resolveAndVerifyOrganization resolves the organization ID from the already-verified session
// and verifies user membership via Better Auth API (with caching).
func resolveAndVerifyOrganization(
	ctx context.Context,
	r *http.Request,
	cache *cache.Cache,
	app *storage.App,
	betterAuthUserID string,
	sessionResp *betterauth.SessionResponse,
	l applogger.Logger,
) (string, error) {
	// Extract organization ID from session (already verified, no duplicate API call)
	organizationID := extractOrgIDFromSession(sessionResp, r)

	// For new users there is a race: the session is created (session.create.before) before
	// the default org is set up (user.create.after → setupNewUser). The session gets
	// activeOrganizationId=null baked into the compact cookie, and patchSessionOrganization
	// only updates the DB row — not the already-issued cookie. Fall back to a direct DB
	// membership lookup so new users can proceed immediately after sign-up.
	if organizationID == "" {
		userUUID, parseErr := uuid.Parse(betterAuthUserID)
		if parseErr == nil {
			var member types.Member
			dbErr := app.Store.DB.NewSelect().
				Model(&member).
				Where("user_id = ?", userUUID).
				OrderExpr("created_at ASC").
				Limit(1).
				Scan(ctx)
			if dbErr == nil && member.OrganizationID != uuid.Nil {
				l.LogCtx(ctx, applogger.Info, "middleware auth: resolved org from DB (session missing activeOrganizationId)", fmt.Sprintf("user_id=%s organization_id=%s", betterAuthUserID, member.OrganizationID))
				return member.OrganizationID.String(), nil
			}
		}
		return "", fmt.Errorf("no organization ID found in Better Auth session for user %s", betterAuthUserID)
	}

	// Check organization membership via Better Auth API (with caching)
	_, err := verifyOrganizationMembership(ctx, r, cache, betterAuthUserID, organizationID, l)
	if err != nil {
		return "", fmt.Errorf("user %s does not belong to organization %s: %w", betterAuthUserID, organizationID, err)
	}

	return organizationID, nil
}

// verifyOrganizationMembership verifies if a user belongs to an organization.
// Uses cache to avoid repeated API calls. When fetching from API, also populates
// RBAC cache so RBAC middleware avoids a duplicate Better Auth API call.
func verifyOrganizationMembership(
	ctx context.Context,
	r *http.Request,
	cache *cache.Cache,
	betterAuthUserID string,
	organizationID string,
	l applogger.Logger,
) (bool, error) {
	// Check cache first
	if cache != nil {
		if cached, err := cache.GetOrgMembership(ctx, betterAuthUserID, organizationID); err == nil && cached {
			return true, nil
		}
	}

	// Verify membership via Better Auth API
	member, err := getBetterAuthOrganizationMember(ctx, r, betterAuthUserID, organizationID, l)
	if err != nil || member == nil {
		// Cache negative result to avoid repeated API calls
		if cache != nil {
			cache.SetOrgMembership(ctx, betterAuthUserID, organizationID, false)
		}
		return false, err
	}

	// Cache positive result
	if cache != nil {
		cache.SetOrgMembership(ctx, betterAuthUserID, organizationID, true)
	}

	// Populate RBAC cache so RBAC middleware avoids duplicate Better Auth API call
	cacheRBACPermissionsFromMember(betterAuthUserID, organizationID, member)

	return true, nil
}
