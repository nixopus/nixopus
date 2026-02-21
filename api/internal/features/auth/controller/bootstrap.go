package auth

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/auth"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// BootstrapUser represents user in bootstrap response
type BootstrapUser struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	IsOnboarded     bool   `json:"isOnboarded"`
	ProvisionStatus string `json:"provisionStatus"`
}

// BootstrapOrg represents org in bootstrap response
type BootstrapOrg struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// BootstrapResponse is the bootstrap API response
type BootstrapResponse struct {
	User                 BootstrapUser  `json:"user"`
	Organizations        []BootstrapOrg `json:"organizations"`
	ActiveOrganizationID *string        `json:"activeOrganizationId"`
	HasServers           bool           `json:"hasServers"`
	ProvisionID          *string        `json:"provisionId,omitempty"`
	ProvisionStep        *string        `json:"provisionStep,omitempty"`
}

// HandleBootstrap returns user, orgs, activeOrgId, isOnboarded, provisionStatus, hasServers.
// Used by Kraken (and Niixopus View) for auth context init.
func (ac *AuthController) HandleBootstrap(c fuego.ContextNoBody) (*BootstrapResponse, error) {
	w, r := c.Response(), c.Request()
	ctx := r.Context()

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{Err: nil, Status: http.StatusUnauthorized}
	}

	// Get session for activeOrganizationId
	sessionResp, err := auth.VerifySession(r)
	if err != nil {
		ac.logger.Log(logger.Error, "bootstrap: verify session failed", err.Error())
		return nil, fuego.HTTPError{Err: err, Status: http.StatusUnauthorized}
	}

	provisionStatus := "NOT_STARTED"
	if user.ProvisionStatus != nil && *user.ProvisionStatus != "" {
		provisionStatus = *user.ProvisionStatus
	}

	// Query member + organization for user's orgs (Better Auth uses member table)
	var members []types.Member
	err = ac.store.DB.NewSelect().
		Model(&members).
		Where("user_id = ?", user.ID).
		Order("created_at ASC").
		Scan(ctx)
	if err != nil {
		ac.logger.Log(logger.Error, "bootstrap: failed to query members", err.Error())
	}

	orgs := make([]BootstrapOrg, 0, len(members))
	var firstOrgID string
	for _, m := range members {
		var org types.Organization
		errOrg := ac.store.DB.NewSelect().Model(&org).Where("id = ?", m.OrganizationID).Scan(ctx)
		if errOrg != nil {
			continue
		}
		orgs = append(orgs, BootstrapOrg{
			ID:   org.ID.String(),
			Name: org.Name,
			Role: m.Role,
		})
		if firstOrgID == "" {
			firstOrgID = org.ID.String()
		}
	}

	// activeOrganizationId: session ?? first org
	activeOrgID := firstOrgID
	if sessionResp.Session.ActiveOrganizationID != nil && *sessionResp.Session.ActiveOrganizationID != "" {
		activeOrgID = *sessionResp.Session.ActiveOrganizationID
	}
	var activeOrgIDPtr *string
	if activeOrgID != "" {
		activeOrgIDPtr = &activeOrgID
	}

	// hasServers: active org has rows in ssh_keys
	hasServers := false
	if activeOrgID != "" {
		orgUUID, _ := uuid.Parse(activeOrgID)
		exists, errExists := ac.store.DB.NewSelect().
			Table("ssh_keys").
			ColumnExpr("1").
			Where("organization_id = ?", orgUUID).
			Where("deleted_at IS NULL").
			Limit(1).
			Exists(ctx)
		if errExists == nil && exists {
			hasServers = true
		}
	}

	// provisionId and provisionStep when PROVISIONING (from user_provision_details)
	var provisionID *string
	var provisionStep *string
	if provisionStatus == "PROVISIONING" {
		var upd types.UserProvisionDetails
		err = ac.store.DB.NewSelect().
			Model(&upd).
			Where("user_id = ?", user.ID).
			Order("created_at DESC").
			Limit(1).
			Scan(ctx)
		if err != nil {
			ac.logger.Log(logger.Error, "bootstrap: failed to query user_provision_details", err.Error())
			return nil, fuego.HTTPError{Err: err, Status: http.StatusInternalServerError}
		}
		id := upd.ID.String()
		provisionID = &id
		if upd.Step != nil {
			step := string(*upd.Step)
			provisionStep = &step
		}
	}

	return &BootstrapResponse{
		User: BootstrapUser{
			ID:              user.ID.String(),
			Name:            user.Name,
			Email:           user.Email,
			IsOnboarded:     user.IsOnboarded,
			ProvisionStatus: provisionStatus,
		},
		Organizations:        orgs,
		ActiveOrganizationID: activeOrgIDPtr,
		HasServers:           hasServers,
		ProvisionID:          provisionID,
		ProvisionStep:        provisionStep,
	}, nil
}
