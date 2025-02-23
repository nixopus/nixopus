package types

import (
	"errors"
	
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// This struct is used for both login and register responses
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	User         shared_types.User   `json:"user"`
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
	ErrRefreshTokenIsRequired                  = errors.New("refresh token is required")
	ErrInvalidRefreshToken                     = errors.New("invalid refresh token")
	ErrPermissionAlreadyExists                 = errors.New("permission already exists")
	ErrPermissionDoesNotExist                  = errors.New("permission does not exist")
	ErrUserNameContainsSpaces                  = errors.New("user name cannot contain spaces")
	ErrUserNameTooLong                         = errors.New("user name is too long")
	ErrInvalidEmail                            = errors.New("invalid email")
	ErrInvalidRequestType                      = errors.New("invalid request type")
	ErrFailedToCreateAccessToken               = errors.New("failed to create access token")
	ErrMissingRefreshToken                     = errors.New("refresh token is required")
)
