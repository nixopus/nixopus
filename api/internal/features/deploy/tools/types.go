package tools

import deploy_types "github.com/raghavyuva/nixopus-api/internal/features/deploy/types"

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

// GetApplicationDeploymentsOutput is the output structure for the MCP tool
type GetApplicationDeploymentsOutput struct {
	Response deploy_types.ListDeploymentsResponse `json:"response"`
}

// GetApplicationInput is the input structure for the MCP tool
type GetApplicationInput struct {
	ID string `json:"id" jsonschema:"required"`
}

// GetApplicationOutput is the output structure for the MCP tool
type GetApplicationOutput struct {
	Response deploy_types.ApplicationResponse `json:"response"`
}

// GetApplicationsInput is the input structure for the MCP tool
type GetApplicationsInput struct {
	Page     string `json:"page,omitempty"`
	PageSize string `json:"page_size,omitempty"`
}

// GetApplicationsOutput is the output structure for the MCP tool
type GetApplicationsOutput struct {
	Response deploy_types.ListApplicationsResponse `json:"response"`
}

// GetDeploymentByIdInput is the input structure for the MCP tool
type GetDeploymentByIdInput struct {
	ID string `json:"id" jsonschema:"required"`
}

// GetDeploymentByIdOutput is the output structure for the MCP tool
type GetDeploymentByIdOutput struct {
	Response deploy_types.DeploymentResponse `json:"response"`
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
	Response deploy_types.LogsResponse `json:"response"`
}

// CreateProjectInput is the input structure for the MCP tool
type CreateProjectInput struct {
	Name                 string            `json:"name" jsonschema:"required"`
	Domain               string            `json:"domain" jsonschema:"required"`
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
	Response deploy_types.ApplicationResponse `json:"response"`
}

// DeployProjectInput is the input structure for the MCP tool
type DeployProjectInput struct {
	ID string `json:"id" jsonschema:"required"`
}

// DeployProjectOutput is the output structure for the MCP tool
type DeployProjectOutput struct {
	Response deploy_types.ApplicationResponse `json:"response"`
}

// DuplicateProjectInput is the input structure for the MCP tool
type DuplicateProjectInput struct {
	SourceProjectID string `json:"source_project_id" jsonschema:"required"`
	Domain          string `json:"domain" jsonschema:"required"`
	Environment     string `json:"environment" jsonschema:"required"`
	Branch          string `json:"branch,omitempty"`
}

// DuplicateProjectOutput is the output structure for the MCP tool
type DuplicateProjectOutput struct {
	Response deploy_types.ApplicationResponse `json:"response"`
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
	Response deploy_types.ApplicationResponse `json:"response"`
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
	Response deploy_types.ApplicationResponse `json:"response"`
}
