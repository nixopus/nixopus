package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

type GetApplicationsRequest struct {
	Page       string `json:"page"`
	PageSize   string `json:"page_size"`
	Repository string `json:"repository"`
}

func (c *DeployController) GetApplications(f fuego.ContextWithBody[GetApplicationsRequest]) (*types.ListApplicationsResponse, error) {
	w, r := f.Response(), f.Request()
	page := r.URL.Query().Get("page")
	pageSize := r.URL.Query().Get("page_size")
	organizationID := utils.GetOrganizationID(r)
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "organization not found", "")
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
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
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	applications, totalCount, err := c.service.GetApplications(page, pageSize, organizationID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
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
