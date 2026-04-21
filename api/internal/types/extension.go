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
	Featured          bool              `json:"featured" bun:"featured,notnull,default:false"`
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

	Extension *Extension `json:"-" bun:"rel:belongs-to,join:extension_id=id"`
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
