package controller

import (
	"github.com/go-fuego/fuego"

	"github.com/raghavyuva/nixopus-api/internal/types"
)

func HealthCheck(fuego.ContextNoBody) (types.Response, error) {
	return types.Response{
		Status:  "success",
		Message: "Server is up and running",
		Data:    nil,
	}, nil
}
