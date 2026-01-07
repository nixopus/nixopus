package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	user_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/supertokens/supertokens-golang/recipe/session"
)

// responseRecorder is a minimal http.ResponseWriter for token verification
type responseRecorder struct {
	statusCode int
}

func (r *responseRecorder) Header() http.Header        { return make(http.Header) }
func (r *responseRecorder) Write([]byte) (int, error)  { return 0, nil }
func (r *responseRecorder) WriteHeader(statusCode int) { r.statusCode = statusCode }

// OrganizationIDExtractor is an interface for types that have an organization ID field
type OrganizationIDExtractor interface {
	GetOrganizationID() string
}

var (
	ErrOrganizationNotProvided = errors.New("organization ID not provided")
	ErrAuthTokenNotProvided    = errors.New("authentication token not provided")
	ErrInvalidAuthToken        = errors.New("invalid authentication token")
)

const (
	// AuthTokenMetaKey is the key used in MCP request metadata to pass authentication tokens
	AuthTokenMetaKey = "auth_token"
)

// VerifyToken verifies a SuperTokens session token and returns the user if valid.
// This function creates a mock HTTP request to leverage SuperTokens session verification.
func VerifyToken(tokenString string, store *shared_storage.Store, ctx context.Context, l logger.Logger) (*shared_types.User, error) {
	if tokenString == "" {
		return nil, ErrAuthTokenNotProvided
	}

	// Create a mock HTTP request with the token to verify the session
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	var user *shared_types.User
	var err error

	// Use a response recorder to capture any errors
	var responseWriter http.ResponseWriter = &responseRecorder{}

	session.VerifySession(nil, func(w http.ResponseWriter, r *http.Request) {
		sessionContainer := session.GetSessionFromRequestContext(r.Context())
		if sessionContainer == nil {
			err = ErrInvalidAuthToken
			return
		}

		userID := sessionContainer.GetUserID()
		if userID == "" {
			err = ErrInvalidAuthToken
			return
		}

		userStorage := user_storage.UserStorage{
			DB:  store.DB,
			Ctx: ctx,
		}

		user, err = userStorage.FindUserBySupertokensID(userID)
		if err != nil {
			l.Log(logger.Error, fmt.Sprintf("failed to find user by SuperTokens ID: %v", err), "")
			err = fmt.Errorf("user not found: %w", err)
		}
	}).ServeHTTP(responseWriter, req)

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrInvalidAuthToken
	}

	return user, nil
}

// AuthenticateUser extracts and validates the user from the context.
// The user should be set in the context by MCP middleware or caller using UserContextKey.
func AuthenticateUser(ctx context.Context, store *shared_storage.Store, l logger.Logger) (*shared_types.User, error) {
	// Try to get user from context (set by MCP middleware or caller)
	userAny := ctx.Value(shared_types.UserContextKey)
	if userAny == nil {
		l.Log(logger.Error, "user not found in context", "")
		return nil, shared_types.ErrFailedToGetUserFromContext
	}

	user, ok := userAny.(*shared_types.User)
	if !ok {
		l.Log(logger.Error, "invalid user type in context", "")
		return nil, shared_types.ErrFailedToGetUserFromContext
	}

	return user, nil
}

// AuthorizeOrganizationAccess verifies that the user belongs to the specified organization.
// Returns an error if the user does not belong to the organization or if verification fails.
func AuthorizeOrganizationAccess(store *shared_storage.Store, ctx context.Context, userID, organizationID string, l logger.Logger) error {
	userStorage := user_storage.UserStorage{
		DB:  store.DB,
		Ctx: ctx,
	}

	belongsToOrg, err := userStorage.UserBelongsToOrganization(userID, organizationID)
	if err != nil {
		l.Log(logger.Error, fmt.Sprintf("failed to verify organization membership: %v", err), "")
		return fmt.Errorf("failed to verify organization membership: %w", err)
	}

	if !belongsToOrg {
		l.Log(logger.Warning, fmt.Sprintf("user %s does not belong to organization %s", userID, organizationID), "")
		return shared_types.ErrUserDoesNotBelongToOrganization
	}

	return nil
}

// WithAuth wraps an MCP tool handler with authentication and authorization middleware.
// The input type must implement OrganizationIDExtractor interface.
// This wrapper ensures all tool calls are authenticated and authorized before execution.
// Authentication tokens should be passed in the request metadata with key "auth_token".
func WithAuth[Input OrganizationIDExtractor, Output any](
	store *shared_storage.Store,
	l logger.Logger,
	handler func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, Output, error),
) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, Output, error) {
	return func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		input Input,
	) (*mcp.CallToolResult, Output, error) {
		// Extract auth token from request metadata
		var authToken string
		if req.Params != nil && req.Params.Meta != nil {
			if tokenAny, ok := req.Params.Meta[AuthTokenMetaKey]; ok {
				if tokenStr, ok := tokenAny.(string); ok {
					authToken = tokenStr
				}
			}
		}

		// Verify token and get user
		user, err := VerifyToken(authToken, store, ctx, l)
		if err != nil {
			l.Log(logger.Error, fmt.Sprintf("authentication failed: %v", err), "")
			var zero Output
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Authentication failed: %v", err)},
				},
			}, zero, nil
		}

		// Set user in context for downstream handlers
		ctx = context.WithValue(ctx, shared_types.UserContextKey, user)

		orgID := input.GetOrganizationID()
		if orgID == "" {
			l.Log(logger.Error, "organization ID not provided", "")
			var zero Output
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Organization ID not provided"},
				},
			}, zero, nil
		}

		if err := AuthorizeOrganizationAccess(store, ctx, user.ID.String(), orgID, l); err != nil {
			l.Log(logger.Error, fmt.Sprintf("authorization failed: %v", err), "")
			var zero Output
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Authorization failed: %v", err)},
				},
			}, zero, nil
		}

		return handler(ctx, req, input)
	}
}
