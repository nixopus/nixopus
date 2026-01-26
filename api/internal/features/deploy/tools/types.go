package tools

import (
	"time"

	"github.com/google/uuid"
	deploy_types "github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/uptrace/bun"
)

// DeleteApplicationInput is the input structure for the MCP tool
type DeleteApplicationInput struct {
	ID string `json:"id" jsonschema:"required"`
}

// DeleteApplicationOutput is the output structure for the MCP tool
type DeleteApplicationOutput struct {
	Response deploy_types.MessageResponse `json:"response"`
}

// GetApplicationDeploymentsInput is the input structure for the MCP tool
type GetApplicationDeploymentsInput struct {
	ID       string `json:"id" jsonschema:"required"`
	Page     string `json:"page,omitempty"`
	PageSize string `json:"page_size,omitempty"`
}

// MCPApplicationDeployment is a simplified ApplicationDeployment without circular references
type MCPApplicationDeployment struct {
	ID              string    `json:"id"`
	ApplicationID   string    `json:"application_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	CommitHash      string    `json:"commit_hash"`
	ContainerID     string    `json:"container_id"`
	ContainerName   string    `json:"container_name"`
	ContainerImage  string    `json:"container_image"`
	ContainerStatus string    `json:"container_status"`
	Status          string    `json:"status,omitempty"`
}

// MCPListDeploymentsResponseData contains the data for list deployments response
type MCPListDeploymentsResponseData struct {
	Deployments []MCPApplicationDeployment `json:"deployments"`
	TotalCount  int                        `json:"total_count"`
	Page        string                     `json:"page"`
	PageSize    string                     `json:"page_size"`
}

// MCPListDeploymentsResponse is the typed response for listing deployments without circular references
type MCPListDeploymentsResponse struct {
	Status  string                         `json:"status"`
	Message string                         `json:"message"`
	Data    MCPListDeploymentsResponseData `json:"data"`
}

// GetApplicationDeploymentsOutput is the output structure for the MCP tool
type GetApplicationDeploymentsOutput struct {
	Response MCPListDeploymentsResponse `json:"response"`
}

// MCPApplication is a simplified Application without circular references
type MCPApplication struct {
	ID                   string                 `json:"id"`
	Name                 string                 `json:"name"`
	Port                 int                    `json:"port"`
	Environment          string                 `json:"environment"`
	ProxyServer          string                 `json:"proxy_server"`
	BuildVariables       string                 `json:"build_variables"`
	EnvironmentVariables string                 `json:"environment_variables"`
	BuildPack            string                 `json:"build_pack"`
	Repository           string                 `json:"repository"`
	Branch               string                 `json:"branch"`
	PreRunCommand        string                 `json:"pre_run_command"`
	PostRunCommand       string                 `json:"post_run_command"`
	Domains              []MCPApplicationDomain `json:"domains"`
	DockerfilePath       string                 `json:"dockerfile_path"`
	BasePath             string                 `json:"base_path"`
	UserID               string                 `json:"user_id"`
	OrganizationID       string                 `json:"organization_id"`
	FamilyID             *string                `json:"family_id,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
	Status               string                 `json:"status,omitempty"`
	Labels               []string               `json:"labels,omitempty"`
}

type MCPApplicationDomain struct {
	bun.BaseModel `bun:"table:application_domains,alias:ad" swaggerignore:"true"`
	ID            uuid.UUID `json:"id" bun:"id,pk,type:uuid"`
	ApplicationID uuid.UUID `json:"application_id" bun:"application_id,notnull,type:uuid"`
	Domain        string    `json:"domain" bun:"domain,notnull"`
	CreatedAt     time.Time `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`

	Application *MCPApplication `json:"application,omitempty" bun:"rel:belongs-to,join:application_id=id"`
}

// MCPApplicationResponse is the typed response for single application without circular references
type MCPApplicationResponse struct {
	Status  string         `json:"status"`
	Message string         `json:"message"`
	Data    MCPApplication `json:"data"`
}

// MCPListApplicationsResponseData contains the data for list applications response
type MCPListApplicationsResponseData struct {
	Applications []MCPApplication `json:"applications"`
	TotalCount   int              `json:"total_count"`
	Page         string           `json:"page"`
	PageSize     string           `json:"page_size"`
}

// MCPListApplicationsResponse is the typed response for listing applications without circular references
type MCPListApplicationsResponse struct {
	Status  string                          `json:"status"`
	Message string                          `json:"message"`
	Data    MCPListApplicationsResponseData `json:"data"`
}

// GetApplicationInput is the input structure for the MCP tool
type GetApplicationInput struct {
	ID string `json:"id" jsonschema:"required"`
}

// GetApplicationOutput is the output structure for the MCP tool
type GetApplicationOutput struct {
	Response MCPApplicationResponse `json:"response"`
}

// GetApplicationsInput is the input structure for the MCP tool
type GetApplicationsInput struct {
	Page     string `json:"page,omitempty"`
	PageSize string `json:"page_size,omitempty"`
}

// GetApplicationsOutput is the output structure for the MCP tool
type GetApplicationsOutput struct {
	Response MCPListApplicationsResponse `json:"response"`
}

// MCPDeploymentResponse is the typed response for single deployment without circular references
type MCPDeploymentResponse struct {
	Status  string                   `json:"status"`
	Message string                   `json:"message"`
	Data    MCPApplicationDeployment `json:"data"`
}

// GetDeploymentByIdInput is the input structure for the MCP tool
type GetDeploymentByIdInput struct {
	ID string `json:"id" jsonschema:"required"`
}

