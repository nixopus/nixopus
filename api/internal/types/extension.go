package types

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ExtensionCategory string

const (
	ExtensionCategorySecurity     ExtensionCategory = "Security"
	ExtensionCategoryContainers   ExtensionCategory = "Containers"
	ExtensionCategoryDatabase     ExtensionCategory = "Database"
	ExtensionCategoryWebServer    ExtensionCategory = "Web Server"
	ExtensionCategoryMaintenance  ExtensionCategory = "Maintenance"
	ExtensionCategoryMonitoring   ExtensionCategory = "Monitoring"
	ExtensionCategoryStorage      ExtensionCategory = "Storage"
	ExtensionCategoryNetwork      ExtensionCategory = "Network"
	ExtensionCategoryDevelopment  ExtensionCategory = "Development"
	ExtensionCategoryMedia        ExtensionCategory = "Media"
	ExtensionCategoryGame         ExtensionCategory = "Game"
	ExtensionCategoryUtility      ExtensionCategory = "Utility"
	ExtensionCategoryOther        ExtensionCategory = "Other"
	ExtensionCategoryProductivity ExtensionCategory = "Productivity"
	ExtensionCategorySocial       ExtensionCategory = "Social"
)

type ValidationStatus string

const (
	ValidationStatusNotValidated ValidationStatus = "not_validated"
	ValidationStatusValid        ValidationStatus = "valid"
	ValidationStatusInvalid      ValidationStatus = "invalid"
)

type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
)

type ExtensionType string

const (
	ExtensionTypeInstall ExtensionType = "install"
	ExtensionTypeRun     ExtensionType = "run"
)

type Extension struct {
	bun.BaseModel     `bun:"table:extensions,alias:e" swaggerignore:"true"`
	ID                uuid.UUID         `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	ExtensionID       string            `json:"extension_id" bun:"extension_id,unique,notnull"`
	ParentExtensionID *uuid.UUID        `json:"parent_extension_id,omitempty" bun:"parent_extension_id,nullzero"`
	Name              string            `json:"name" bun:"name,notnull"`
	Description       string            `json:"description" bun:"description,notnull"`
	Author            string            `json:"author" bun:"author,notnull"`
	Icon              string            `json:"icon" bun:"icon,notnull"`
	Category          ExtensionCategory `json:"category" bun:"category,notnull"`
	ExtensionType     ExtensionType     `json:"extension_type" bun:"extension_type,notnull"`
	Version           string            `json:"version" bun:"version"`
	IsVerified        bool              `json:"is_verified" bun:"is_verified,notnull,default:false"`
	YAMLContent       string            `json:"yaml_content" bun:"yaml_content,notnull"`
	ParsedContent     string            `json:"parsed_content" bun:"parsed_content,notnull,type:jsonb"`
	ContentHash       string            `json:"content_hash" bun:"content_hash,notnull"`
	ValidationStatus  ValidationStatus  `json:"validation_status" bun:"validation_status,default:'not_validated'"`
	ValidationErrors  string            `json:"validation_errors" bun:"validation_errors,type:jsonb"`
	CreatedAt         time.Time         `json:"created_at" bun:"created_at,notnull,default:now()"`
	UpdatedAt         time.Time         `json:"updated_at" bun:"updated_at,notnull,default:now()"`
	DeletedAt         *time.Time        `json:"deleted_at,omitempty" bun:"deleted_at"`

	Variables []ExtensionVariable `json:"variables,omitempty" bun:"rel:has-many,join:id=extension_id"`
}

type ExtensionVariable struct {
	bun.BaseModel     `bun:"table:extension_variables,alias:ev" swaggerignore:"true"`
	ID                uuid.UUID       `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	ExtensionID       uuid.UUID       `json:"extension_id" bun:"extension_id,notnull,type:uuid"`
	VariableName      string          `json:"variable_name" bun:"variable_name,notnull"`
	VariableType      string          `json:"variable_type" bun:"variable_type,notnull"`
	Description       string          `json:"description" bun:"description"`
	DefaultValue      json.RawMessage `json:"default_value" bun:"default_value,type:jsonb"`
	IsRequired        bool            `json:"is_required" bun:"is_required,default:false"`
	ValidationPattern string          `json:"validation_pattern" bun:"validation_pattern"`
	CreatedAt         time.Time       `json:"created_at" bun:"created_at,notnull,default:now()"`

	Extension *Extension `json:"extension,omitempty" bun:"rel:belongs-to,join:extension_id=id"`
}

