package mcp

import "embed"

//go:embed icons
var IconFS embed.FS

type ProviderField struct {
	Key          string `json:"key"`
	Label        string `json:"label"`
	Required     bool   `json:"required"`
	HeaderName   string `json:"header_name,omitempty"`
	HeaderPrefix string `json:"header_prefix,omitempty"`
	IsQueryParam bool   `json:"is_query_param,omitempty"`
	Sensitive    bool   `json:"sensitive"`
}

type MCPProvider struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	URL         string          `json:"-"` // resolved at serve time, never stored
	Transport   string          `json:"-"` // "http" or "sse"
	LogoURL     string          `json:"logo_url"`
	Fields      []ProviderField `json:"fields"`
}

var Catalog = []MCPProvider{
	{
		ID:          "supabase",
		Name:        "Supabase",
		Description: "Connect to your Supabase project via MCP",
		URL:         "https://mcp.supabase.com/mcp",
		Transport:   "http",
		Fields: []ProviderField{
			{Key: "access_token", Label: "Access Token", Required: true,
				HeaderName: "Authorization", HeaderPrefix: "Bearer", Sensitive: true},
			{Key: "project_ref", Label: "Project Ref (optional)", Required: false,
				IsQueryParam: true, Sensitive: false},
		},
	},
	{
		ID:          "github",
		Name:        "GitHub",
		Description: "Access GitHub via MCP",
		URL:         "https://api.githubcopilot.com/mcp/",
		Transport:   "http",
		Fields: []ProviderField{
			{Key: "token", Label: "Personal Access Token", Required: true,
				HeaderName: "Authorization", HeaderPrefix: "Bearer", Sensitive: true},
		},
	},
	{
		ID:          "linear",
		Name:        "Linear",
		Description: "Access Linear issues, projects, and teams via MCP",
		URL:         "https://mcp.linear.app/mcp",
		Transport:   "http",
		Fields: []ProviderField{
			{Key: "api_key", Label: "API Key", Required: true,
				HeaderName: "Authorization", HeaderPrefix: "Bearer", Sensitive: true},
		},
	},
	{
		ID:          "sentry",
		Name:        "Sentry",
		Description: "Access Sentry issues, errors, and debugging data via MCP",
		URL:         "https://mcp.sentry.dev/mcp",
		Transport:   "http",
		Fields: []ProviderField{
			{Key: "auth_token", Label: "Auth Token", Required: true,
				HeaderName: "Authorization", HeaderPrefix: "Bearer", Sensitive: true},
		},
	},
	{
		ID:          "atlassian",
		Name:        "Atlassian",
		Description: "Access Jira, Confluence, and Compass via MCP",
		URL:         "https://mcp.atlassian.com/v1/mcp",
		Transport:   "http",
		Fields: []ProviderField{
			{Key: "api_token", Label: "API Token", Required: true,
				HeaderName: "Authorization", HeaderPrefix: "Bearer", Sensitive: true},
		},
	},
	{
		ID:          "semgrep",
		Name:        "Semgrep",
		Description: "Scan code for security vulnerabilities via MCP",
		URL:         "https://mcp.semgrep.ai/mcp",
		Transport:   "http",
		Fields: []ProviderField{
			{Key: "app_token", Label: "App Token", Required: true,
				HeaderName: "SEMGREP_APP_TOKEN", Sensitive: true},
		},
	},
	{
		ID:          "neon",
		Name:        "Neon",
		Description: "Manage Neon Postgres databases via MCP",
		URL:         "https://mcp.neon.tech/mcp",
		Transport:   "http",
		Fields: []ProviderField{
			{Key: "api_key", Label: "API Key", Required: true,
				HeaderName: "Authorization", HeaderPrefix: "Bearer", Sensitive: true},
		},
	},
	{
		ID:          "planetscale",
		Name:        "PlanetScale",
		Description: "Access PlanetScale MySQL databases via MCP",
		URL:         "https://mcp.pscale.dev/mcp/planetscale",
		Transport:   "http",
		Fields: []ProviderField{
			{Key: "api_token", Label: "Service Token", Required: true,
				HeaderName: "Authorization", HeaderPrefix: "Bearer", Sensitive: true},
		},
	},
	{
		ID:          "custom",
		Name:        "Custom",
		Description: "Connect to any hosted MCP server",
		URL:         "",
		Transport:   "http",
		Fields:      []ProviderField{},
	},
}

func GetProvider(id string) *MCPProvider {
	for i := range Catalog {
		if Catalog[i].ID == id {
			return &Catalog[i]
		}
	}
	return nil
}
