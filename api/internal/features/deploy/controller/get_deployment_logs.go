package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

type GetDeploymentLogsRequest struct {
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	Level      string    `json:"level"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	SearchTerm string    `json:"search_term"`
}

func (c *DeployController) GetDeploymentLogs(f fuego.ContextNoBody) (*types.LogsResponse, error) {
	deploymentID := f.PathParam("deployment_id")
	page, _ := strconv.Atoi(f.QueryParam("page"))
	pageSize, _ := strconv.Atoi(f.QueryParam("page_size"))
	level := f.QueryParam("level")
	startTimeStr := f.QueryParam("start_time")
	endTimeStr := f.QueryParam("end_time")
	searchTerm := f.QueryParam("search_term")

	if page == 0 {
		page = 1
	}

	if pageSize == 0 {
		pageSize = 100
	}

	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return nil, fuego.BadRequestError{
				Detail: err.Error(),
				Err:    err,
			}
		}
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return nil, fuego.BadRequestError{
				Detail: err.Error(),
				Err:    err,
			}
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "deploy: user not found", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "deploy: organization not found", "")
		return nil, fuego.UnauthorizedError{
			Detail: "organization not found",
		}
	}

	logData := deployRequestData(f.Request(), user)
	if deploymentID != "" {
		if logData == "" {
			logData = fmt.Sprintf("deployment_id=%s", deploymentID)
		} else {
			logData = fmt.Sprintf("%s deployment_id=%s", logData, deploymentID)
		}
	}

	logs, totalCount, err := c.service.GetDeploymentLogs(f.Request().Context(), deploymentID, page, pageSize, level, startTime, endTime, searchTerm)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("deploy: GetDeploymentLogs: %v", err), logData)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.LogsResponse{
		Status:  "success",
		Message: "Deployment logs retrieved successfully",
		Data: types.LogsResponseData{
			Logs:       logs,
			TotalCount: totalCount,
			Page:       page,
			PageSize:   pageSize,
		},
	}, nil
}
