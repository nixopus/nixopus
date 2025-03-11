package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	organization_types "github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *AuthService) Register(registration_request types.RegisterRequest) (types.AuthResponse, error) {
	var user shared_types.User
	c.logger.Log(logger.Info, "registering user", registration_request.Email)
	hashedPassword, err := HashPassword(registration_request.Password)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToHashPassword.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToHashPassword
	}

	user = shared_types.NewUser(registration_request.Email, hashedPassword, registration_request.Username, "", registration_request.Type, user.Type == "admin")

	if db_user, err := c.storage.FindUserByEmail(registration_request.Email); err == nil {
		c.logger.Log(logger.Error, types.ErrUserWithEmailAlreadyExists.Error(), "")
		if db_user.ID != uuid.Nil {
			return types.AuthResponse{}, types.ErrUserWithEmailAlreadyExists
		}
	}

	err = c.storage.CreateUser(&user)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToRegisterUser.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToRegisterUser
	}

	refreshToken, err := c.storage.CreateRefreshToken(user.ID)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToCreateRefreshToken.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToCreateToken
	}

	accessToken, err := createToken(user.Email, time.Minute*15)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToCreateAccessToken.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToCreateToken
	}

	// for admin user we need to add them to a default organization with appropriate role
	if user.Type == "admin" {
		organization, err := c.createDefaultOrganization(user)
		if err != nil {
			c.logger.Log(logger.Error, types.ErrFailedToCreateDefaultOrganization.Error(), err.Error())
			return types.AuthResponse{}, types.ErrFailedToCreateDefaultOrganization
		}

		err = c.addUserToOrganizationWithAdminRole(user, organization)
		if err != nil {
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
	c.logger.Log(logger.Info, "creating default organization for admin user", user.Email)
	org, err := c.organization_service.CreateOrganization(&organization_types.CreateOrganizationRequest{
		Name:        user.Username + "'s Team",
		Description: "My Team",
	})
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToCreateDefaultOrganization.Error(), err.Error())
		return shared_types.Organization{}, types.ErrFailedToCreateDefaultOrganization
	}

	c.logger.Log(logger.Info, "created default organization for admin user", user.Email)
	return org, nil
}

func (c *AuthService) addUserToOrganizationWithAdminRole(user shared_types.User, organization shared_types.Organization) error {
	c.logger.Log(logger.Info, "adding user to organization with admin role", user.Email)

	roles, err := c.role_service.GetRoleByName("admin")
	if err != nil {
		return err
	}
	if roles == nil {
		return types.ErrNoRolesFound
	}

	user_organization := organization_types.AddUserToOrganizationRequest{
		OrganizationID: organization.ID.String(),
		UserID:         user.ID.String(),
		RoleId:         roles.ID.String(),
	}
	return c.organization_service.AddUserToOrganization(user_organization)
}
