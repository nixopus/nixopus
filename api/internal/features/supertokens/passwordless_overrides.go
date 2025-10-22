package supertokens

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	user_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/supertokens/supertokens-golang/recipe/passwordless/plessmodels"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/recipe/userroles"
	"github.com/supertokens/supertokens-golang/recipe/userroles/userrolesclaims"
	"github.com/supertokens/supertokens-golang/supertokens"
)

// createPasswordlessUser creates a passwordless user in our database
func createPasswordlessUser(supertokensUserID, email string) (*shared_types.User, error) {
	if app == nil {
		return nil, fmt.Errorf("app not initialized")
	}

	// Check if user already exists
	userStorage := &user_storage.UserStorage{DB: app.Store.DB, Ctx: app.Ctx}
	existingUser, err := userStorage.FindUserBySupertokensID(supertokensUserID)
	if err == nil && existingUser != nil {
		return existingUser, nil
	}

	if err != nil {
		// Continue to create new user
	}

	// Create new user
	userID := uuid.New()
	user := &shared_types.User{
		ID:                userID,
		SupertokensUserID: supertokensUserID,
		Email:             email,
		Username:          strings.Split(email, "@")[0], // Use email prefix as username
		Type:              shared_types.UserTypeUser,
		IsVerified:        true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := userStorage.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user in database: %w", err)
	}

	return user, nil
}

