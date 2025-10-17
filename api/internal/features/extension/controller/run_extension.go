package controller

import (
	"net/http"
	"strconv"

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

type ListLogsResponse struct {
	Logs      []types.ExtensionLog `json:"logs"`
	NextAfter int64                `json:"next_after"`
}

func (c *ExtensionsController) ListExecutionLogs(ctx fuego.ContextNoBody) (*ListLogsResponse, error) {
	execID := ctx.PathParam("execution_id")
	if execID == "" {
		return nil, fuego.HTTPError{Err: nil, Status: http.StatusBadRequest}
	}
	afterSeq := int64(0)
	if v := ctx.QueryParam("afterSeq"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			afterSeq = parsed
		}
	}
	limit := 200
	if v := ctx.QueryParam("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			limit = parsed
		}
	}
	logs, err := c.service.ListExecutionLogs(execID, afterSeq, limit)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Status: http.StatusInternalServerError}
	}
	var next int64 = afterSeq
	if len(logs) > 0 {
		next = logs[len(logs)-1].Sequence
	}
	return &ListLogsResponse{Logs: logs, NextAfter: next}, nil
}
