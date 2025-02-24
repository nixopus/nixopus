package auth

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// Register handles HTTP requests to register a new user.
//
// It expects a JSON body of type types.RegisterRequest containing the user's
// username, email, and password.
//
// If the request body cannot be decoded, it responds with a 400 error.
// If any of the fields are empty, it responds with a 400 error.
// If the password is invalid, it responds with a 400 error.
// If the password hashing fails, it responds with a 500 error.
// If a user with the provided email already exists, it responds with a 400 error.
// If the user cannot be registered, it responds with a 500 error.
// If a token cannot be created, it responds with a 500 error.
//
// On successful registration, it responds with a 200 status code and a JSON
// response containing the authentication token and user information.
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var registration_request types.RegisterRequest

	err := c.validator.ParseRequestBody(r, r.Body, &registration_request)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToDecodeRequest.Error(), err.Error())
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	err = c.validator.ValidateRequest(registration_request)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	userResponse, err := c.service.Register(registration_request)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.SendJSONResponse(w, "success", "User registered successfully", userResponse)
}
