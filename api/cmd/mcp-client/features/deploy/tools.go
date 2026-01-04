package deploy

import (
	"context"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	client_types "github.com/raghavyuva/nixopus-api/cmd/mcp-client/types"
	"github.com/raghavyuva/nixopus-api/cmd/mcp-client/utils"
)

// ToolHandler handles deploy feature tool calls
type ToolHandler struct{}

// NewToolHandler creates a new deploy tool handler
func NewToolHandler() *ToolHandler {
	return &ToolHandler{}
}

// GetToolParams returns the tool parameters for a given tool name
func (h *ToolHandler) GetToolParams(toolName string) (*mcp.CallToolParams, error) {
	applicationID := os.Getenv("APPLICATION_ID")
	authToken := os.Getenv("AUTH_TOKEN")

	if applicationID == "" {
		applicationID = "test-application-id"
	}
	if authToken == "" {
		fmt.Println("Warning: AUTH_TOKEN not set. Authentication will fail.")
		fmt.Println("   Set AUTH_TOKEN environment variable with a valid API key.")
	}

	var params *mcp.CallToolParams

	switch toolName {
	case "delete_application":
		params = &mcp.CallToolParams{
			Name: "delete_application",
			Arguments: map[string]any{
				"id": applicationID,
			},
		}
	case "get_application_deployments":
		arguments := map[string]any{
			"id": applicationID,
		}
		// Add optional pagination parameters if set
		if page := os.Getenv("PAGE"); page != "" {
			arguments["page"] = page
		}
		if pageSize := os.Getenv("PAGE_SIZE"); pageSize != "" {
			arguments["page_size"] = pageSize
		}
		params = &mcp.CallToolParams{
			Name:      "get_application_deployments",
			Arguments: arguments,
		}
	case "get_application":
		params = &mcp.CallToolParams{
			Name: "get_application",
			Arguments: map[string]any{
				"id": applicationID,
			},
		}
	case "get_applications":
		arguments := map[string]any{}
		// Add optional pagination parameters if set
		if page := os.Getenv("PAGE"); page != "" {
			arguments["page"] = page
		}
		if pageSize := os.Getenv("PAGE_SIZE"); pageSize != "" {
			arguments["page_size"] = pageSize
		}
		params = &mcp.CallToolParams{
			Name:      "get_applications",
			Arguments: arguments,
		}
	case "get_deployment_by_id":
		deploymentID := os.Getenv("DEPLOYMENT_ID")
		if deploymentID == "" {
			deploymentID = "test-deployment-id"
		}
		params = &mcp.CallToolParams{
			Name: "get_deployment_by_id",
			Arguments: map[string]any{
				"id": deploymentID,
			},
		}
	case "get_deployment_logs":
		deploymentID := os.Getenv("DEPLOYMENT_ID")
		if deploymentID == "" {
			deploymentID = "test-deployment-id"
		}
		arguments := map[string]any{
			"id": deploymentID,
		}
		// Add optional parameters if set
		if page := os.Getenv("PAGE"); page != "" {
			arguments["page"] = page
		}
		if pageSize := os.Getenv("PAGE_SIZE"); pageSize != "" {
			arguments["page_size"] = pageSize
		}
		if level := os.Getenv("LOG_LEVEL"); level != "" {
			arguments["level"] = level
		}
		if startTime := os.Getenv("START_TIME"); startTime != "" {
			arguments["start_time"] = startTime
		}
		if endTime := os.Getenv("END_TIME"); endTime != "" {
			arguments["end_time"] = endTime
		}
		if searchTerm := os.Getenv("SEARCH_TERM"); searchTerm != "" {
			arguments["search_term"] = searchTerm
		}
		params = &mcp.CallToolParams{
			Name:      "get_deployment_logs",
			Arguments: arguments,
		}
	case "create_project":
		arguments := map[string]any{
			"name":       os.Getenv("PROJECT_NAME"),
			"domain":     os.Getenv("PROJECT_DOMAIN"),
			"repository": os.Getenv("PROJECT_REPOSITORY"),
		}
		if env := os.Getenv("PROJECT_ENVIRONMENT"); env != "" {
			arguments["environment"] = env
		}
		if buildPack := os.Getenv("PROJECT_BUILD_PACK"); buildPack != "" {
			arguments["build_pack"] = buildPack
		}
		if branch := os.Getenv("PROJECT_BRANCH"); branch != "" {
			arguments["branch"] = branch
		}
		if port := os.Getenv("PROJECT_PORT"); port != "" {
			arguments["port"] = port
		}
		if dockerfilePath := os.Getenv("PROJECT_DOCKERFILE_PATH"); dockerfilePath != "" {
			arguments["dockerfile_path"] = dockerfilePath
		}
		if basePath := os.Getenv("PROJECT_BASE_PATH"); basePath != "" {
			arguments["base_path"] = basePath
		}
		if preRunCommand := os.Getenv("PROJECT_PRE_RUN_COMMAND"); preRunCommand != "" {
			arguments["pre_run_command"] = preRunCommand
		}
		if postRunCommand := os.Getenv("PROJECT_POST_RUN_COMMAND"); postRunCommand != "" {
			arguments["post_run_command"] = postRunCommand
		}
		params = &mcp.CallToolParams{
			Name:      "create_project",
			Arguments: arguments,
		}
	case "deploy_project":
		projectID := os.Getenv("PROJECT_ID")
		if projectID == "" {
			projectID = applicationID
		}
		params = &mcp.CallToolParams{
			Name: "deploy_project",
			Arguments: map[string]any{
				"id": projectID,
			},
		}
	case "duplicate_project":
		sourceProjectID := os.Getenv("SOURCE_PROJECT_ID")
		if sourceProjectID == "" {
			sourceProjectID = applicationID
		}
		arguments := map[string]any{
			"source_project_id": sourceProjectID,
			"domain":            os.Getenv("PROJECT_DOMAIN"),
			"environment":       os.Getenv("PROJECT_ENVIRONMENT"),
		}
		if branch := os.Getenv("PROJECT_BRANCH"); branch != "" {
			arguments["branch"] = branch
		}
		params = &mcp.CallToolParams{
			Name:      "duplicate_project",
			Arguments: arguments,
		}
	case "restart_deployment":
		deploymentID := os.Getenv("DEPLOYMENT_ID")
		if deploymentID == "" {
			deploymentID = "test-deployment-id"
		}
		params = &mcp.CallToolParams{
			Name: "restart_deployment",
			Arguments: map[string]any{
				"id": deploymentID,
			},
		}
	case "rollback_deployment":
		deploymentID := os.Getenv("DEPLOYMENT_ID")
		if deploymentID == "" {
			deploymentID = "test-deployment-id"
		}
		params = &mcp.CallToolParams{
			Name: "rollback_deployment",
			Arguments: map[string]any{
				"id": deploymentID,
			},
		}
	case "redeploy_application":
		arguments := map[string]any{
			"id": applicationID,
		}
		if force := os.Getenv("FORCE"); force == "true" {
			arguments["force"] = true
		}
		if forceWithoutCache := os.Getenv("FORCE_WITHOUT_CACHE"); forceWithoutCache == "true" {
			arguments["force_without_cache"] = true
		}
		params = &mcp.CallToolParams{
			Name:      "redeploy_application",
			Arguments: arguments,
		}
	case "update_project":
		arguments := map[string]any{
			"id": applicationID,
		}
		if name := os.Getenv("PROJECT_NAME"); name != "" {
			arguments["name"] = name
		}
		if env := os.Getenv("PROJECT_ENVIRONMENT"); env != "" {
			arguments["environment"] = env
		}
		if preRunCommand := os.Getenv("PROJECT_PRE_RUN_COMMAND"); preRunCommand != "" {
			arguments["pre_run_command"] = preRunCommand
		}
		if postRunCommand := os.Getenv("PROJECT_POST_RUN_COMMAND"); postRunCommand != "" {
			arguments["post_run_command"] = postRunCommand
		}
		if port := os.Getenv("PROJECT_PORT"); port != "" {
			arguments["port"] = port
		}
		if dockerfilePath := os.Getenv("PROJECT_DOCKERFILE_PATH"); dockerfilePath != "" {
			arguments["dockerfile_path"] = dockerfilePath
		}
		if basePath := os.Getenv("PROJECT_BASE_PATH"); basePath != "" {
			arguments["base_path"] = basePath
		}
		if force := os.Getenv("FORCE"); force == "true" {
			arguments["force"] = true
		}
		params = &mcp.CallToolParams{
			Name:      "update_project",
			Arguments: arguments,
		}
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}

	// Add auth token to metadata if provided
	if authToken != "" {
		params.Meta = mcp.Meta{
			"auth_token": authToken,
		}
	}

	return params, nil
}

