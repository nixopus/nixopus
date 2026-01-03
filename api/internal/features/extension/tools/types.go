package tools

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// MCPExtensionVariable is a simplified ExtensionVariable without the circular Extension reference
type MCPExtensionVariable struct {
	ID                uuid.UUID       `json:"id"`
	ExtensionID       uuid.UUID       `json:"extension_id"`
	VariableName      string          `json:"variable_name"`
	VariableType      string          `json:"variable_type"`
	Description       string          `json:"description"`
	DefaultValue      json.RawMessage `json:"default_value"`
	IsRequired        bool            `json:"is_required"`
	ValidationPattern string          `json:"validation_pattern"`
	CreatedAt         time.Time       `json:"created_at"`
}

// MCPExtension is a simplified Extension without the circular Extension reference in Variables
type MCPExtension struct {
	ID                uuid.UUID                      `json:"id"`
	ExtensionID       string                         `json:"extension_id"`
	ParentExtensionID *uuid.UUID                     `json:"parent_extension_id,omitempty"`
	Name              string                         `json:"name"`
	Description       string                         `json:"description"`
	Author            string                         `json:"author"`
	Icon              string                         `json:"icon"`
	Category          shared_types.ExtensionCategory `json:"category"`
	ExtensionType     shared_types.ExtensionType     `json:"extension_type"`
	Version           string                         `json:"version"`
	IsVerified        bool                           `json:"is_verified"`
	YAMLContent       string                         `json:"yaml_content"`
	ParsedContent     string                         `json:"parsed_content"`
	ContentHash       string                         `json:"content_hash"`
	ValidationStatus  shared_types.ValidationStatus  `json:"validation_status"`
	ValidationErrors  string                         `json:"validation_errors"`
	CreatedAt         time.Time                      `json:"created_at"`
	UpdatedAt         time.Time                      `json:"updated_at"`
	DeletedAt         *time.Time                     `json:"deleted_at,omitempty"`
	Variables         []MCPExtensionVariable         `json:"variables,omitempty"`
}

// MCPExtensionListResponse is the response structure for listing extensions without circular references
type MCPExtensionListResponse struct {
	Extensions []MCPExtension `json:"extensions"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// MCPExecutionStep is a simplified ExecutionStep without the circular Execution reference
type MCPExecutionStep struct {
	ID          uuid.UUID                    `json:"id"`
	ExecutionID uuid.UUID                    `json:"execution_id"`
	StepName    string                       `json:"step_name"`
	Phase       string                       `json:"phase"`
	StepOrder   int                          `json:"step_order"`
	StartedAt   time.Time                    `json:"started_at"`
	CompletedAt *time.Time                   `json:"completed_at,omitempty"`
	Status      shared_types.ExecutionStatus `json:"status"`
	ExitCode    int                          `json:"exit_code"`
	Output      string                       `json:"output"`
	CreatedAt   time.Time                    `json:"created_at"`
}

// MCPExtensionExecution is a simplified ExtensionExecution without the circular Extension reference
type MCPExtensionExecution struct {
	ID             uuid.UUID                    `json:"id"`
	ExtensionID    uuid.UUID                    `json:"extension_id"`
	ServerHostname string                       `json:"server_hostname"`
	VariableValues string                       `json:"variable_values"`
	Status         shared_types.ExecutionStatus `json:"status"`
	StartedAt      time.Time                    `json:"started_at"`
	CompletedAt    *time.Time                   `json:"completed_at,omitempty"`
	ExitCode       int                          `json:"exit_code"`
	ErrorMessage   string                       `json:"error_message"`
	ExecutionLog   string                       `json:"execution_log"`
	LogSeq         int64                        `json:"log_seq"`
	CreatedAt      time.Time                    `json:"created_at"`
	Steps          []MCPExecutionStep           `json:"steps,omitempty"`
}

// ListExtensionsInput is the input structure for the MCP tool
type ListExtensionsInput struct {
	OrganizationID string `json:"organization_id" jsonschema:"required"`
	Category       string `json:"category,omitempty"`
	Type           string `json:"type,omitempty"`
	Search         string `json:"search,omitempty"`
	SortBy         string `json:"sort_by,omitempty"`
	SortDir        string `json:"sort_dir,omitempty"`
	Page           *int   `json:"page,omitempty"`
	PageSize       *int   `json:"page_size,omitempty"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i ListExtensionsInput) GetOrganizationID() string {
	return i.OrganizationID
}

// ListExtensionsOutput is the output structure for the MCP tool
type ListExtensionsOutput struct {
	Response MCPExtensionListResponse `json:"response"`
}

// GetExtensionInput is the input structure for the MCP tool
type GetExtensionInput struct {
	ID             string `json:"id" jsonschema:"required"`
	OrganizationID string `json:"organization_id" jsonschema:"required"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i GetExtensionInput) GetOrganizationID() string {
	return i.OrganizationID
}

// GetExtensionOutput is the output structure for the MCP tool
type GetExtensionOutput struct {
	Extension MCPExtension `json:"extension"`
}

// RunExtensionInput is the input structure for the MCP tool
type RunExtensionInput struct {
	ExtensionID    string                 `json:"extension_id" jsonschema:"required"`
	OrganizationID string                 `json:"organization_id" jsonschema:"required"`
	Variables      map[string]interface{} `json:"variables,omitempty"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i RunExtensionInput) GetOrganizationID() string {
	return i.OrganizationID
}

// RunExtensionOutput is the output structure for the MCP tool
type RunExtensionOutput struct {
	Execution MCPExtensionExecution `json:"execution"`
}

// GetExecutionInput is the input structure for the MCP tool
type GetExecutionInput struct {
	ExecutionID    string `json:"execution_id" jsonschema:"required"`
	OrganizationID string `json:"organization_id" jsonschema:"required"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i GetExecutionInput) GetOrganizationID() string {
	return i.OrganizationID
}

// GetExecutionOutput is the output structure for the MCP tool
type GetExecutionOutput struct {
	Execution MCPExtensionExecution `json:"execution"`
}

// ListExecutionLogsInput is the input structure for the MCP tool
type ListExecutionLogsInput struct {
	ExecutionID    string `json:"execution_id" jsonschema:"required"`
	OrganizationID string `json:"organization_id" jsonschema:"required"`
	AfterSeq       *int64 `json:"after_seq,omitempty"`
	Limit          *int   `json:"limit,omitempty"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i ListExecutionLogsInput) GetOrganizationID() string {
	return i.OrganizationID
}

// ListExecutionLogsOutput is the output structure for the MCP tool
type ListExecutionLogsOutput struct {
	Logs            []shared_types.ExtensionLog   `json:"logs"`
	NextAfter       int64                         `json:"next_after"`
	ExecutionStatus *shared_types.ExecutionStatus `json:"execution_status,omitempty"`
}

// CancelExecutionInput is the input structure for the MCP tool
type CancelExecutionInput struct {
	ExecutionID    string `json:"execution_id" jsonschema:"required"`
	OrganizationID string `json:"organization_id" jsonschema:"required"`
}

// GetOrganizationID implements OrganizationIDExtractor interface
func (i CancelExecutionInput) GetOrganizationID() string {
	return i.OrganizationID
}

// CancelExecutionOutput is the output structure for the MCP tool
type CancelExecutionOutput struct {
	Message string `json:"message"`
}
