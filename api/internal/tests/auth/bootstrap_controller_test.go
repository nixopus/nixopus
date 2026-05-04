package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-fuego/fuego"
	"github.com/google/uuid"

	"github.com/nixopus/nixopus/api/internal/cache"
	"github.com/nixopus/nixopus/api/internal/config"
	sessions "github.com/nixopus/nixopus/api/internal/features/auth"
	ctl "github.com/nixopus/nixopus/api/internal/features/auth/controller"
	authsvc "github.com/nixopus/nixopus/api/internal/features/auth/service"
	authstor "github.com/nixopus/nixopus/api/internal/features/auth/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/testutils"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/require"
)

func sessionPayload(uid, activeOrg string) []byte {
	session := map[string]any{"id": "sid", "userId": uid, "expiresAt": "", "token": "t"}
	if activeOrg != "" {
		session["activeOrganizationId"] = activeOrg
	}
	m := map[string]any{"session": session}
	m["user"] = map[string]any{"id": uid, "email": "e@x", "name": "n", "emailVerified": true}
	b, _ := json.Marshal(m)
	return b
}

func swapBetterAuth(saved types.BetterAuthConfig) func() {
	prev := config.AppConfig.BetterAuth
	config.AppConfig.BetterAuth = saved
	return func() { config.AppConfig.BetterAuth = prev }
}

func swapSessionsClient(c *http.Client) func() {
	prev := sessions.HTTPClient
	sessions.HTTPClient = c
	return func() { sessions.HTTPClient = prev }
}

func ctrlFromSetup(t *testing.T, setup *testutils.TestSetup, redisURL string) *ctl.AuthController {
	t.Helper()
	log := logger.NewLogger()
	st := authstor.UserStorage{DB: setup.DB, Ctx: setup.Ctx}
	svc := authsvc.NewAuthService(&st, st.DB, log, setup.Ctx, redisURL)
	return ctl.NewAuthController(setup.Ctx, log, svc)
}

func mockFuegoCtx(r *http.Request) fuego.ContextNoBody {
	mc := fuego.NewMockContextNoBody()
	mc.SetRequest(r)
	return mc
}

func withCtxUser(u *types.User) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	return req.WithContext(context.WithValue(req.Context(), types.UserContextKey, u))
}

func sessSrv(t *testing.T, uid, activeOrg string) (*httptest.Server, *http.Client) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/get-session" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write(sessionPayload(uid, activeOrg))
	}))
	t.Cleanup(srv.Close)
	return srv, srv.Client()
}

func installSessionMocks(t *testing.T, srvURL string, cl *http.Client) func() {
	swapU := swapBetterAuth(types.BetterAuthConfig{URL: srvURL, Secret: "x"})
	swapC := swapSessionsClient(cl)
	return func() {
		swapC()
		swapU()
	}
}

func TestAuthController_New(t *testing.T) {
	require.NotNil(t, ctrlFromSetup(t, testutils.NewTestSetup(), ""))
}

func TestAuthController_HandleBootstrap_requiresUser(t *testing.T) {
	c := ctrlFromSetup(t, testutils.NewTestSetup(), "")
	_, err := c.HandleBootstrap(mockFuegoCtx(httptest.NewRequest(http.MethodGet, "/", nil)))
	require.ErrorAs(t, err, new(fuego.UnauthorizedError))
}

func TestAuthController_HandleBootstrap_newRequestFails(t *testing.T) {
	setup := testutils.NewTestSetup()
	swap := swapBetterAuth(types.BetterAuthConfig{URL: "http://[\n", Secret: "x"})
	defer swap()

	u, _, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	_, err = ctrlFromSetup(t, setup, "").HandleBootstrap(mockFuegoCtx(withCtxUser(u)))
	require.ErrorAs(t, err, new(fuego.UnauthorizedError))
}

func TestAuthController_HandleBootstrap_nullSession(t *testing.T) {
	setup := testutils.NewTestSetup()
	u, _, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("null"))
	}))
	t.Cleanup(srv.Close)

	cleanup := installSessionMocks(t, srv.URL, srv.Client())
	defer cleanup()

	resp, err := ctrlFromSetup(t, setup, "").HandleBootstrap(mockFuegoCtx(withCtxUser(u)))
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestAuthController_HandleBootstrap_happy(t *testing.T) {
	setup := testutils.NewTestSetup()
	u, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	srv, cl := sessSrv(t, u.ID.String(), "")
	cleanup := installSessionMocks(t, srv.URL, cl)
	defer cleanup()

	got, err := ctrlFromSetup(t, setup, "").HandleBootstrap(mockFuegoCtx(withCtxUser(u)))
	require.NoError(t, err)
	require.Equal(t, org.ID.String(), *got.ActiveOrganizationID)
	require.False(t, got.HasServers)
}

func TestAuthController_HandleBootstrap_activeOrgOverride(t *testing.T) {
	setup := testutils.NewTestSetup()
	u, _, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	other := uuid.New()

	srv, cl := sessSrv(t, u.ID.String(), other.String())
	cleanup := installSessionMocks(t, srv.URL, cl)
	defer cleanup()

	got, err := ctrlFromSetup(t, setup, "").HandleBootstrap(mockFuegoCtx(withCtxUser(u)))
	require.NoError(t, err)
	require.Equal(t, other.String(), *got.ActiveOrganizationID)
}

