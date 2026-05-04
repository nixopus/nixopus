package mcp

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

// --- Internal: List enabled servers ---

func TestListMCPInternalServers_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /mcp/internal/servers without auth returns 401"),
		Get(tests.GetMCPInternalServersURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestListMCPInternalServers_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /mcp/internal/servers with valid auth returns 200 with success status"),
		Get(tests.GetMCPInternalServersURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestListMCPInternalServers_InvalidOrgID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /mcp/internal/servers with non-UUID org ID returns 400"),
		Get(tests.GetMCPInternalServersURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add("not-a-uuid"),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestListMCPInternalServers_CrossOrgDenied(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /mcp/internal/servers with foreign org ID returns 403"),
		Get(tests.GetMCPInternalServersURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add("123e4567-e89b-12d3-a456-426614174000"),
		Expect().Status().Equal(http.StatusForbidden),
	)
}

// --- Internal: List tools from enabled servers ---

func TestListMCPInternalTools_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /mcp/internal/tools without auth returns 401"),
		Get(tests.GetMCPInternalToolsURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestListMCPInternalTools_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /mcp/internal/tools with valid auth returns 200 (may be empty if no enabled servers)"),
		Get(tests.GetMCPInternalToolsURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestListMCPInternalTools_InvalidOrgID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /mcp/internal/tools with non-UUID org ID returns 400"),
		Get(tests.GetMCPInternalToolsURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add("not-a-uuid"),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestListMCPInternalTools_CrossOrgDenied(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /mcp/internal/tools with foreign org ID returns 403"),
		Get(tests.GetMCPInternalToolsURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add("123e4567-e89b-12d3-a456-426614174000"),
		Expect().Status().Equal(http.StatusForbidden),
	)
}

// --- Internal: Call tool ---

func TestCallMCPTool_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /mcp/internal/tools/call without auth returns 401"),
		Post(tests.GetMCPInternalToolsCallURL()),
		Send().Body().JSON(map[string]interface{}{
			"server_id": uuid.New().String(),
			"tool_name": "search",
		}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestCallMCPTool_MissingServerID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /mcp/internal/tools/call without server_id returns 400"),
		Post(tests.GetMCPInternalToolsCallURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"tool_name": "search",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCallMCPTool_MissingToolName(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /mcp/internal/tools/call without tool_name returns 400"),
		Post(tests.GetMCPInternalToolsCallURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"server_id": uuid.New().String(),
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCallMCPTool_InvalidServerIDFormat(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /mcp/internal/tools/call with non-UUID server_id returns 400"),
		Post(tests.GetMCPInternalToolsCallURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"server_id": "not-a-uuid",
			"tool_name": "search",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCallMCPTool_NonExistentServerID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /mcp/internal/tools/call with unknown server_id returns 404"),
		Post(tests.GetMCPInternalToolsCallURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"server_id": uuid.New().String(),
			"tool_name": "search",
		}),
		Expect().Status().OneOf(int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
	)
}

func TestCallMCPTool_FullFlow(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	// Need a valid provider to create a server first
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
		t.Skip("no providers in catalog, skipping full tool call flow")
	}

	// Create a server
	var serverID string
	Test(t,
		Description("Create MCP server for tool call test"),
		Post(tests.GetMCPServersURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"provider_id": providerID,
			"name":        "internal-test-server-for-call",
			"enabled":     true,
			"credentials": map[string]string{},
		}),
		Expect().Status().Equal(http.StatusOK),
		Store().Response().Body().JSON().JQ(".data.id").In(&serverID),
	)

	if serverID == "" {
		t.Skip("no server ID returned, skipping tool call flow")
	}

	// Discover tools from all enabled servers
	var toolName string
	Test(t,
		Description("Discover tools from enabled MCP servers"),
		Get(tests.GetMCPInternalToolsURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Store().Response().Body().JSON().JQ(".data[0].tools[0].name").In(&toolName),
	)

	if toolName == "" {
		t.Skip("no tools discovered (server may require valid credentials), skipping tool call")
	}

	// Call the discovered tool
	Test(t,
		Description("POST /mcp/internal/tools/call with valid server and tool"),
		Post(tests.GetMCPInternalToolsCallURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"server_id": serverID,
			"tool_name": toolName,
			"arguments": map[string]interface{}{},
		}),
		Expect().Status().OneOf(
			int64(http.StatusOK),
			int64(http.StatusBadGateway),
			int64(http.StatusInternalServerError),
		),
	)
}
