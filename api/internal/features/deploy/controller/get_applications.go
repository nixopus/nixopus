package controller

import (
	"fmt"
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
		c.logger.Log(logger.Error, "deploy: organization not found", "")
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
		c.logger.Log(logger.Error, "deploy: user not found", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	var serverID *uuid.UUID
	if s := r.URL.Query().Get("server_id"); s != "" {
		if parsed, err := uuid.Parse(s); err == nil {
			serverID = &parsed
		}
	}

	logData := deployRequestData(w, r, user)

	applications, totalCount, err := c.service.GetApplications(page, pageSize, sortBy, sortDirection, organizationID, serverID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("deploy: GetApplications: %v", err), logData)
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
