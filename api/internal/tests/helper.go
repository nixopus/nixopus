package tests

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
