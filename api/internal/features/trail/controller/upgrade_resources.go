package controller

import (
	"errors"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/trail/types"
)

func (c *TrailController) UpgradeResources(f fuego.ContextWithBody[types.UpgradeResourcesRequest]) (*types.UpgradeResourcesResponse, error) {
	r := f.Request()

	secret := r.Header.Get("X-Internal-Secret")
	if secret == "" || secret != config.AppConfig.BetterAuth.Secret {
		return nil, fuego.HTTPError{
			Err:    errors.New("unauthorized"),
			Status: http.StatusUnauthorized,
		}
	}

	body, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	if body.UserID == "" || body.OrgID == "" {
		return nil, fuego.HTTPError{
			Err:    errors.New("user_id and org_id are required"),
			Status: http.StatusBadRequest,
		}
	}

	if body.VcpuCount <= 0 || body.MemoryMB <= 0 {
		return nil, fuego.HTTPError{
			Err:    errors.New("vcpu_count and memory_mb must be positive"),
			Status: http.StatusBadRequest,
		}
	}

	if err := c.service.UpgradeResources(body.UserID, body.OrgID, body.VcpuCount, body.MemoryMB); err != nil {
		c.logger.Log(logger.Error, err.Error(), body.UserID)
		status := mapErrorToStatus(err)
		return nil, fuego.HTTPError{
			Err:    err,
			Status: status,
		}
	}

	return &types.UpgradeResourcesResponse{
		Status:  "success",
		Message: "Resource upgrade enqueued",
	}, nil
}
