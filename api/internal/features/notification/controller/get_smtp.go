package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type GetWithID struct {
	ID string `query:"id"`
}

func (c *NotificationController) GetSmtp(f fuego.ContextWithBody[GetWithID]) (*shared_types.Response, error) {
	params, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	SMTPConfigs, err := c.service.GetSmtp(user.ID.String(), params.ID)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "SMTP configs fetched successfully",
		Data:    SMTPConfigs,
	}, nil
}
