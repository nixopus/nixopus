package controller

import (
	"github.com/go-fuego/fuego"
)

// HealthCheckResponse is the typed response for health check
type HealthCheckResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func HealthCheck(fuego.ContextNoBody) (*HealthCheckResponse, error) {
	return &HealthCheckResponse{
		Status:  "success",
		Message: "Server is up and running",
	}, nil
}
