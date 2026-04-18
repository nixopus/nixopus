package mcp

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

// --- Catalog ---

func TestListMCPCatalog_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /mcp/catalog without auth returns 401"),
		Get(tests.GetMCPCatalogURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestListMCPCatalog_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /mcp/catalog returns catalog list"),
		Get(tests.GetMCPCatalogURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestListMCPCatalog_WithPagination(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /mcp/catalog with pagination params returns 200"),
		Get(tests.GetMCPCatalogURL()+"?page=1&limit=10"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".data.items").NotEqual(nil),
	)
}

func TestListMCPCatalog_WithSearch(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /mcp/catalog?q=github returns filtered results"),
		Get(tests.GetMCPCatalogURL()+"?q=github"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
	)
}

func TestListMCPCatalog_InvalidOrgID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /mcp/catalog with invalid org ID returns 400"),
		Get(tests.GetMCPCatalogURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add("not-a-uuid"),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

// --- List servers ---

func TestListMCPServers_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /mcp/servers without auth returns 401"),
		Get(tests.GetMCPServersURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestListMCPServers_ValidAuth_EmptyList(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /mcp/servers with valid auth returns 200 with empty list"),
		Get(tests.GetMCPServersURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".data.items").NotEqual(nil),
	)
}

func TestListMCPServers_WithPagination(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /mcp/servers with page and limit returns 200"),
		Get(tests.GetMCPServersURL()+"?page=1&limit=20"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
	)
}

func TestListMCPServers_CrossOrgDenied(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /mcp/servers for different org returns 403"),
		Get(tests.GetMCPServersURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add("123e4567-e89b-12d3-a456-426614174000"),
		Expect().Status().Equal(http.StatusForbidden),
	)
}

// --- Add server ---

func TestAddMCPServer_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /mcp/servers without auth returns 401"),
		Post(tests.GetMCPServersURL()),
		Send().Body().JSON(map[string]interface{}{
			"provider_id": "github",
			"name":        "my-github",
		}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestAddMCPServer_MissingName(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /mcp/servers without name returns 400"),
		Post(tests.GetMCPServersURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"provider_id": "github"}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestAddMCPServer_MissingProviderID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /mcp/servers without provider_id returns 400"),
		Post(tests.GetMCPServersURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"name": "my-server"}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestAddMCPServer_UnknownProvider(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /mcp/servers with unknown provider_id returns 400"),
		Post(tests.GetMCPServersURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"provider_id": "definitely-not-a-real-provider",
			"name":        "my-server",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestAddMCPServer_ValidProvider(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	// Use a catalog entry — fetch catalog first to find a valid provider ID
	var providerID string
	Test(t,
		Description("Fetch catalog to get a valid provider ID"),
		Get(tests.GetMCPCatalogURL()+"?limit=1"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Store().Response().Body().JSON().JQ(".data.items[0].provider_id").In(&providerID),
	)

	if providerID == "" {
		t.Skip("no providers in catalog, skipping")
	}

	Test(t,
		Description("POST /mcp/servers with valid provider creates server"),
		Post(tests.GetMCPServersURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"provider_id": providerID,
			"name":        "integration-test-server",
			"enabled":     true,
			"credentials": map[string]string{},
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

// --- Update server ---

func TestUpdateMCPServer_NoAuth(t *testing.T) {
	Test(t,
		Description("PUT /mcp/servers/:id without auth returns 401"),
		Put(tests.GetMCPServerURL(uuid.New().String())),
		Send().Body().JSON(map[string]interface{}{"name": "updated"}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestUpdateMCPServer_InvalidID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PUT /mcp/servers/not-a-uuid returns 400"),
		Put(tests.GetMCPServerURL("not-a-uuid")),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"name": "updated"}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestUpdateMCPServer_NonExistentID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PUT /mcp/servers/:id with non-existent ID returns error"),
		Put(tests.GetMCPServerURL(uuid.New().String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"id":      uuid.New().String(),
			"name":    "updated-name",
			"enabled": true,
		}),
		Expect().Status().OneOf(int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
	)
}

func TestUpdateMCPServer_ValidFlow(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	var providerID string
	Test(t,
		Description("Fetch catalog for valid provider"),
		Get(tests.GetMCPCatalogURL()+"?limit=1"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Store().Response().Body().JSON().JQ(".data.items[0].provider_id").In(&providerID),
	)

	if providerID == "" {
		t.Skip("no providers in catalog")
	}

	var serverID string
	Test(t,
		Description("Create MCP server for update"),
		Post(tests.GetMCPServersURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"provider_id": providerID,
			"name":        "server-to-update",
			"enabled":     true,
			"credentials": map[string]string{},
		}),
		Expect().Status().Equal(http.StatusOK),
		Store().Response().Body().JSON().JQ(".data.id").In(&serverID),
	)

	if serverID == "" {
		t.Skip("no server ID returned")
	}

	Test(t,
		Description("PUT /mcp/servers/:id updates server name"),
		Put(tests.GetMCPServerURL(serverID)),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"id":          serverID,
			"name":        "server-updated",
			"enabled":     false,
			"credentials": map[string]string{},
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

// --- Delete server ---

func TestDeleteMCPServer_NoAuth(t *testing.T) {
	Test(t,
		Description("DELETE /mcp/servers without auth returns 401"),
		Delete(tests.GetMCPServersURL()),
		Send().Body().JSON(map[string]interface{}{"id": uuid.New().String()}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestDeleteMCPServer_MissingID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("DELETE /mcp/servers without id returns 400"),
		Delete(tests.GetMCPServersURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestDeleteMCPServer_NonExistentID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("DELETE /mcp/servers with non-existent ID returns error"),
		Delete(tests.GetMCPServersURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"id": uuid.New().String()}),
		Expect().Status().OneOf(int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
	)
}

func TestDeleteMCPServer_ValidFlow(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	var providerID string
	Test(t,
		Description("Fetch catalog for valid provider"),
		Get(tests.GetMCPCatalogURL()+"?limit=1"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Store().Response().Body().JSON().JQ(".data.items[0].provider_id").In(&providerID),
	)

	if providerID == "" {
		t.Skip("no providers in catalog")
	}

	var serverID string
	Test(t,
		Description("Create server for deletion"),
		Post(tests.GetMCPServersURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"provider_id": providerID,
			"name":        "server-to-delete",
			"enabled":     true,
			"credentials": map[string]string{},
		}),
		Expect().Status().Equal(http.StatusOK),
		Store().Response().Body().JSON().JQ(".data.id").In(&serverID),
	)

	if serverID == "" {
		t.Skip("no server ID returned")
	}

	Test(t,
		Description("DELETE /mcp/servers with valid ID returns 200"),
		Delete(tests.GetMCPServersURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"id": serverID}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)

	// Verify it's gone from the list
	Test(t,
		Description("Deleted server no longer appears in list"),
		Get(tests.GetMCPServersURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data.total_count").Equal(0),
	)
}

// --- Test server connection ---

func TestTestMCPServer_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /mcp/servers/test without auth returns 401"),
		Post(tests.GetMCPServerTestURL()),
		Send().Body().JSON(map[string]interface{}{"provider_id": "github"}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestTestMCPServer_MissingProviderID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /mcp/servers/test without provider_id returns 400"),
		Post(tests.GetMCPServerTestURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestTestMCPServer_UnknownProvider(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /mcp/servers/test with unknown provider returns 400"),
		Post(tests.GetMCPServerTestURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"provider_id": "not-a-real-provider",
			"credentials": map[string]string{},
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestTestMCPServer_ValidProvider_ConnectionFails(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	var providerID string
	Test(t,
		Description("Fetch catalog for valid provider"),
		Get(tests.GetMCPCatalogURL()+"?limit=1"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Store().Response().Body().JSON().JQ(".data.items[0].provider_id").In(&providerID),
	)

	if providerID == "" {
		t.Skip("no providers in catalog")
	}

	// No real credentials → connection will fail, but endpoint accepts and processes the request
	Test(t,
		Description("POST /mcp/servers/test with empty credentials — request accepted, connection may fail"),
		Post(tests.GetMCPServerTestURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"provider_id": providerID,
			"credentials": map[string]string{},
		}),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusBadRequest), int64(http.StatusInternalServerError)),
	)
}
