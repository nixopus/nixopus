package features

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	container_feature "github.com/raghavyuva/nixopus-api/cmd/mcp-client/features/container"
	dashboard_feature "github.com/raghavyuva/nixopus-api/cmd/mcp-client/features/dashboard"
	deploy_feature "github.com/raghavyuva/nixopus-api/cmd/mcp-client/features/deploy"
	extension_feature "github.com/raghavyuva/nixopus-api/cmd/mcp-client/features/extension"
	file_manager_feature "github.com/raghavyuva/nixopus-api/cmd/mcp-client/features/file-manager"
	ssh_feature "github.com/raghavyuva/nixopus-api/cmd/mcp-client/features/ssh"
	client_types "github.com/raghavyuva/nixopus-api/cmd/mcp-client/types"
)

// FeatureHandler interface for feature-specific tool handlers
type FeatureHandler interface {
	GetToolParams(toolName string) (*mcp.CallToolParams, error)
	TestTool(ctx context.Context, session client_types.Session, toolName string) error
	GetAvailableTools() []string
	GetToolDescription(toolName string) string
}

// Registry manages feature handlers
type Registry struct {
	handlers map[string]FeatureHandler
}

// NewRegistry creates a new feature registry
func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[string]FeatureHandler),
	}
}

// RegisterFeature registers a feature handler
func (r *Registry) RegisterFeature(featureName string, handler FeatureHandler) {
	r.handlers[featureName] = handler
}

// GetHandler returns the handler for a feature
func (r *Registry) GetHandler(featureName string) (FeatureHandler, error) {
	handler, ok := r.handlers[featureName]
	if !ok {
		return nil, fmt.Errorf("unknown feature: %s", featureName)
	}
	return handler, nil
}

// GetToolFeature extracts the feature from a tool name
// e.g., "get_container" -> "container", "get_container_logs" -> "container"
func (r *Registry) GetToolFeature(toolName string) (string, error) {
	// Map tool names to features
	toolFeatureMap := map[string]string{
		"get_container":               "container",
		"get_container_logs":          "container",
		"list_containers":             "container",
		"list_images":                 "container",
		"prune_images":                "container",
		"prune_build_cache":           "container",
		"remove_container":            "container",
		"restart_container":           "container",
		"start_container":             "container",
		"stop_container":              "container",
		"update_container_resources":  "container",
		"list_files":                  "file-manager",
		"create_directory":            "file-manager",
		"delete_file":                 "file-manager",
		"move_file":                   "file-manager",
		"copy_directory":              "file-manager",
		"get_system_stats":            "dashboard",
		"run_command":                 "ssh",
		"list_extensions":             "extension",
		"get_extension":               "extension",
		"run_extension":               "extension",
		"get_execution":               "extension",
		"list_execution_logs":         "extension",
		"cancel_execution":            "extension",
		"delete_application":          "deploy",
		"get_application_deployments": "deploy",
		"get_application":             "deploy",
		"get_applications":            "deploy",
		"get_deployment_by_id":        "deploy",
		"get_deployment_logs":         "deploy",
		"create_project":              "deploy",
		"deploy_project":              "deploy",
		"duplicate_project":           "deploy",
		"restart_deployment":          "deploy",
		"rollback_deployment":         "deploy",
		"redeploy_application":        "deploy",
		"update_project":              "deploy",
	}

	feature, ok := toolFeatureMap[toolName]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}

	return feature, nil
}

// TestTool tests a tool by finding its feature and calling the appropriate handler
func (r *Registry) TestTool(ctx context.Context, session client_types.Session, toolName string) error {
	feature, err := r.GetToolFeature(toolName)
	if err != nil {
		return err
	}

	handler, err := r.GetHandler(feature)
	if err != nil {
		return err
	}

	return handler.TestTool(ctx, session, toolName)
}

// GetToolDescription returns the description for a tool
func (r *Registry) GetToolDescription(toolName string) string {
	feature, err := r.GetToolFeature(toolName)
	if err != nil {
		return ""
	}

	handler, err := r.GetHandler(feature)
	if err != nil {
		return ""
	}

	return handler.GetToolDescription(toolName)
}

// InitializeRegistry initializes the registry with all feature handlers
func InitializeRegistry() *Registry {
	registry := NewRegistry()

	// Register container feature
	containerHandler := container_feature.NewToolHandler()
	registry.RegisterFeature("container", containerHandler)

	// Register file-manager feature
	fileManagerHandler := file_manager_feature.NewToolHandler()
	registry.RegisterFeature("file-manager", fileManagerHandler)

	// Register dashboard feature
	dashboardHandler := dashboard_feature.NewToolHandler()
	registry.RegisterFeature("dashboard", dashboardHandler)

	// Register SSH feature
	sshHandler := ssh_feature.NewToolHandler()
	registry.RegisterFeature("ssh", sshHandler)

	// Register extension feature
	extensionHandler := extension_feature.NewToolHandler()
	registry.RegisterFeature("extension", extensionHandler)

	// Register deploy feature
	deployHandler := deploy_feature.NewToolHandler()
	registry.RegisterFeature("deploy", deployHandler)

	return registry
}
