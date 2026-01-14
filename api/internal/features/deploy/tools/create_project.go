package tools

import (
	"context"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/service"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	mcp_middleware "github.com/raghavyuva/nixopus-api/internal/mcp/middleware"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// CreateProjectHandler returns the handler function for creating a project
// Auth middleware is applied automatically during registration
func CreateProjectHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	deployService *service.DeployService,
) func(context.Context, *mcp.CallToolRequest, CreateProjectInput) (*mcp.CallToolResult, CreateProjectOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input CreateProjectInput,
	) (*mcp.CallToolResult, CreateProjectOutput, error) {
		orgID, err := mcp_middleware.GetOrganizationIDFromContext(toolCtx)
		if err != nil {
			var zero CreateProjectOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		organizationID, err := uuid.Parse(orgID)
		if err != nil {
			return nil, CreateProjectOutput{}, err
		}

		user, err := mcp_middleware.AuthenticateUser(toolCtx, store, l)
		if err != nil {
			var zero CreateProjectOutput
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, zero, nil
		}
		userID := user.ID

		// Convert string inputs to proper types
		var environment shared_types.Environment
		if input.Environment != "" {
			environment = shared_types.Environment(input.Environment)
		} else {
			environment = shared_types.Production
		}

		var buildPack shared_types.BuildPack
		if input.BuildPack != "" {
			buildPack = shared_types.BuildPack(input.BuildPack)
		} else {
			buildPack = shared_types.DockerFile
		}

		createRequest := types.CreateProjectRequest{
			Name:                 input.Name,
			Domains:              input.Domains,
			Repository:           input.Repository,
			Environment:          environment,
			BuildPack:            buildPack,
			Branch:               input.Branch,
			PreRunCommand:        input.PreRunCommand,
			PostRunCommand:       input.PostRunCommand,
			BuildVariables:       input.BuildVariables,
			EnvironmentVariables: input.EnvironmentVariables,
			Port:                 input.Port,
			DockerfilePath:       input.DockerfilePath,
			BasePath:             input.BasePath,
		}

		application, err := deployService.CreateProject(&createRequest, userID, organizationID)
		if err != nil {
			l.Log(logger.Error, "Failed to create project", err.Error())
			return nil, CreateProjectOutput{}, err
		}

		return nil, CreateProjectOutput{
			Response: types.ApplicationResponse{
				Status:  "success",
				Message: "Project created successfully. You can deploy it when ready.",
				Data:    application,
			},
		}, nil
	}
}
