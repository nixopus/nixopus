package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// HealthCheck godoc
// @Summary Check if the server is up
// @Description Simple health check
// @Tags health
// @Produce json
// @Success 200 {object} types.Response "Success response"
// @Router /health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.SendJSONResponse(w, "success", "Server is healthy", nil)
}