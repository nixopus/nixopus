package llm

import (
	"encoding/json"
	"fmt"
)

// ValidateToolArgs validates raw JSON arguments against a JSON Schema (subset).
//
// Enforced constraints:
//   - args must be a JSON object
//   - all "required" fields must be present
//   - declared property types (string/integer/number/boolean/object/array) must match
//   - enum constraints on string properties must be satisfied
func ValidateToolArgs(args json.RawMessage, schema json.RawMessage) error {
	var s toolSchema
	if err := json.Unmarshal(schema, &s); err != nil {
		return fmt.Errorf("schema parse error: %w", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(args, &fields); err != nil {
		return fmt.Errorf("args must be a JSON object: %w", err)
	}
	if fields == nil {
		return fmt.Errorf("args must be a JSON object, got null")
	}

	for _, req := range s.Required {
		if _, ok := fields[req]; !ok {
			return fmt.Errorf("missing required field %q", req)
		}
	}

	for name, prop := range s.Properties {
		val, ok := fields[name]
		if !ok {
			continue
		}
		if prop.Type != "" {
			if err := checkJSONType(name, val, prop.Type); err != nil {
				return err
			}
		}
		if len(prop.Enum) > 0 {
			if err := checkJSONEnum(name, val, prop.Enum); err != nil {
				return err
			}
		}
	}

	return nil
}

type toolSchema struct {
	Required   []string              `json:"required"`
	Properties map[string]schemaProp `json:"properties"`
}

type schemaProp struct {
	Type string   `json:"type"`
	Enum []string `json:"enum"`
}

func checkJSONType(field string, raw json.RawMessage, want string) error {
	if len(raw) == 0 {
		return fmt.Errorf("field %q is empty", field)
	}
	got := jsonKind(raw)
	switch want {
	case "string":
		if got != "string" {
			return fmt.Errorf("field %q: expected string, got %s", field, got)
		}
	case "integer", "number":
		if got != "number" {
			return fmt.Errorf("field %q: expected number, got %s", field, got)
		}
	case "boolean":
		if got != "boolean" {
			return fmt.Errorf("field %q: expected boolean, got %s", field, got)
		}
	case "object":
		if got != "object" {
			return fmt.Errorf("field %q: expected object, got %s", field, got)
		}
	case "array":
		if got != "array" {
			return fmt.Errorf("field %q: expected array, got %s", field, got)
		}
	}
	return nil
}

func checkJSONEnum(field string, raw json.RawMessage, allowed []string) error {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return fmt.Errorf("field %q: cannot check enum for non-string value", field)
	}
	for _, v := range allowed {
		if s == v {
			return nil
		}
	}
	return fmt.Errorf("field %q: value %q not in allowed values %v", field, s, allowed)
}

func jsonKind(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "empty"
	}
	switch raw[0] {
	case '"':
		return "string"
	case '{':
		return "object"
	case '[':
		return "array"
	case 't', 'f':
		return "boolean"
	case 'n':
		return "null"
	default:
		return "number"
	}
}
