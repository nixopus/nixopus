package auth

import (
	"context"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/validation"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	organization_storage "github.com/raghavyuva/nixopus-api/internal/features/organization/storage"
	permissions_storage "github.com/raghavyuva/nixopus-api/internal/features/permission/storage"
	role_storage "github.com/raghavyuva/nixopus-api/internal/features/role/storage"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

type AuthController struct {
	store        *shared_storage.Store
	validator    *validation.Validator
	service      *service.AuthService
	ctx          context.Context
	logger       logger.Logger
	notification *notification.NotificationManager
}

// NewAuthController creates a new AuthController with the given App.
//
// This function creates a new AuthController with the given App and returns a pointer to it.
//
// The App passed to this function should be a valid App that has been created with storage.NewApp.
func NewAuthController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	notificationManager *notification.NotificationManager,
) *AuthController {
	return &AuthController{
		store:     store,
		validator: validation.NewValidator(),
		service: service.NewAuthService(store, ctx, l,
			&storage.UserStorage{DB: store.DB, Ctx: ctx},
			&permissions_storage.PermissionStorage{DB: store.DB, Ctx: ctx},
			&role_storage.RoleStorage{DB: store.DB, Ctx: ctx},
			organization_storage.OrganizationStore{DB: store.DB, Ctx: ctx}),
		ctx:          ctx,
		logger:       l,
		notification: notificationManager,
	}
}

// parseAndValidate parses and validates the request body.
//
// This method attempts to parse the request body into the provided 'req' interface
// using the controller's validator. If parsing fails, an error response is sent
// and the method returns false. It also validates the parsed request object and
// returns false if validation fails. If both operations are successful, it returns true.
//
// Parameters:
//
//	w - the HTTP response writer to send error responses.
//	r - the HTTP request containing the body to parse.
//	req - the interface to populate with the parsed request body.
//
// Returns:
//
//	bool - true if parsing and validation succeed, false otherwise.
func (c *AuthController) parseAndValidate(w http.ResponseWriter, r *http.Request, req interface{}) bool {
	if err := c.validator.ParseRequestBody(r, r.Body, req); err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToDecodeRequest.Error(), err.Error())
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return false
	}

	if err := c.validator.ValidateRequest(req); err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}

// GetUser retrieves the current user from the request context.
//
// This method extracts the user from the request context using the UserContextKey.
// If the user cannot be retrieved or is of the wrong type, an error is logged,
// an error response is sent to the client, and the method returns nil.
//
// Parameters:
//
//	w - the HTTP response writer used to send error responses.
//	r - the HTTP request containing the context from which to retrieve the user.
//
// Returns:
//
//	*shared_types.User - a pointer to the User object retrieved from the context,
//	or nil if the user could not be retrieved.
func (c *AuthController) GetUser(w http.ResponseWriter, r *http.Request) *shared_types.User {
	userAny := r.Context().Value(shared_types.UserContextKey)
	user, ok := userAny.(*shared_types.User)

	if !ok {
		c.logger.Log(logger.Error, types.ErrFailedToGetUserFromContext.Error(), types.ErrFailedToGetUserFromContext.Error())
		utils.SendErrorResponse(w, types.ErrFailedToGetUserFromContext.Error(), http.StatusInternalServerError)
		return nil
	}

	return user
}

// Notify sends a notification to the user for the given payload type.
//
// This method constructs a new NotificationPayload object with the given user and request data,
// and sends it to the notification manager.
func (c *AuthController) Notify(payloadType notification.NotificationPayloadType, user *shared_types.User, r *http.Request) {
	c.notification.SendNotification(notification.NewNotificationPayload(
		payloadType,
		user.ID.String(),
		notification.NotificationAuthenticationData{
			Email: user.Email,
			NotificationBaseData: notification.NotificationBaseData{
				IP:      r.RemoteAddr,
				Browser: r.UserAgent(),
			},
			UserName: user.Username,
		},
		notification.NotificationCategoryAuthentication,
	))
}
