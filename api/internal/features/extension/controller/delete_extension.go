package controller

import (
	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/extension/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

func (c *ExtensionsController) DeleteFork(ctx fuego.ContextNoBody) (*types.MessageResponse, error) {
	id := ctx.PathParam("id")
	if id == "" {
		return nil, fuego.BadRequestError{Detail: "extension ID is required"}
	}
	if err := c.service.DeleteFork(id); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}
	return &types.MessageResponse{Status: "success", Message: "Fork deleted successfully"}, nil
}
