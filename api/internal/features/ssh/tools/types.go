package tools

// RunCommandInput is the input structure for the MCP tool
type RunCommandInput struct {
	Command  string `json:"command" jsonschema:"required"`
	ClientID string `json:"client_id,omitempty"`
}

// RunCommandOutput is the output structure for the MCP tool
type RunCommandOutput struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code,omitempty"`
}
