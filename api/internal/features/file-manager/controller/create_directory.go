package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *FileManagerController) CreateDirectory(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	if path == "" {
		c.logger.Log(logger.Error, "path is required", "")
		utils.SendErrorResponse(w, "path is required", http.StatusBadRequest)
		return
	}

	err := c.service.CreateDirectory(path)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Directory created successfully", nil)
}
