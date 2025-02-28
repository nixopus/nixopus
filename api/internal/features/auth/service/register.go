package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	organization_service "github.com/raghavyuva/nixopus-api/internal/features/organization/service"
	organization_types "github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	permission_service "github.com/raghavyuva/nixopus-api/internal/features/permission/service"
	permission_types "github.com/raghavyuva/nixopus-api/internal/features/permission/types"
	role_service "github.com/raghavyuva/nixopus-api/internal/features/role/service"
	role_types "github.com/raghavyuva/nixopus-api/internal/features/role/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// Register registers a new user and returns an authentication response.
//
// The function takes a types.RegisterRequest as input, which includes the user's
// username, email, and password. It first hashes the password and checks if a user
// with the provided email already exists. If a user exists, it returns an error.
// If not, it creates a new user in the database. It then creates a refresh token
// and an access token for the user.
//
// Returns a types.AuthResponse containing the access token, refresh token, expiration
// time, and user information. If any step fails, it returns an appropriate error.
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

	// for admin user we need to create a default organization, roles and permissions, and add the user to the organization
	if user.Type == "admin" {
		organization, err := c.createDefaultOrganization(user)
		if err != nil {
			c.logger.Log(logger.Error, types.ErrFailedToCreateDefaultOrganization.Error(), err.Error())
			return types.AuthResponse{}, types.ErrFailedToCreateDefaultOrganization
		}

		err = c.createDefaultPermissions(organization)
		if err != nil {
			c.logger.Log(logger.Error, types.ErrFailedToCreateDefaultPermissions.Error(), err.Error())
			return types.AuthResponse{}, types.ErrFailedToCreateDefaultPermissions
		}

		err = c.createDefaultRoles(organization)
		if err != nil {
			c.logger.Log(logger.Error, types.ErrFailedToCreateDefaultRoles.Error(), err.Error())
			return types.AuthResponse{}, types.ErrFailedToCreateDefaultRoles
		}

		err = c.addDefaultPermissionsToDefaultRole()
		if err != nil {
			c.logger.Log(logger.Error, types.ErrFailedToAddDefaultPermissionsToDefaultRole.Error(), err.Error())
			return types.AuthResponse{}, types.ErrFailedToAddDefaultPermissionsToDefaultRole
		}

		err = c.addUserToOrganization(user)
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
	organiztion_service := organization_service.NewOrganizationService(c.store, c.storage.Ctx, c.logger)

	err := organiztion_service.CreateOrganization(&organization_types.CreateOrganizationRequest{
		Name:        user.Username + "'s Team",
		Description: "My Team",
	})
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToCreateDefaultOrganization.Error(), err.Error())
		return shared_types.Organization{}, types.ErrFailedToCreateDefaultOrganization
	}

	c.logger.Log(logger.Info, "created default organization for admin user", user.Email)

	return shared_types.Organization{}, nil
}

func (c *AuthService) createDefaultPermissions(organization shared_types.Organization) error {
	c.logger.Log(logger.Info, "creating default permissions for organization", organization.Name)
	permissions_service := permission_service.NewPermissionService(c.store, c.storage.Ctx, c.logger)
	for _, permission := range permissions_to_create {
		err := permissions_service.CreatePermission(&permission)
		if err != nil {
			c.logger.Log(logger.Error, types.ErrFailedToCreateDefaultPermissions.Error(), err.Error())
			return types.ErrFailedToCreateDefaultPermissions
		}
	}
	return nil
}

func (c *AuthService) createDefaultRoles(organization shared_types.Organization) error {
	c.logger.Log(logger.Info, "creating default roles for organization", organization.Name)
	roles_service := role_service.NewRoleService(c.store, c.storage.Ctx, c.logger)

	for _, role := range roles_to_create {
		err := roles_service.CreateRole(&role)
		if err != nil {
			c.logger.Log(logger.Error, types.ErrFailedToCreateDefaultRoles.Error(), err.Error())
			return types.ErrFailedToCreateDefaultRoles
		}
	}
	return nil
}

