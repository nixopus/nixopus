package controller

import (
	"net/http"
	"strconv"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

type GetApplicationDeploymentsRequest struct {
	ID       string `json:"id"`
	Page     string `json:"page"`
	PageSize string `json:"limit"`
}

func (c *DeployController) GetApplicationDeployments(f fuego.ContextWithBody[GetApplicationDeploymentsRequest]) (*types.ListDeploymentsResponse, error) {
	r := f.Request()
	id := r.URL.Query().Get("id")
	page := r.URL.Query().Get("page")
	pageSize := r.URL.Query().Get("limit")

	if id == "" {
		c.logger.Log(logger.Error, "application ID is required", "")
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	if page == "" {
		page = "1"
	}

	if pageSize == "" {
		pageSize = "10"
	}

	applicationID, err := uuid.Parse(id)
	if err != nil {
		c.logger.Log(logger.Error, "Invalid application ID", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = 1
	}

	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		pageSizeInt = 10
	}

	deployments, totalCount, err := c.service.GetApplicationDeployments(applicationID, pageInt, pageSizeInt)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to get application deployments", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.ListDeploymentsResponse{
		Status:  "success",
		Message: "Application deployments retrieved successfully",
		Data: types.ListDeploymentsResponseData{
			Deployments: deployments,
			TotalCount:  totalCount,
			Page:        page,
			PageSize:    pageSize,
		},
	}, nil
}
