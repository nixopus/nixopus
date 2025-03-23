package auth

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// Register godoc
// @Summary Register a new user
// @Description Registers a new user in the application.
// @Tags auth
// @Accept json
// @Produce json
// @Param register body types.RegisterRequest true "Register request"
// @Success 200 {object} types.AuthResponse "Success response with token"
// @Failure 400 {object} types.Response "Bad request"
// @Router /auth/register [post]
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var registration_request types.RegisterRequest
	if !c.parseAndValidate(w, r, &registration_request) {
		return
	}

	userResponse, err := c.service.Register(registration_request)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.SendJSONResponse(w, "success", "User registered successfully", userResponse)
}

func (c *AuthController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var registration_request types.RegisterRequest
	if !c.parseAndValidate(w, r, &registration_request) {
		return
	}

	userResponse, err := c.service.Register(registration_request)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.SendJSONResponse(w, "success", "User created successfully", userResponse)
}
