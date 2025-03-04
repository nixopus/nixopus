package controller

import (
	"github.com/raghavyuva/nixopus-api/internal/utils"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// UpdateGithubConnectorRequest godoc
// @Summary Update a GitHub connector request
// @Description Updates a GitHub connector request for the authenticated user
// @Tags github-connector
// @Accept json
// @Produce json
// @Param connector body types.UpdateGithubConnectorRequest true "GitHub Connector request update"
// @Success 200 {object} types.Response "Success response"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /github-connector-request [put]
func (c *GithubConnectorController) UpdateGithubConnectorRequest(w http.ResponseWriter, r *http.Request) {
	var UpdateConnectorRequest types.UpdateGithubConnectorRequest

	if !c.parseAndValidate(w, r, &UpdateConnectorRequest) {
		return
	}

	user := c.GetUser(w, r)

	if user == nil {
		return
	}

	err := c.service.UpdateGithubConnectorRequest(UpdateConnectorRequest.InstallationID, user.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Github connector request updated", nil)
}
