package service

import (
	"fmt"
	"strings"

	exttypes "github.com/nixopus/nixopus/api/internal/features/extension/types"
)

func (p *extensionYAMLParser) validateExtension(ext *exttypes.ExtensionYAML) error {
	if ext.Metadata.ID == "" {
		return fmt.Errorf("metadata.id is required")
	}
	if ext.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if ext.Metadata.Description == "" {
		return fmt.Errorf("metadata.description is required")
	}
	if ext.Metadata.Author == "" {
		return fmt.Errorf("metadata.author is required")
	}
	if ext.Metadata.Icon == "" {
		return fmt.Errorf("metadata.icon is required")
	}
	if ext.Metadata.Category == "" {
		return fmt.Errorf("metadata.category is required")
	}

	if ext.Metadata.Type == "" {
		return fmt.Errorf("metadata.type is required (install or run)")
	}
	if ext.Metadata.Type != "install" && ext.Metadata.Type != "run" {
		return fmt.Errorf("invalid metadata.type: %s", ext.Metadata.Type)
	}

	if !p.isValidCategory(ext.Metadata.Category) {
		return fmt.Errorf("invalid category: %s", ext.Metadata.Category)
	}

	if !p.isValidExtensionID(ext.Metadata.ID) {
		return fmt.Errorf("invalid extension_id format: %s", ext.Metadata.ID)
	}

	if ext.Metadata.Version != "" && !p.isValidVersion(ext.Metadata.Version) {
		return fmt.Errorf("invalid version format: %s", ext.Metadata.Version)
	}

	for varName, variable := range ext.Variables {
		if !p.isValidVariableName(varName) {
			return fmt.Errorf("invalid variable name: %s", varName)
		}
		if !p.isValidVariableType(variable.Type) {
			return fmt.Errorf("invalid variable type for %s: %s", varName, variable.Type)
		}
	}

	return nil
}

func (p *extensionYAMLParser) isValidCategory(category string) bool {
	validCategories := []string{
		"Security", "Containers", "Database", "Web Server",
		"Maintenance", "Monitoring", "Storage", "Network",
		"Development", "Media", "Game", "Utility", "Productivity", "Other",
		"Social",
	}
	for _, valid := range validCategories {
		if category == valid {
			return true
		}
	}
	return false
}

func (p *extensionYAMLParser) isValidExtensionID(id string) bool {
	if len(id) < 3 || len(id) > 50 {
		return false
	}
	for _, char := range id {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			return false
		}
	}
	return !strings.HasPrefix(id, "-") && !strings.HasSuffix(id, "-")
}

func (p *extensionYAMLParser) isValidVersion(version string) bool {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 {
			return false
		}
		for _, char := range part {
			if char < '0' || char > '9' {
				return false
			}
		}
	}
	return true
}

func (p *extensionYAMLParser) isValidVariableName(name string) bool {
	if len(name) == 0 || len(name) > 100 {
		return false
	}
	if !((name[0] >= 'a' && name[0] <= 'z') || (name[0] >= 'A' && name[0] <= 'Z') || name[0] == '_') {
		return false
	}
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}
	return true
}

func (p *extensionYAMLParser) isValidVariableType(varType string) bool {
	validTypes := []string{"string", "integer", "boolean", "array"}
	for _, valid := range validTypes {
		if varType == valid {
			return true
		}
	}
	return false
}
