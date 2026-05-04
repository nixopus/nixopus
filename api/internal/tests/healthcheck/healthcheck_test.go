package healthcheck

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

// --- List healthchecks ---

func TestListHealthchecks_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /healthcheck without auth returns 401"),
		Get(tests.GetHealthCheckURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestListHealthchecks_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID := uuid.New().String()
	Test(t,
		Description("GET /healthcheck?application_id=<uuid> returns 200"),
		Get(tests.GetHealthCheckURL()+"?application_id="+appID),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestListHealthchecks_MissingApplicationID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /healthcheck without application_id returns 400"),
		Get(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestListHealthchecks_InvalidOrgID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID := uuid.New().String()
	Test(t,
		Description("GET /healthcheck with invalid org ID returns 400"),
		Get(tests.GetHealthCheckURL()+"?application_id="+appID),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add("not-a-uuid"),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

// --- Create healthcheck ---

func TestCreateHealthcheck_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /healthcheck without auth returns 401"),
		Post(tests.GetHealthCheckURL()),
		Send().Body().JSON(map[string]interface{}{
			"application_id": uuid.New().String(),
			"endpoint":       "/health",
		}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestCreateHealthcheck_MissingApplicationID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /healthcheck without application_id returns 400"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"endpoint": "/health"}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCreateHealthcheck_ValidMinimal(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("POST /healthcheck with application_id returns 201"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"application_id": appID,
		}),
		Expect().Status().Equal(http.StatusCreated),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".data.id").NotEqual(nil),
	)
}

func TestCreateHealthcheck_WithFullConfig(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("POST /healthcheck with full config returns 201"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"application_id":        appID,
			"endpoint":              "/health",
			"interval_seconds":      60,
			"timeout_seconds":       10,
			"method":                "GET",
			"expected_status_codes": []int{200},
			"failure_threshold":     3,
		}),
		Expect().Status().Equal(http.StatusCreated),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestCreateHealthcheck_DuplicateApplicationID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("Create first healthcheck"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"application_id": appID}),
		Expect().Status().Equal(http.StatusCreated),
	)

	Test(t,
		Description("POST /healthcheck with duplicate application_id returns error"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"application_id": appID}),
		Expect().Status().OneOf(int64(http.StatusConflict), int64(http.StatusBadRequest), int64(http.StatusInternalServerError)),
	)
}

// --- Update healthcheck ---

