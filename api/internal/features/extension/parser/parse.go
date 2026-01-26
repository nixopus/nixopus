package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/raghavyuva/nixopus-api/internal/types"
	"gopkg.in/yaml.v3"
)

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
		Featured:         extYAML.Metadata.Featured,
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
	files, err := filepath.Glob(filepath.Join(dirPath, "*.yaml"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Filter out rfc.yaml
	validFiles := make([]string, 0, len(files))
	for _, file := range files {
		if filepath.Base(file) != "rfc.yaml" {
			validFiles = append(validFiles, file)
		}
	}

	if len(validFiles) == 0 {
		return []*types.Extension{}, [][]types.ExtensionVariable{}, nil
	}

	// Process files in parallel
	type result struct {
		extension *types.Extension
		variables []types.ExtensionVariable
		err       error
		index     int
	}

	results := make([]result, len(validFiles))
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Process files concurrently
	for i, file := range validFiles {
		wg.Add(1)
		go func(idx int, filePath string) {
			defer wg.Done()
			ext, vars, err := p.ParseExtensionFile(filePath)
			mu.Lock()
			results[idx] = result{extension: ext, variables: vars, err: err, index: idx}
			mu.Unlock()
		}(i, file)
	}

	wg.Wait()

	// Collect results and check for errors
	extensions := make([]*types.Extension, 0, len(validFiles))
	allVariables := make([][]types.ExtensionVariable, 0, len(validFiles))

	for _, res := range results {
		if res.err != nil {
			return nil, nil, fmt.Errorf("failed to parse file at index %d: %w", res.index, res.err)
		}
		extensions = append(extensions, res.extension)
		allVariables = append(allVariables, res.variables)
	}

	return extensions, allVariables, nil
}
