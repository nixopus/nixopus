package utils

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	server_storage "github.com/raghavyuva/nixopus-api/internal/features/servers/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

// ValidateServerAccess checks if a user has access to a specific server and returns the server details
func ValidateServerAccess(db *bun.DB, ctx context.Context, userID, serverID string) (*types.Server, error) {
	if serverID == "" {
		return nil, nil
	}

	_, err := uuid.Parse(serverID)
	if err != nil {
		return nil, err
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	serverStorage := server_storage.ServerStorage{
		DB:  db,
		Ctx: ctx,
	}

	server, err := serverStorage.GetServer(serverID)
	if err != nil {
		return nil, err
	}

	if server == nil {
		return nil, errors.New("server not found")
	}

	if server.UserID != userUUID {
		return nil, errors.New("user does not have access to this server")
	}

	return server, nil
}

// GetServerDetails extracts server details from request context
func GetServerDetailsWithErr(r *http.Request) (*types.Server, error) {
	serverAny := r.Context().Value(types.ServerIDKey)
	server, ok := serverAny.(*types.Server)

	if !ok {
		return nil, errors.New("server details not found")
	}

	return server, nil
}

// GetServer retrieves the current server from the request context (similar to GetUser pattern)
func GetServer(w http.ResponseWriter, r *http.Request) *types.Server {
	serverAny := r.Context().Value(types.ServerIDKey)
	server, ok := serverAny.(*types.Server)

	if !ok {
		return nil
	}

	return server
}

// GetServerID extracts server ID from request context
func GetServerID(r *http.Request) string {
	server, err := GetServerDetailsWithErr(r)
	if err != nil || server == nil {
		return ""
	}
	return server.ID.String()
}