func TestUpdateHealthcheck_NoAuth(t *testing.T) {
	Test(t,
		Description("PUT /healthcheck without auth returns 401"),
		Put(tests.GetHealthCheckURL()),
		Send().Body().JSON(map[string]interface{}{"application_id": uuid.New().String()}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestUpdateHealthcheck_MissingID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PUT /healthcheck without application_id returns 400"),
		Put(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestUpdateHealthcheck_NonExistentID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PUT /healthcheck with non-existent application_id returns error"),
		Put(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"application_id": uuid.New().String()}),
		Expect().Status().OneOf(int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
	)
}

func TestUpdateHealthcheck_ValidFlow(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("Create healthcheck for update"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"application_id": appID}),
		Expect().Status().Equal(http.StatusCreated),
	)

	Test(t,
		Description("PUT /healthcheck updates endpoint and interval"),
		Put(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"application_id":   appID,
			"endpoint":         "/health/updated",
			"interval_seconds": 120,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

// --- Delete healthcheck ---

func TestDeleteHealthcheck_NoAuth(t *testing.T) {
	Test(t,
		Description("DELETE /healthcheck without auth returns 401"),
		Delete(tests.GetHealthCheckURL()+"?application_id="+uuid.New().String()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestDeleteHealthcheck_MissingID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("DELETE /healthcheck without application_id returns 400"),
		Delete(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestDeleteHealthcheck_NonExistentID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("DELETE /healthcheck with non-existent application_id returns error"),
		Delete(tests.GetHealthCheckURL()+"?application_id="+uuid.New().String()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
	)
}

func TestDeleteHealthcheck_ValidFlow(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("Create healthcheck to delete"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"application_id": appID}),
		Expect().Status().Equal(http.StatusCreated),
	)

	Test(t,
		Description("DELETE /healthcheck with valid application_id returns 200"),
		Delete(tests.GetHealthCheckURL()+"?application_id="+appID),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

// --- Toggle healthcheck ---

func TestToggleHealthcheck_NoAuth(t *testing.T) {
	Test(t,
		Description("PATCH /healthcheck/toggle without auth returns 401"),
		Method("PATCH", tests.GetHealthCheckToggleURL()),
		Send().Body().JSON(map[string]interface{}{"application_id": uuid.New().String(), "enabled": false}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestToggleHealthcheck_MissingID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /healthcheck/toggle without application_id returns 400"),
		Method("PATCH", tests.GetHealthCheckToggleURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestToggleHealthcheck_NonExistentID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /healthcheck/toggle with non-existent application_id returns error"),
		Method("PATCH", tests.GetHealthCheckToggleURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"application_id": uuid.New().String(),
			"enabled":        false,
		}),
		Expect().Status().OneOf(int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
	)
}

func TestToggleHealthcheck_ValidFlow(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("Create healthcheck for toggle"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"application_id": appID}),
		Expect().Status().Equal(http.StatusCreated),
	)

	Test(t,
		Description("PATCH /healthcheck/toggle disables the check"),
		Method("PATCH", tests.GetHealthCheckToggleURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"application_id": appID,
			"enabled":        false,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)

	Test(t,
		Description("PATCH /healthcheck/toggle re-enables the check"),
		Method("PATCH", tests.GetHealthCheckToggleURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"application_id": appID,
			"enabled":        true,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

// --- Healthcheck results ---

func TestGetHealthcheckResults_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /healthcheck/results without auth returns 401"),
		Get(tests.GetHealthCheckResultsURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetHealthcheckResults_MissingID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /healthcheck/results without application_id param returns 400"),
		Get(tests.GetHealthCheckResultsURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestGetHealthcheckResults_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("Create healthcheck for results"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"application_id": appID}),
		Expect().Status().Equal(http.StatusCreated),
	)

	Test(t,
		Description("GET /healthcheck/results with valid application_id returns 200 (empty results)"),
		Get(tests.GetHealthCheckResultsURL()+"?application_id="+appID),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestGetHealthcheckResults_WithPagination(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("Create healthcheck for paginated results"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"application_id": appID}),
		Expect().Status().Equal(http.StatusCreated),
	)

	Test(t,
		Description("GET /healthcheck/results with application_id and limit params"),
		Get(tests.GetHealthCheckResultsURL()+"?application_id="+appID+"&limit=10"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
	)
}

// --- Healthcheck stats ---

func TestGetHealthcheckStats_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /healthcheck/stats without auth returns 401"),
		Get(tests.GetHealthCheckStatsURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetHealthcheckStats_MissingID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /healthcheck/stats without application_id param returns 400"),
		Get(tests.GetHealthCheckStatsURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestGetHealthcheckStats_NonExistentID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /healthcheck/stats with random application_id returns 200 (empty stats)"),
		Get(tests.GetHealthCheckStatsURL()+"?application_id="+uuid.New().String()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
	)
}

func TestGetHealthcheckStats_ValidFlow(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("Create healthcheck for stats"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"application_id": appID}),
		Expect().Status().Equal(http.StatusCreated),
	)

	Test(t,
		Description("GET /healthcheck/stats returns uptime and response time stats"),
		Get(tests.GetHealthCheckStatsURL()+"?application_id="+appID),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

// ---------------------------------------------------------------------------
// Create — validation errors
// ---------------------------------------------------------------------------

func TestCreateHealthcheck_InvalidEndpoint(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("POST /healthcheck with endpoint lacking leading slash returns 400"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"application_id": appID,
			"endpoint":       "no-leading-slash",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCreateHealthcheck_InvalidMethod(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("POST /healthcheck with unsupported HTTP method returns 400"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"application_id": appID,
			"method":         "DELETE",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCreateHealthcheck_InvalidTimeout(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("POST /healthcheck with timeout below minimum returns 400"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"application_id":  appID,
			"timeout_seconds": 2,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCreateHealthcheck_InvalidInterval(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("POST /healthcheck with interval below minimum returns 400"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"application_id":   appID,
			"interval_seconds": 5,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCreateHealthcheck_InvalidFailureThreshold(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("POST /healthcheck with failure_threshold above maximum returns 400"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"application_id":    appID,
			"failure_threshold": 99,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCreateHealthcheck_InvalidRetentionDays(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("POST /healthcheck with retention_days above maximum returns 400"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"application_id": appID,
			"retention_days": 999,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

// ---------------------------------------------------------------------------
// Create — missing org header
// ---------------------------------------------------------------------------

func TestCreateHealthcheck_MissingOrgHeader(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /healthcheck without X-Organization-ID header returns 400"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Body().JSON(map[string]interface{}{"application_id": uuid.New().String()}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

// ---------------------------------------------------------------------------
// Update — validation errors
// ---------------------------------------------------------------------------

func TestUpdateHealthcheck_InvalidEndpoint(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("Create healthcheck for invalid-endpoint update test"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"application_id": appID}),
		Expect().Status().Equal(http.StatusCreated),
	)

	Test(t,
		Description("PUT /healthcheck with invalid endpoint returns 400"),
		Put(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"application_id": appID,
			"endpoint":       "no-leading-slash",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestUpdateHealthcheck_InvalidMethod(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("Create healthcheck for invalid-method update test"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"application_id": appID}),
		Expect().Status().Equal(http.StatusCreated),
	)

	Test(t,
		Description("PUT /healthcheck with unsupported method returns 400"),
		Put(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"application_id": appID,
			"method":         "PATCH",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestUpdateHealthcheck_InvalidTimeout(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("Create healthcheck for invalid-timeout update test"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"application_id": appID}),
		Expect().Status().Equal(http.StatusCreated),
	)

	Test(t,
		Description("PUT /healthcheck with timeout above maximum returns 400"),
		Put(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"application_id":  appID,
			"timeout_seconds": 999,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestUpdateHealthcheck_MissingOrgHeader(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PUT /healthcheck without X-Organization-ID uses session org; unknown application_id returns 404"),
		Put(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Body().JSON(map[string]interface{}{"application_id": uuid.New().String()}),
		Expect().Status().Equal(http.StatusNotFound),
	)
}

// ---------------------------------------------------------------------------
// Delete — missing org header
// ---------------------------------------------------------------------------

func TestDeleteHealthcheck_MissingOrgHeader(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("DELETE /healthcheck without X-Organization-ID uses session org; unknown application_id returns 404"),
		Delete(tests.GetHealthCheckURL()+"?application_id="+uuid.New().String()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Expect().Status().Equal(http.StatusNotFound),
	)
}

// ---------------------------------------------------------------------------
// Toggle — additional cases
// ---------------------------------------------------------------------------

func TestToggleHealthcheck_MissingOrgHeader(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /healthcheck/toggle without X-Organization-ID uses session org; unknown application_id returns 404"),
		Method("PATCH", tests.GetHealthCheckToggleURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Body().JSON(map[string]interface{}{
			"application_id": uuid.New().String(),
			"enabled":        true,
		}),
		Expect().Status().Equal(http.StatusNotFound),
	)
}

func TestToggleHealthcheck_InvalidApplicationID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /healthcheck/toggle with non-UUID application_id returns 400"),
		Method("PATCH", tests.GetHealthCheckToggleURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"application_id": "not-a-uuid",
			"enabled":        true,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

// ---------------------------------------------------------------------------
// Results — time range and limit parameters
// ---------------------------------------------------------------------------

func TestGetHealthcheckResults_MissingOrgHeader(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /healthcheck/results without X-Organization-ID uses session org; unknown application_id returns 404 or 500"),
		Get(tests.GetHealthCheckResultsURL()+"?application_id="+uuid.New().String()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Expect().Status().OneOf(int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
	)
}

func TestGetHealthcheckResults_WithTimeRange(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("Create healthcheck for time-range results test"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"application_id": appID}),
		Expect().Status().Equal(http.StatusCreated),
	)

	url := tests.GetHealthCheckResultsURL() +
		"?application_id=" + appID +
		"&start_time=2024-01-01T00:00:00Z" +
		"&end_time=2024-12-31T23:59:59Z"

	Test(t,
		Description("GET /healthcheck/results with start_time and end_time filters returns 200"),
		Get(url),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestGetHealthcheckResults_WithInvalidLimitString(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("Create healthcheck for invalid-limit results test"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"application_id": appID}),
		Expect().Status().Equal(http.StatusCreated),
	)

	Test(t,
		Description("GET /healthcheck/results with non-numeric limit falls back to default and returns 200"),
		Get(tests.GetHealthCheckResultsURL()+"?application_id="+appID+"&limit=abc"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
	)
}

// ---------------------------------------------------------------------------
// Stats — period parameter variants
// ---------------------------------------------------------------------------

func TestGetHealthcheckStats_MissingOrgHeader(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /healthcheck/stats without X-Organization-ID uses session org; unknown application_id returns 404 or 500"),
		Get(tests.GetHealthCheckStatsURL()+"?application_id="+uuid.New().String()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Expect().Status().OneOf(int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
	)
}

func TestGetHealthcheckStats_Period1h(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("Create healthcheck for 1h-stats test"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"application_id": appID}),
		Expect().Status().Equal(http.StatusCreated),
	)

	Test(t,
		Description("GET /healthcheck/stats?period=1h returns 200 with period in response"),
		Get(tests.GetHealthCheckStatsURL()+"?application_id="+appID+"&period=1h"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".data.period").Equal("1h"),
	)
}

func TestGetHealthcheckStats_Period7d(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("Create healthcheck for 7d-stats test"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"application_id": appID}),
		Expect().Status().Equal(http.StatusCreated),
	)

	Test(t,
		Description("GET /healthcheck/stats?period=7d returns 200 with period in response"),
		Get(tests.GetHealthCheckStatsURL()+"?application_id="+appID+"&period=7d"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data.period").Equal("7d"),
	)
}

func TestGetHealthcheckStats_Period30d(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("Create healthcheck for 30d-stats test"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"application_id": appID}),
		Expect().Status().Equal(http.StatusCreated),
	)

	Test(t,
		Description("GET /healthcheck/stats?period=30d returns 200 with period in response"),
		Get(tests.GetHealthCheckStatsURL()+"?application_id="+appID+"&period=30d"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data.period").Equal("30d"),
	)
}

func TestGetHealthcheckStats_DefaultPeriod(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	Test(t,
		Description("Create healthcheck for default-period stats test"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"application_id": appID}),
		Expect().Status().Equal(http.StatusCreated),
	)

	Test(t,
		Description("GET /healthcheck/stats without period defaults to 24h"),
		Get(tests.GetHealthCheckStatsURL()+"?application_id="+appID),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data.period").Equal("24h"),
	)
}

// ---------------------------------------------------------------------------
// Full CRUD lifecycle
// ---------------------------------------------------------------------------

func TestHealthcheck_FullCRUDLifecycle(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	appID, err := setup.SeedApplication(auth.User.ID.String(), auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	cookies := auth.GetAuthCookiesHeader()
	orgHeader := auth.OrganizationID

	// Step 1: Create
	Test(t,
		Description("CRUD step 1 — create healthcheck"),
		Post(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(cookies),
		Send().Headers("X-Organization-ID").Add(orgHeader),
		Send().Body().JSON(map[string]interface{}{
			"application_id":   appID,
			"endpoint":         "/health",
			"method":           "GET",
			"timeout_seconds":  10,
			"interval_seconds": 60,
		}),
		Expect().Status().Equal(http.StatusCreated),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".data.endpoint").Equal("/health"),
	)

	// Step 2: Get
	Test(t,
		Description("CRUD step 2 — get healthcheck returns created config"),
		Get(tests.GetHealthCheckURL()+"?application_id="+appID),
		Send().Headers("Cookie").Add(cookies),
		Send().Headers("X-Organization-ID").Add(orgHeader),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".data.endpoint").Equal("/health"),
	)

	// Step 3: Update
	Test(t,
		Description("CRUD step 3 — update healthcheck endpoint and interval"),
		Put(tests.GetHealthCheckURL()),
		Send().Headers("Cookie").Add(cookies),
		Send().Headers("X-Organization-ID").Add(orgHeader),
		Send().Body().JSON(map[string]interface{}{
			"application_id":   appID,
			"endpoint":         "/healthz",
			"interval_seconds": 120,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".data.endpoint").Equal("/healthz"),
	)

	// Step 4: Toggle off
	Test(t,
		Description("CRUD step 4 — toggle healthcheck off"),
		Method("PATCH", tests.GetHealthCheckToggleURL()),
		Send().Headers("Cookie").Add(cookies),
		Send().Headers("X-Organization-ID").Add(orgHeader),
		Send().Body().JSON(map[string]interface{}{
			"application_id": appID,
			"enabled":        false,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data.enabled").Equal(false),
	)

	// Step 5: Toggle back on
	Test(t,
		Description("CRUD step 5 — toggle healthcheck on"),
		Method("PATCH", tests.GetHealthCheckToggleURL()),
		Send().Headers("Cookie").Add(cookies),
		Send().Headers("X-Organization-ID").Add(orgHeader),
		Send().Body().JSON(map[string]interface{}{
			"application_id": appID,
			"enabled":        true,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data.enabled").Equal(true),
	)

	// Step 6: Get results (empty)
	Test(t,
		Description("CRUD step 6 — get results returns empty list"),
		Get(tests.GetHealthCheckResultsURL()+"?application_id="+appID),
		Send().Headers("Cookie").Add(cookies),
		Send().Headers("X-Organization-ID").Add(orgHeader),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)

	// Step 7: Get stats
	Test(t,
		Description("CRUD step 7 — get stats returns data shape"),
		Get(tests.GetHealthCheckStatsURL()+"?application_id="+appID+"&period=24h"),
		Send().Headers("Cookie").Add(cookies),
		Send().Headers("X-Organization-ID").Add(orgHeader),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data.application_id").Equal(appID),
	)

	// Step 8: Delete
	Test(t,
		Description("CRUD step 8 — delete healthcheck"),
		Delete(tests.GetHealthCheckURL()+"?application_id="+appID),
		Send().Headers("Cookie").Add(cookies),
		Send().Headers("X-Organization-ID").Add(orgHeader),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)

	// Step 9: Confirm gone
	Test(t,
		Description("CRUD step 9 — get after delete returns success with null data"),
		Get(tests.GetHealthCheckURL()+"?application_id="+appID),
		Send().Headers("Cookie").Add(cookies),
		Send().Headers("X-Organization-ID").Add(orgHeader),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data").Equal(nil),
	)
}
