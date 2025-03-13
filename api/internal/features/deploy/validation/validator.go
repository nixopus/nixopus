package validation

import (
	"encoding/json"
	"io"

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
		return validateDeploymentRequest(*r)
	case *types.UpdateDeploymentRequest:
		return validateUpdateDeploymentRequest(*r)
	case *types.DeleteDeploymentRequest:
		return validateDeleteDeploymentRequest(*r)
	case *types.ReDeployApplicationRequest:
		return validateRedeployApplicationRequest(*r)
	default:
		return types.ErrInvalidRequestType
	}
}

func validateDeploymentRequest(req types.CreateDeploymentRequest) error {
	if req.Name == "" {
		return types.ErrMissingName
	}
	if req.DomainID == uuid.Nil {
		return types.ErrMissingDomainID
	}
	if req.Repository == "" {
		return types.ErrMissingRepository
	}
	if req.Branch == "" {
		return types.ErrMissingBranch
	}
	if req.Port <= 0 {
		return types.ErrMissingPort
	}

	if !isValidEnvironment(req.Environment) {
		return types.ErrInvalidEnvironment
	}
	if !isValidBuildPack(req.BuildPack) {
		return types.ErrInvalidBuildPack
	}

	return nil
}

func validateUpdateDeploymentRequest(req types.UpdateDeploymentRequest) error {
	// here we need to validate whether user has access to update the deployment
	if req.ID == uuid.Nil {
		return types.ErrMissingID
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

func isValidEnvironment(env shared_types.Environment) bool {
	validEnvs := []shared_types.Environment{
		shared_types.Development,
		shared_types.Staging,
		shared_types.Production,
	}
	for _, v := range validEnvs {
		if env == v {
			return true
		}
	}
	return false
}

func isValidBuildPack(bp shared_types.BuildPack) bool {
	validBPs := []shared_types.BuildPack{
		shared_types.DockerFile,
		shared_types.DockerCompose,
		shared_types.Static,
	}
	for _, v := range validBPs {
		if bp == v {
			return true
		}
	}
	return false
}
