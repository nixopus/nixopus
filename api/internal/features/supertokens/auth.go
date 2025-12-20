package supertokens

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/config"
	user_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/supertokens/supertokens-golang/recipe/emailpassword"
	"github.com/supertokens/supertokens-golang/recipe/emailpassword/epmodels"
	"github.com/supertokens/supertokens-golang/recipe/passwordless"
	"github.com/supertokens/supertokens-golang/recipe/passwordless/plessmodels"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
	"github.com/supertokens/supertokens-golang/recipe/userroles"
	"github.com/supertokens/supertokens-golang/recipe/userroles/userrolesclaims"
	"github.com/supertokens/supertokens-golang/recipe/userroles/userrolesmodels"
	"github.com/supertokens/supertokens-golang/supertokens"
	"github.com/uptrace/bun"
)

var app *storage.App

// Permission constants to avoid duplication
var (
	adminPermissions = []string{
		"user:create", "user:read", "user:update", "user:delete",
		"organization:create", "organization:read", "organization:update", "organization:delete",
		"role:create", "role:read", "role:update", "role:delete",
		"permission:create", "permission:read", "permission:update", "permission:delete",
		"domain:create", "domain:read", "domain:update", "domain:delete",
		"github-connector:create", "github-connector:read", "github-connector:update", "github-connector:delete",
		"notification:create", "notification:read", "notification:update", "notification:delete",
		"file-manager:create", "file-manager:read", "file-manager:update", "file-manager:delete",
		"deploy:create", "deploy:read", "deploy:update", "deploy:delete",
		"container:create", "container:read", "container:update", "container:delete",
		"audit:create", "audit:read", "audit:update", "audit:delete",
		"terminal:create", "terminal:read", "terminal:update", "terminal:delete",
		"feature_flags:read", "feature_flags:update",
		"dashboard:read", "extension:read", "extension:create", "extension:update", "extension:delete",
	}

	memberPermissions = []string{
		"user:read", "user:update",
		"organization:read", "organization:update",
		"container:read",
		"audit:read",
		"domain:read",
		"notification:read",
		"file-manager:read",
		"deploy:read",
		"feature_flags:read",
		"dashboard:read",
		"extension:read",
	}

	viewerPermissions = []string{
		"user:read", "organization:read", "container:read", "audit:read", "domain:read", "notification:read", "file-manager:read", "deploy:read", "feature_flags:read", "dashboard:read",
		"extension:read",
	}
)

// Init initializes the SuperTokens authentication system
func Init(appInstance *storage.App) {
	app = appInstance
	config := config.AppConfig
	apiBasePath := "/auth"
	websiteBasePath := "/auth"

	// Disable debug logs in development mode
	isDevelopment := strings.ToLower(config.App.Environment) == "development" || strings.ToLower(config.App.Environment) == "dev"
	debugEnabled := !isDevelopment

	err := supertokens.Init(supertokens.TypeInput{
		Debug: debugEnabled,
		Supertokens: &supertokens.ConnectionInfo{
			ConnectionURI: config.Supertokens.ConnectionURI,
			APIKey:        config.Supertokens.APIKey,
		},
		AppInfo: supertokens.AppInfo{
			AppName:         "Nixopus",
			APIDomain:       config.Supertokens.APIDomain,
			WebsiteDomain:   config.Supertokens.WebsiteDomain,
			APIBasePath:     &apiBasePath,
			WebsiteBasePath: &websiteBasePath,
		},
		RecipeList: []supertokens.Recipe{
			userroles.Init(&userrolesmodels.TypeInput{}),
			emailpassword.Init(&epmodels.TypeInput{
				Override: &epmodels.OverrideStruct{
					APIs: func(originalImplementation epmodels.APIInterface) epmodels.APIInterface {
						originalSignUpPOST := *originalImplementation.SignUpPOST
						newSignUpPOST := func(formFields []epmodels.TypeFormField, tenantId string, options epmodels.APIOptions, userContext supertokens.UserContext) (epmodels.SignUpPOSTResponse, error) {
							// If an admin already exists, disable sign up attempts
							if app != nil {
								userStorage := &user_storage.UserStorage{DB: app.Store.DB, Ctx: app.Ctx}
								adminUser, findErr := userStorage.FindUserByType(shared_types.UserTypeAdmin)
								if findErr == nil && adminUser != nil {
									return epmodels.SignUpPOSTResponse{
										GeneralError: &supertokens.GeneralErrorResponse{Message: "Sign up is disabled"},
									}, nil
								}
							}

							// Call the original sign up API
							response, err := originalSignUpPOST(formFields, tenantId, options, userContext)

							// If sign up was successful, create user in our database
							if err == nil && response.OK != nil {
								createUserInDatabase(response.OK.User.ID, response.OK.User.Email)

								// Add roles and permissions to the newly created session
								if options.Req != nil {
									ctx := options.Req.Context()
									sessContainer := session.GetSessionFromRequestContext(ctx)
									if sessContainer != nil {
										_ = addRolesAndPermissionsToSession(sessContainer)
									}
								}
							}

							return response, err
						}
						originalImplementation.SignUpPOST = &newSignUpPOST
						return originalImplementation
					},
				},
			}),
			passwordless.Init(plessmodels.TypeInput{
				FlowType: "MAGIC_LINK",
				ContactMethodEmail: plessmodels.ContactMethodEmailConfig{
					Enabled: true,
				},
				Override: createPasswordlessOverrides(),
			}),
			session.Init(&sessmodels.TypeInput{
				ExposeAccessTokenToFrontendInCookieBasedAuth: true,
			}),
		},
	})

	if err != nil {
		panic(err.Error())
	}

	if seedErr := seedDefaultRolesAndPermissions(); seedErr != nil {
		log.Printf("Failed to seed roles and permissions via SuperTokens: %v", seedErr)
	}
}

