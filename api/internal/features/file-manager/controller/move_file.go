package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/file-manager/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

type MoveDirectory struct {
	FromPath string `json:"from_path"`
	ToPath   string `json:"to_path"`
}

func (c *FileManagerController) MoveDirectory(f fuego.ContextWithBody[MoveDirectory]) (*types.MessageResponse, error) {
	request, err := f.Body()

	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	err = c.service.MoveDirectory(request.FromPath, request.ToPath)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.MessageResponse{
		Status:  "success",
		Message: "Directory moved successfully",
	}, nil
}
