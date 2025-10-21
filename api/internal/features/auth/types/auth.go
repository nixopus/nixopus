package types

import (
	"errors"
	"time"

	"github.com/google/uuid"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type AuthResponse struct {
	AccessToken  string            `json:"access_token"`
	RefreshToken string            `json:"refresh_token"`
	ExpiresIn    int64             `json:"expires_in"`
	User         shared_types.User `json:"user"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Type         string `json:"type"`
	Organization string `json:"organization"`
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

type ResetPasswordRequest struct {
	Password string `json:"password"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type VerificationToken struct {
	ID        uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	UserID    uuid.UUID `bun:"user_id,type:uuid,notnull"`
	Token     string    `bun:"token,type:text,notnull,unique"`
	ExpiresAt time.Time `bun:"expires_at,type:timestamp,notnull"`
	CreatedAt time.Time `bun:"created_at,type:timestamp,notnull,default:now()"`
}

type TwoFactorSetupResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qr_code"`
}

type TwoFactorVerifyRequest struct {
	Code string `json:"code"`
}

type TwoFactorLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

var (
	ErrInvalidUser                             = errors.New("invalid user")
	ErrEmptyPassword                           = errors.New("password cannot be empty")
	ErrPasswordMustHaveAtLeast8Chars           = errors.New("password must have at least 8 characters")
	ErrPasswordMustHaveAtLeast1Number          = errors.New("password must have at least 1 number")
	ErrPasswordMustHaveAtLeast1SpecialChar     = errors.New("password must have at least 1 special character")
	ErrPasswordMustHaveAtLeast1UppercaseLetter = errors.New("password must have at least 1 uppercase letter")
	ErrPasswordMustHaveAtLeast1LowercaseLetter = errors.New("password must have at least 1 lowercase letter")
	ErrFailedToDecodeRequest                   = errors.New("failed to decode request body")
	ErrMissingRequiredFields                   = errors.New("missing required fields")
	ErrUserWithEmailAlreadyExists              = errors.New("user with email already exists")
	ErrUserWithUsernameAlreadyExists           = errors.New("user with username already exists")
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
	ErrRefreshTokenIsRequired                  = errors.New("refresh token is required")
	ErrInvalidRefreshToken                     = errors.New("invalid refresh token")
	ErrRefreshTokenAlreadyRevoked              = errors.New("refresh token is already revoked")
	ErrPermissionAlreadyExists                 = errors.New("permission already exists")
	ErrPermissionDoesNotExist                  = errors.New("permission does not exist")
	ErrUserNameContainsSpaces                  = errors.New("user name cannot contain spaces")
	ErrUserNameTooLong                         = errors.New("user name is too long")
	ErrInvalidEmail                            = errors.New("invalid email")
	ErrInvalidRequestType                      = errors.New("invalid request type")
	ErrFailedToCreateAccessToken               = errors.New("failed to create access token")
	ErrMissingRefreshToken                     = errors.New("refresh token is required")
	ErrInvalidUserType                         = errors.New("invalid user type")
	ErrFailedToCreateDefaultOrganization       = errors.New("failed to create default organization")
	ErrFailedToCreateDefaultPermissions        = errors.New("failed to create default permissions")
	ErrNoOrganizationsFound                    = errors.New("no organizations found")
	ErrFailedToAddUserToOrganization           = errors.New("failed to add user to organization")
	ErrFailedToGetOrganization                 = errors.New("failed to get organization")
	ErrInvalidAccess                           = errors.New("invalid access")
	ErrFailedToSetup2FA                        = errors.New("failed to setup two-factor authentication")
	ErrFailedToEnable2FA                       = errors.New("failed to enable two-factor authentication")
	ErrFailedToDisable2FA                      = errors.New("failed to disable two-factor authentication")
	ErrInvalid2FACode                          = errors.New("invalid two-factor authentication code")
)
