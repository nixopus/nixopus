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