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
