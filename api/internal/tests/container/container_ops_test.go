package container

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

// All container ops require Docker/SSH on the target machine.
// In CI the expected result for authenticated requests is 500 (no SSH infra).
// 401 for no auth and 400 for bad params are always testable.

// --- GET container by ID ---

func TestGetContainerByID_NoAuth(t *testing.T) {
	containerID := uuid.New().String()
	Test(t,
		Description("GET /container/:id without auth returns 401"),
		Get(tests.GetContainerURL(containerID)),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetContainerByID_InvalidID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /container/not-a-uuid returns 400"),
		Get(tests.GetContainerURL("not-a-uuid")),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestGetContainerByID_ValidAuth_NoSSH(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /container/:id with valid UUID — 500 expected (no SSH in CI)"),
		Get(tests.GetContainerURL(uuid.New().String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
	)
}

// --- POST container logs ---

func TestGetContainerLogs_NoAuth(t *testing.T) {
	containerID := uuid.New().String()
	Test(t,
		Description("POST /container/:id/logs without auth returns 401"),
		Post(tests.GetContainerLogsURL(containerID)),
		Send().Body().JSON(map[string]interface{}{"id": containerID}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetContainerLogs_InvalidID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /container/:id/logs with non-UUID id returns 400"),
		Post(tests.GetContainerLogsURL("not-a-uuid")),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"id": "not-a-uuid"}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestGetContainerLogs_ValidAuth_NoSSH(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	containerID := uuid.New().String()
	Test(t,
		Description("POST /container/:id/logs — 500 expected (no SSH in CI)"),
		Post(tests.GetContainerLogsURL(containerID)),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"id": containerID}),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

func TestGetContainerLogs_WithTailParam(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	containerID := uuid.New().String()
	Test(t,
		Description("POST /container/:id/logs with tail param is accepted"),
		Post(tests.GetContainerLogsURL(containerID)),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"id": containerID, "tail": 100}),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

// --- POST container start ---

func TestStartContainer_NoAuth(t *testing.T) {
	containerID := uuid.New().String()
	Test(t,
		Description("POST /container/:id/start without auth returns 401"),
		Post(tests.GetContainerStartURL(containerID)),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestStartContainer_InvalidID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /container/not-a-uuid/start returns 400"),
		Post(tests.GetContainerStartURL("not-a-uuid")),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestStartContainer_ValidAuth_NoSSH(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /container/:id/start — 500 expected (no SSH in CI)"),
		Post(tests.GetContainerStartURL(uuid.New().String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

// --- POST container stop ---

func TestStopContainer_NoAuth(t *testing.T) {
	containerID := uuid.New().String()
	Test(t,
		Description("POST /container/:id/stop without auth returns 401"),
		Post(tests.GetContainerStopURL(containerID)),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestStopContainer_InvalidID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /container/not-a-uuid/stop returns 400"),
		Post(tests.GetContainerStopURL("not-a-uuid")),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestStopContainer_ValidAuth_NoSSH(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /container/:id/stop — 500 expected (no SSH in CI)"),
		Post(tests.GetContainerStopURL(uuid.New().String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

// --- POST container restart ---

func TestRestartContainer_NoAuth(t *testing.T) {
	containerID := uuid.New().String()
	Test(t,
		Description("POST /container/:id/restart without auth returns 401"),
		Post(tests.GetContainerRestartURL(containerID)),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestRestartContainer_InvalidID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /container/not-a-uuid/restart returns 400"),
		Post(tests.GetContainerRestartURL("not-a-uuid")),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestRestartContainer_ValidAuth_NoSSH(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /container/:id/restart — 500 expected (no SSH in CI)"),
		Post(tests.GetContainerRestartURL(uuid.New().String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

// --- DELETE container (remove) ---

func TestRemoveContainer_NoAuth(t *testing.T) {
	containerID := uuid.New().String()
	Test(t,
		Description("DELETE /container/:id without auth returns 401"),
		Delete(tests.GetContainerURL(containerID)),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestRemoveContainer_ValidAuth_NoSSH(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("DELETE /container/:id — OK or 500 (no SSH in CI); remove does not reject non-UUID in controller"),
		Delete(tests.GetContainerURL(uuid.New().String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

func TestRemoveContainer_InvalidOrgHeader(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("DELETE /container/:id with invalid org header returns 400"),
		Delete(tests.GetContainerURL(uuid.New().String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add("not-a-uuid"),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestRemoveContainer_CrossOrgDenied(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("DELETE /container/:id for different org returns 403"),
		Delete(tests.GetContainerURL(uuid.New().String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add("123e4567-e89b-12d3-a456-426614174000"),
		Expect().Status().Equal(http.StatusForbidden),
	)
}

// --- PUT container resources ---

func TestGetContainerResources_NoAuth(t *testing.T) {
	containerID := uuid.New().String()
	Test(t,
		Description("PUT /container/:id/resources without auth returns 401"),
		Put(tests.GetContainerResourcesURL(containerID)),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetContainerResources_InvalidID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PUT /container/not-a-uuid/resources returns 400"),
		Put(tests.GetContainerResourcesURL("not-a-uuid")),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestGetContainerResources_ValidAuth_NoSSH(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PUT /container/:id/resources — 500 expected (no SSH in CI)"),
		Put(tests.GetContainerResourcesURL(uuid.New().String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

// --- GET container images ---

func TestListContainerImages_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /container/images without auth returns 401"),
		Post(tests.GetContainerImagesURL()),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestListContainerImages_ValidAuth_NoSSH(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /container/images — 500 expected (no Docker in CI)"),
		Post(tests.GetContainerImagesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

func TestListContainerImages_InvalidOrgID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /container/images with invalid org ID returns 400"),
		Post(tests.GetContainerImagesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add("not-a-uuid"),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestListContainerImages_CrossOrgDenied(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /container/images for different org returns 403"),
		Post(tests.GetContainerImagesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add("123e4567-e89b-12d3-a456-426614174000"),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Equal(http.StatusForbidden),
	)
}

// --- POST prune images ---

func TestPruneImages_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /container/prune/images without auth returns 401"),
		Post(tests.GetContainerPruneImagesURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestPruneImages_ValidAuth_NoSSH(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /container/prune/images — 500 expected (no Docker in CI)"),
		Post(tests.GetContainerPruneImagesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

func TestPruneImages_InvalidOrgID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /container/prune/images with invalid org ID returns 400"),
		Post(tests.GetContainerPruneImagesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add("not-a-uuid"),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

// --- POST prune build cache ---

func TestPruneBuildCache_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /container/prune/build-cache without auth returns 401"),
		Post(tests.GetContainerPruneBuildCacheURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestPruneBuildCache_ValidAuth_NoSSH(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /container/prune/build-cache — 500 expected (no Docker in CI)"),
		Post(tests.GetContainerPruneBuildCacheURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

func TestPruneBuildCache_InvalidOrgID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /container/prune/build-cache with invalid org ID returns 400"),
		Post(tests.GetContainerPruneBuildCacheURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add("not-a-uuid"),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestPruneBuildCache_CrossOrgDenied(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /container/prune/build-cache for different org returns 403"),
		Post(tests.GetContainerPruneBuildCacheURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add("123e4567-e89b-12d3-a456-426614174000"),
		Expect().Status().Equal(http.StatusForbidden),
	)
}
