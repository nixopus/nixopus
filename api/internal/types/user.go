package types

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel     `bun:"table:users,alias:u" swaggerignore:"true"`
	ID                uuid.UUID            `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	SupertokensUserID string               `json:"supertokens_user_id" bun:"supertokens_user_id,type:text,unique"`
	Username          string               `json:"username" bun:"username,type:text,notnull"`
	Email             string               `json:"email" bun:"email,type:text,notnull,unique"`
	Password          string               `json:"-" bun:"password,type:text"` // Optional for SuperTokens users
	Avatar            string               `json:"avatar" bun:"avatar,type:text"`
	CreatedAt         time.Time            `json:"created_at" bun:"created_at,type:timestamp,notnull,default:now()"`
	UpdatedAt         time.Time            `json:"updated_at" bun:"updated_at,type:timestamp,notnull,default:now()"`
	DeletedAt         *time.Time           `json:"deleted_at,omitempty" bun:"deleted_at"`
	IsVerified        bool                 `json:"is_verified" bun:"is_verified,type:boolean,notnull,default:false"`
	ResetToken        string               `json:"-" bun:"reset_token"`
	Type              string               `json:"type" bun:"type,type:text,notnull"`
	TwoFactorEnabled  bool                 `json:"two_factor_enabled" bun:"two_factor_enabled,type:boolean,notnull,default:false"`
	TwoFactorSecret   string               `json:"two_factor_secret" bun:"two_factor_secret,type:text"`
	Organizations     []*Organization      `json:"organizations,omitempty" bun:"m2m:organization_users,join:User=Organization"`
	OrganizationUsers []*OrganizationUsers `json:"organization_users,omitempty" bun:"m2m:organization_users,join:User=Organization"`
}

type RefreshToken struct {
	ID        uuid.UUID  `json:"id" bson:"_id"`
	UserID    uuid.UUID  `json:"user_id" bson:"user_id"`
	Token     string     `json:"token" bson:"token"`
	ExpiresAt time.Time  `json:"expires_at" bson:"expires_at"`
	CreatedAt time.Time  `json:"created_at" bson:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" bson:"revoked_at,omitempty"`
}

// NewUser returns a new User with default values set. If the provided User has empty strings for Role, CreatedAt, UpdatedAt, DeletedAt, or IsVerified, the corresponding fields in the returned User will be set with default values.
func (u User) NewUser() User {
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = time.Now()
	}

	return User{
		ID:         uuid.New(),
		Username:   u.Username,
		Email:      u.Email,
		Password:   u.Password,
		Avatar:     u.Avatar,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
		DeletedAt:  u.DeletedAt,
		IsVerified: u.IsVerified,
	}
}

func NewUser(email string, password string, username string, avatar string, role string, isVerified bool) User {
	return User{
		ID:         uuid.New(),
		Username:   username,
		Email:      email,
		Password:   password,
		Avatar:     avatar,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		DeletedAt:  nil,
		IsVerified: isVerified,
		Type:       role,
	}
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
