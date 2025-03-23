package controller

import (
	"github.com/raghavyuva/nixopus-api/internal/utils"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// GetGithubConnectors godoc
// @Summary Retrieve all GitHub connectors for the authenticated user
// @Description Retrieves a list of all GitHub connectors associated with the authenticated user
// @Tags github-connector
// @Produce json
// @Success 200 {array} types.GithubConnector "Success response with connectors"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /github-connectors [get]
func (c *GithubConnectorController) GetGithubConnectors(w http.ResponseWriter, r *http.Request) {
	user := c.GetUser(w, r)

	if user == nil {
		return
	}

	connectors, err := c.service.GetAllConnectors(user.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Github connectors", connectors)
}
