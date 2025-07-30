package feature_flags

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
)

func TestGetFeatureFlags(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	testCases := []struct {
		name           string
		token          string
		organizationID string
		expectedStatus int
		description    string
	}{
		{
			name:           "successfully fetch feature flags with valid token",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "given valid credentials, get all the feature flags for organization",
		},
		{
			name:           "deny unauthorized access",
			token:          "",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "return unauthorized error without token",
		},
		{
			name:           "unauthorized with invalid token",
			token:          "invalid-token",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "throw unauthorized error with invalid token or expired",
		},
		{
			name:           "request without organization header",
			token:          user.AccessToken,
			organizationID: "",
			expectedStatus: http.StatusBadRequest,
			description:    "throws error when organization header is missing from request",
		},
		{
			name:           "request with invalid organization ID",
			token:          user.AccessToken,
			organizationID: "invalid-org-id",
			expectedStatus: http.StatusInternalServerError,
			description:    "throws 500 error when organization ID is invalid format",
		},
		{
			name:           "cross organization access denied",
			token:          user.AccessToken,
			organizationID: "123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusForbidden,
			description:    "should deny access to feature flags from different organization that user is not part of",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSteps := []IStep{
				Description(tc.description),
				Get(tests.GetFeatureFlagsURL()),
			}

			if tc.token != "" {
				testSteps = append(testSteps, Send().Headers("Authorization").Add("Bearer "+tc.token))
			}

			if tc.organizationID != "" {
				testSteps = append(testSteps, Send().Headers("X-Organization-Id").Add(tc.organizationID))
			}

			testSteps = append(testSteps, Expect().Status().Equal(int64(tc.expectedStatus)))

			if tc.expectedStatus == http.StatusOK {
				testSteps = append(testSteps,
					Expect().Body().JSON().JQ(".status").Equal("success"),
					Expect().Body().JSON().JQ(".message").Equal("Feature flags retrieved successfully"),
					Expect().Body().JSON().JQ(".data").NotEqual(nil),
					Expect().Body().JSON().JQ(".data | map(select(.feature_name == \"terminal\")) | length").Equal(1),
					Expect().Body().JSON().JQ(".data | map(select(.feature_name == \"container\")) | length").Equal(1),
					Expect().Body().JSON().JQ(".data | map(select(.feature_name == \"domain\")) | length").Equal(1),
				)
			}

			Test(t, testSteps...)
		})
	}
}

