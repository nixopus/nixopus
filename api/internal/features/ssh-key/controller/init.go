package controller

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	auth_controller "github.com/raghavyuva/nixopus-api/internal/features/auth/controller"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	ssh_key_service "github.com/raghavyuva/nixopus-api/internal/features/ssh-key/service"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

type SSHKeyController struct {
	service *ssh_key_service.SSHKeyService
	store   *shared_storage.Store
	ctx     context.Context
	logger  logger.Logger
}

func NewSSHKeyController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	service *ssh_key_service.SSHKeyService,
) *SSHKeyController {
	return &SSHKeyController{
		service: service,
		store:   store,
		ctx:     ctx,
		logger:  l,
	}
}

// SetupSelfHostedSSHKey sets up SSH keys for a self-hosted user after registration
func (c *SSHKeyController) SetupSelfHostedSSHKey(f fuego.ContextNoBody) (interface{}, error) {
	w, r := f.Response(), f.Request()

	// Get authenticated user
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	// Check if user is email/password user
	var account auth_controller.Account
	err := c.store.DB.NewSelect().
		Model(&account).
		Where("user_id = ?", user.ID).
		Where("password IS NOT NULL").
		Where("provider_id = ?", "credential").
		Scan(c.ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			// User is not email/password (not self-hosted), skip SSH key setup
			c.logger.Log(logger.Info, "User is not email/password user, skipping SSH key setup", user.ID.String())
			return map[string]interface{}{
				"status":  "success",
				"message": "User is not self-hosted, SSH key setup skipped",
			}, nil
		}
		c.logger.Log(logger.Error, "Failed to check user account type", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	// Get user's default organization from Better Auth
	orgIDStr, err := utils.GetOrganizationIDFromBetterAuth(r)

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.logger.Log(logger.Error, "Invalid organization ID format", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	// Setup SSH key for self-hosted user
	if err := c.service.SetupSelfHostedSSHKey(user.ID, orgID); err != nil {
		c.logger.Log(logger.Error, "Failed to setup SSH key", err.Error())
		// Don't fail the request, just log the error
		// SSH key setup is not critical for user registration
		return map[string]interface{}{
			"status":  "warning",
			"message": "SSH key setup failed, but registration completed successfully",
		}, nil
	}

	return map[string]interface{}{
		"status":  "success",
		"message": "SSH key setup completed successfully",
	}, nil
}
