package validation

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type Validator struct {
}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ParseRequestBody(req interface{}, body io.ReadCloser, decoded interface{}) error {
	return json.NewDecoder(body).Decode(decoded)
}

func (v *Validator) ValidateRequest(req interface{}) error {
	switch r := req.(type) {
	case *types.CreateDeploymentRequest:
		return validateDeploymentRequest(r)
	case *types.UpdateDeploymentRequest:
		return validateUpdateDeploymentRequest(r)
	case *types.DeleteDeploymentRequest:
		return validateDeleteDeploymentRequest(*r)
	case *types.ReDeployApplicationRequest:
		return validateRedeployApplicationRequest(*r)
	case *types.RollbackDeploymentRequest:
		return validateRollbackDeploymentRequest(*r)
	case *types.RestartDeploymentRequest:
		return validateRestartDeploymentRequest(*r)
	default:
		return types.ErrInvalidRequestType
	}
}

func validateDeploymentRequest(req *types.CreateDeploymentRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.Domain == "" {
		return errors.New("domain is required")
	}
	if req.Environment == "" {
		return errors.New("environment is required")
	}
	if req.BuildPack == "" {
		return errors.New("build_pack is required")
	}

	// Docker Compose has different validation rules
	if req.BuildPack == shared_types.DockerCompose {
		return validateDockerComposeRequest(req)
	}

	// Standard Dockerfile validation
	if req.Repository == "" {
		return errors.New("repository is required")
	}
	if req.Branch == "" {
		return errors.New("branch is required")
	}
	if req.Port == 0 {
		return errors.New("port is required")
	}
	if req.BasePath == "" {
		req.BasePath = "/"
	} else if req.BasePath[0] != '/' {
		req.BasePath = "/" + req.BasePath
	}
	return nil
}

// validateDockerComposeRequest handles validation for docker compose deployments
// Supports three source types: repository, URL, or raw content
func validateDockerComposeRequest(req *types.CreateDeploymentRequest) error {
	hasRepository := req.Repository != ""
	hasURL := req.ComposeFileURL != ""
	hasContent := req.ComposeFileContent != ""

	sourceCount := 0
	if hasRepository {
		sourceCount++
	}
	if hasURL {
		sourceCount++
	}
	if hasContent {
		sourceCount++
	}

	if sourceCount == 0 {
		return errors.New("docker-compose requires one of: repository, compose_file_url, or compose_file_content")
	}
	if sourceCount > 1 {
		return errors.New("docker-compose accepts only one source: repository, compose_file_url, or compose_file_content")
	}

	if hasRepository {
		if req.Branch == "" {
			return errors.New("branch is required when using repository")
		}
	}

	if hasURL {
		if !strings.HasPrefix(req.ComposeFileURL, "http://") && !strings.HasPrefix(req.ComposeFileURL, "https://") {
			return errors.New("compose_file_url must be a valid HTTP/HTTPS URL")
		}
	}

	if hasContent {
		if len(req.ComposeFileContent) < 10 {
			return errors.New("compose_file_content appears to be too short")
		}
	}

	if req.BasePath == "" {
		req.BasePath = "/"
	} else if req.BasePath[0] != '/' {
		req.BasePath = "/" + req.BasePath
	}

	return nil
}

func validateUpdateDeploymentRequest(req *types.UpdateDeploymentRequest) error {
	if req.Name != "" {
		if len(req.Name) < 3 {
			return errors.New("name must be at least 3 characters")
		}
	}
	if req.Port != 0 {
		if req.Port < 1 || req.Port > 65535 {
			return errors.New("port must be between 1 and 65535")
		}
	}
	if req.BasePath != "" {
		if req.BasePath[0] != '/' {
			req.BasePath = "/" + req.BasePath
		}
	}
	return nil
}

func validateDeleteDeploymentRequest(req types.DeleteDeploymentRequest) error {
	// 	// here we need to validate whether user has access to delete the deployment
	if req.ID == uuid.Nil {
		return types.ErrMissingID
	}
	return nil
}

func validateRedeployApplicationRequest(req types.ReDeployApplicationRequest) error {
	if req.ID == uuid.Nil {
		return types.ErrMissingID
	}
	return nil
}

func validateRollbackDeploymentRequest(req types.RollbackDeploymentRequest) error {
	if req.ID == uuid.Nil {
		return types.ErrMissingID
	}
	return nil
}

func validateRestartDeploymentRequest(req types.RestartDeploymentRequest) error {
	if req.ID == uuid.Nil {
		return types.ErrMissingID
	}
	return nil
}
