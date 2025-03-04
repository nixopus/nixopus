package auth

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// Login godoc
// @Summary User login endpoint
// @Description Authenticates a user and returns a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param login body types.LoginRequest true "Login credentials"
// @Success 200 {object} types.Response{data=types.AuthResponse} "Success response with token"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 401 {object} types.Response "Unauthorized"
// @Router /auth/login [post]
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest types.LoginRequest
	if !c.parseAndValidate(w, r, &loginRequest) {
		return
	}

	response, err := c.service.Login(loginRequest.Email, loginRequest.Password)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	}

	c.Notify(notification.NotificationPayloadTypeLogin, &response.User, r)

	utils.SendJSONResponse(w, "success", "User logged in successfully", response)
}
