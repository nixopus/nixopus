package types

import (
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// ExtensionMetadata is the metadata block from extension metadata.yaml.
type ExtensionMetadata struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Author      string `yaml:"author"`
	Icon        string `yaml:"icon"`
	Category    string `yaml:"category"`
	Type        string `yaml:"type"`
	Version     string `yaml:"version"`
	IsVerified  bool   `yaml:"isVerified"`
	Featured    bool   `yaml:"featured"`
}

// ExtensionVariable is one variable definition from metadata.yaml.
type ExtensionVariable struct {
	Type              string      `yaml:"type"`
	Description       string      `yaml:"description"`
	Default           interface{} `yaml:"default"`
	IsRequired        bool        `yaml:"is_required"`
	ValidationPattern string      `yaml:"validation_pattern"`
}

// ExtensionYAML is the root structure of extension metadata.yaml.
type ExtensionYAML struct {
	Metadata  ExtensionMetadata            `yaml:"metadata"`
	Variables map[string]ExtensionVariable `yaml:"variables"`
}

// MessageResponse is a generic response with just status and message
type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ExtensionResponse is the typed response for single extension operations
type ExtensionResponse struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Data    shared_types.Extension `json:"data"`
}

// CategoriesResponse is the typed response for listing categories
type CategoriesResponse struct {
	Status  string                           `json:"status"`
	Message string                           `json:"message"`
	Data    []shared_types.ExtensionCategory `json:"data"`
}

// ListExtensionsResponse wraps the extension list response
type ListExtensionsResponse struct {
	Status  string                             `json:"status"`
	Message string                             `json:"message"`
	Data    shared_types.ExtensionListResponse `json:"data"`
}
