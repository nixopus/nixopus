package types

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// User represents the user table from Better Auth schema (auth_schema.ts)
// Table: "user" with columns: id, name, email, email_verified, image, created_at, updated_at
type User struct {
	bun.BaseModel     `bun:"table:user,alias:u" swaggerignore:"true"`
	ID                uuid.UUID            `json:"id" bun:"id,pk,type:uuid"`
	Name              string               `json:"name" bun:"name,type:text,notnull"`
	Email             string               `json:"email" bun:"email,type:text,notnull"`
	EmailVerified     bool                 `json:"email_verified" bun:"email_verified,type:boolean,notnull,default:false"`
	Image             *string              `json:"image,omitempty" bun:"image,type:text"`
	IsOnboarded       bool                 `json:"is_onboarded" bun:"is_onboarded,type:boolean,notnull,default:false"`
	CreatedAt         time.Time            `json:"created_at" bun:"created_at,type:timestamp,notnull"`
	UpdatedAt         time.Time            `json:"updated_at" bun:"updated_at,type:timestamp,notnull"`
	Organizations     []*Organization      `json:"organizations,omitempty" bun:"m2m:organization_users,join:User=Organization"`
	OrganizationUsers []*OrganizationUsers `json:"organization_users,omitempty" bun:"m2m:organization_users,join:User=Organization"`

	// Backward compatibility fields (computed, not persisted)
	Username          string `json:"username" bun:"-"`            // Computed from Name
	Avatar            string `json:"avatar" bun:"-"`              // Computed from Image
	IsVerified        bool   `json:"is_verified" bun:"-"`         // Computed from EmailVerified
	SupertokensUserID string `json:"supertokens_user_id" bun:"-"` // Computed from ID
}

type RefreshToken struct {
	ID        uuid.UUID  `json:"id" bson:"_id"`
	UserID    uuid.UUID  `json:"user_id" bson:"user_id"`
	Token     string     `json:"token" bson:"token"`
	ExpiresAt time.Time  `json:"expires_at" bson:"expires_at"`
	CreatedAt time.Time  `json:"created_at" bson:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" bson:"revoked_at,omitempty"`
}

// ComputeCompatibilityFields computes backward compatibility fields from Better Auth fields
func (u *User) ComputeCompatibilityFields() {
	u.Username = u.Name
	if u.Username == "" {
		// Derive username from email if name is empty
		if atIndex := strings.Index(u.Email, "@"); atIndex > 0 {
			u.Username = u.Email[:atIndex]
		} else {
			u.Username = u.Email
		}
	}

	if u.Image != nil {
		u.Avatar = *u.Image
	}
	u.IsVerified = u.EmailVerified
	u.SupertokensUserID = u.ID.String()
}

// NewUser returns a new User with default values set (for backward compatibility)
func (u User) NewUser() User {
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = time.Now()
	}

	user := User{
		ID:            uuid.New(),
		Name:          u.Name,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		Image:         u.Image,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
	user.ComputeCompatibilityFields()
	return user
}

func NewUser(email string, password string, username string, avatar string, role string, isVerified bool) User {
	var imagePtr *string
	if avatar != "" {
		imagePtr = &avatar
	}
	user := User{
		ID:            uuid.New(),
		Name:          username,
		Email:         email,
		EmailVerified: isVerified,
		Image:         imagePtr,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	user.ComputeCompatibilityFields()
	return user
}

var (
	ErrFailedToDecodeRequest              = errors.New("failed to decode request")
	ErrFailedToGetUserFromContext         = errors.New("failed to get user from context")
	ErrFailedToGetOrganizationFromContext = errors.New("failed to get organization from context")
	ErrUserDoesNotBelongToOrganization    = errors.New("user does not belong to organization")
	ErrNoRoleAssigned                     = errors.New("no role assigned")
)

const (
	UserTypeAdmin = "admin"
	UserTypeUser  = "app_user"
)
