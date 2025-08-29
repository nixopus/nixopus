package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/audit/service"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	org_types "github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
	"github.com/uptrace/bun"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
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

// GetRecentAuditLogs returns the last 10 audit logs for an organization
func (c *AuditController) GetRecentAuditLogs(f fuego.ContextNoBody) (*shared_types.Response, error) {
	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    ErrUnauthorized,
			Status: http.StatusUnauthorized,
		}
	}

	page := f.QueryParam("page")
	pageSize := f.QueryParam("pageSize")

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

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = 1
	}

	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		pageSizeInt = 4
	}

	logs, _, err := c.service.GetAuditLogsByOrganization(orgID, pageInt, pageSizeInt)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to get audit logs", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Audit logs retrieved successfully",
		Data:    logs,
	}, nil
}
