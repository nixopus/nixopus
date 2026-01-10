package tools

import (
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// convertToMCPApplication converts shared_types.Application to MCPApplication
// removing circular references (User, Organization, Status, Logs, Deployments)
func convertToMCPApplication(app shared_types.Application) MCPApplication {
	status := ""
	if app.Status != nil {
		status = string(app.Status.Status)
	}

	familyID := (*string)(nil)
	if app.FamilyID != nil {
		id := app.FamilyID.String()
		familyID = &id
	}

	return MCPApplication{
		ID:                   app.ID.String(),
		Name:                 app.Name,
		Port:                 app.Port,
		Environment:          string(app.Environment),
		ProxyServer:          string(app.ProxyServer),
		BuildVariables:       app.BuildVariables,
		EnvironmentVariables: app.EnvironmentVariables,
		BuildPack:            string(app.BuildPack),
		Repository:           app.Repository,
		Branch:               app.Branch,
		PreRunCommand:        app.PreRunCommand,
		PostRunCommand:       app.PostRunCommand,
		Domain:               app.Domain,
		DockerfilePath:       app.DockerfilePath,
		BasePath:             app.BasePath,
		UserID:               app.UserID.String(),
		OrganizationID:       app.OrganizationID.String(),
		FamilyID:             familyID,
		CreatedAt:            app.CreatedAt,
		UpdatedAt:            app.UpdatedAt,
		Status:               status,
		Labels:               app.Labels,
	}
}

// convertToMCPApplicationDeployment converts shared_types.ApplicationDeployment to MCPApplicationDeployment
// removing circular references (Application, Status.ApplicationDeployment, Logs.ApplicationDeployment)
func convertToMCPApplicationDeployment(dep shared_types.ApplicationDeployment) MCPApplicationDeployment {
	status := ""
	if dep.Status != nil {
		status = string(dep.Status.Status)
	}

	return MCPApplicationDeployment{
		ID:              dep.ID.String(),
		ApplicationID:   dep.ApplicationID.String(),
		CreatedAt:       dep.CreatedAt,
		UpdatedAt:       dep.UpdatedAt,
		CommitHash:      dep.CommitHash,
		ContainerID:     dep.ContainerID,
		ContainerName:   dep.ContainerName,
		ContainerImage:  dep.ContainerImage,
		ContainerStatus: dep.ContainerStatus,
		Status:          status,
	}
}

// convertToMCPApplicationLogs converts shared_types.ApplicationLogs to MCPApplicationLogs
// removing circular references (ApplicationDeployment, Application)
func convertToMCPApplicationLogs(log shared_types.ApplicationLogs) MCPApplicationLogs {
	return MCPApplicationLogs{
		ID:                      log.ID.String(),
		ApplicationID:           log.ApplicationID.String(),
		CreatedAt:               log.CreatedAt,
		UpdatedAt:               log.UpdatedAt,
		Log:                     log.Log,
		ApplicationDeploymentID: log.ApplicationDeploymentID.String(),
	}
}
