package auth

import (
	"fmt"
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
		ac.logger.Log(logger.Error, "bootstrap: authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	userCtx := fmt.Sprintf("user_id=%s", user.ID)
	ac.logger.Log(logger.Info, "bootstrap: start", userCtx)

	sessionResp, err := auth.VerifySession(r)
	if err != nil {
		ac.logger.Log(logger.Error, fmt.Sprintf("bootstrap: verify session failed: %v", err), userCtx)
		return nil, fuego.UnauthorizedError{Detail: err.Error(), Err: err}
	}

	activeOrgID := "none"
	if p := sessionResp.Session.ActiveOrganizationID; p != nil && *p != "" {
		activeOrgID = *p
	}
	ctxStr := fmt.Sprintf("%s active_org_id=%s", userCtx, activeOrgID)
	ac.logger.Log(logger.Info, "bootstrap: session verified", ctxStr)

	resp, err := ac.service.BuildBootstrap(ctx, user, sessionResp.Session.ActiveOrganizationID)
	if err != nil {
		ac.logger.Log(logger.Error, fmt.Sprintf("bootstrap: build failed: %v", err), ctxStr)
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}

	ac.logger.Log(logger.Info, "bootstrap: success", fmt.Sprintf(
		"%s org_count=%d has_servers=%v is_onboarded=%v provision_status=%q",
		ctxStr,
		len(resp.Organizations),
		resp.HasServers,
		resp.User.IsOnboarded,
		resp.User.ProvisionStatus,
	))
	return resp, nil
}
