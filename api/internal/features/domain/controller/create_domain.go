package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *DomainsController) CreateDomain(f fuego.ContextWithBody[types.CreateDomainRequest]) (*shared_types.Response, error) {
	domainRequest, err := f.Body()

	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	w, r := f.Response(), f.Request()
	if !c.parseAndValidate(w, r, &domainRequest) {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	created, err := c.service.CreateDomain(domainRequest, user.ID.String())

	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")

		if isInvalidDomainError(err) {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusBadRequest,
			}
		}

		if err == types.ErrDomainAlreadyExists {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusConflict,
			}
		}

		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Domain created successfully",
		Data:    created,
	}, nil
}
