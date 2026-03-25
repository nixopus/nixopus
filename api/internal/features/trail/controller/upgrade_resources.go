package controller

import (
	"errors"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/trail/types"
)

func (c *TrailController) UpgradeResources(f fuego.ContextWithBody[types.UpgradeResourcesRequest]) (*types.UpgradeResourcesResponse, error) {
	r := f.Request()

	secret := r.Header.Get("X-Internal-Secret")
	if secret == "" || secret != config.AppConfig.BetterAuth.Secret {
		return nil, fuego.UnauthorizedError{Detail: "unauthorized", Err: errors.New("unauthorized")}
	}

	body, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	if body.UserID == "" || body.OrgID == "" {
		return nil, fuego.BadRequestError{Detail: "user_id and org_id are required", Err: errors.New("user_id and org_id are required")}
	}

	if body.VcpuCount <= 0 || body.MemoryMB <= 0 {
		return nil, fuego.BadRequestError{Detail: "vcpu_count and memory_mb must be positive", Err: errors.New("vcpu_count and memory_mb must be positive")}
	}

	if err := c.service.UpgradeResources(body.UserID, body.OrgID, body.VcpuCount, body.MemoryMB); err != nil {
		c.logger.Log(logger.Error, err.Error(), body.UserID)
		status := mapErrorToStatus(err)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: status,
		}
	}

	return &types.UpgradeResourcesResponse{
		Status:  "success",
		Message: "Resource upgrade enqueued",
	}, nil
}
