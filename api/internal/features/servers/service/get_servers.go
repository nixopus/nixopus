package service

import (
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/types"
)

// GetServers retrieves servers for a specific user within their organization with pagination, search, and sorting
func (s *ServersService) GetServers(userID string, organizationID string, queryParams *types.ServerQueryParams) (*types.ServerListResponse, error) {
	s.logger.Log(logger.Info, "get servers request received", fmt.Sprintf("user_id=%s, organization_id=%s, page=%d, page_size=%d, search=%s, sort_by=%s, sort_order=%s",
		userID, organizationID, queryParams.Page, queryParams.PageSize, queryParams.Search, queryParams.SortBy, queryParams.SortOrder))

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Log(logger.Error, "invalid user id", fmt.Sprintf("user_id=%s", userID))
		return nil, types.ErrInvalidUserID
	}

	if userID == "" {
		s.logger.Log(logger.Error, "invalid user id", fmt.Sprintf("user_id=%s", userID))
		return nil, types.ErrInvalidUserID
	}

	totalCount, err := s.storage.GetServersCount(organizationID, userUUID, queryParams.Search)
	if err != nil {
		s.logger.Log(logger.Error, "error while retrieving servers count", err.Error())
		return nil, err
	}

	servers, err := s.storage.GetServersPaginated(organizationID, userUUID, queryParams)
	if err != nil {
		s.logger.Log(logger.Error, "error while retrieving servers", err.Error())
		return nil, err
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(queryParams.PageSize)))

	pagination := types.Pagination{
		CurrentPage: queryParams.Page,
		PageSize:    queryParams.PageSize,
		TotalPages:  totalPages,
		TotalItems:  totalCount,
		HasNext:     queryParams.Page < totalPages,
		HasPrev:     queryParams.Page > 1,
	}

	return &types.ServerListResponse{
		Servers:    servers,
		Pagination: pagination,
	}, nil
}