// addPasswordlessUserToOrganization adds a passwordless user to an organization with a specific role
func addPasswordlessUserToOrganization(userID, email, organizationID, role string) error {
	if app == nil {
		return fmt.Errorf("app not initialized")
	}

	// Create the user in database
	user, err := createPasswordlessUser(userID, email)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Get the organization
	var organization shared_types.Organization
	err = app.Store.DB.NewSelect().Model(&organization).Where("id = ?", organizationID).Scan(app.Ctx)
	if err != nil {
		return fmt.Errorf("failed to find organization: %w", err)
	}

	// Start transaction for organization assignment
	tx, err := app.Store.DB.BeginTx(app.Ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Add user to organization with role
	if err := addUserToOrganization(*user, organization, &tx); err != nil {
		return fmt.Errorf("failed to add user to organization: %w", err)
	}

	// Create organization specific role and assign it to the user
	roleName := fmt.Sprintf("orgid_%s_%s", organizationID, role)

	// Determine permissions based on role
	var permissions []string
	switch role {
	case "admin":
		permissions = GetAdminPermissions()
	case "member":
		permissions = GetMemberPermissions()
	case "viewer":
		permissions = GetViewerPermissions()
	default:
		permissions = GetViewerPermissions() // Default to viewer permissions
	}

	// Create the organization specific role first
	if _, createRoleErr := userroles.CreateNewRoleOrAddPermissions(roleName, permissions, nil); createRoleErr != nil {
		// Log error but don't fail the entire operation for role creation failure
		fmt.Printf("Failed to create organization specific role %s: %v", roleName, createRoleErr)
	}

	// Then assign the role to the user
	if _, roleErr := userroles.AddRoleToUser("public", userID, roleName, nil); roleErr != nil {
		// Log error but don't fail the entire operation for role assignment failure
		fmt.Printf("Failed to assign SuperTokens role %s to user %s: %v", roleName, userID, roleErr)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ensurePasswordlessUserOrganizationAccess ensures an existing passwordless user has access to the organization
func ensurePasswordlessUserOrganizationAccess(userID, email, organizationID, role string) error {
	if app == nil {
		return fmt.Errorf("app not initialized")
	}

	// Check if user is already in the organization
	var userOrg shared_types.OrganizationUsers
	err := app.Store.DB.NewSelect().Model(&userOrg).
		Where("user_id = ? AND organization_id = ?", userID, organizationID).
		Scan(app.Ctx)

	if err == sql.ErrNoRows {
		// User not in organization, add them
		return addPasswordlessUserToOrganization(userID, email, organizationID, role)
	} else if err != nil {
		return fmt.Errorf("failed to check user organization membership: %w", err)
	}

	// User is already in organization
	return nil
}

// isOrganizationInvitation checks if the user context contains organization invitation data
func isOrganizationInvitation(userContext supertokens.UserContext) bool {
	if userContext == nil {
		return false
	}

	// Check for organization_id and role in userContext
	_, hasOrgID := (*userContext)["organization_id"]
	_, hasRole := (*userContext)["role"]

	return hasOrgID && hasRole
}

// extractInvitationData extracts organization invitation data from user context
func extractInvitationData(userContext supertokens.UserContext) (orgID, role, email string, ok bool) {
	// Extract from _default.request if it exists
	if userContext != nil {
		if defaultData, exists := (*userContext)["_default"]; exists {
			if castData, ok := defaultData.(map[string]interface{}); ok {
				if requestVal, reqExists := castData["request"]; reqExists {
					// Try to extract from request body
					if request, reqOk := requestVal.(*http.Request); reqOk {
						// Read the request body
						bodyBytes, err := io.ReadAll(request.Body)
						if err != nil {
							return "", "", "", false
						}

						// Parse JSON body
						var bodyData map[string]interface{}
						if err := json.Unmarshal(bodyBytes, &bodyData); err != nil {
							return "", "", "", false
						}

						// Extract organization data
						if orgIDVal, orgExists := bodyData["organization_id"]; orgExists {
							if roleVal, roleExists := bodyData["role"]; roleExists {
								if emailVal, emailExists := bodyData["email"]; emailExists {
									orgID, orgOk := orgIDVal.(string)
									role, roleOk := roleVal.(string)
									email, emailOk := emailVal.(string)

									if orgOk && roleOk && emailOk {
										return orgID, role, email, true
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return "", "", "", false
}

// handleNewUserSignup processes organization assignment for new users
func handleNewUserSignup(user plessmodels.User, userContext supertokens.UserContext) {
	orgID, role, email, ok := extractInvitationData(userContext)
	if !ok {
		return
	}

	addPasswordlessUserToOrganization(user.ID, email, orgID, role)
}

// handleExistingUserSignin processes organization access verification for existing users
func handleExistingUserSignin(user plessmodels.User, userContext supertokens.UserContext) {
	orgID, role, email, ok := extractInvitationData(userContext)
	if !ok {
		return
	}

	ensurePasswordlessUserOrganizationAccess(user.ID, email, orgID, role)
}

// createCodeOverride returns the CreateCode override function
func createCodeOverride(originalCreateCode func(email *string, phoneNumber *string, userInputCode *string, tenantId string, userContext supertokens.UserContext) (plessmodels.CreateCodeResponse, error)) func(email *string, phoneNumber *string, userInputCode *string, tenantId string, userContext supertokens.UserContext) (plessmodels.CreateCodeResponse, error) {
	return func(email *string, phoneNumber *string, userInputCode *string, tenantId string, userContext supertokens.UserContext) (plessmodels.CreateCodeResponse, error) {
		// Check if this is an organization invitation by checking the user context that we will set in the send invite endpoint
		isOrgInvite := isOrganizationInvitation(userContext)

		if isOrgInvite {
			// This is an organization invitation, allow it
			return originalCreateCode(email, phoneNumber, userInputCode, tenantId, userContext)
		}

		// Block unauthorized passwordless signups (because we use email password for admin registration by default and passwordless invites are only available through organization invitations)
		return plessmodels.CreateCodeResponse{}, fmt.Errorf("passwordless authentication is only available through organization invitations")
	}
}

// consumeCodeOverride overrides the ConsumeCode function based on SuperTokens documentation
func consumeCodeOverride(originalConsumeCode func(userInput *plessmodels.UserInputCodeWithDeviceID, linkCode *string, preAuthSessionID string, tenantId string, userContext supertokens.UserContext) (plessmodels.ConsumeCodeResponse, error)) func(userInput *plessmodels.UserInputCodeWithDeviceID, linkCode *string, preAuthSessionID string, tenantId string, userContext supertokens.UserContext) (plessmodels.ConsumeCodeResponse, error) {
	return func(userInput *plessmodels.UserInputCodeWithDeviceID, linkCode *string, preAuthSessionID string, tenantId string, userContext supertokens.UserContext) (plessmodels.ConsumeCodeResponse, error) {
		// First call the original ConsumeCode implementation
		response, err := originalConsumeCode(userInput, linkCode, preAuthSessionID, tenantId, userContext)
		if err != nil {
			return plessmodels.ConsumeCodeResponse{}, err
		}

		if response.OK != nil {
			user := response.OK.User

			if response.OK.CreatedNewUser {
				// Post sign up logic
				handleNewUserSignup(user, userContext)
			} else {
				// Post sign in logic
				handleExistingUserSignin(user, userContext)
			}

			// Refresh session with roles and permissions for both new and existing users
			// We need to get the session from the user context
			if userContext != nil {
				if defaultData, exists := (*userContext)["_default"]; exists {
					if castData, ok := defaultData.(map[string]interface{}); ok {
						if requestVal, reqExists := castData["request"]; reqExists {
							if request, reqOk := requestVal.(*http.Request); reqOk {
								ctx := request.Context()
								sessContainer := session.GetSessionFromRequestContext(ctx)
								if sessContainer != nil {
									// Refresh roles and permissions in the session
									if err := sessContainer.FetchAndSetClaim(userrolesclaims.UserRoleClaim); err != nil {
										// Log error but don't fail the operation
									}
									if err := sessContainer.FetchAndSetClaim(userrolesclaims.PermissionClaim); err != nil {
										// Log error but don't fail the operation
									}
								}
							}
						}
					}
				}
			}
		}

		return response, nil
	}
}

// createPasswordlessOverrides returns the passwordless recipe overrides for organization-based authentication
func createPasswordlessOverrides() *plessmodels.OverrideStruct {
	return &plessmodels.OverrideStruct{
		Functions: func(originalImplementation plessmodels.RecipeInterface) plessmodels.RecipeInterface {
			originalCreateCode := *originalImplementation.CreateCode
			originalConsumeCode := *originalImplementation.ConsumeCode

			// Override CreateCode to prevent unauthorized signups
			(*originalImplementation.CreateCode) = createCodeOverride(originalCreateCode)

			// Override ConsumeCode to handle organization assignment
			(*originalImplementation.ConsumeCode) = consumeCodeOverride(originalConsumeCode)

			return originalImplementation
		},
	}
}
