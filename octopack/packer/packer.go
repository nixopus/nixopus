package packer

import (
	"encoding/json"
	"fmt"
	"io"
)

// FromJSON reads JSON data and returns a DockerfileSpec.
// This is a convenience function for library users.
func FromJSON(data []byte) (*DockerfileSpec, error) {
	var spec DockerfileSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}
	return &spec, nil
}

// FromJSONReader reads JSON from an io.Reader and returns a DockerfileSpec.
// This is a convenience function for library users.
func FromJSONReader(reader io.Reader) (*DockerfileSpec, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return FromJSON(data)
}

// ProcessJSON is a convenience function that reads JSON, validates it, and generates a Dockerfile.
// If validate is true, it will return an error if validation fails.
// Returns the generated Dockerfile string and any errors.
// If validation fails, returns a ValidationError with details about all errors.
func ProcessJSON(data []byte, validate bool) (string, error) {
	spec, err := FromJSON(data)
	if err != nil {
		return "", err
	}

	if validate {
		result := Validate(spec)
		if result.HasErrors() {
			// Return first error with context
			return "", fmt.Errorf("validation failed: %s", result.Errors[0].Error())
		}
	}

	return Generate(spec)
}

// ProcessJSONReader is a convenience function that reads JSON from an io.Reader,
// validates it, and generates a Dockerfile.
// If validate is true, it will return an error if validation fails.
// Returns the generated Dockerfile string and any errors.
func ProcessJSONReader(reader io.Reader, validate bool) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return ProcessJSON(data, validate)
}
