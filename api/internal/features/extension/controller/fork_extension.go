package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/extension/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type ForkExtensionRequest struct {
	YAMLContent *string `json:"yaml_content"`
}

func (c *ExtensionsController) ForkExtension(ctx fuego.ContextWithBody[ForkExtensionRequest]) (*types.ExtensionResponse, error) {
	extensionID := ctx.PathParam("extension_id")
	if extensionID == "" {
		return nil, fuego.HTTPError{Err: nil, Status: http.StatusBadRequest}
	}
	req, err := ctx.Body()
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Status: http.StatusBadRequest}
	}
	var yamlOverride string
	if req.YAMLContent != nil {
		yamlOverride = *req.YAMLContent
	}
	authorName := ""
	if userAny := ctx.Request().Context().Value(shared_types.UserContextKey); userAny != nil {
		if u, ok := userAny.(*shared_types.User); ok && u != nil {
			authorName = u.Username
		}
	}
	newExt, err := c.service.ForkExtension(extensionID, yamlOverride, authorName)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Status: http.StatusInternalServerError}
	}
	return &types.ExtensionResponse{
		Status:  "success",
		Message: "Extension forked successfully",
		Data:    *newExt,
	}, nil
}
