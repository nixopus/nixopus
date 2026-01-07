package middleware

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api_key_service "github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	api_key_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	user_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

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

// VerifyAPIKey verifies an API key and returns the user if valid.
// The API key format is: nixopus_<prefix>_<rest>
// It can be provided as:
//   - "nixopus_<prefix>_<rest>" (direct)
//   - "Bearer nixopus_<prefix>_<rest>" (with Bearer prefix)
func VerifyAPIKey(apiKeyString string, store *shared_storage.Store, ctx context.Context, l logger.Logger) (*shared_types.User, error) {
	if apiKeyString == "" {
		return nil, ErrAuthTokenNotProvided
	}

	// Remove "Bearer " prefix if present
	apiKeyString = strings.TrimPrefix(apiKeyString, "Bearer ")
	apiKeyString = strings.TrimSpace(apiKeyString)

	// Initialize API key service
	apiKeyStorage := api_key_storage.APIKeyStorage{
		DB:  store.DB,
		Ctx: ctx,
	}
	apiKeyService := api_key_service.NewAPIKeyService(apiKeyStorage, l)

	// Verify the API key
	apiKey, err := apiKeyService.VerifyAPIKey(apiKeyString)
	if err != nil {
		l.Log(logger.Error, fmt.Sprintf("failed to verify API key: %v", err), "")
		return nil, ErrInvalidAuthToken
	}

	// Get user from API key's user ID
	userStorage := user_storage.UserStorage{
		DB:  store.DB,
		Ctx: ctx,
	}

	user, err := userStorage.FindUserByID(apiKey.UserID.String())
	if err != nil {
		l.Log(logger.Error, fmt.Sprintf("failed to find user by ID: %v", err), "")
		return nil, fmt.Errorf("user not found: %w", err)
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
// API keys should be passed in the request metadata with key "auth_token" or in the context.
// The API key format is: nixopus_<prefix>_<rest>
// It can be provided as:
//   - "nixopus_<prefix>_<rest>" (direct)
//   - "Bearer nixopus_<prefix>_<rest>" (with Bearer prefix)
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
		// Extract API key from request metadata first
		var apiKey string
		if req.Params != nil && req.Params.Meta != nil {
			if tokenAny, ok := req.Params.Meta[AuthTokenMetaKey]; ok {
				if tokenStr, ok := tokenAny.(string); ok {
					apiKey = tokenStr
				}
			}
		}

		// If not found in metadata, try to get from context (set by HTTP middleware)
		if apiKey == "" {
			if tokenAny := ctx.Value(AuthTokenMetaKey); tokenAny != nil {
				if tokenStr, ok := tokenAny.(string); ok {
					apiKey = tokenStr
				}
			}
		}

		// If still not found, try environment variable (for stdio MCP clients like Cursor)
		if apiKey == "" {
			apiKey = os.Getenv("AUTH_TOKEN")
		}

		// Verify API key and get user
		user, err := VerifyAPIKey(apiKey, store, ctx, l)
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
		// Only perform organization authorization if organization ID is provided
		if orgID != "" {
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
		}

		return handler(ctx, req, input)
	}
}
