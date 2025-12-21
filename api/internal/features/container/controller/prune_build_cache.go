package controller

import (
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/go-fuego/fuego"
	container_types "github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
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
	err = c.dockerService.PruneBuildCache(types.BuildCachePruneOptions{
		All: req.All,
	})
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &container_types.MessageResponse{
		Status:  "success",
		Message: "Build cache pruned successfully",
	}, nil
}
