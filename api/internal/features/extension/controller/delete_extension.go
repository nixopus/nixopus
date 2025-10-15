package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func (c *ExtensionsController) DeleteFork(ctx fuego.ContextNoBody) (*struct {
	Status string `json:"status"`
}, error) {
	id := ctx.PathParam("id")
	if id == "" {
		return nil, fuego.HTTPError{Err: nil, Status: http.StatusBadRequest}
	}
	if err := c.service.DeleteFork(id); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Status: http.StatusBadRequest}
	}
	return &struct {
		Status string `json:"status"`
	}{Status: "ok"}, nil
}
