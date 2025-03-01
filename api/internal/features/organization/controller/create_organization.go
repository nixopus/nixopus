package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// CreateOrganization godoc
// @Summary Create a new organization
// @Description Creates a new organization in the application.
// @Tags organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param create_organization body types.CreateOrganizationRequest true "Create organization request"
// @Success 200 {object} types.Response "Success response with organization"
// @Failure 400 {object} types.Response "Bad request"
// @Router /organization/create [post]
func (c *OrganizationsController) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var organization types.CreateOrganizationRequest

	userAny := r.Context().Value(shared_types.UserContextKey)
	loggedInUser, ok := userAny.(*shared_types.User)

	if !ok {
		utils.SendErrorResponse(w, shared_types.ErrFailedToGetUserFromContext.Error(), http.StatusInternalServerError)
		return
	}

	if err := c.validator.ParseRequestBody(r, r.Body, &organization); err != nil {
		c.logger.Log(logger.Error, shared_types.ErrFailedToDecodeRequest.Error(), err.Error())
		utils.SendErrorResponse(w, shared_types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validator.ValidateRequest(organization); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdOrganization, err := c.service.CreateOrganization(&organization)

	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.notification.SendNotification(notification.NewNotificationPayload(
		notification.NortificationPayloadTypeCreateOrganization,
		loggedInUser.ID.String(),
		notification.NotificationOrganizationData{
			NotificationBaseData: notification.NotificationBaseData{
				IP:      r.RemoteAddr,
				Browser: r.UserAgent(),
			},
			OrganizationID: organization.Name,
		},
		notification.NotificationCategoryOrganization,
	))

	utils.SendJSONResponse(w, "success", "Organization created successfully", createdOrganization)
}
