package auth

import (
	"net/http"

	"github.com/go-fuego/fuego"
	auth_types "github.com/nixopus/nixopus/api/internal/features/auth/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

// IsAdminRegistered checks if an admin user is already registered via email/password.
// The result is cached in Redis with an asymmetric TTL:
//   - true  (admin exists)  -> 24 h  — the value is permanent once set.
//   - false (no admin yet)  -> 30 s  — re-checked quickly so first signup is detected fast.
//
// On cache errors the handler falls through to the database transparently.
func (ar *AuthController) IsAdminRegistered(s fuego.ContextNoBody) (*auth_types.AdminRegisteredResponse, error) {
	ar.logger.Log(logger.Info, "checking if admin is registered", "")

	registered, err := ar.service.GetAdminRegistered(ar.ctx)
	if err != nil {
		ar.logger.Log(logger.Error, "failed to check admin registration", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &auth_types.AdminRegisteredResponse{
		Status:  "success",
		Message: "Admin registration status retrieved successfully",
		Data:    auth_types.AdminRegisteredData{AdminRegistered: registered},
	}, nil
}
