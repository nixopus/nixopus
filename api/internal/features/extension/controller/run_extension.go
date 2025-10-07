package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *ExtensionsController) RunExtension(ctx fuego.ContextWithBody[RunExtensionRequest]) (*types.ExtensionExecution, error) {
	extensionID := ctx.PathParam("extension_id")
	if extensionID == "" {
		return nil, fuego.HTTPError{Err: nil, Status: http.StatusBadRequest}
	}
	contentType := ctx.Request().Header.Get("Content-Type")
	if contentType != "" && (len(contentType) >= 19 && contentType[:19] == "multipart/form-data") {
		vars, err := c.service.ParseMultipartRunRequest(ctx.Request())
		if err != nil {
			return nil, fuego.HTTPError{Err: err, Status: http.StatusBadRequest}
		}
		exec, err := c.service.StartRun(extensionID, vars)
		if err != nil {
			c.logger.Log(logger.Error, err.Error(), "")
			return nil, fuego.HTTPError{Err: err, Status: http.StatusInternalServerError}
		}
		return exec, nil
	}

	req, err := ctx.Body()
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Status: http.StatusBadRequest}
	}
	exec, err := c.service.StartRun(extensionID, req.Variables)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Status: http.StatusInternalServerError}
	}
	return exec, nil
}

func (c *ExtensionsController) CancelExecution(ctx fuego.ContextNoBody) (*types.Response, error) {
	execID := ctx.PathParam("execution_id")
	if execID == "" {
		return nil, fuego.HTTPError{Err: nil, Status: http.StatusBadRequest}
	}
	if err := c.service.CancelExecution(execID); err != nil {
		return nil, fuego.HTTPError{Err: err, Status: http.StatusInternalServerError}
	}
	return &types.Response{Status: "success", Message: "Execution cancelled"}, nil
}
