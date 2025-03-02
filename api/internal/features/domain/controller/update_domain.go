package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *DomainsController) UpdateDomain(w http.ResponseWriter, r *http.Request) {
	var domainRequest types.UpdateDomainRequest

	err := c.validator.ParseRequestBody(r, r.Body, &domainRequest)
	if err != nil {
		c.logger.Log(logger.Error, shared_types.ErrFailedToDecodeRequest.Error(), err.Error())
		utils.SendErrorResponse(w, shared_types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	err = c.validator.ValidateRequest(domainRequest)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	userAny := r.Context().Value(shared_types.UserContextKey)
	user, ok := userAny.(*shared_types.User)

	if !ok {
		c.logger.Log(logger.Error, shared_types.ErrFailedToGetUserFromContext.Error(), shared_types.ErrFailedToGetUserFromContext.Error())
		utils.SendErrorResponse(w, shared_types.ErrFailedToGetUserFromContext.Error(), http.StatusInternalServerError)
		return
	}

	updated,err := c.service.UpdateDomain(domainRequest.Name, user.ID.String(), domainRequest.ID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Domain updated successfully", updated)
}
