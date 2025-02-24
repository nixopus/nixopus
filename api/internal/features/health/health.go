package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// HealthCheck handles HTTP requests to check the health status of the server.
// It responds with a JSON message indicating that the server is healthy.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.SendJSONResponse(w, "success", "Server is healthy", nil)
}
