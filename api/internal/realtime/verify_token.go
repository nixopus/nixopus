package realtime

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	betterauth "github.com/raghavyuva/nixopus-api/internal/features/betterauth"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

// verifyToken verifies the Better Auth session token and returns the user and organization ID if the token is valid.
// This function uses Better Auth's VerifySession directly and creates a User object from the session response,
// avoiding database queries for organization_users since Better Auth is the source of truth for organization membership.
//
// Parameters:
//
//	tokenString - the Better Auth session token string to verify (from query param, fallback).
//	originalRequest - the original HTTP request containing cookies (preferred method).
//
// Returns:
//   - the user if the token is valid.
//   - the organization ID if available.
//   - an error if the token is invalid.
func (s *SocketServer) verifyToken(tokenString string, originalRequest *http.Request) (*types.User, string, error) {
	var req *http.Request

	// Prefer using the original request with actual cookies from the browser
	// WebSocket upgrade requests include cookies, which Better Auth needs
	if originalRequest != nil {
		// Clone the request to avoid modifying the original
		req = originalRequest.Clone(originalRequest.Context())
		// Also add the token as Authorization header as fallback
		if tokenString != "" {
			req.Header.Set("Authorization", "Bearer "+tokenString)
		}
	} else {
		// Fallback: create a mock request with the token
		req, _ = http.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		// Also set as cookie if it's a cookie-based session
		req.AddCookie(&http.Cookie{
			Name:  "better-auth.session_token",
			Value: tokenString,
		})
	}

	// Verify Better Auth session - this is the source of truth for authentication and organization membership
	sessionResp, err := betterauth.VerifySession(req)
	if err != nil {
		return nil, "", fmt.Errorf("session verification failed: %v", err)
	}

	// Parse Better Auth user ID
	betterAuthUserID, err := uuid.Parse(sessionResp.User.ID)
	if err != nil {
		return nil, "", fmt.Errorf("invalid user ID format: %v", err)
	}

	// Extract organization ID from session
	var orgID string
	if sessionResp.Session.ActiveOrganizationID != nil && *sessionResp.Session.ActiveOrganizationID != "" {
		orgID = *sessionResp.Session.ActiveOrganizationID
	} else {
		// Fallback to header if not in session
		orgID = originalRequest.Header.Get("X-Organization-Id")
	}

	// Create User object directly from Better Auth session response
	// We don't need to query the database for organization_users since Better Auth provides organization info
	user := &types.User{
		ID:            betterAuthUserID,
		Name:          sessionResp.User.Name,
		Email:         sessionResp.User.Email,
		EmailVerified: sessionResp.User.EmailVerified,
		CreatedAt:     time.Now(), // Better Auth doesn't provide this, use current time
		UpdatedAt:     time.Now(), // Better Auth doesn't provide this, use current time
	}

	// Compute backward compatibility fields
	user.ComputeCompatibilityFields()

	return user, orgID, nil
}