// GetDeploymentByIdOutput is the output structure for the MCP tool
type GetDeploymentByIdOutput struct {
	Response MCPDeploymentResponse `json:"response"`
}

// MCPApplicationLogs is a simplified ApplicationLogs without circular references
type MCPApplicationLogs struct {
	ID                      string    `json:"id"`
	ApplicationID           string    `json:"application_id"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
	Log                     string    `json:"log"`
	ApplicationDeploymentID string    `json:"application_deployment_id"`
}

// MCPLogsResponseData contains the data for logs response
type MCPLogsResponseData struct {
	Logs       []MCPApplicationLogs `json:"logs"`
	TotalCount int64                `json:"total_count"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
}

// MCPLogsResponse is the typed response for logs without circular references
type MCPLogsResponse struct {
	Status  string              `json:"status"`
	Message string              `json:"message"`
	Data    MCPLogsResponseData `json:"data"`
}

// GetDeploymentLogsInput is the input structure for the MCP tool
type GetDeploymentLogsInput struct {
	ID         string `json:"id" jsonschema:"required"`
	Page       string `json:"page,omitempty"`
	PageSize   string `json:"page_size,omitempty"`
	Level      string `json:"level,omitempty"`
	StartTime  string `json:"start_time,omitempty"`
	EndTime    string `json:"end_time,omitempty"`
	SearchTerm string `json:"search_term,omitempty"`
}

// GetDeploymentLogsOutput is the output structure for the MCP tool
type GetDeploymentLogsOutput struct {
	Response MCPLogsResponse `json:"response"`
}

// CreateProjectInput is the input structure for the MCP tool
type CreateProjectInput struct {
	Name                 string            `json:"name" jsonschema:"required"`
	Domains              []string          `json:"domains,omitempty"`
	Repository           string            `json:"repository" jsonschema:"required"`
	Environment          string            `json:"environment,omitempty"`
	BuildPack            string            `json:"build_pack,omitempty"`
	Branch               string            `json:"branch,omitempty"`
	PreRunCommand        string            `json:"pre_run_command,omitempty"`
	PostRunCommand       string            `json:"post_run_command,omitempty"`
	BuildVariables       map[string]string `json:"build_variables,omitempty"`
	EnvironmentVariables map[string]string `json:"environment_variables,omitempty"`
	Port                 int               `json:"port,omitempty"`
	DockerfilePath       string            `json:"dockerfile_path,omitempty"`
	BasePath             string            `json:"base_path,omitempty"`
}

// CreateProjectOutput is the output structure for the MCP tool
type CreateProjectOutput struct {
	Response MCPApplicationResponse `json:"response"`
}

// DeployProjectInput is the input structure for the MCP tool
type DeployProjectInput struct {
	ID string `json:"id" jsonschema:"required"`
}

// DeployProjectOutput is the output structure for the MCP tool
type DeployProjectOutput struct {
	Response MCPApplicationResponse `json:"response"`
}

// DuplicateProjectInput is the input structure for the MCP tool
type DuplicateProjectInput struct {
	SourceProjectID string   `json:"source_project_id" jsonschema:"required"`
	Domains         []string `json:"domains,omitempty"`
	Environment     string   `json:"environment" jsonschema:"required"`
	Branch          string   `json:"branch,omitempty"`
}

// DuplicateProjectOutput is the output structure for the MCP tool
type DuplicateProjectOutput struct {
	Response MCPApplicationResponse `json:"response"`
}

// RestartDeploymentInput is the input structure for the MCP tool
type RestartDeploymentInput struct {
	ID string `json:"id" jsonschema:"required"`
}

// RestartDeploymentOutput is the output structure for the MCP tool
type RestartDeploymentOutput struct {
	Response deploy_types.MessageResponse `json:"response"`
}

// RollbackDeploymentInput is the input structure for the MCP tool
type RollbackDeploymentInput struct {
	ID string `json:"id" jsonschema:"required"`
}

// RollbackDeploymentOutput is the output structure for the MCP tool
type RollbackDeploymentOutput struct {
	Response deploy_types.MessageResponse `json:"response"`
}

// RedeployApplicationInput is the input structure for the MCP tool
type RedeployApplicationInput struct {
	ID                string `json:"id" jsonschema:"required"`
	Force             bool   `json:"force,omitempty"`
	ForceWithoutCache bool   `json:"force_without_cache,omitempty"`
}

// RedeployApplicationOutput is the output structure for the MCP tool
type RedeployApplicationOutput struct {
	Response MCPApplicationResponse `json:"response"`
}

// UpdateProjectInput is the input structure for the MCP tool
type UpdateProjectInput struct {
	ID                   string            `json:"id" jsonschema:"required"`
	Name                 string            `json:"name,omitempty"`
	Environment          string            `json:"environment,omitempty"`
	PreRunCommand        string            `json:"pre_run_command,omitempty"`
	PostRunCommand       string            `json:"post_run_command,omitempty"`
	BuildVariables       map[string]string `json:"build_variables,omitempty"`
	EnvironmentVariables map[string]string `json:"environment_variables,omitempty"`
	Port                 int               `json:"port,omitempty"`
	Force                bool              `json:"force,omitempty"`
	DockerfilePath       string            `json:"dockerfile_path,omitempty"`
	BasePath             string            `json:"base_path,omitempty"`
}

// UpdateProjectOutput is the output structure for the MCP tool
type UpdateProjectOutput struct {
	Response MCPApplicationResponse `json:"response"`
}
