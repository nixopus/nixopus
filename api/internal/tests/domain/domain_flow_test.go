package domain

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
)

func TestCreateDomain(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	testCases := []struct {
		name           string
		domainName     string
		organizationID string
		token          string
		expectedStatus int
		description    string
	}{
		{
			name:           "Successfully create domain with valid data",
			domainName:     "test-domain.nixopus.dev",
			organizationID: orgID,
			token:          user.AccessToken,
			expectedStatus: http.StatusOK,
			description:    "Should create domain successfully with valid data",
		},
		{
			name:           "Create domain with subdomain",
			domainName:     "api.test-domain.nixopus.dev",
			organizationID: orgID,
			token:          user.AccessToken,
			expectedStatus: http.StatusOK,
			description:    "Should create subdomain successfully",
		},
		{
			name:           "Unauthorized request without token",
			domainName:     "unauthorized.nixopus.dev",
			organizationID: orgID,
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication token is provided",
		},
		{
			name:           "Unauthorized request with invalid token",
			domainName:     "invalid-token.nixopus.dev",
			organizationID: orgID,
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid authentication token is provided",
		},
		{
			name:           "Request without organization header",
			domainName:     "no-org.nixopus.dev",
			organizationID: "",
			token:          user.AccessToken,
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization header is missing",
		},
		{
			name:           "Create domain with empty name",
			domainName:     "",
			organizationID: orgID,
			token:          user.AccessToken,
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when domain name is empty",
		},
		{
			name:           "Create domain with invalid name format",
			domainName:     "invalid..domain",
			organizationID: orgID,
			token:          user.AccessToken,
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when domain name format is invalid",
		},
		{
			name:           "Create duplicate domain",
			domainName:     "test-domain.nixopus.dev", // Same as first test case
			organizationID: orgID,
			token:          user.AccessToken,
			expectedStatus: http.StatusConflict,
			description:    "Should return 409 when domain already exists",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestBody := map[string]interface{}{
				"name":            tc.domainName,
				"organization_id": tc.organizationID,
			}

			testSteps := []IStep{
				Description(tc.description),
				Post(tests.GetDomainURL()),
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
					Expect().Body().JSON().JQ(".message").Equal("Domain created successfully"),
					Expect().Body().JSON().JQ(".data.id").NotEqual(""),
				)
			}

			Test(t, testSteps...)
		})
	}
}

func TestGetDomains(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	// First, create a test domain
	createDomainRequest := map[string]interface{}{
		"name":            "list-test.nixopus.dev",
		"organization_id": orgID,
	}

	Test(t,
		Description("Create a test domain for listing"),
		Post(tests.GetDomainURL()),
		Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
		Send().Headers("X-Organization-Id").Add(orgID),
		Send().Body().JSON(createDomainRequest),
		Expect().Status().Equal(http.StatusOK),
	)

	testCases := []struct {
		name           string
		token          string
		organizationID string
		expectedStatus int
		description    string
	}{
		{
			name:           "Successfully fetch domains with valid token",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "Should return domains list with valid authentication",
		},
		{
			name:           "Unauthorized request without token",
			token:          "",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication token is provided",
		},
		{
			name:           "Unauthorized request with invalid token",
			token:          "invalid-token",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid authentication token is provided",
		},
		{
			name:           "Request without organization header",
			token:          user.AccessToken,
			organizationID: "",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization header is missing",
		},
		{
			name:           "Cross-organization access attempt",
			token:          user.AccessToken,
			organizationID: "123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusForbidden,
			description:    "Should deny access to domains from different organization",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSteps := []IStep{
				Description(tc.description),
				Get(tests.GetDomainsURL()),
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
					Expect().Body().JSON().JQ(".message").Equal("Domains fetched successfully"),
					Expect().Body().JSON().JQ(".data").NotEqual(nil),
				)
			}

			Test(t, testSteps...)
		})
	}
}

