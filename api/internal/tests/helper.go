package tests

import "fmt"

var baseURL = "http://localhost:8080/api/v1"

func GetHealthURL() string {
	return baseURL + "/health"
}

func GetRegisterURL() string {
	return baseURL + "/auth/register"
}

func GetLoginURL() string {
	return baseURL + "/auth/login"
}

func GetRefreshTokenURL() string {
	return baseURL + "/auth/refresh-token"
}

func GetRequestPasswordResetURL() string {
	return baseURL + "/auth/request-password-reset"
}

func GetResetPasswordURL() string {
	return baseURL + "/auth/reset-password"
}

func GetCreateUserURL() string {
	return baseURL + "/auth/create-user"
}

func GetSendVerificationEmailURL() string {
	return baseURL + "/auth/send-verification-email"
}

func GetSetup2FAURL() string {
	return baseURL + "/auth/setup-2fa"
}

func GetVerify2FAURL() string {
	return baseURL + "/auth/verify-2fa"
}

func GetDisable2FAURL() string {
	return baseURL + "/auth/disable-2fa"
}

func Get2FALoginURL() string {
	return baseURL + "/auth/2fa-login"
}

func GetVerifyEmailURL() string {
	return baseURL + "/auth/verify-email"
}

func GetLogoutURL() string {
	return baseURL + "/auth/logout"
}

func GetUserDetailsURL() string {
	return baseURL + "/user"
}

func GetIsAdminRegisteredURL() string {
	return baseURL + "/auth/is-admin-registered"
}

func GetContainersURL() string {
	return baseURL + "/container"
}

func GetContainerURL(containerID string) string {
	return baseURL + "/container/" + containerID
}

func GetContainerLogsURL(containerID string) string {
	return baseURL + "/container/" + containerID + "/logs"
}

// JQFirstContainerIDFromList resolves a container id from ListContainers grouped JSON (.groups / .ungrouped).
func JQFirstContainerIDFromList() string {
	return `(.data.ungrouped[0].id // .data.groups[0].containers[0].id)`
}

// JQContainerIDNamed builds jq that finds a container id by exact name among ungrouped and grouped containers.
func JQContainerIDNamed(containerName string) string {
	return fmt.Sprintf(
		`[.data.ungrouped[]?, (.data.groups // [])[].containers[]?] | map(select(.name == %q)) | .[0].id`,
		containerName,
	)
}

func GetDomainURL() string {
	return baseURL + "/domain"
}

func GetDomainsURL() string {
	return baseURL + "/domains"
}

func GetDomainGenerateURL() string {
	return baseURL + "/domain/generate"
}

func GetFeatureFlagsURL() string {
	return baseURL + "/feature-flags"
}

func GetFeatureFlagCheckURL() string {
	return baseURL + "/feature-flags/check"
}

func GetDeployApplicationURL() string {
	return baseURL + "/deploy/application"
}

func GetDeployApplicationsURL() string {
	return baseURL + "/deploy/applications"
}

func GetDeployApplicationRedeployURL() string {
	return baseURL + "/deploy/application/redeploy"
}

func GetDeployApplicationRestartURL() string {
	return baseURL + "/deploy/application/restart"
}

func GetDeployApplicationRollbackURL() string {
	return baseURL + "/deploy/application/rollback"
}

func GetDeployApplicationDeploymentsURL() string {
	return baseURL + "/deploy/application/deployments"
}

func GetDeployApplicationDeploymentByIDURL(deploymentID string) string {
	return baseURL + "/deploy/application/deployments/" + deploymentID
}

func GetDeployApplicationDeploymentLogsURL(deploymentID string) string {
	return baseURL + "/deploy/application/deployments/" + deploymentID + "/logs"
}

func GetDeployApplicationLogsURL(applicationID string) string {
	return baseURL + "/deploy/application/logs/" + applicationID
}

func GetDeployApplicationCancelURL() string {
	return baseURL + "/deploy/application/cancel-deployment"
}

func GetCreateOrganizationURL() string {
	return baseURL + "/organizations"
}

// Machine routes
func GetMachinesURL() string {
	return baseURL + "/machines"
}

func GetMachineByIDURL(id string) string {
	return baseURL + "/machines/" + id
}

func GetMachineSetDefaultURL(id string) string {
	return baseURL + "/machines/" + id + "/set-default"
}

func GetMachineSSHStatusURL() string {
	return baseURL + "/machines/ssh/status"
}

func GetMachineVerifyURL(id string) string {
	return baseURL + "/machines/" + id + "/verify"
}

func GetMachineRenameURL(id string) string {
	return baseURL + "/machines/" + id + "/rename"
}

func GetMachineSSHKeyStatusURL(id string) string {
	return baseURL + "/machines/" + id + "/ssh/status"
}

func GetMachineStatusURL(serverID string) string {
	return baseURL + "/machines/status?server_id=" + serverID
}

func GetMachineRestartURL(serverID string) string {
	return baseURL + "/machines/restart?server_id=" + serverID
}

func GetMachinePauseURL(serverID string) string {
	return baseURL + "/machines/pause?server_id=" + serverID
}

func GetMachineResumeURL(serverID string) string {
	return baseURL + "/machines/resume?server_id=" + serverID
}

func GetMachineBackupsURL() string {
	return baseURL + "/machines/backups"
}

func GetMachineTriggerBackupURL(serverID string) string {
	return baseURL + "/machines/backup?server_id=" + serverID
}

func GetMachineBackupScheduleURL() string {
	return baseURL + "/machines/backup/schedule"
}

func GetMachineStatsURL() string {
	return baseURL + "/machines/stats"
}

// Notification
func GetNotificationSMTPURL() string {
	return baseURL + "/notification/smtp"
}

