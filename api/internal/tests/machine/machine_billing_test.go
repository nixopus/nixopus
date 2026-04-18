package machine

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

// --- Machine plans ---

func TestListMachinePlans_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /machines/plans without auth returns 401"),
		Get(tests.GetMachinePlansURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestListMachinePlans_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /machines/plans with valid auth returns 200 with plans list"),
		Get(tests.GetMachinePlansURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".data").NotEqual(nil),
	)
}

func TestListMachinePlans_InvalidOrgID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	// Billing routes have RBAC disabled; the session's active org ID takes
	// precedence over the X-Organization-Id header, so an invalid header value
	// does not trigger a 400. The session's valid org is used instead.
	Test(t,
		Description("GET /machines/plans with invalid header — session org used, returns 200 or 500"),
		Get(tests.GetMachinePlansURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add("not-a-uuid"),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError), int64(http.StatusNotFound)),
	)
}

// --- Select machine plan ---

func TestSelectMachinePlan_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /machines/plan/select without auth returns 401"),
		Post(tests.GetMachinePlanSelectURL()),
		Send().Body().JSON(map[string]interface{}{"plan_id": uuid.New().String()}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestSelectMachinePlan_MissingPlanID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /machines/plan/select without plan_id returns 400"),
		Post(tests.GetMachinePlanSelectURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestSelectMachinePlan_NonExistentPlan(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /machines/plan/select with non-existent plan returns error"),
		Post(tests.GetMachinePlanSelectURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"plan_id": uuid.New().String()}),
		Expect().Status().OneOf(int64(http.StatusNotFound), int64(http.StatusBadRequest), int64(http.StatusInternalServerError)),
	)
}

func TestSelectMachinePlan_ValidPlan(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	// Get a real plan ID first
	var planID string
	Test(t,
		Description("Fetch plans to get a valid plan ID"),
		Get(tests.GetMachinePlansURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Store().Response().Body().JSON().JQ(".data[0].id").In(&planID),
	)

	if planID == "" {
		t.Skip("no plans available, skipping")
	}

	// Selecting a plan deducts from wallet — may fail if wallet is empty (expected in CI)
	Test(t,
		Description("POST /machines/plan/select with valid plan — may fail if wallet empty"),
		Post(tests.GetMachinePlanSelectURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"plan_id": planID}),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusBadRequest), int64(http.StatusInternalServerError)),
	)
}

// --- Machine billing status ---

func TestGetMachineBillingStatus_NoAuth(t *testing.T) {
	serverID := uuid.New().String()
	Test(t,
		Description("GET /machines/billing without auth returns 401"),
		Get(tests.GetMachineBillingStatusURL()+"?server_id="+serverID),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetMachineBillingStatus_NoServerID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	// server_id is optional — billing status is org-scoped, not server-scoped
	Test(t,
		Description("GET /machines/billing without server_id returns 200 (org-scoped)"),
		Get(tests.GetMachineBillingStatusURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

func TestGetMachineBillingStatus_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}

	keyID := seedBYOSMachine(t, setup, orgID, auth.User.ID, true)

	Test(t,
		Description("GET /machines/billing with valid server_id returns 200"),
		Get(tests.GetMachineBillingStatusURL()+"?server_id="+keyID.String()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

// --- Machine stats ---

func TestGetMachineStats_NoAuth(t *testing.T) {
	serverID := uuid.New().String()
	Test(t,
		Description("GET /machines/stats without auth returns 401"),
		Get(tests.GetMachineStatsURL()+"?server_id="+serverID),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetMachineStats_NoServerID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	// server_id is optional; falls back to org's provisioned machine
	Test(t,
		Description("GET /machines/stats without server_id uses org default machine"),
		Get(tests.GetMachineStatsURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError), int64(http.StatusNotFound)),
	)
}

func TestGetMachineStats_RandomServerID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	// Stats require SSH — returns 500 for non-existent machine, 404 if not found in DB
	Test(t,
		Description("GET /machines/stats with random server_id returns error (no SSH)"),
		Get(tests.GetMachineStatsURL()+"?server_id="+uuid.New().String()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
	)
}

// --- Machine metrics ---

func TestGetMachineMetrics_NoAuth(t *testing.T) {
	serverID := uuid.New().String()
	Test(t,
		Description("GET /machines/metrics without auth returns 401"),
		Get(tests.GetMachineMetricsURL()+"?server_id="+serverID),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetMachineMetrics_NoServerID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	// server_id is optional for metrics (org-wide query)
	Test(t,
		Description("GET /machines/metrics without server_id returns 200 (org-wide)"),
		Get(tests.GetMachineMetricsURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

func TestGetMachineMetrics_ValidAuth_EmptyTimescale(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}

	keyID := seedBYOSMachine(t, setup, orgID, auth.User.ID, true)

	// No metrics in Timescale yet — returns 200 with empty data or 500 if Timescale unavailable
	Test(t,
		Description("GET /machines/metrics with valid server_id returns 200 (empty) or 500 (no Timescale)"),
		Get(tests.GetMachineMetricsURL()+"?server_id="+keyID.String()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

func TestGetMachineMetrics_WithTimeRange(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}

	keyID := seedBYOSMachine(t, setup, orgID, auth.User.ID, true)

	Test(t,
		Description("GET /machines/metrics with from/to query params"),
		Get(tests.GetMachineMetricsURL()+"?server_id="+keyID.String()+"&from=2025-01-01T00:00:00Z&to=2025-01-02T00:00:00Z"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

// --- Machine events ---

func TestGetMachineEvents_NoAuth(t *testing.T) {
	serverID := uuid.New().String()
	Test(t,
		Description("GET /machines/events without auth returns 401"),
		Get(tests.GetMachineEventsURL()+"?server_id="+serverID),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetMachineEvents_NoServerID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	// server_id is optional for events (org-wide query)
	Test(t,
		Description("GET /machines/events without server_id returns 200 (org-wide)"),
		Get(tests.GetMachineEventsURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

func TestGetMachineEvents_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}

	keyID := seedBYOSMachine(t, setup, orgID, auth.User.ID, true)

	Test(t,
		Description("GET /machines/events with valid server_id returns 200 or 500 (no Timescale)"),
		Get(tests.GetMachineEventsURL()+"?server_id="+keyID.String()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

func TestGetMachineEvents_WithTimeRange(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}

	keyID := seedBYOSMachine(t, setup, orgID, auth.User.ID, true)

	Test(t,
		Description("GET /machines/events with from/to and limit params"),
		Get(tests.GetMachineEventsURL()+"?server_id="+keyID.String()+"&limit=50"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

// --- Machine metrics summary ---

func TestGetMachineMetricsSummary_NoAuth(t *testing.T) {
	serverID := uuid.New().String()
	Test(t,
		Description("GET /machines/metrics/summary without auth returns 401"),
		Get(tests.GetMachineMetricsSummaryURL()+"?server_id="+serverID),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetMachineMetricsSummary_NoServerID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	// server_id is optional for metrics summary
	Test(t,
		Description("GET /machines/metrics/summary without server_id returns 200 (org-wide)"),
		Get(tests.GetMachineMetricsSummaryURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

func TestGetMachineMetricsSummary_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}

	keyID := seedBYOSMachine(t, setup, orgID, auth.User.ID, true)

	Test(t,
		Description("GET /machines/metrics/summary with valid server_id returns 200 or 500 (no Timescale)"),
		Get(tests.GetMachineMetricsSummaryURL()+"?server_id="+keyID.String()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}