func TestAuthController_HandleBootstrap_provisionLabelFromUser(t *testing.T) {
	setup := testutils.NewTestSetup()
	u, _, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	ready := "ready"
	u.ProvisionStatus = &ready

	srv, cl := sessSrv(t, u.ID.String(), "")
	cleanup := installSessionMocks(t, srv.URL, cl)
	defer cleanup()

	got, err := ctrlFromSetup(t, setup, "").HandleBootstrap(mockFuegoCtx(withCtxUser(u)))
	require.NoError(t, err)
	require.Equal(t, ready, got.User.ProvisionStatus)
}

func TestAuthController_HandleBootstrap_provisioningNoRow(t *testing.T) {
	setup := testutils.NewTestSetup()
	u, _, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	ps := "provisioning"
	u.ProvisionStatus = &ps

	srv, cl := sessSrv(t, u.ID.String(), "")
	cleanup := installSessionMocks(t, srv.URL, cl)
	defer cleanup()

	_, err = ctrlFromSetup(t, setup, "").HandleBootstrap(mockFuegoCtx(withCtxUser(u)))
	require.Error(t, err)

	var he fuego.HTTPError
	require.ErrorAs(t, err, &he)
}

func TestAuthController_HandleBootstrap_provisioningWithDetails(t *testing.T) {
	setup := testutils.NewTestSetup()
	u, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	ps := "provisioning"
	u.ProvisionStatus = &ps

	step := types.ProvisionStepInitializing
	details := types.UserProvisionDetails{
		UserID:         u.ID,
		OrganizationID: org.ID,
		Type:           "trial",
		VcpuCount:      1,
		MemoryMB:       512,
		DiskSizeGB:     10,
		Step:           &step,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err = setup.DB.NewInsert().Model(&details).Exec(setup.Ctx)
	require.NoError(t, err)

	srv, cl := sessSrv(t, u.ID.String(), "")
	cleanup := installSessionMocks(t, srv.URL, cl)
	defer cleanup()

	got, err := ctrlFromSetup(t, setup, "").HandleBootstrap(mockFuegoCtx(withCtxUser(u)))
	require.NoError(t, err)
	require.NotNil(t, got.ProvisionID)
	require.NotNil(t, got.ProvisionStep)
}

func TestAuthController_HandleBootstrap_hasServers(t *testing.T) {
	setup := testutils.NewTestSetup()
	u, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	key := types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "k",
		AuthMethod:     "key",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err = setup.DB.NewInsert().Model(&key).Exec(setup.Ctx)
	require.NoError(t, err)

	srv, cl := sessSrv(t, u.ID.String(), "")
	cleanup := installSessionMocks(t, srv.URL, cl)
	defer cleanup()

	got, err := ctrlFromSetup(t, setup, "").HandleBootstrap(mockFuegoCtx(withCtxUser(u)))
	require.NoError(t, err)
	require.True(t, got.HasServers)
}

func TestAuthController_IsAdminRegistered_dbFalseCachesRedis(t *testing.T) {
	setup := testutils.NewTestSetup()
	mr := miniredis.RunT(t)
	url := "redis://" + mr.Addr()

	c := ctrlFromSetup(t, setup, url)
	r1, err := c.IsAdminRegistered(mockFuegoCtx(httptest.NewRequest(http.MethodGet, "/", nil)))
	require.NoError(t, err)
	require.False(t, r1.Data.AdminRegistered)

	r2, err := c.IsAdminRegistered(mockFuegoCtx(httptest.NewRequest(http.MethodGet, "/", nil)))
	require.NoError(t, err)
	require.False(t, r2.Data.AdminRegistered)
}

func TestAuthController_IsAdminRegistered_cacheWarmTrueFastPath(t *testing.T) {
	setup := testutils.NewTestSetup()
	mr := miniredis.RunT(t)
	url := "redis://" + mr.Addr()

	redisCache, err := cache.NewCache(url)
	require.NoError(t, err)
	require.NoError(t, redisCache.SetAdminRegistered(setup.Ctx, true))

	c := ctrlFromSetup(t, setup, url)
	resp, err := c.IsAdminRegistered(mockFuegoCtx(httptest.NewRequest(http.MethodGet, "/", nil)))
	require.NoError(t, err)
	require.True(t, resp.Data.AdminRegistered)
}

func TestAuthController_IsAdminRegistered_afterSeedCredentialAccount(t *testing.T) {
	setup := testutils.NewTestSetup()
	u, _, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	require.NoError(t, setup.SeedCredentialAccount(u.ID.String()))

	c := ctrlFromSetup(t, setup, "")
	resp, err := c.IsAdminRegistered(mockFuegoCtx(httptest.NewRequest(http.MethodGet, "/", nil)))
	require.NoError(t, err)
	require.True(t, resp.Data.AdminRegistered)
}

func TestAuthController_IsAdminRegistered_cacheReadErrorFallsThroughToDB(t *testing.T) {
	setup := testutils.NewTestSetup()
	mr := miniredis.RunT(t)
	url := "redis://" + mr.Addr()

	c := ctrlFromSetup(t, setup, url)
	_, err := c.IsAdminRegistered(mockFuegoCtx(httptest.NewRequest(http.MethodGet, "/", nil)))
	require.NoError(t, err)

	mr.Close()

	resp, err := c.IsAdminRegistered(mockFuegoCtx(httptest.NewRequest(http.MethodGet, "/", nil)))
	require.NoError(t, err)
	require.False(t, resp.Data.AdminRegistered)
}
