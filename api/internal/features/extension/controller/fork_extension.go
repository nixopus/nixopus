package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type ForkExtensionRequest struct {
	YAMLContent *string `json:"yaml_content"`
}

func (c *ExtensionsController) ForkExtension(ctx fuego.ContextWithBody[ForkExtensionRequest]) (types.Extension, error) {
	extensionID := ctx.PathParam("extension_id")
	if extensionID == "" {
		return types.Extension{}, fuego.HTTPError{Err: nil, Status: http.StatusBadRequest}
	}
	req, err := ctx.Body()
	if err != nil {
		return types.Extension{}, fuego.HTTPError{Err: err, Status: http.StatusBadRequest}
	}
	var yamlOverride string
	if req.YAMLContent != nil {
		yamlOverride = *req.YAMLContent
	}
	authorName := ""
	if userAny := ctx.Request().Context().Value(types.UserContextKey); userAny != nil {
		if u, ok := userAny.(*types.User); ok && u != nil {
			authorName = u.Username
		}
	}
	newExt, err := c.service.ForkExtension(extensionID, yamlOverride, authorName)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return types.Extension{}, fuego.HTTPError{Err: err, Status: http.StatusInternalServerError}
	}
	return *newExt, nil
}