// TestTool tests a deploy tool
func (h *ToolHandler) TestTool(ctx context.Context, session client_types.Session, toolName string) error {
	fmt.Printf("\nTesting %s tool...\n", toolName)

	params, err := h.GetToolParams(toolName)
	if err != nil {
		return err
	}

	res, err := session.CallTool(ctx, params)
	if err != nil {
		return fmt.Errorf("CallTool failed: %w", err)
	}

	utils.PrintToolResponse(res)

	if res.IsError {
		return fmt.Errorf("tool returned an error")
	}

	return nil
}

// GetAvailableTools returns the list of available deploy tools
func (h *ToolHandler) GetAvailableTools() []string {
	return []string{
		"delete_application",
		"get_application_deployments",
		"get_application",
		"get_applications",
		"get_deployment_by_id",
		"get_deployment_logs",
		"create_project",
		"deploy_project",
		"duplicate_project",
		"restart_deployment",
		"rollback_deployment",
		"redeploy_application",
		"update_project",
	}
}

// GetToolDescription returns the description for a tool
func (h *ToolHandler) GetToolDescription(toolName string) string {
	descriptions := map[string]string{
		"delete_application":          "Delete a deployed application. Requires application ID.",
		"get_application_deployments": "Get deployments for an application with pagination. Requires application ID. Optionally specify page and page_size.",
		"get_application":             "Get a single application by ID. Requires application ID.",
		"get_applications":            "Get all applications with pagination. Optionally specify page and page_size.",
		"get_deployment_by_id":        "Get a single deployment by ID. Requires deployment ID.",
		"get_deployment_logs":         "Get logs for a deployment with pagination and filtering. Requires deployment ID. Optionally specify page, page_size, level, start_time (RFC3339), end_time (RFC3339), and search_term.",
		"create_project":              "Create a new project (application) without triggering deployment. Requires name, domain, and repository. Optionally specify environment, build_pack, branch, port, dockerfile_path, base_path, pre_run_command, post_run_command, build_variables, and environment_variables.",
		"deploy_project":              "Deploy an existing project (application) that was saved as a draft. Requires project ID.",
		"duplicate_project":           "Duplicate an existing project with a different environment. Requires source_project_id, domain, and environment. Optionally specify branch.",
		"restart_deployment":          "Restart a deployment. Requires deployment ID.",
		"rollback_deployment":         "Rollback a deployment to a previous version. Requires deployment ID.",
		"redeploy_application":        "Redeploy an application. Requires application ID. Optionally specify force and force_without_cache.",
		"update_project":              "Update a project configuration without triggering deployment. Requires application ID. Optionally specify name, environment, pre_run_command, post_run_command, build_variables, environment_variables, port, force, dockerfile_path, and base_path.",
	}
	return descriptions[toolName]
}
