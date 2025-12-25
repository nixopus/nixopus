package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/utils"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	organization_types "github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

// Deprecated: Use SupertokensRegister instead
func (c *AuthService) Register(registrationRequest types.RegisterRequest, userTypeype string) (types.AuthResponse, error) {
	c.logger.Log(logger.Info, "registering user", registrationRequest.Email)
	userType := registrationRequest.Type
	if userType == "" {
		userType = shared_types.RoleViewer
	}

	if userType != shared_types.RoleAdmin && userType != shared_types.RoleMember && userType != shared_types.RoleViewer {
		c.logger.Log(logger.Error, types.ErrInvalidUserType.Error(), "")
		return types.AuthResponse{}, types.ErrInvalidUserType
	}

	if dbUser, err := c.storage.FindUserByEmail(registrationRequest.Email); err == nil && dbUser.ID != uuid.Nil {
		c.logger.Log(logger.Error, types.ErrUserWithEmailAlreadyExists.Error(), "")
		return types.AuthResponse{}, types.ErrUserWithEmailAlreadyExists
	}

	if dbUser, err := c.storage.FindUserByUsername(registrationRequest.Username); err == nil && dbUser.ID != uuid.Nil {
		c.logger.Log(logger.Error, types.ErrUserWithUsernameAlreadyExists.Error(), "")
		return types.AuthResponse{}, types.ErrUserWithUsernameAlreadyExists
	}

	hashedPassword, err := utils.HashPassword(registrationRequest.Password)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToHashPassword.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToHashPassword
	}

	user := shared_types.NewUser(
		registrationRequest.Email,
		hashedPassword,
		registrationRequest.Username,
		"",
		userType,
		false,
	)

	tx, err := c.storage.BeginTx()
	if err != nil {
		c.logger.Log(logger.Error, "failed to start transaction", err.Error())
		return types.AuthResponse{}, types.ErrFailedToRegisterUser
	}
	defer tx.Rollback()

	txStorage := c.storage.WithTx(tx)

	if err := txStorage.CreateUser(&user); err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToRegisterUser.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToRegisterUser
	}

	refreshToken, err := txStorage.CreateRefreshToken(user.ID)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToCreateRefreshToken.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToCreateToken
	}

	accessToken, err := utils.CreateToken(user.Email, time.Minute*15, user.TwoFactorEnabled, true)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToCreateAccessToken.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToCreateToken
	}

	// If the user is an admin, create a default organization and add the user to it,
	// else add the user to the requested organization
	// this makes sure that newly registered admin users always have an organization to work with
	// and if the user is not an admin, add them to the requested organization so that they can start working on their projects without having to create an organization first
	if userType == shared_types.RoleAdmin {
		organization, err := c.createDefaultOrganization(user, tx)
		if err != nil {
			c.logger.Log(logger.Error, types.ErrFailedToCreateDefaultOrganization.Error(), err.Error())
			return types.AuthResponse{}, types.ErrFailedToCreateDefaultOrganization
		}

		if err := c.addUserToOrganizationWithRole(user, organization, "admin", tx); err != nil {
			c.logger.Log(logger.Error, types.ErrFailedToAddUserToOrganization.Error(), err.Error())
			return types.AuthResponse{}, types.ErrFailedToAddUserToOrganization
		}
	}

	// if registrationRequest.Organization != "" {
	// 	requestedOrganization, err := c.organization_service.GetOrganization(registrationRequest.Organization)
	// 	if err != nil {
	// 		c.logger.Log(logger.Error, types.ErrFailedToGetOrganization.Error(), err.Error())
	// 		return types.AuthResponse{}, types.ErrFailedToGetOrganization
	// 	}

	// 	if err := c.addUserToOrganizationWithRole(user, requestedOrganization, userType, tx); err != nil {
	// 		c.logger.Log(logger.Error, types.ErrFailedToAddUserToOrganization.Error(), err.Error())
	// 		return types.AuthResponse{}, types.ErrFailedToAddUserToOrganization
	// 	}
	// }

	if err := tx.Commit(); err != nil {
		c.logger.Log(logger.Error, "failed to commit transaction", err.Error())
		return types.AuthResponse{}, types.ErrFailedToRegisterUser
	}

	return types.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
		ExpiresIn:    refreshToken.ExpiresAt.Unix(),
		User:         user,
	}, nil
}

// Deprecated: Use SupertokensRegister instead
func (c *AuthService) createDefaultOrganization(user shared_types.User, tx bun.Tx) (shared_types.Organization, error) {
	c.logger.Log(logger.Info, "creating default organization for user", user.Email)

	orgRequest := &organization_types.CreateOrganizationRequest{
		Name:        user.Username + "'s Team",
		Description: "My Team",
	}

	org, err := c.organization_service.CreateOrganization(orgRequest, tx)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToCreateDefaultOrganization.Error(), err.Error())
		return shared_types.Organization{}, types.ErrFailedToCreateDefaultOrganization
	}

	c.logger.Log(logger.Info, "created default organization for user", user.Email)
	return org, nil
}

func (c *AuthService) addUserToOrganizationWithRole(user shared_types.User, organization shared_types.Organization, roleName string, tx bun.Tx) error {
	c.logger.Log(logger.Info, "adding user to organization with role", roleName)

	userOrganization := organization_types.AddUserToOrganizationRequest{
		OrganizationID: organization.ID.String(),
		UserID:         user.ID.String(),
	}

	return c.organization_service.AddUserToOrganization(userOrganization, tx)
}
