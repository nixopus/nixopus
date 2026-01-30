package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/server/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/server/types"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type ServerService struct {
	store   *shared_storage.Store
	ctx     context.Context
	logger  logger.Logger
	storage storage.ServerRepository
}

func NewServerService(store *shared_storage.Store, ctx context.Context, l logger.Logger) *ServerService {
	serverStorage := &storage.ServerStorage{DB: store.DB, Ctx: ctx}
	return &ServerService{
		store:   store,
		ctx:     ctx,
		logger:  l,
		storage: serverStorage,
	}
}

// ListServers retrieves a paginated, filtered, and sorted list of servers (SSH keys)
// for an organization with optional user_provision_details.
func (s *ServerService) ListServers(orgID uuid.UUID, params types.ServerListParams) (*types.ListServersResponse, error) {
	// Set defaults for pagination
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	// Set default sorting
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	// Call storage layer
	servers, totalCount, err := s.storage.ListServersByOrganizationID(orgID, params)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), orgID.String())
		return nil, err
	}

	// Return empty slice if nil
	if servers == nil {
		servers = []types.ServerResponse{}
	}

	return &types.ListServersResponse{
		Status:  "success",
		Message: "Servers fetched successfully",
		Data: types.ListServersResponseData{
			Servers:    servers,
			TotalCount: totalCount,
			Page:       params.Page,
			PageSize:   params.PageSize,
			SortBy:     params.SortBy,
			SortOrder:  params.SortOrder,
			Search:     params.Search,
			Status:     params.Status,
			IsActive:   params.IsActive,
		},
	}, nil
}
