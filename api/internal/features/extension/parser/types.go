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

type ExtensionYAML struct {
	Metadata  ExtensionMetadata            `yaml:"metadata"`
	Variables map[string]ExtensionVariable `yaml:"variables"`
}

type Parser struct{}
