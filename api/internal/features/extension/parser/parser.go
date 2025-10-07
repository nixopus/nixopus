package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/types"
	"gopkg.in/yaml.v3"
)

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
}

type ExtensionVariable struct {
	Type              string      `yaml:"type"`
	Description       string      `yaml:"description"`
	Default           interface{} `yaml:"default"`
	IsRequired        bool        `yaml:"is_required"`
	ValidationPattern string      `yaml:"validation_pattern"`
}

type ExecutionStep struct {
	Name         string                 `yaml:"name"`
	Type         string                 `yaml:"type"`
	Properties   map[string]interface{} `yaml:"properties"`
	Conditions   []string               `yaml:"conditions,omitempty"`
	IgnoreErrors bool                   `yaml:"ignore_errors,omitempty"`
	Timeout      int                    `yaml:"timeout,omitempty"`
}

type ExtensionExecution struct {
	Run      []ExecutionStep `yaml:"run"`
	Validate []ExecutionStep `yaml:"validate"`
}

type ExtensionYAML struct {
	Metadata  ExtensionMetadata            `yaml:"metadata"`
	Variables map[string]ExtensionVariable `yaml:"variables"`
	Execution ExtensionExecution           `yaml:"execution"`
}

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseExtensionFile(filePath string) (*types.Extension, []types.ExtensionVariable, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var extYAML ExtensionYAML
	if err := yaml.Unmarshal(data, &extYAML); err != nil {
		return nil, nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := p.validateExtension(&extYAML); err != nil {
		return nil, nil, fmt.Errorf("validation failed: %w", err)
	}

	extension := p.convertToExtension(&extYAML, string(data))
	variables := p.convertToVariables(&extYAML, extension.ExtensionID)

	return extension, variables, nil
}

func (p *Parser) ParseExtensionContent(content string) (*types.Extension, []types.ExtensionVariable, error) {
	var extYAML ExtensionYAML
	if err := yaml.Unmarshal([]byte(content), &extYAML); err != nil {
		return nil, nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	if err := p.validateExtension(&extYAML); err != nil {
		return nil, nil, fmt.Errorf("validation failed: %w", err)
	}
	extension := p.convertToExtension(&extYAML, content)
	variables := p.convertToVariables(&extYAML, extension.ExtensionID)
	return extension, variables, nil
}

func (p *Parser) validateExtension(ext *ExtensionYAML) error {
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

	if len(ext.Execution.Run) == 0 && len(ext.Execution.Validate) == 0 {
		return fmt.Errorf("execution must have at least one step")
	}
	for _, step := range append([]ExecutionStep{}, ext.Execution.Run...) {
		if err := p.validateStep(step); err != nil {
			return err
		}
	}
	for _, step := range append([]ExecutionStep{}, ext.Execution.Validate...) {
		if err := p.validateStep(step); err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) validateStep(step ExecutionStep) error {
	if strings.TrimSpace(step.Name) == "" {
		return fmt.Errorf("execution step name is required")
	}
	allowedTypes := map[string]bool{"command": true, "file": true, "service": true, "user": true}
	if !allowedTypes[step.Type] {
		return fmt.Errorf("invalid execution step type: %s", step.Type)
	}
	if step.Type == "command" {
		if _, ok := step.Properties["cmd"]; !ok {
			return fmt.Errorf("command step requires 'cmd' property")
		}
	}
	if step.Type == "file" {
		if _, ok := step.Properties["action"]; !ok {
			return fmt.Errorf("file step requires 'action' property")
		}
		if _, ok := step.Properties["src"]; !ok {
			return fmt.Errorf("file step requires 'src' property")
		}
		if _, ok := step.Properties["dest"]; !ok {
			return fmt.Errorf("file step requires 'dest' property")
		}
	}
	if step.Type == "service" {
		if _, ok := step.Properties["name"]; !ok {
			return fmt.Errorf("service step requires 'name' property")
		}
		if _, ok := step.Properties["action"]; !ok {
			return fmt.Errorf("service step requires 'action' property")
		}
	}
	if step.Type == "user" {
		if _, ok := step.Properties["username"]; !ok {
			return fmt.Errorf("user step requires 'username' property")
		}
		if _, ok := step.Properties["action"]; !ok {
			return fmt.Errorf("user step requires 'action' property")
		}
	}
	if step.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}
	return nil
}

func (p *Parser) isValidCategory(category string) bool {
	validCategories := []string{
		"Security", "Containers", "Database", "Web Server",
		"Maintenance", "Monitoring", "Storage", "Network",
		"Development", "Other",
	}
	for _, valid := range validCategories {
		if category == valid {
			return true
		}
	}
	return false
}

func (p *Parser) isValidExtensionID(id string) bool {
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

func (p *Parser) isValidVersion(version string) bool {
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

func (p *Parser) isValidVariableName(name string) bool {
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

func (p *Parser) isValidVariableType(varType string) bool {
	validTypes := []string{"string", "integer", "boolean", "array"}
	for _, valid := range validTypes {
		if varType == valid {
			return true
		}
	}
	return false
}

func (p *Parser) convertToExtension(extYAML *ExtensionYAML, yamlContent string) *types.Extension {
	parsedContent, err := json.Marshal(extYAML)
	if err != nil {
		parsedContent = []byte("{}")
	}
	hash := sha256.Sum256([]byte(yamlContent))

	return &types.Extension{
		ExtensionID:      extYAML.Metadata.ID,
		Name:             extYAML.Metadata.Name,
		Description:      extYAML.Metadata.Description,
		Author:           extYAML.Metadata.Author,
		Icon:             extYAML.Metadata.Icon,
		Category:         types.ExtensionCategory(extYAML.Metadata.Category),
		ExtensionType:    types.ExtensionType(extYAML.Metadata.Type),
		Version:          extYAML.Metadata.Version,
		IsVerified:       extYAML.Metadata.IsVerified,
		YAMLContent:      yamlContent,
		ParsedContent:    string(parsedContent),
		ContentHash:      hex.EncodeToString(hash[:]),
		ValidationStatus: types.ValidationStatusValid,
	}
}

func (p *Parser) convertToVariables(extYAML *ExtensionYAML, extensionID string) []types.ExtensionVariable {
	var variables []types.ExtensionVariable

	for varName, variable := range extYAML.Variables {
		defaultValueJSON, err := json.Marshal(variable.Default)
		if err != nil {
			defaultValueJSON = []byte("null")
		}

		variables = append(variables, types.ExtensionVariable{
			VariableName:      varName,
			VariableType:      variable.Type,
			Description:       variable.Description,
			DefaultValue:      json.RawMessage(defaultValueJSON),
			IsRequired:        variable.IsRequired,
			ValidationPattern: variable.ValidationPattern,
		})
	}

	return variables
}

func (p *Parser) LoadExtensionsFromDirectory(dirPath string) ([]*types.Extension, [][]types.ExtensionVariable, error) {
	var extensions []*types.Extension
	var allVariables [][]types.ExtensionVariable

	files, err := filepath.Glob(filepath.Join(dirPath, "*.yaml"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, file := range files {
		// Skip rfc.yaml file
		if filepath.Base(file) == "rfc.yaml" {
			continue
		}

		extension, variables, err := p.ParseExtensionFile(file)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse %s: %w", file, err)
		}

		extensions = append(extensions, extension)
		allVariables = append(allVariables, variables)
	}

	return extensions, allVariables, nil
}
