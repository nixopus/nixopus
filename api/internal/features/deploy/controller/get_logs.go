package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

type GetLogsRequest struct {
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	Level      string    `json:"level"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	SearchTerm string    `json:"search_term"`
}

func (c *DeployController) GetLogs(f fuego.ContextNoBody) (*shared_types.Response, error) {
	applicationID := f.PathParam("application_id")
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
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusBadRequest,
			}
		}
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusBadRequest,
			}
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "user not found", "")
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "organization not found", "")
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	logs, totalCount, err := c.service.GetLogs(applicationID, page, pageSize, level, startTime, endTime, searchTerm)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Logs retrieved successfully",
		Data: map[string]interface{}{
			"logs":        logs,
			"total_count": totalCount,
			"page":        page,
			"page_size":   pageSize,
		},
	}, nil
}
