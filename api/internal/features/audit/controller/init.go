package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/audit/service"
	"github.com/raghavyuva/nixopus-api/internal/features/audit/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	org_types "github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
	"github.com/uptrace/bun"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
)

type AuditController struct {
	service *service.AuditService
	ctx     context.Context
	logger  logger.Logger
}

func NewAuditController(db *bun.DB, ctx context.Context, l logger.Logger) *AuditController {
	return &AuditController{
		service: service.NewAuditService(db, ctx, l),
		ctx:     ctx,
		logger:  l,
	}
}

// GetRecentAuditLogs returns human-readable activities with pagination, search, and filtering
func (c *AuditController) GetRecentAuditLogs(f fuego.ContextNoBody) (*types.GetActivitiesResponse, error) {
	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    ErrUnauthorized,
			Status: http.StatusUnauthorized,
		}
	}

	// Get query parameters
	page := f.QueryParam("page")
	pageSize := f.QueryParam("pageSize")
	search := f.QueryParam("search")
	resourceType := f.QueryParam("resource_type")

	orgIDStr := f.Request().Header.Get("X-ORGANIZATION-ID")
	if orgIDStr == "" {
		return nil, fuego.HTTPError{
			Err:    org_types.ErrMissingOrganizationID,
			Status: http.StatusBadRequest,
		}
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    org_types.ErrInvalidOrganizationID,
			Status: http.StatusBadRequest,
		}
	}

	// Parse pagination parameters
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}

	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeInt < 1 {
		pageSizeInt = 10
	}

	if pageSizeInt > 100 {
		pageSizeInt = 100
	}

	activities, totalCount, err := c.service.GetActivitiesByOrganization(orgID, pageInt, pageSizeInt, search, resourceType)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to get activities", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	totalPages := (totalCount + pageSizeInt - 1) / pageSizeInt
	hasNext := pageInt < totalPages
	hasPrev := pageInt > 1

	return &types.GetActivitiesResponse{
		Status:  "success",
		Message: "Activities retrieved successfully",
		Data: types.GetActivitiesResponseData{
			Activities: activities,
			Pagination: types.PaginationInfo{
				CurrentPage: pageInt,
				PageSize:    pageSizeInt,
				TotalCount:  totalCount,
				TotalPages:  totalPages,
				HasNext:     hasNext,
				HasPrev:     hasPrev,
			},
		},
	}, nil
}
