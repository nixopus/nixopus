package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification"
	"github.com/nixopus/nixopus/api/internal/features/notification/controller/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *NotificationController) UpdatePreference(f fuego.ContextWithBody[notification.UpdatePreferenceRequest]) (*types.MessageResponse, error) {
	prefRequest, err := f.Body()

	if err != nil {
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	w, r := f.Response(), f.Request()

	jsonData, err := json.Marshal(prefRequest)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	r.Body = io.NopCloser(bytes.NewBuffer(jsonData))

	if !c.parseAndValidate(w, r, &prefRequest) {
		return nil, fuego.BadRequestError{
			Detail: "validation failed",
		}
	}

	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	err = c.service.UpdatePreference(prefRequest, user.ID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.MessageResponse{
		Status:  "success",
		Message: "Preference updated successfully",
	}, nil
}
