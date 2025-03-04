package controller

import (
	"github.com/raghavyuva/nixopus-api/internal/utils"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// CreateGithubConnector godoc
// @Summary Create a new GitHub connector
// @Description Creates a new GitHub connector for the authenticated user
// @Tags github-connector
// @Accept json
// @Produce json
// @Param connector body types.CreateGithubConnectorRequest true "GitHub Connector creation request"
// @Success 200 {object} types.Response "Success response"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /github-connector [post]
func (c *GithubConnectorController) CreateGithubConnector(w http.ResponseWriter, r *http.Request) {
	var githubConnectorRequest types.CreateGithubConnectorRequest

	if !c.parseAndValidate(w, r, &githubConnectorRequest) {
		return
	}

	user := c.GetUser(w, r)

	if user == nil {
		return
	}

	err := c.service.CreateConnector(&githubConnectorRequest, user.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Github connector created", nil)
}
