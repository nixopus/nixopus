package parser

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
