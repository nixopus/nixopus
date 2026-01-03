package features

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	container_feature "github.com/raghavyuva/nixopus-api/cmd/mcp-client/features/container"
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
		"get_container":      "container",
		"get_container_logs": "container",
		"list_containers":    "container",
		"list_images":        "container",
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

	return registry
}