func GetNotificationPreferencesURL() string {
	return baseURL + "/notification/preferences"
}

func GetNotificationWebhookURL(webhookType string) string {
	return baseURL + "/notification/webhook/" + webhookType
}

func GetNotificationWebhookBaseURL() string {
	return baseURL + "/notification/webhook"
}

func GetNotificationSendURL() string {
	return baseURL + "/notification/send"
}

// Domain (custom/verify/dns-check)
func GetDomainCustomURL() string {
	return baseURL + "/domain/custom"
}

func GetDomainVerifyURL() string {
	return baseURL + "/domain/verify"
}

func GetDomainDNSCheckURL(id string) string {
	return baseURL + "/domain/dns-check?id=" + id
}

// Extensions
func GetExtensionsURL() string {
	return baseURL + "/extensions"
}

func GetExtensionCategoriesURL() string {
	return baseURL + "/extensions/categories"
}

func GetExtensionByIDURL(id string) string {
	return baseURL + "/extensions/" + id
}

func GetExtensionByExtensionIDURL(extensionID string) string {
	return baseURL + "/extensions/by-extension-id/" + extensionID
}

// Healthcheck
func GetHealthCheckURL() string {
	return baseURL + "/healthcheck"
}

func GetHealthCheckToggleURL() string {
	return baseURL + "/healthcheck/toggle"
}

func GetHealthCheckResultsURL() string {
	return baseURL + "/healthcheck/results"
}

func GetHealthCheckStatsURL() string {
	return baseURL + "/healthcheck/stats"
}

// User settings
func GetUserNameURL() string {
	return baseURL + "/user/name"
}

func GetUserSettingsURL() string {
	return baseURL + "/user/settings"
}

func GetUserSettingsFontURL() string {
	return baseURL + "/user/settings/font"
}

func GetUserSettingsThemeURL() string {
	return baseURL + "/user/settings/theme"
}

func GetUserSettingsLanguageURL() string {
	return baseURL + "/user/settings/language"
}

func GetUserSettingsAutoUpdateURL() string {
	return baseURL + "/user/settings/auto-update"
}

func GetUserAvatarURL() string {
	return baseURL + "/user/avatar"
}

func GetUserPreferencesURL() string {
	return baseURL + "/user/preferences"
}

func GetUserOnboardedURL() string {
	return baseURL + "/user/onboarded"
}

// Update
func GetUpdateCheckURL() string {
	return baseURL + "/update/check"
}

func GetUpdateURL() string {
	return baseURL + "/update"
}

// MCP
func GetMCPCatalogURL() string {
	return baseURL + "/mcp/catalog"
}

func GetMCPServersURL() string {
	return baseURL + "/mcp/servers"
}

func GetMCPServerURL(id string) string {
	return baseURL + "/mcp/servers/" + id
}

func GetMCPServerTestURL() string {
	return baseURL + "/mcp/servers/test"
}

// Machine billing / metrics
func GetMachinePlansURL() string {
	return baseURL + "/machines/plans"
}

func GetMachinePlanSelectURL() string {
	return baseURL + "/machines/plan/select"
}

func GetMachineBillingStatusURL() string {
	return baseURL + "/machines/billing"
}

func GetMachineMetricsURL() string {
	return baseURL + "/machines/metrics"
}

func GetMachineEventsURL() string {
	return baseURL + "/machines/events"
}

func GetMachineMetricsSummaryURL() string {
	return baseURL + "/machines/metrics/summary"
}

// Trial machine
func GetMachineTrialProvisionURL() string {
	return baseURL + "/machines/trial/provision"
}

func GetMachineTrialStatusURL(sessionID string) string {
	return baseURL + "/machines/trial/status/" + sessionID
}

func GetTrailUpgradeResourcesURL() string {
	return baseURL + "/trail/upgrade-resources"
}

// Container ops
func GetContainerStartURL(containerID string) string {
	return baseURL + "/container/" + containerID + "/start"
}

func GetContainerStopURL(containerID string) string {
	return baseURL + "/container/" + containerID + "/stop"
}

func GetContainerRestartURL(containerID string) string {
	return baseURL + "/container/" + containerID + "/restart"
}

func GetContainerResourcesURL(containerID string) string {
	return baseURL + "/container/" + containerID + "/resources"
}

func GetContainerImagesURL() string {
	return baseURL + "/container/images"
}

func GetContainerPruneImagesURL() string {
	return baseURL + "/container/prune/images"
}

func GetContainerPruneBuildCacheURL() string {
	return baseURL + "/container/prune/build-cache"
}

// Artifacts
func GetDeployArtifactsURL(applicationID string) string {
	return baseURL + "/deploy/artifacts?application_id=" + applicationID
}

func GetDeployArtifactDownloadURL(deploymentID string) string {
	return baseURL + "/deploy/artifacts/" + deploymentID + "/download"
}

func GetDeployArtifactDeleteURL(deploymentID string) string {
	return baseURL + "/deploy/artifacts/" + deploymentID
}

// Audit
func GetAuditURL() string {
	return baseURL + "/audit"
}

// Auth bootstrap
func GetAuthBootstrapURL() string {
	return baseURL + "/auth/bootstrap"
}

// GitHub connector
func GetGithubConnectorURL() string {
	return baseURL + "/api/v1/github-connector"
}

func GetGithubConnectorsURL() string {
	return baseURL + "/api/v1/github-connector/all"
}

func GetGithubRepositoriesURL() string {
	return baseURL + "/api/v1/github-connector/repositories"
}

func GetGithubRepositoryBranchesURL() string {
	return baseURL + "/api/v1/github-connector/repository/branches"
}
