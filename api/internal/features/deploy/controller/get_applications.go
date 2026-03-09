package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *DeployController) GetApplications(f fuego.ContextNoBody) (*types.ListApplicationsResponse, error) {
	w, r := f.Response(), f.Request()
	page := r.URL.Query().Get("page")
	pageSize := r.URL.Query().Get("page_size")
	sortBy := r.URL.Query().Get("sort_by")
	sortDirection := r.URL.Query().Get("sort_direction")
	organizationID := utils.GetOrganizationID(r)
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "organization not found", "")
		return nil, fuego.UnauthorizedError{
			Detail: "organization not found",
		}
	}

	if page == "" {
		page = "1"
	}

	if pageSize == "" {
		pageSize = "10"
	}

	user := utils.GetUser(w, r)

	if user == nil {
		c.logger.Log(logger.Error, "user not found", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	applications, totalCount, err := c.service.GetApplications(page, pageSize, sortBy, sortDirection, organizationID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}
	return &types.ListApplicationsResponse{
		Status:  "success",
		Message: "Applications",
		Data: types.ListApplicationsResponseData{
			Applications: applications,
			TotalCount:   totalCount,
			Page:         page,
			PageSize:     pageSize,
		},
	}, nil
}