func TestUpdateFeatureFlag(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	// First, ensure feature flags exist by fetching them
	Test(t,
		Description("Initialize feature flags by fetching them"),
		Get(tests.GetFeatureFlagsURL()),
		Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
		Send().Headers("X-Organization-Id").Add(orgID),
		Expect().Status().Equal(http.StatusOK),
	)

	testCases := []struct {
		name           string
		featureName    string
		isEnabled      bool
		token          string
		organizationID string
		expectedStatus int
		description    string
	}{
		{
			name:           "sucessfully enable terminal feature flag",
			featureName:    "terminal",
			isEnabled:      true,
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "should enable terminal feature flag successfully",
		},
		{
			name:           "Successfully disable terminal feature flag",
			featureName:    "terminal",
			isEnabled:      false,
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "should disable terminal feature flag successfully",
		},
		{
			name:           "update container feature flag",
			featureName:    "container",
			isEnabled:      true,
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "should update container feature flag successfully",
		},
		{
			name:           "update domain feature flag",
			featureName:    "domain",
			isEnabled:      false,
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "should update domain feature flag successfully",
		},
		{
			name:           "update file_manager feature flag",
			featureName:    "file_manager",
			isEnabled:      true,
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "should update file_manager feature flag successfully",
		},
		{
			name:           "update notifications feature flag",
			featureName:    "notifications",
			isEnabled:      false,
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "should update notifications feature flag successfully",
		},
		{
			name:           "successfully update monitoring feature flag",
			featureName:    "monitoring",
			isEnabled:      true,
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "updates monitoring feature flag successfully",
		},
		{
			name:           "update github_connector feature flag",
			featureName:    "github_connector",
			isEnabled:      false,
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "update github_connector feature flag successfully",
		},
		{
			name:           "update audit feature flag",
			featureName:    "audit",
			isEnabled:      true,
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "update audit feature flag successfully",
		},
		{
			name:           "update self_hosted feature flag",
			featureName:    "self_hosted",
			isEnabled:      false,
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "update self_hosted feature flag successfully",
		},
		{
			name:           "unauthorized request without token",
			featureName:    "terminal",
			isEnabled:      true,
			token:          "",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "throw 401 when no authentication token is provided",
		},
		{
			name:           "unauthorized request with invalid token",
			featureName:    "terminal",
			isEnabled:      true,
			token:          "invalid-token",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "throw 401 when invalid authentication token is provided",
		},
		{
			name:           "request without organization header",
			featureName:    "terminal",
			isEnabled:      true,
			token:          user.AccessToken,
			organizationID: "",
			expectedStatus: http.StatusBadRequest,
			description:    "throw 400 when organization header is missing",
		},
		{
			name:           "on update of non-existent feature flag, create new one",
			featureName:    "non_existent_feature",
			isEnabled:      true,
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "create new feature flag for non-existent feature name",
		},
		{
			name:           "update feature flag with empty name",
			featureName:    "",
			isEnabled:      true,
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusBadRequest,
			description:    "throws 400 when feature name is empty",
		},
		{
			name:           "cross-organization update attempt",
			featureName:    "terminal",
			isEnabled:      true,
			token:          user.AccessToken,
			organizationID: "123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusForbidden,
			description:    "throw error for updating feature flags from different organization",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestBody := map[string]interface{}{
				"feature_name": tc.featureName,
				"is_enabled":   tc.isEnabled,
			}

			testSteps := []IStep{
				Description(tc.description),
				Put(tests.GetFeatureFlagsURL()),
				Send().Body().JSON(requestBody),
			}

			if tc.token != "" {
				testSteps = append(testSteps, Send().Headers("Authorization").Add("Bearer "+tc.token))
			}

			if tc.organizationID != "" {
				testSteps = append(testSteps, Send().Headers("X-Organization-Id").Add(tc.organizationID))
			}

			testSteps = append(testSteps, Expect().Status().Equal(int64(tc.expectedStatus)))

			if tc.expectedStatus == http.StatusOK {
				testSteps = append(testSteps,
					Expect().Body().JSON().JQ(".status").Equal("success"),
					Expect().Body().JSON().JQ(".message").Equal("Feature flag updated successfully"),
				)
			}

			Test(t, testSteps...)
		})
	}
}

