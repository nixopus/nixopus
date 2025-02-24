package auth

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// Login handles HTTP requests to authenticate a user and provide a token.
//
// It expects a JSON body of type types.LoginRequest containing the user's email and password.
//
// If the request body cannot be decoded, it responds with a 400 error.
// If the email or password is missing, it responds with a 400 error.
// If the user is not found, it responds with a 404 error.
// If the password is incorrect, it responds with a 401 error.
// If a token cannot be created, it responds with a 500 error.
//
// On successful authentication, it responds with a 200 status code and a JSON response
// containing the authentication token and user information.
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
