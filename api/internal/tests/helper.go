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
