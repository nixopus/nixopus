package auth

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// Example from your login.go file
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
	var login_request types.LoginRequest
	err := c.validator.ParseRequestBody(r, r.Body, &login_request)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToDecodeRequest.Error(), err.Error())
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	err = c.validator.ValidateRequest(login_request)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := c.service.Login(login_request.Email, login_request.Password)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	}

	utils.SendJSONResponse(w, "success", "User logged in successfully", response)
}
