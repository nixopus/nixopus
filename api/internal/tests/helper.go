package tests

var baseURL = "http://localhost:8080/api/v1"

func GetHealthURL() string {
	return baseURL + "/health"
}