type ExtensionExecution struct {
	bun.BaseModel  `bun:"table:extension_executions,alias:ee" swaggerignore:"true"`
	ID             uuid.UUID       `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	ExtensionID    uuid.UUID       `json:"extension_id" bun:"extension_id,notnull,type:uuid"`
	ServerHostname string          `json:"server_hostname" bun:"server_hostname"`
	VariableValues string          `json:"variable_values" bun:"variable_values,type:jsonb"`
	Status         ExecutionStatus `json:"status" bun:"status,default:'pending'"`
	StartedAt      time.Time       `json:"started_at" bun:"started_at,notnull,default:now()"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty" bun:"completed_at"`
	ExitCode       int             `json:"exit_code" bun:"exit_code"`
	ErrorMessage   string          `json:"error_message" bun:"error_message"`
	ExecutionLog   string          `json:"execution_log" bun:"execution_log"`
	LogSeq         int64           `json:"log_seq" bun:"log_seq"`
	CreatedAt      time.Time       `json:"created_at" bun:"created_at,notnull,default:now()"`

	Extension *Extension      `json:"extension,omitempty" bun:"rel:belongs-to,join:extension_id=id"`
	Steps     []ExecutionStep `json:"steps,omitempty" bun:"rel:has-many,join:id=execution_id"`
}

type ExecutionStep struct {
	bun.BaseModel `bun:"table:execution_steps,alias:es" swaggerignore:"true"`
	ID            uuid.UUID       `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	ExecutionID   uuid.UUID       `json:"execution_id" bun:"execution_id,notnull,type:uuid"`
	StepName      string          `json:"step_name" bun:"step_name,notnull"`
	Phase         string          `json:"phase" bun:"phase,notnull"`
	StepOrder     int             `json:"step_order" bun:"step_order,notnull"`
	StartedAt     time.Time       `json:"started_at" bun:"started_at,notnull,default:now()"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty" bun:"completed_at"`
	Status        ExecutionStatus `json:"status" bun:"status,default:'pending'"`
	ExitCode      int             `json:"exit_code" bun:"exit_code"`
	Output        string          `json:"output" bun:"output"`
	CreatedAt     time.Time       `json:"created_at" bun:"created_at,notnull,default:now()"`

	Execution *ExtensionExecution `json:"execution,omitempty" bun:"rel:belongs-to,join:execution_id=id"`
}

type ExtensionLog struct {
	bun.BaseModel `bun:"table:extension_logs,alias:el" swaggerignore:"true"`
	ID            uuid.UUID       `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	ExecutionID   uuid.UUID       `json:"execution_id" bun:"execution_id,notnull,type:uuid"`
	StepID        *uuid.UUID      `json:"step_id,omitempty" bun:"step_id,nullzero,type:uuid"`
	Level         string          `json:"level" bun:"level,notnull"`
	Message       string          `json:"message" bun:"message,notnull"`
	Data          json.RawMessage `json:"data" bun:"data,notnull,type:jsonb"`
	Sequence      int64           `json:"sequence" bun:"sequence,notnull"`
	CreatedAt     time.Time       `json:"created_at" bun:"created_at,notnull,default:now()"`
}

// SpecStep defines a single step in the extension spec (parsed from YAML/JSON)
type SpecStep struct {
	Name         string                 `json:"Name"`
	Type         string                 `json:"Type"`
	Properties   map[string]interface{} `json:"Properties"`
	IgnoreErrors bool                   `json:"IgnoreErrors"`
	Timeout      int                    `json:"Timeout"`
}

// ExtensionSpec is the parsed extension content used for execution
type ExtensionSpec struct {
	Metadata struct {
		ID          string `json:"ID"`
		Name        string `json:"Name"`
		Description string `json:"Description"`
		Author      string `json:"Author"`
		Icon        string `json:"Icon"`
		Category    string `json:"Category"`
		Type        string `json:"Type"`
		Version     string `json:"Version"`
		IsVerified  bool   `json:"IsVerified"`
	} `json:"Metadata"`
	Variables map[string]struct {
		Type              string      `json:"Type"`
		Description       string      `json:"Description"`
		Default           interface{} `json:"Default"`
		IsRequired        bool        `json:"IsRequired"`
		ValidationPattern string      `json:"ValidationPattern"`
	} `json:"Variables"`
	Execution struct {
		Run      []SpecStep `json:"Run"`
		Validate []SpecStep `json:"Validate"`
	} `json:"Execution"`
}

type SortDirection string

const (
	SortDirectionAsc  SortDirection = "asc"
	SortDirectionDesc SortDirection = "desc"
)

type ExtensionSortField string

const (
	ExtensionSortFieldName       ExtensionSortField = "name"
	ExtensionSortFieldAuthor     ExtensionSortField = "author"
	ExtensionSortFieldCategory   ExtensionSortField = "category"
	ExtensionSortFieldIsVerified ExtensionSortField = "is_verified"
	ExtensionSortFieldCreatedAt  ExtensionSortField = "created_at"
	ExtensionSortFieldUpdatedAt  ExtensionSortField = "updated_at"
)

type ExtensionListParams struct {
	Category *ExtensionCategory `json:"category,omitempty"`
	Type     *ExtensionType     `json:"type,omitempty"`
	Search   string             `json:"search,omitempty"`
	SortBy   ExtensionSortField `json:"sort_by,omitempty"`
	SortDir  SortDirection      `json:"sort_dir,omitempty"`
	Page     int                `json:"page,omitempty"`
	PageSize int                `json:"page_size,omitempty"`
}

type ExtensionListResponse struct {
	Extensions []Extension `json:"extensions"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}
