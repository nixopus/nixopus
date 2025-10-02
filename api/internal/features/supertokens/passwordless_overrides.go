package supertokens

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	user_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/supertokens/supertokens-golang/ingredients/emaildelivery"
	"github.com/supertokens/supertokens-golang/recipe/passwordless/plessmodels"
	"github.com/supertokens/supertokens-golang/recipe/userroles"
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
		log.Printf("User with SuperTokens ID %s already exists", supertokensUserID)
		return existingUser, nil
	}

	// Create new user
	user := &shared_types.User{
		ID:                uuid.New(),
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

	log.Printf("Successfully created user %s (ID: %s) in database with SuperTokens ID %s",
		email, user.ID, supertokensUserID)
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
	if err := addUserToOrganizationWithRole(*user, organization, role, &tx); err != nil {
		return fmt.Errorf("failed to add user to organization: %w", err)
	}

	// Assign SuperTokens role to the user in the default tenant
	if _, roleErr := userroles.AddRoleToUser("public", userID, role, nil); roleErr != nil {
		log.Printf("Failed to assign SuperTokens role to user: %v", roleErr)
		// Don't fail the entire operation for role assignment failure
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Successfully added user %s to organization %s with role %s",
		email, organization.Name, role)
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

	_, hasOrgID := (*userContext)["organization_id"]
	_, hasRole := (*userContext)["role"]

	return hasOrgID && hasRole
}

// extractInvitationData extracts organization invitation data from user context
func extractInvitationData(userContext supertokens.UserContext) (orgID, role, email string, ok bool) {
	if userContext == nil {
		return "", "", "", false
	}

	orgIDVal, orgExists := (*userContext)["organization_id"]
	roleVal, roleExists := (*userContext)["role"]
	emailVal, emailExists := (*userContext)["email"]

	if !orgExists || !roleExists || !emailExists {
		return "", "", "", false
	}

	orgID, orgOk := orgIDVal.(string)
	role, roleOk := roleVal.(string)
	email, emailOk := emailVal.(string)

	return orgID, role, email, orgOk && roleOk && emailOk
}

// handleNewUserSignup processes organization assignment for new users
func handleNewUserSignup(user plessmodels.User, userContext supertokens.UserContext) {
	orgID, role, email, ok := extractInvitationData(userContext)
	if !ok {
		log.Printf("Failed to extract invitation data for new user: %s", user.ID)
		return
	}

	err := addPasswordlessUserToOrganization(user.ID, email, orgID, role)
	if err != nil {
		log.Printf("Failed to add new user to organization: %v", err)
	}
}

// handleExistingUserSignin processes organization access verification for existing users
func handleExistingUserSignin(user plessmodels.User, userContext supertokens.UserContext) {
	orgID, role, email, ok := extractInvitationData(userContext)
	if !ok {
		log.Printf("Failed to extract invitation data for existing user: %s", user.ID)
		return
	}

	err := ensurePasswordlessUserOrganizationAccess(user.ID, email, orgID, role)
	if err != nil {
		log.Printf("Failed to verify user organization access: %v", err)
	}
}

// createCodeOverride returns the CreateCode override function
func createCodeOverride(originalCreateCode func(email *string, phoneNumber *string, userInputCode *string, tenantId string, userContext supertokens.UserContext) (plessmodels.CreateCodeResponse, error)) func(email *string, phoneNumber *string, userInputCode *string, tenantId string, userContext supertokens.UserContext) (plessmodels.CreateCodeResponse, error) {
	return func(email *string, phoneNumber *string, userInputCode *string, tenantId string, userContext supertokens.UserContext) (plessmodels.CreateCodeResponse, error) {
		// Check if this is an organization invitation by checking the user context that we will set in the send invite endpoint
		if isOrganizationInvitation(userContext) {
			// This is an organization invitation, allow it
			return originalCreateCode(email, phoneNumber, userInputCode, tenantId, userContext)
		}

		// Block unauthorized passwordless signups (because we use email password for admin registration by default and passwordless invites are only available through organization invitations)
		return plessmodels.CreateCodeResponse{}, fmt.Errorf("passwordless authentication is only available through organization invitations")
	}
}

// consumeCodeOverride returns the ConsumeCode override function
func consumeCodeOverride(originalConsumeCode func(userInput *plessmodels.UserInputCodeWithDeviceID, linkCode *string, preAuthSessionID string, tenantId string, userContext supertokens.UserContext) (plessmodels.ConsumeCodeResponse, error)) func(userInput *plessmodels.UserInputCodeWithDeviceID, linkCode *string, preAuthSessionID string, tenantId string, userContext supertokens.UserContext) (plessmodels.ConsumeCodeResponse, error) {
	return func(userInput *plessmodels.UserInputCodeWithDeviceID, linkCode *string, preAuthSessionID string, tenantId string, userContext supertokens.UserContext) (plessmodels.ConsumeCodeResponse, error) {
		// First call the original implementation
		response, err := originalConsumeCode(userInput, linkCode, preAuthSessionID, tenantId, userContext)
		if err != nil {
			return plessmodels.ConsumeCodeResponse{}, err
		}

		if response.OK != nil {
			user := response.OK.User

			if response.OK.CreatedNewUser {
				// New user signup - assign to organization with role
				handleNewUserSignup(user, userContext)
			} else {
				// Existing user signin - ensure they have access to the organization
				handleExistingUserSignin(user, userContext)
			}
		}

		return response, nil
	}
}

// createPasswordlessOverrides returns the passwordless recipe overrides for organization-based authentication
func createPasswordlessOverrides() *plessmodels.OverrideStruct {
	return &plessmodels.OverrideStruct{
		Functions: func(originalImplementation plessmodels.RecipeInterface) plessmodels.RecipeInterface {
			originalConsumeCode := *originalImplementation.ConsumeCode
			originalCreateCode := *originalImplementation.CreateCode

			// Override CreateCode to prevent unauthorized signups
			(*originalImplementation.CreateCode) = createCodeOverride(originalCreateCode)

			// Override ConsumeCode to handle organization assignment
			(*originalImplementation.ConsumeCode) = consumeCodeOverride(originalConsumeCode)

			return originalImplementation
		},
	}
}

// createPasswordlessEmailDeliveryOverride returns the email delivery override for organization invitations
func createPasswordlessEmailDeliveryOverride() *emaildelivery.TypeInput {
	return &emaildelivery.TypeInput{
		Override: func(originalImplementation emaildelivery.EmailDeliveryInterface) emaildelivery.EmailDeliveryInterface {
			ogSendEmail := *originalImplementation.SendEmail
			(*originalImplementation.SendEmail) = func(input emaildelivery.EmailType, userContext supertokens.UserContext) error {
				// Customize magic link URL for organization invitations
				if input.PasswordlessLogin != nil && input.PasswordlessLogin.UrlWithLinkCode != nil {
					// Check if this is an organization invitation by looking for custom data
					if userContext != nil {
						if orgID, exists := (*userContext)["organization_id"]; exists {
							newUrl := strings.Replace(
								*input.PasswordlessLogin.UrlWithLinkCode,
								"/auth/verify",
								fmt.Sprintf("/auth/organization-invite?org_id=%s", orgID),
								1,
							)
							input.PasswordlessLogin.UrlWithLinkCode = &newUrl
						}
					}
				}
				return ogSendEmail(input, userContext)
			}
			return originalImplementation
		},
	}
}
