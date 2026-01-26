package auth

import (
	"net/http"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	auth_types "github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/uptrace/bun"
)

// Account represents the account table from Better Auth schema
// This table stores authentication credentials (email/password, OAuth, etc.)
type Account struct {
	bun.BaseModel `bun:"table:account,alias:a"`
	ID             uuid.UUID `bun:"id,pk,type:uuid"`
	UserID         uuid.UUID `bun:"user_id,type:uuid,notnull"`
	AccountID      string    `bun:"account_id,type:text"`
	ProviderID     string    `bun:"provider_id,type:text,notnull"` // "credential" for email/password
	Password       *string   `bun:"password,type:text"`            // Password hash for email/password accounts
	CreatedAt      time.Time `bun:"created_at,type:timestamp,notnull"`
	UpdatedAt      time.Time `bun:"updated_at,type:timestamp,notnull"`
}

// IsAdminRegistered checks if an admin user is already registered via email/password
// With Better Auth, this checks if any users exist with email/password accounts
// by checking for accounts where the password field is not null (email/password accounts only)
func (ar *AuthController) IsAdminRegistered(s fuego.ContextNoBody) (*auth_types.AdminRegisteredResponse, error) {
	ar.logger.Log(logger.Info, "checking if admin is registered", "")

	// Check if any users exist with email/password accounts
	// Email/password accounts have a password field set (not null)
	// We check for accounts where password is not null, which indicates email/password authentication
	count, err := ar.store.DB.NewSelect().
		Model((*Account)(nil)).
		Where("password IS NOT NULL").
		Count(ar.ctx)

	if err != nil {
		ar.logger.Log(logger.Error, "failed to check admin registration", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	adminRegistered := count > 0

	return &auth_types.AdminRegisteredResponse{
		Status:  "success",
		Message: "Admin registration status retrieved successfully",
		Data: auth_types.AdminRegisteredData{
			AdminRegistered: adminRegistered,
		},
	}, nil
}