func (c *AuthService) addDefaultPermissionsToDefaultRole() error {
	permission_service := permission_service.NewPermissionService(c.store, c.storage.Ctx, c.logger)
	all_permissions, err := permission_service.GetAllPermissions()
	if err != nil {
		return err
	}

	role_service := role_service.NewRoleService(c.store, c.storage.Ctx, c.logger)
	all_roles, err := role_service.GetRoles()
	if err != nil {
		return err
	}

	roleNameToID := make(map[string]string)
	for _, role := range all_roles {
		roleNameToID[role.Name] = role.ID.String()
	}

	permKeyToID := make(map[string]string)
	for _, perm := range all_permissions {
		key := perm.Resource + ":" + perm.Name
		permKeyToID[key] = perm.ID.String()
	}

	for roleName, permissions := range rolePermissions {
		roleID, ok := roleNameToID[roleName]
		if !ok {
			c.logger.Log(logger.Error, "Role not found", roleName)
			continue
		}

		for _, permKey := range permissions {
			permID, ok := permKeyToID[permKey]
			if !ok {
				c.logger.Log(logger.Error, "Permission not found", permKey)
				continue
			}

			req := permission_types.AddPermissionToRoleRequest{
				RoleID:       roleID,
				PermissionID: permID,
			}

			err := permission_service.AddPermissionToRole(req.PermissionID, req.RoleID)
			if err != nil {
				c.logger.Log(logger.Error, "Failed to add permission to role", err.Error())
			}
		}
	}

	return nil
}

func (c *AuthService) addUserToOrganization(user shared_types.User) error {
	c.logger.Log(logger.Info, "adding user to organization", user.Email)
	organization_service := organization_service.NewOrganizationService(c.store, c.storage.Ctx, c.logger)
	organizations, err := organization_service.GetOrganizations()
	if err != nil {
		return err
	}
	if len(organizations) == 0 {
		return types.ErrNoOrganizationsFound
	}
	roles_service := role_service.NewRoleService(c.store, c.storage.Ctx, c.logger)
	roles, err := roles_service.GetRoles()
	if err != nil {
		return err
	}
	if len(roles) == 0 {
		return types.ErrNoRolesFound
	}
	user_organization := organization_types.AddUserToOrganizationRequest{
		OrganizationID: organizations[0].ID.String(),
		UserID:         user.ID.String(),
		RoleId:         roles[0].ID.String(),
	}
	return organization_service.AddUserToOrganization(user_organization)
}

var permissions_to_create = []permission_types.CreatePermissionRequest{
	{
		Name:        "READ",
		Resource:    "organization",
		Description: "Read Organization",
	},
	{
		Name:        "UPDATE",
		Resource:    "organization",
		Description: "Update Organization",
	},
	{
		Name:        "DELETE",
		Resource:    "organization",
		Description: "Delete Organization",
	},
	{
		Name:        "READ",
		Resource:    "user",
		Description: "Read User",
	},
	{
		Name:        "UPDATE",
		Resource:    "user",
		Description: "Update User",
	},
	{
		Name:        "DELETE",
		Resource:    "user",
		Description: "Delete User",
	},
	{
		Name:        "READ",
		Resource:    "role",
		Description: "Read Role",
	},
	{
		Name:        "UPDATE",
		Resource:    "role",
		Description: "Update Role",
	},
	{
		Name:        "DELETE",
		Resource:    "role",
		Description: "Delete Role",
	},
}

var roles_to_create = []role_types.CreateRoleRequest{
	{
		Name:        "Owner",
		Description: "Owner Role",
	},
	{
		Name:        "admin",
		Description: "Admin Role",
	},
	{
		Name:        "user",
		Description: "User Role",
	},
}

var rolePermissions = map[string][]string{
	"Owner": {
		"organization:READ", "organization:UPDATE", "organization:DELETE",
		"user:READ", "user:UPDATE", "user:DELETE",
		"role:READ", "role:UPDATE", "role:DELETE",
	},
	"admin": {
		"organization:READ", "organization:UPDATE",
		"user:READ", "user:UPDATE", "user:DELETE",
		"role:READ",
	},
	"user": {
		"organization:READ",
		"user:READ",
	},
}
