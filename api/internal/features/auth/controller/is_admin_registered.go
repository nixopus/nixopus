package auth

import (
	"net/http"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	auth_types "github.com/nixopus/nixopus/api/internal/features/auth/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/uptrace/bun"
)

// Account represents the account table from Better Auth schema
// This table stores authentication credentials (email/password, OAuth, etc.)
type Account struct {
	bun.BaseModel `bun:"table:account,alias:a"`
	ID            uuid.UUID `bun:"id,pk,type:uuid"`
	UserID        uuid.UUID `bun:"user_id,type:uuid,notnull"`
	AccountID     string    `bun:"account_id,type:text"`
	ProviderID    string    `bun:"provider_id,type:text,notnull"` // "credential" for email/password
	Password      *string   `bun:"password,type:text"`            // Password hash for email/password accounts
	CreatedAt     time.Time `bun:"created_at,type:timestamp,notnull"`
	UpdatedAt     time.Time `bun:"updated_at,type:timestamp,notnull"`
}

// IsAdminRegistered checks if an admin user is already registered via email/password.
// The result is cached in Redis with an asymmetric TTL:
//   - true  (admin exists)  -> 24 h  — the value is permanent once set.
//   - false (no admin yet)  -> 30 s  — re-checked quickly so first signup is detected fast.
//
// On cache errors the handler falls through to the database transparently.
func (ar *AuthController) IsAdminRegistered(s fuego.ContextNoBody) (*auth_types.AdminRegisteredResponse, error) {
	ar.logger.Log(logger.Info, "checking if admin is registered", "")

	if ar.cache != nil {
		registered, hit, err := ar.cache.GetAdminRegistered(ar.ctx)
		if err != nil {
			ar.logger.Log(logger.Error, "cache read failed, falling through to db", err.Error())
		}
		if hit && err == nil {
			return &auth_types.AdminRegisteredResponse{
				Status:  "success",
				Message: "Admin registration status retrieved successfully",
				Data:    auth_types.AdminRegisteredData{AdminRegistered: registered},
			}, nil
		}
	}

	count, err := ar.store.DB.NewSelect().
		Model((*Account)(nil)).
		Where("password IS NOT NULL").
		Count(ar.ctx)

	if err != nil {
		ar.logger.Log(logger.Error, "failed to check admin registration", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	adminRegistered := count > 0

	if ar.cache != nil {
		if cacheErr := ar.cache.SetAdminRegistered(ar.ctx, adminRegistered); cacheErr != nil {
			ar.logger.Log(logger.Error, "failed to cache admin registration status", cacheErr.Error())
		}
	}

	return &auth_types.AdminRegisteredResponse{
		Status:  "success",
		Message: "Admin registration status retrieved successfully",
		Data:    auth_types.AdminRegisteredData{AdminRegistered: adminRegistered},
	}, nil
}