// addRolesAndPermissionsToSession fetches and sets the user's roles and permissions claims in the session
func addRolesAndPermissionsToSession(sessionContainer sessmodels.SessionContainer) error {
	if err := sessionContainer.FetchAndSetClaim(userrolesclaims.UserRoleClaim); err != nil {
		return err
	}
	if err := sessionContainer.FetchAndSetClaim(userrolesclaims.PermissionClaim); err != nil {
		return err
	}
	return nil
}

// createUserInDatabase creates a user in our database when they sign up through SuperTokens
func createUserInDatabase(supertokensUserID, email string) {
	if app == nil {
		log.Printf("Warning: App instance not available for user creation")
		return
	}

	// Check if user already exists
	userStorage := &user_storage.UserStorage{DB: app.Store.DB, Ctx: app.Ctx}
	existingUser, err := userStorage.FindUserBySupertokensID(supertokensUserID)
	if err == nil && existingUser != nil {
		log.Printf("User with SuperTokens ID %s already exists", supertokensUserID)
		return
	}

	// Create new user
	user := &types.User{
		ID:                uuid.New(),
		SupertokensUserID: supertokensUserID,
		Email:             email,
		Username:          strings.Split(email, "@")[0], // Use email as username for now
		Type:              shared_types.UserTypeAdmin,
		IsVerified:        true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := userStorage.CreateUser(user); err != nil {
		log.Printf("Failed to create user in database: %v", err)
		return
	}

	// Start transaction for organization creation
	tx, err := app.Store.DB.BeginTx(app.Ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		return
	}
	defer tx.Rollback()

	// Create default organization for the user (similar to admin registration flow)
	organization, err := createDefaultOrganizationForUser(*user, &tx)
	if err != nil {
		log.Printf("Failed to create default organization: %v", err)
		return
	}

	// Add user to organization as admin
	if err := addUserToOrganization(*user, organization, &tx); err != nil {
		log.Printf("Failed to add user to organization: %v", err)
		return
	}

	// Create organization specific admin role and assign it to the user
	roleName := fmt.Sprintf("orgid_%s_admin", organization.ID.String())

	// Create the organization specific role first
	if _, createRoleErr := userroles.CreateNewRoleOrAddPermissions(roleName, GetAdminPermissions(), nil); createRoleErr != nil {
		log.Printf("Failed to create organization-specific role %s: %v", roleName, createRoleErr)
		return
	}

	// Then assign the role to the user
	if _, roleErr := userroles.AddRoleToUser("public", supertokensUserID, roleName, nil); roleErr != nil {
		log.Printf("Failed to assign SuperTokens role %s to user %s: %v", roleName, supertokensUserID, roleErr)
		return
	}

	if err := createDefaultFeatureFlags(organization.ID, &tx); err != nil {
		log.Printf("Failed to create default feature flags: %v", err)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return
	}

	log.Printf("Successfully created user %s (ID: %s) in database with SuperTokens ID %s and default organization %s (ID: %s)",
		email, user.ID, supertokensUserID, organization.Name, organization.ID)
}

// seedDefaultRolesAndPermissions creates initial roles and permissions in SuperTokens
func seedDefaultRolesAndPermissions() error {
	if _, err := userroles.CreateNewRoleOrAddPermissions("admin", GetAdminPermissions(), nil); err != nil {
		return err
	}

	if _, err := userroles.CreateNewRoleOrAddPermissions("member", GetMemberPermissions(), nil); err != nil {
		return err
	}

	if _, err := userroles.CreateNewRoleOrAddPermissions("viewer", GetViewerPermissions(), nil); err != nil {
		return err
	}

	return nil
}

// createDefaultOrganizationForUser creates a default organization for a user
func createDefaultOrganizationForUser(user types.User, tx *bun.Tx) (types.Organization, error) {
	log.Printf("Creating default organization for user %s", user.Email)

	// Create a simple organization structure
	organization := types.Organization{
		ID:          uuid.New(),
		Name:        user.Username + "'s Team",
		Description: "My Team",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err := tx.NewInsert().Model(&organization).Exec(app.Ctx)
	if err != nil {
		return types.Organization{}, fmt.Errorf("failed to create organization: %w", err)
	}

	log.Printf("Created default organization for user %s", user.Email)
	return organization, nil
}

// addUserToOrganization adds a user to an organization
func addUserToOrganization(user types.User, organization types.Organization, tx *bun.Tx) error {
	orgUser := types.OrganizationUsers{
		ID:             uuid.New(),
		UserID:         user.ID,
		OrganizationID: organization.ID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err := tx.NewInsert().Model(&orgUser).Exec(app.Ctx)
	if err != nil {
		return fmt.Errorf("failed to add user to organization: %w", err)
	}

	log.Printf("Added user %s (ID: %s) to organization %s (ID: %s)",
		user.Email, user.ID, organization.Name, organization.ID)
	return nil
}

// createDefaultFeatureFlags creates default feature flags for a new organization
func createDefaultFeatureFlags(organizationID uuid.UUID, tx *bun.Tx) error {
	log.Printf("Creating default feature flags for organization %s", organizationID)

	defaultFeatures := []types.FeatureName{
		types.FeatureDomain,
		types.FeatureTerminal,
		types.FeatureNotifications,
		types.FeatureFileManager,
		types.FeatureSelfHosted,
		types.FeatureAudit,
		types.FeatureGithubConnector,
		types.FeatureMonitoring,
		types.FeatureContainer,
	}

	for _, feature := range defaultFeatures {
		featureFlag := types.FeatureFlag{
			ID:             uuid.New(),
			OrganizationID: organizationID,
			FeatureName:    string(feature),
			IsEnabled:      true,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		_, err := tx.NewInsert().Model(&featureFlag).Exec(app.Ctx)
		if err != nil {
			return fmt.Errorf("failed to create feature flag %s: %w", feature, err)
		}
	}

	log.Printf("Created %d default feature flags for organization %s", len(defaultFeatures), organizationID)
	return nil
}

// GetAdminPermissions returns the admin permissions list
func GetAdminPermissions() []string {
	return adminPermissions
}

// GetMemberPermissions returns the member permissions list
func GetMemberPermissions() []string {
	return memberPermissions
}

// GetViewerPermissions returns the viewer permissions list
func GetViewerPermissions() []string {
	return viewerPermissions
}

// GetRolesAndPermissionsForUserInOrganization retrieves roles and permissions for a user from SuperTokens, filtered by organization
func GetRolesAndPermissionsForUserInOrganization(userId, organizationId string) ([]string, []string, error) {
	// Get roles for the user
	rolesResponse, err := userroles.GetRolesForUser("public", userId, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get roles for user %s: %w", userId, err)
	}

	allRoles := rolesResponse.OK.Roles

	// Filter roles to only include organization specific roles
	var orgRoles []string
	var allPermissions []string
	permissionSet := make(map[string]bool)

	for _, role := range allRoles {
		// Check if this role is organization specific
		if strings.HasPrefix(role, "orgid_") && strings.Contains(role, organizationId) {
			orgRoles = append(orgRoles, role)
		} else if role == "admin" || role == "member" || role == "viewer" {
			orgRoles = append(orgRoles, role)
		} else {
			// Skip roles that don't belong to the organization
			continue
		}

		// Get permissions for the role
		permissionsResponse, err := userroles.GetPermissionsForRole(role, nil)
		if err != nil {
			continue
		}

		if permissionsResponse.UnknownRoleError != nil {
			continue
		}

		for _, permission := range permissionsResponse.OK.Permissions {
			if !permissionSet[permission] {
				permissionSet[permission] = true
				allPermissions = append(allPermissions, permission)
			}
		}
	}

	return orgRoles, allPermissions, nil
}
