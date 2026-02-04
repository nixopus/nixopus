package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	container_types "github.com/raghavyuva/nixopus-api/internal/features/container/types"
)

type PruneBuildCacheRequest struct {
	All     bool   `json:"all,omitempty"`
	Filters string `json:"filters,omitempty"`
}

func (c *ContainerController) PruneBuildCache(f fuego.ContextWithBody[PruneBuildCacheRequest]) (*container_types.MessageResponse, error) {
	req, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	ctx := f.Request().Context()
	dockerService, err := c.getDockerService(ctx)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	opts := service.PruneBuildCacheOptions{
		All: req.All,
	}

	response, err := service.PruneBuildCache(dockerService, c.logger, opts)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &response, nil
}
