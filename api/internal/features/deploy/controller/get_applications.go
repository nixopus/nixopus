package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *DeployController) GetApplications(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	pageSize := r.URL.Query().Get("page_size")

	if page == "" {
		page = "1"
	}

	if pageSize == "" {
		pageSize = "10"
	}

	user := c.GetUser(w, r)

	if user == nil {
		c.logger.Log(logger.Error, "user not found", "")
		return
	}

	applications, totalCount, err := c.service.GetApplications(page, pageSize)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Applications", map[string]interface{}{
		"applications": applications,
		"total_count":  totalCount,
		"page":         page,
		"page_size":    pageSize,
	})
}
