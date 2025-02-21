package types

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`
	ID            uuid.UUID  `json:"id" bun:"id,pk,type:uuid"`
	Username      string     `json:"username" bun:"username,notnull"`
	Email         string     `json:"email" bun:"email,unique,notnull"`
	Password      string     `json:"-" bun:"password,notnull"`
	Role          string     `json:"role" bun:"role,notnull"`
	Avatar        string     `json:"avatar" bun:"avatar"`
	CreatedAt     time.Time  `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time  `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`
	IsVerified    bool       `json:"is_verified" bun:"is_verified,notnull,default:false"`
	ResetToken    string     `json:"-" bun:"reset_token"`
}

type RefreshToken struct {
	ID        uuid.UUID  `json:"id" bson:"_id"`
	UserID    uuid.UUID  `json:"user_id" bson:"user_id"`
	Token     string     `json:"token" bson:"token"`
	ExpiresAt time.Time  `json:"expires_at" bson:"expires_at"`
	CreatedAt time.Time  `json:"created_at" bson:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" bson:"revoked_at,omitempty"`
}

func (u User) SetUserName(username string) User {
	u.Username = username
	return u
}

func (u User) SetEmail(email string) User {
	u.Email = email
	return u
}

func (u User) SetPassword(password string) User {
	u.Password = password
	return u
}

func (u User) SetRole(role string) User {
	u.Role = role
	return u
}

func (u User) SetAvatar(avatar string) User {
	u.Avatar = avatar
	return u
}

func (u User) SetIsVerified(isVerified bool) User {
	u.IsVerified = isVerified
	return u
}

func (u User) IsValidPassword(password string) error {
	if password == "" {
		return ErrEmptyPassword
	}
	if len(password) < 8 {
		return ErrPasswordMustHaveAtLeast8Chars
	}
	if !containsNumber(password) {
		return ErrPasswordMustHaveAtLeast1Number
	}
	if !containsSpecialChar(password) {
		return ErrPasswordMustHaveAtLeast1SpecialChar
	}
	if !containsUppercaseLetter(password) {
		return ErrPasswordMustHaveAtLeast1UppercaseLetter
	}
	if !containsLowercaseLetter(password) {
		return ErrPasswordMustHaveAtLeast1LowercaseLetter
	}
	return nil
}

// NewUser returns a new User with default values set. If the provided User has empty strings for Role, CreatedAt, UpdatedAt, DeletedAt, or IsVerified, the corresponding fields in the returned User will be set with default values.
func (u User) NewUser() User {
	if u.Role == "" {
		u.Role = "user"
	}
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
		Role:       u.Role,
		Avatar:     u.Avatar,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
		DeletedAt:  u.DeletedAt,
		IsVerified: u.IsVerified,
	}
}

// This struct is used for both login and register responses
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	User         User   `json:"user"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
}

type DeleteUserRequest struct {
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

var (
	ErrEmptyPassword                           = errors.New("password cannot be empty")
	ErrPasswordMustHaveAtLeast8Chars           = errors.New("password must have at least 8 characters")
	ErrPasswordMustHaveAtLeast1Number          = errors.New("password must have at least 1 number")
	ErrPasswordMustHaveAtLeast1SpecialChar     = errors.New("password must have at least 1 special character")
	ErrPasswordMustHaveAtLeast1UppercaseLetter = errors.New("password must have at least 1 uppercase letter")
	ErrPasswordMustHaveAtLeast1LowercaseLetter = errors.New("password must have at least 1 lowercase letter")
	ErrFailedToDecodeRequest                   = errors.New("failed to decode request body")
	ErrMissingRequiredFields                   = errors.New("missing required fields")
	ErrUserWithEmailAlreadyExists              = errors.New("user with email already exists")
	ErrFailedToRegisterUser                    = errors.New("failed to register user")
	ErrFailedToHashPassword                    = errors.New("failed to hash password")
	ErrFailedToCreateToken                     = errors.New("failed to create token")
	ErrInvalidPassword                         = errors.New("invalid password")
	ErrUserNotFound                            = errors.New("user not found")
	ErrFailedToGetUserFromContext              = errors.New("failed to get user from context")
	ErrFailedToUpdateUser                      = errors.New("failed to update user")
	ErrSamePassword                            = errors.New("passwords must be different")
	ErrFailedToSendEmail                       = errors.New("failed to send email")
	ErrInvalidResetToken                       = errors.New("invalid reset token")
	ErrFailedToCreateRefreshToken              = errors.New("failed to create refresh token")
)

func containsNumber(password string) bool {
	for _, char := range password {
		if char >= '0' && char <= '9' {
			return true
		}
	}
	return false
}

func containsSpecialChar(password string) bool {
	for _, char := range password {
		if char >= '!' && char <= '/' || char >= ':' && char <= '@' || char >= '[' && char <= '`' || char >= '{' && char <= '~' {
			return true
		}
	}
	return false
}

func containsUppercaseLetter(password string) bool {
	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			return true
		}
	}
	return false
}

func containsLowercaseLetter(password string) bool {
	for _, char := range password {
		if char >= 'a' && char <= 'z' {
			return true
		}
	}
	return false
}