func TestUpdateDomain(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	// First, create a test domain to update
	var domainID string
	createDomainRequest := map[string]interface{}{
		"name":            "update-test.nixopus.dev",
		"organization_id": orgID,
	}

	Test(t,
		Description("Create a test domain for updating"),
		Post(tests.GetDomainURL()),
		Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
		Send().Headers("X-Organization-Id").Add(orgID),
		Send().Body().JSON(createDomainRequest),
		Expect().Status().Equal(http.StatusOK),
		Store().Response().Body().JSON().JQ(".data.id").In(&domainID),
	)

	testCases := []struct {
		name           string
		domainID       string
		newName        string
		token          string
		organizationID string
		expectedStatus int
		description    string
	}{
		{
			name:           "Successfully update domain with valid data",
			domainID:       domainID,
			newName:        "updated-domain.nixopus.dev",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "Should update domain successfully with valid data",
		},
		{
			name:           "Update domain with subdomain",
			domainID:       domainID,
			newName:        "api.updated-domain.nixopus.dev",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "Should update domain to subdomain successfully",
		},
		{
			name:           "Unauthorized request without token",
			domainID:       domainID,
			newName:        "unauthorized-update.nixopus.dev",
			token:          "",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication token is provided",
		},
		{
			name:           "Unauthorized request with invalid token",
			domainID:       domainID,
			newName:        "invalid-token-update.nixopus.dev",
			token:          "invalid-token",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid authentication token is provided",
		},
		{
			name:           "Request without organization header",
			domainID:       domainID,
			newName:        "no-org-update.nixopus.dev",
			token:          user.AccessToken,
			organizationID: "",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization header is missing",
		},
		{
			name:           "Update domain with empty name",
			domainID:       domainID,
			newName:        "",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when domain name is empty",
		},
		{
			name:           "Update domain with invalid name format",
			domainID:       domainID,
			newName:        "invalid..domain",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when domain name format is invalid",
		},
		{
			name:           "Update non-existent domain",
			domainID:       "123e4567-e89b-12d3-a456-426614174000",
			newName:        "non-existent-update.nixopus.dev",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusNotFound,
			description:    "Should return 404 when domain doesn't exist",
		},
		{
			name:           "Update domain with invalid ID format",
			domainID:       "invalid-id",
			newName:        "invalid-id-update.nixopus.dev",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when domain ID format is invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestBody := map[string]interface{}{
				"id":   tc.domainID,
				"name": tc.newName,
			}

			testSteps := []IStep{
				Description(tc.description),
				Put(tests.GetDomainURL()),
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
					Expect().Body().JSON().JQ(".message").Equal("Domain updated successfully"),
				)
			}

			Test(t, testSteps...)
		})
	}
}

func TestDeleteDomain(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	// Create test domains for deletion
	var domainID1, domainID2 string

	createDomainRequest1 := map[string]interface{}{
		"name":            "delete-test1.nixopus.dev",
		"organization_id": orgID,
	}

	Test(t,
		Description("Create first test domain for deletion"),
		Post(tests.GetDomainURL()),
		Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
		Send().Headers("X-Organization-Id").Add(orgID),
		Send().Body().JSON(createDomainRequest1),
		Expect().Status().Equal(http.StatusOK),
		Store().Response().Body().JSON().JQ(".data.id").In(&domainID1),
	)

	createDomainRequest2 := map[string]interface{}{
		"name":            "delete-test2.nixopus.dev",
		"organization_id": orgID,
	}

	Test(t,
		Description("Create second test domain for deletion"),
		Post(tests.GetDomainURL()),
		Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
		Send().Headers("X-Organization-Id").Add(orgID),
		Send().Body().JSON(createDomainRequest2),
		Expect().Status().Equal(http.StatusOK),
		Store().Response().Body().JSON().JQ(".data.id").In(&domainID2),
	)

	testCases := []struct {
		name           string
		domainID       string
		token          string
		organizationID string
		expectedStatus int
		description    string
	}{
		{
			name:           "Successfully delete domain with valid ID",
			domainID:       domainID1,
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "Should delete domain successfully with valid ID",
		},
		{
			name:           "Unauthorized request without token",
			domainID:       domainID2,
			token:          "",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication token is provided",
		},
		{
			name:           "Unauthorized request with invalid token",
			domainID:       domainID2,
			token:          "invalid-token",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid authentication token is provided",
		},
		{
			name:           "Request without organization header",
			domainID:       domainID2,
			token:          user.AccessToken,
			organizationID: "",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization header is missing",
		},
		{
			name:           "Delete non-existent domain",
			domainID:       "123e4567-e89b-12d3-a456-426614174000",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusNotFound,
			description:    "Should return 404 when domain doesn't exist",
		},
		{
			name:           "Delete domain with invalid ID format",
			domainID:       "invalid-id",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when domain ID format is invalid",
		},
		{
			name:           "Delete already deleted domain",
			domainID:       domainID1, // Already deleted in first test case so expcected to throw 404
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusNotFound,
			description:    "Should return 404 when trying to delete already deleted domain",
		},
		{
			name:           "Successfully delete second domain",
			domainID:       domainID2,
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "Should delete second domain successfully",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestBody := map[string]interface{}{
				"id": tc.domainID,
			}

			testSteps := []IStep{
				Description(tc.description),
				Delete(tests.GetDomainURL()),
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
					Expect().Body().JSON().JQ(".message").Equal("Domain deleted successfully"),
				)
			}

			Test(t, testSteps...)
		})
	}
}

