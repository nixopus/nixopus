package tools

import (
	"context"
	"strconv"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/service"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

// GetDeploymentLogsHandler returns the handler function for getting deployment logs
// Auth middleware is applied automatically during registration
func GetDeploymentLogsHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	deployService *service.DeployService,
) func(context.Context, *mcp.CallToolRequest, GetDeploymentLogsInput) (*mcp.CallToolResult, GetDeploymentLogsOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input GetDeploymentLogsInput,
	) (*mcp.CallToolResult, GetDeploymentLogsOutput, error) {
		deploymentID := input.ID

		page := 1
		if input.Page != "" {
			if p, err := strconv.Atoi(input.Page); err == nil && p > 0 {
				page = p
			}
		}

		pageSize := 100
		if input.PageSize != "" {
			if ps, err := strconv.Atoi(input.PageSize); err == nil && ps > 0 {
				pageSize = ps
			}
		}

		var startTime, endTime time.Time
		if input.StartTime != "" {
			parsedTime, err := time.Parse(time.RFC3339, input.StartTime)
			if err != nil {
				var zero GetDeploymentLogsOutput
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{Text: "invalid start_time format, expected RFC3339"},
					},
				}, zero, nil
			}
			startTime = parsedTime
		}

		if input.EndTime != "" {
			parsedTime, err := time.Parse(time.RFC3339, input.EndTime)
			if err != nil {
				var zero GetDeploymentLogsOutput
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{Text: "invalid end_time format, expected RFC3339"},
					},
				}, zero, nil
			}
			endTime = parsedTime
		}

		logs, totalCount, err := deployService.GetDeploymentLogs(toolCtx, deploymentID, page, pageSize, input.Level, startTime, endTime, input.SearchTerm)
		if err != nil {
			l.Log(logger.Error, "Failed to get deployment logs", err.Error())
			return nil, GetDeploymentLogsOutput{}, err
		}

		return nil, GetDeploymentLogsOutput{
			Response: types.LogsResponse{
				Status:  "success",
				Message: "Deployment logs retrieved successfully",
				Data: types.LogsResponseData{
					Logs:       logs,
					TotalCount: totalCount,
					Page:       page,
					PageSize:   pageSize,
				},
			},
		}, nil
	}
}
