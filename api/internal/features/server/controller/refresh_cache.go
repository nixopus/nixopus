package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/server/types"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// RefreshSSHCache invalidates SSH and Docker caches for the org; new requests fetch fresh config.
func (c *ServerController) RefreshSSHCache(f fuego.ContextNoBody) (*types.MessageResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{Err: nil, Status: http.StatusUnauthorized}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logger.Log(logger.Error, "Organization ID not found in context", "")
		return nil, fuego.HTTPError{Err: nil, Status: http.StatusBadRequest}
	}

	ssh.InvalidateSSHManagerCache(orgID)
	docker.InvalidateDockerServiceCache(orgID)

	return &types.MessageResponse{
		Status:  "success",
		Message: "SSH and Docker caches invalidated for organization",
	}, nil
}
