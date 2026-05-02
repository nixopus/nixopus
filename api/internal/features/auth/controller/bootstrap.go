package auth

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/auth"
	auth_types "github.com/nixopus/nixopus/api/internal/features/auth/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

// HandleBootstrap returns user, orgs, activeOrgId, isOnboarded, provisionStatus, hasServers.
func (ac *AuthController) HandleBootstrap(c fuego.ContextNoBody) (*auth_types.BootstrapResponse, error) {
	w, r := c.Response(), c.Request()
	ctx := r.Context()

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	sessionResp, err := auth.VerifySession(r)
	if err != nil {
		ac.logger.Log(logger.Error, "bootstrap: verify session failed", err.Error())
		return nil, fuego.UnauthorizedError{Detail: err.Error(), Err: err}
	}

	resp, err := ac.service.BuildBootstrap(ctx, user, sessionResp.Session.ActiveOrganizationID)
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}
	return resp, nil
}