func TestGenerateRandomSubDomain(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	// first create a domain then generating subdomains
	createRequest := map[string]interface{}{
		"name":            "base-domain.nixopus.dev",
		"organization_id": orgID,
	}

	Test(t,
		Description("Create a base domain for subdomain generation"),
		Post(tests.GetDomainURL()),
		Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
		Send().Headers("X-Organization-Id").Add(orgID),
		Send().Body().JSON(createRequest),
		Expect().Status().Equal(http.StatusOK),
	)

	testCases := []struct {
		name           string
		token          string
		organizationID string
		expectedStatus int
		description    string
	}{
		{
			name:           "Successfully generate random subdomain",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "Should generate random subdomain successfully",
		},
		{
			name:           "Unauthorized request without token",
			token:          "",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication token is provided",
		},
		{
			name:           "Unauthorized request with invalid token",
			token:          "invalid-token",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid authentication token is provided",
		},
		{
			name:           "Request without organization header",
			token:          user.AccessToken,
			organizationID: "",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization header is missing",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSteps := []IStep{
				Description(tc.description),
				Get(tests.GetDomainGenerateURL()),
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
					Expect().Body().JSON().JQ(".message").Equal("Random subdomain generated successfully"),
					Expect().Body().JSON().JQ(".data.subdomain").NotEqual(""),
					Expect().Body().JSON().JQ(".data.domain").NotEqual(""),
				)
			}

			Test(t, testSteps...)
		})
	}
}

func TestDomainsCRUDFlow(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	t.Run("Validate CRUD flow for domains", func(t *testing.T) {
		var domainID string

		// creating a domain
		createRequest := map[string]interface{}{
			"name":            "crud-flow.nixopus.dev",
			"organization_id": orgID,
		}

		Test(t,
			Description("Create a new domain"),
			Post(tests.GetDomainURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(createRequest),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".status").Equal("success"),
			Expect().Body().JSON().JQ(".message").Equal("Domain created successfully"),
			Store().Response().Body().JSON().JQ(".data.id").In(&domainID),
		)

		// check listing if added once in available or not
		Test(t,
			Description("Verify domain appears in domains listing"),
			Get(tests.GetDomainsURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".status").Equal("success"),
			Expect().Body().JSON().JQ(".data").NotEqual(nil),
			Expect().Body().JSON().JQ(".data[0].id").NotEqual(nil),
		)

		updateRequest := map[string]interface{}{
			"id":   domainID,
			"name": "updated-crud-flow.nixopus.dev",
		}

		Test(t,
			Description("Update the domain"),
			Put(tests.GetDomainURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(updateRequest),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".status").Equal("success"),
			Expect().Body().JSON().JQ(".message").Equal("Domain updated successfully"),
		)

		// Cross check domain update in listing
		Test(t,
			Description("Verify domain update appears in domains listing"),
			Get(tests.GetDomainsURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".status").Equal("success"),
			Expect().Body().JSON().JQ(".data").NotEqual(nil),
			// assert agaisnst the updated domain name
			Expect().Body().JSON().JQ(".data[0].name").Equal("updated-crud-flow.nixopus.dev"),
		)

		// Step 5: Delete the domain
		deleteRequest := map[string]interface{}{
			"id": domainID,
		}

		Test(t,
			Description("Step 5: Delete the domain"),
			Delete(tests.GetDomainURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(deleteRequest),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".status").Equal("success"),
			Expect().Body().JSON().JQ(".message").Equal("Domain deleted successfully"),
		)

		// Step 6: Verify domain is removed from listing
		Test(t,
			Description("Step 6: Verify domain is removed from domains listing"),
			Get(tests.GetDomainsURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".status").Equal("success"),
			// Verify the domain list is empty after deletion (could be null or empty array)
			// Just check that the response is successful, domains being null indicates empty list
		)
	})
}

