package validation

import (
	"errors"
	"net"
	"net/url"
	"strings"

	mcp "github.com/nixopus/nixopus/api/internal/features/mcp"
)

var (
	ErrNameRequired         = errors.New("name is required")
	ErrProviderRequired     = errors.New("provider_id is required")
	ErrUnknownProvider      = errors.New("unknown provider_id")
	ErrCustomURLRequired    = errors.New("custom_url is required for custom provider")
	ErrInvalidURL           = errors.New("url must use https scheme")
	ErrPrivateURL           = errors.New("url must not point to a private or loopback address")
	ErrMissingRequiredField = errors.New("missing required credential field")
)

func ValidateURL(rawURL string) error {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return ErrInvalidURL
	}
	if u.Scheme != "https" {
		return ErrInvalidURL
	}
	host := u.Hostname()
	ips, err := net.LookupHost(host)
	if err != nil {
		return nil
	}
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		if isPrivate(ip) {
			return ErrPrivateURL
		}
	}
	return nil
}

func isPrivate(ip net.IP) bool {
	private := []string{
		"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
		"127.0.0.0/8", "169.254.0.0/16", "::1/128", "fc00::/7",
	}
	for _, cidr := range private {
		_, block, _ := net.ParseCIDR(cidr)
		if block != nil && block.Contains(ip) {
			return true
		}
	}
	return false
}

type CreateServerRequest struct {
	ProviderID  string            `json:"provider_id" validate:"required" description:"MCP provider identifier (e.g. 'openai', 'anthropic', 'custom')" example:"openai"`
	Name        string            `json:"name" validate:"required" description:"Display name for this MCP server" example:"My OpenAI Server"`
	Credentials map[string]string `json:"credentials" description:"Provider-specific credential key-value pairs"`
	CustomURL   string            `json:"custom_url,omitempty" description:"Custom server URL. Required when provider_id is 'custom'. Must use HTTPS"`
	Enabled     bool              `json:"enabled" description:"Whether the server is active"`
}

type UpdateServerRequest struct {
	ID          string            `json:"id" validate:"required" description:"MCP server ID to update"`
	Name        string            `json:"name" validate:"required" description:"Updated display name"`
	Credentials map[string]string `json:"credentials" description:"Updated provider credentials"`
	CustomURL   string            `json:"custom_url,omitempty" description:"Updated custom server URL. Must use HTTPS if provided"`
	Enabled     bool              `json:"enabled" description:"Whether the server is active"`
}

type TestServerRequest struct {
	ProviderID  string            `json:"provider_id" validate:"required" description:"MCP provider to test connection for"`
	Credentials map[string]string `json:"credentials" description:"Credentials to use for the test"`
	CustomURL   string            `json:"custom_url,omitempty" description:"Custom URL to test. Required when provider_id is 'custom'"`
}

func ValidateCreateRequest(req *CreateServerRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return ErrNameRequired
	}
	if strings.TrimSpace(req.ProviderID) == "" {
		return ErrProviderRequired
	}
	provider := mcp.GetProvider(req.ProviderID)
	if provider == nil {
		return ErrUnknownProvider
	}
	if req.ProviderID == "custom" {
		if strings.TrimSpace(req.CustomURL) == "" {
			return ErrCustomURLRequired
		}
		if err := ValidateURL(req.CustomURL); err != nil {
			return err
		}
	}
	for _, field := range provider.Fields {
		if field.Required {
			if v, ok := req.Credentials[field.Key]; !ok || strings.TrimSpace(v) == "" {
				return ErrMissingRequiredField
			}
		}
	}
	return nil
}

func ValidateUpdateRequest(req *UpdateServerRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return ErrNameRequired
	}
	if req.CustomURL != "" {
		if err := ValidateURL(req.CustomURL); err != nil {
			return err
		}
	}
	return nil
}

type CallToolRequest struct {
	ServerID  string         `json:"server_id" validate:"required" description:"ID of the MCP server to call the tool on"`
	ToolName  string         `json:"tool_name" validate:"required" description:"Name of the MCP tool to invoke" example:"search"`
	Arguments map[string]any `json:"arguments,omitempty" description:"Tool-specific arguments as key-value pairs"`
}

func ValidateCallToolRequest(req *CallToolRequest) error {
	if strings.TrimSpace(req.ServerID) == "" {
		return errors.New("server_id is required")
	}
	if strings.TrimSpace(req.ToolName) == "" {
		return errors.New("tool_name is required")
	}
	return nil
}

func ValidateTestRequest(req *TestServerRequest) error {
	if strings.TrimSpace(req.ProviderID) == "" {
		return ErrProviderRequired
	}
	provider := mcp.GetProvider(req.ProviderID)
	if provider == nil {
		return ErrUnknownProvider
	}
	if req.ProviderID == "custom" {
		if strings.TrimSpace(req.CustomURL) == "" {
			return ErrCustomURLRequired
		}
		if err := ValidateURL(req.CustomURL); err != nil {
			return err
		}
	}
	return nil
}
