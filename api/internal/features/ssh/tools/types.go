package tools

// RunCommandInput is the input structure for the MCP tool
type RunCommandInput struct {
	Command        string `json:"command" jsonschema:"required"`
	OrganizationID string `json:"organization_id" jsonschema:"required"`
	ClientID       string `json:"client_id,omitempty"` // Optional SSH client ID for multi-client support
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i RunCommandInput) GetOrganizationID() string {
	return i.OrganizationID
}

// RunCommandOutput is the output structure for the MCP tool
type RunCommandOutput struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code,omitempty"`
}