func TestIsFeatureEnabled(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	// First, ensure feature flags exist and set known states
	Test(t,
		Description("Initialize feature flags"),
		Get(tests.GetFeatureFlagsURL()),
		Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
		Send().Headers("X-Organization-Id").Add(orgID),
		Expect().Status().Equal(http.StatusOK),
	)

	// Enable terminal feature for testing
	Test(t,
		Description("enable terminal feature for testing"),
		Put(tests.GetFeatureFlagsURL()),
		Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
		Send().Headers("X-Organization-Id").Add(orgID),
		Send().Body().JSON(map[string]interface{}{
			"feature_name": "terminal",
			"is_enabled":   true,
		}),
		Expect().Status().Equal(http.StatusOK),
	)

	// Disable container feature for testing
	Test(t,
		Description("disable container feature for testing"),
		Put(tests.GetFeatureFlagsURL()),
		Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
		Send().Headers("X-Organization-Id").Add(orgID),
		Send().Body().JSON(map[string]interface{}{
			"feature_name": "container",
			"is_enabled":   false,
		}),
		Expect().Status().Equal(http.StatusOK),
	)

	testCases := []struct {
		name           string
		featureName    string
		token          string
		organizationID string
		expectedStatus int
		expectedResult bool
		description    string
	}{
		{
			name:           "Check enabled feature (terminal)",
			featureName:    "terminal",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			expectedResult: true,
			description:    "return true for enabled terminal feature",
		},
		{
			name:           "Check disabled feature (container)",
			featureName:    "container",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			expectedResult: false,
			description:    " return false for disabled container feature",
		},
		{
			name:           "Check non-existent feature",
			featureName:    "non_existent_feature",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK, // API returns default enabled state instead of 404
			expectedResult: true,          // Default state is enabled
			description:    "return default enabled state for non-existent feature",
		},
		{
			name:           "Unauthorized request without token",
			featureName:    "terminal",
			token:          "",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "throw 401 when no authentication token is provided",
		},
		{
			name:           "Unauthorized request with invalid token",
			featureName:    "terminal",
			token:          "invalid-token",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "throw 401 when invalid authentication token is provided",
		},
		{
			name:           "Request without organization header",
			featureName:    "terminal",
			token:          user.AccessToken,
			organizationID: "",
			expectedStatus: http.StatusBadRequest,
			description:    "throw 400 when organization header is missing",
		},
		{
			name:           "Check feature with empty name",
			featureName:    "",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK, // API returns default state instead of 400
			expectedResult: true,          // Default state when no feature name provided
			description:    "return default state when feature name is empty",
		},
		{
			name:           "Cross-organization feature check",
			featureName:    "terminal",
			token:          user.AccessToken,
			organizationID: "123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusForbidden,
			description:    "deny checking feature flags from different organization",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := tests.GetFeatureFlagCheckURL()
			if tc.featureName != "" {
				url += "?feature_name=" + tc.featureName
			}

			testSteps := []IStep{
				Description(tc.description),
				Get(url),
			}

			if tc.token != "" {
				testSteps = append(testSteps, Send().Headers("Authorization").Add("Bearer "+tc.token))
			}

			if tc.organizationID != "" {
				testSteps = append(testSteps, Send().Headers("X-Organization-Id").Add(tc.organizationID))
			}

			testSteps = append(testSteps, Expect().Status().Equal(int64(tc.expectedStatus)))

			if tc.expectedStatus == http.StatusOK {
				testSteps = append(testSteps,
					Expect().Body().JSON().JQ(".status").Equal("success"),
					Expect().Body().JSON().JQ(".data.is_enabled").Equal(tc.expectedResult),
				)
			}

			Test(t, testSteps...)
		})
	}
}

func TestFeatureFlagsCRUDFlow(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	t.Run("Validate CRUD flow for feature flags", func(t *testing.T) {
		// Step 1: Get initial feature flags (should create defaults)
		Test(t,
			Description("Step 1: Get initial feature flags - should create defaults"),
			Get(tests.GetFeatureFlagsURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".status").Equal("success"),
			Expect().Body().JSON().JQ(".data").NotEqual(nil),
		)

		// Step 2: Check initial state of terminal feature (should be enabled by default)
		Test(t,
			Description("Step 2: Check initial state of terminal feature"),
			Get(tests.GetFeatureFlagCheckURL()+"?feature_name=terminal"),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".data.is_enabled").Equal(true),
		)

		// Step 3: Disable terminal feature
		Test(t,
			Description("Step 3: Disable terminal feature"),
			Put(tests.GetFeatureFlagsURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(map[string]interface{}{
				"feature_name": "terminal",
				"is_enabled":   false,
			}),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".status").Equal("success"),
		)

		// Step 4: Verify terminal feature is now disabled
		Test(t,
			Description("Step 4: Verify terminal feature is now disabled"),
			Get(tests.GetFeatureFlagCheckURL()+"?feature_name=terminal"),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".data.is_enabled").Equal(false),
		)

		// Step 5: Re-enable terminal feature
		Test(t,
			Description("Step 5: Re-enable terminal feature"),
			Put(tests.GetFeatureFlagsURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(map[string]interface{}{
				"feature_name": "terminal",
				"is_enabled":   true,
			}),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".status").Equal("success"),
		)

		// Step 6: Verify terminal feature is enabled again
		Test(t,
			Description("Step 6: Verify terminal feature is enabled again"),
			Get(tests.GetFeatureFlagCheckURL()+"?feature_name=terminal"),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".data.is_enabled").Equal(true),
		)

		// Step 7: Get all feature flags and verify terminal is in the list and enabled
		Test(t,
			Description("Step 7: Get all feature flags and verify terminal state"),
			Get(tests.GetFeatureFlagsURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".data | map(select(.feature_name == \"terminal\")) | .[0].is_enabled").Equal(true),
		)
	})
}

