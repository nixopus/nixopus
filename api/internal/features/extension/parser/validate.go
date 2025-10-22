package parser

import (
	"fmt"
	"strings"
)

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
	validators := map[string]func(ExecutionStep) error{
		"command":        p.validateCommandStep,
		"file":           p.validateFileStep,
		"service":        p.validateServiceStep,
		"user":           p.validateUserStep,
		"docker":         p.validateDockerStep,
		"docker_compose": p.validateDockerComposeStep,
		"proxy":          p.validateProxyStep,
	}
	v, ok := validators[step.Type]
	if !ok {
		return fmt.Errorf("invalid execution step type: %s", step.Type)
	}
	if err := v(step); err != nil {
		return err
	}
	if step.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}
	return nil
}

func (p *Parser) validateCommandStep(step ExecutionStep) error {
	return p.requireProps(step, map[string]bool{"cmd": true})
}

func (p *Parser) validateFileStep(step ExecutionStep) error {
	if err := p.requireProps(step, map[string]bool{"action": true}); err != nil {
		return err
	}
	action, _ := step.Properties["action"].(string)
	if action == "mkdir" {
		return p.requireProps(step, map[string]bool{"dest": true})
	}
	return p.requireProps(step, map[string]bool{"src": true, "dest": true})
}

func (p *Parser) validateServiceStep(step ExecutionStep) error {
	return p.requireProps(step, map[string]bool{"name": true, "action": true})
}

func (p *Parser) validateUserStep(step ExecutionStep) error {
	return p.requireProps(step, map[string]bool{"username": true, "action": true})
}

func (p *Parser) validateDockerStep(step ExecutionStep) error {
	if err := p.requireProps(step, map[string]bool{"action": true}); err != nil {
		return err
	}
	action, _ := step.Properties["action"].(string)
	switch action {
	case "pull":
		return p.requireProps(step, map[string]bool{"image": true})
	case "run":
		// name and image are required; optional: tag, ports, restart, env, volumes, networks, cmd
		return p.requireProps(step, map[string]bool{"name": true, "image": true})
	case "stop", "start", "rm":
		return p.requireProps(step, map[string]bool{"name": true})
	default:
		return fmt.Errorf("unsupported docker action: %s", action)
	}
}

func (p *Parser) validateDockerComposeStep(step ExecutionStep) error {
	if err := p.requireProps(step, map[string]bool{"action": true, "file": true}); err != nil {
		return err
	}
	action, _ := step.Properties["action"].(string)
	switch action {
	case "up", "down", "build":
		return nil
	default:
		return fmt.Errorf("unsupported docker_compose action: %s", action)
	}
}

func (p *Parser) validateProxyStep(step ExecutionStep) error {
	if err := p.requireProps(step, map[string]bool{"action": true}); err != nil {
		return err
	}
	action, _ := step.Properties["action"].(string)
	switch action {
	case "add", "update":
		return p.requireProps(step, map[string]bool{"domain": true, "port": true})
	case "remove":
		return p.requireProps(step, map[string]bool{"domain": true})
	default:
		return fmt.Errorf("unsupported proxy action: %s", action)
	}
}

func (p *Parser) requireProps(step ExecutionStep, required map[string]bool) error {
	for key := range required {
		if _, ok := step.Properties[key]; !ok {
			return fmt.Errorf("%s step requires '%s' property", step.Type, key)
		}
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