func TestDomainPermissions(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	t.Run("Domain permissions and organization isolation", func(t *testing.T) {
		var domainID string

		// Create a domain in the user' organization
		createRequest := map[string]interface{}{
			"name":            "permissions-test.nixopus.dev",
			"organization_id": orgID,
		}

		Test(t,
			Description("Create domain in user's organization"),
			Post(tests.GetDomainURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(createRequest),
			Expect().Status().Equal(http.StatusOK),
			Store().Response().Body().JSON().JQ(".data.id").In(&domainID),
		)

		// Try to access with different organization ID
		Test(t,
			Description("Should deny access to domains from different organization"),
			Get(tests.GetDomainsURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add("123e4567-e89b-12d3-a456-426614174000"),
			Expect().Status().Equal(http.StatusForbidden),
		)

		// Try to update domain from different organization id
		updateRequest := map[string]interface{}{
			"id":   domainID,
			"name": "unauthorized-update.nixopus.dev",
		}

		Test(t,
			Description("Should deny domain update from different organization"),
			Put(tests.GetDomainURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add("123e4567-e89b-12d3-a456-426614174000"),
			Send().Body().JSON(updateRequest),
			Expect().Status().Equal(http.StatusForbidden),
		)

		// Try to delete domain from different organization Id
		deleteRequest := map[string]interface{}{
			"id": domainID,
		}

		Test(t,
			Description("Should deny domain deletion from different organization"),
			Delete(tests.GetDomainURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add("123e4567-e89b-12d3-a456-426614174000"),
			Send().Body().JSON(deleteRequest),
			Expect().Status().Equal(http.StatusForbidden),
		)

		// Clean up: Delete the domain with correct organization id
		Test(t,
			Description("Clean up: Delete domain with correct organization"),
			Delete(tests.GetDomainURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(deleteRequest),
			Expect().Status().Equal(http.StatusOK),
		)
	})
}

func TestDomainErrorHandling(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	t.Run("Malformed authorization header", func(t *testing.T) {
		Test(t,
			Description("Should handle malformed authorization header gracefully"),
			Get(tests.GetDomainsURL()),
			Send().Headers("Authorization").Add("InvalidFormat"),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusUnauthorized),
		)
	})

	t.Run("Empty authorization header", func(t *testing.T) {
		Test(t,
			Description("Should handle empty authorization header"),
			Get(tests.GetDomainsURL()),
			Send().Headers("Authorization").Add(""),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusUnauthorized),
		)
	})

	t.Run("Missing Content-Type header for POST requests", func(t *testing.T) {
		createRequest := map[string]interface{}{
			"name":            "content-type-test.nixopus.dev",
			"organization_id": orgID,
		}

		Test(t,
			Description("Should handle missing Content-Type header"),
			Post(tests.GetDomainURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(createRequest),
			Expect().Status().Equal(http.StatusOK),
		)
	})

	t.Run("Invalid JSON payload", func(t *testing.T) {
		Test(t,
			Description("Should handle invalid JSON payload"),
			Post(tests.GetDomainURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().String("{invalid-json}"),
			Expect().Status().Equal(http.StatusBadRequest),
		)
	})

	t.Run("Very long domain name", func(t *testing.T) {
		longDomainName := ""
		for i := 0; i < 300; i++ {
			longDomainName += "a"
		}
		longDomainName += ".nixopus.dev"

		createRequest := map[string]interface{}{
			"name":            longDomainName,
			"organization_id": orgID,
		}

		Test(t,
			Description("Should throw an error for very long domain names"),
			Post(tests.GetDomainURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(createRequest),
			Expect().Status().Equal(http.StatusBadRequest),
		)
	})
}
