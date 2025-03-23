package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	organization_types "github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *AuthService) Register(registrationRequest types.RegisterRequest) (types.AuthResponse, error) {
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

	hashedPassword, err := HashPassword(registrationRequest.Password)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToHashPassword.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToHashPassword
	}

	user := shared_types.NewUser(
		registrationRequest.Email,
		hashedPassword,
		registrationRequest.Username,
		"",
		"",
		false,
	)

	if err := c.storage.CreateUser(&user); err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToRegisterUser.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToRegisterUser
	}

	refreshToken, err := c.storage.CreateRefreshToken(user.ID)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToCreateRefreshToken.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToCreateToken
	}

	accessToken, err := CreateToken(user.Email, time.Minute*15)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToCreateAccessToken.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToCreateToken
	}

	organization, err := c.createDefaultOrganization(user)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToCreateDefaultOrganization.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToCreateDefaultOrganization
	}

	if err := c.addUserToOrganizationWithRole(user, organization, "admin"); err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToAddUserToOrganization.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToAddUserToOrganization
	}

	if registrationRequest.Organization != "" {
		requestedOrganization, err := c.organization_service.GetOrganization(registrationRequest.Organization)
		if err != nil {
			c.logger.Log(logger.Error, types.ErrFailedToGetOrganization.Error(), err.Error())
			return types.AuthResponse{}, types.ErrFailedToGetOrganization
		}

		if err := c.addUserToOrganizationWithRole(user, requestedOrganization, userType); err != nil {
			c.logger.Log(logger.Error, types.ErrFailedToAddUserToOrganization.Error(), err.Error())
			return types.AuthResponse{}, types.ErrFailedToAddUserToOrganization
		}
	}

	return types.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
		ExpiresIn:    refreshToken.ExpiresAt.Unix(),
		User:         user,
	}, nil
}

func (c *AuthService) createDefaultOrganization(user shared_types.User) (shared_types.Organization, error) {
	c.logger.Log(logger.Info, "creating default organization for user", user.Email)

	orgRequest := &organization_types.CreateOrganizationRequest{
		Name:        user.Username + "'s Team",
		Description: "My Team",
	}

	org, err := c.organization_service.CreateOrganization(orgRequest)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToCreateDefaultOrganization.Error(), err.Error())
		return shared_types.Organization{}, types.ErrFailedToCreateDefaultOrganization
	}

	c.logger.Log(logger.Info, "created default organization for user", user.Email)
	return org, nil
}

func (c *AuthService) addUserToOrganizationWithRole(user shared_types.User, organization shared_types.Organization, roleName string) error {
	c.logger.Log(logger.Info, "adding user to organization with role", roleName)

	roles, err := c.role_service.GetRoleByName(roleName)
	if err != nil {
		c.logger.Log(logger.Error, "failed to get role by name", err.Error())
		return err
	}

	if roles == nil {
		c.logger.Log(logger.Error, types.ErrNoRolesFound.Error(), "")
		return types.ErrNoRolesFound
	}

	userOrganization := organization_types.AddUserToOrganizationRequest{
		OrganizationID: organization.ID.String(),
		UserID:         user.ID.String(),
		RoleId:         roles.ID.String(),
	}

	return c.organization_service.AddUserToOrganization(userOrganization)
}