func TestFeatureFlagPermissions(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	t.Run("Feature flag permissions and organization isolation", func(t *testing.T) {
		// Initialize feature flags in user's organization
		Test(t,
			Description("Initialize feature flags in user's organization"),
			Get(tests.GetFeatureFlagsURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusOK),
		)

		// Try to access feature flags with different organization ID
		Test(t,
			Description("deny access to feature flags from different organization"),
			Get(tests.GetFeatureFlagsURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add("123e4567-e89b-12d3-a456-426614174000"),
			Expect().Status().Equal(http.StatusForbidden),
		)

		// Try to update feature flag from different organization
		Test(t,
			Description("deny feature flag update from different organization"),
			Put(tests.GetFeatureFlagsURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add("123e4567-e89b-12d3-a456-426614174000"),
			Send().Body().JSON(map[string]interface{}{
				"feature_name": "terminal",
				"is_enabled":   false,
			}),
			Expect().Status().Equal(http.StatusForbidden),
		)

		// Try to check feature flag from different organization
		Test(t,
			Description("deny feature flag check from different organization"),
			Get(tests.GetFeatureFlagCheckURL()+"?feature_name=terminal"),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add("123e4567-e89b-12d3-a456-426614174000"),
			Expect().Status().Equal(http.StatusForbidden),
		)
	})
}

func TestFeatureFlagErrorHandling(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	t.Run("Malformed authorization header", func(t *testing.T) {
		Test(t,
			Description("handle malformed authorization header gracefully"),
			Get(tests.GetFeatureFlagsURL()),
			Send().Headers("Authorization").Add("InvalidFormat"),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusUnauthorized),
		)
	})

	t.Run("Empty authorization header", func(t *testing.T) {
		Test(t,
			Description("handle empty authorization header"),
			Get(tests.GetFeatureFlagsURL()),
			Send().Headers("Authorization").Add(""),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusUnauthorized),
		)
	})

	t.Run("Missing Content-Type header for PUT requests", func(t *testing.T) {
		Test(t,
			Description("handle missing Content-Type header"),
			Put(tests.GetFeatureFlagsURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(map[string]interface{}{
				"feature_name": "terminal",
				"is_enabled":   true,
			}),
			Expect().Status().Equal(http.StatusOK),
		)
	})

	t.Run("Invalid JSON payload", func(t *testing.T) {
		Test(t,
			Description("handle invalid JSON payload"),
			Put(tests.GetFeatureFlagsURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().String("{invalid-json}"),
			Expect().Status().Equal(http.StatusBadRequest),
		)
	})

	t.Run("Very long feature name", func(t *testing.T) {
		longFeatureName := ""
		for i := 0; i < 300; i++ {
			longFeatureName += "a"
		}

		Test(t,
			Description("handle very long feature names"),
			Put(tests.GetFeatureFlagsURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(map[string]interface{}{
				"feature_name": longFeatureName,
				"is_enabled":   true,
			}),
			Expect().Status().Equal(http.StatusInternalServerError), // Database returns 500 for varchar length constraint
		)
	})

	t.Run("Special characters in feature name", func(t *testing.T) {
		Test(t,
			Description("handle special characters in feature name"),
			Put(tests.GetFeatureFlagsURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(map[string]interface{}{
				"feature_name": "feature@#$%^&*()",
				"is_enabled":   true,
			}),
			Expect().Status().Equal(http.StatusOK),
		)
	})
}
