package validation

import (
	"strings"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/healthcheck/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/healthcheck/types"
)

type Validator struct {
	storage storage.HealthCheckRepository
}

func NewValidator(repository storage.HealthCheckRepository) *Validator {
	return &Validator{
		storage: repository,
	}
}

func (v *Validator) ValidateRequest(req interface{}) error {
	switch r := req.(type) {
	case *types.CreateHealthCheckRequest:
		return v.validateCreateHealthCheckRequest(*r)
	case *types.UpdateHealthCheckRequest:
		return v.validateUpdateHealthCheckRequest(*r)
	case *types.ToggleHealthCheckRequest:
		return v.validateToggleHealthCheckRequest(*r)
	case *types.GetHealthCheckResultsRequest:
		return v.validateGetHealthCheckResultsRequest(*r)
	case *types.GetHealthCheckStatsRequest:
		return v.validateGetHealthCheckStatsRequest(*r)
	default:
		return types.ErrInvalidRequestType
	}
}

func (v *Validator) validateCreateHealthCheckRequest(req types.CreateHealthCheckRequest) error {
	if req.ApplicationID == "" {
		return types.ErrInvalidApplicationID
	}

	if _, err := uuid.Parse(req.ApplicationID); err != nil {
		return types.ErrInvalidApplicationID
	}

	if req.Endpoint == "" {
		req.Endpoint = "/"
	}

	// Accept either a path (starting with "/") or a full URL (starting with "http://" or "https://")
	if !strings.HasPrefix(req.Endpoint, "/") && !strings.HasPrefix(req.Endpoint, "http://") && !strings.HasPrefix(req.Endpoint, "https://") {
		return types.ErrInvalidEndpoint
	}

	if req.Method == "" {
		req.Method = "GET"
	}

	validMethods := map[string]bool{"GET": true, "POST": true, "HEAD": true}
	if !validMethods[strings.ToUpper(req.Method)] {
		return types.ErrInvalidMethod
	}

	if req.TimeoutSeconds == 0 {
		req.TimeoutSeconds = 30
	}
	if req.TimeoutSeconds < 5 || req.TimeoutSeconds > 120 {
		return types.ErrInvalidTimeout
	}

	if req.IntervalSeconds == 0 {
		req.IntervalSeconds = 60
	}
	if req.IntervalSeconds < 30 || req.IntervalSeconds > 3600 {
		return types.ErrInvalidInterval
	}

	if req.FailureThreshold == 0 {
		req.FailureThreshold = 3
	}
	if req.FailureThreshold < 1 || req.FailureThreshold > 10 {
		return types.ErrInvalidThreshold
	}

	if req.SuccessThreshold == 0 {
		req.SuccessThreshold = 1
	}
	if req.SuccessThreshold < 1 || req.SuccessThreshold > 10 {
		return types.ErrInvalidThreshold
	}

	if req.RetentionDays == 0 {
		req.RetentionDays = 30
	}
	if req.RetentionDays < 1 || req.RetentionDays > 365 {
		return types.ErrInvalidRetentionDays
	}

	if len(req.ExpectedStatus) == 0 {
		req.ExpectedStatus = []int{200}
	}

	return nil
}

func (v *Validator) validateUpdateHealthCheckRequest(req types.UpdateHealthCheckRequest) error {
	if req.ApplicationID == "" {
		return types.ErrInvalidApplicationID
	}

	if _, err := uuid.Parse(req.ApplicationID); err != nil {
		return types.ErrInvalidApplicationID
	}

	// Accept either a path (starting with "/") or a full URL (starting with "http://" or "https://")
	if req.Endpoint != "" && !strings.HasPrefix(req.Endpoint, "/") && !strings.HasPrefix(req.Endpoint, "http://") && !strings.HasPrefix(req.Endpoint, "https://") {
		return types.ErrInvalidEndpoint
	}

	if req.Method != "" {
		validMethods := map[string]bool{"GET": true, "POST": true, "HEAD": true}
		if !validMethods[strings.ToUpper(req.Method)] {
			return types.ErrInvalidMethod
		}
	}

	if req.TimeoutSeconds != 0 && (req.TimeoutSeconds < 5 || req.TimeoutSeconds > 120) {
		return types.ErrInvalidTimeout
	}

	if req.IntervalSeconds != 0 && (req.IntervalSeconds < 30 || req.IntervalSeconds > 3600) {
		return types.ErrInvalidInterval
	}

	if req.FailureThreshold != 0 && (req.FailureThreshold < 1 || req.FailureThreshold > 10) {
		return types.ErrInvalidThreshold
	}

	if req.SuccessThreshold != 0 && (req.SuccessThreshold < 1 || req.SuccessThreshold > 10) {
		return types.ErrInvalidThreshold
	}

	if req.RetentionDays != 0 && (req.RetentionDays < 1 || req.RetentionDays > 365) {
		return types.ErrInvalidRetentionDays
	}

	return nil
}

func (v *Validator) validateToggleHealthCheckRequest(req types.ToggleHealthCheckRequest) error {
	if req.ApplicationID == "" {
		return types.ErrInvalidApplicationID
	}

	if _, err := uuid.Parse(req.ApplicationID); err != nil {
		return types.ErrInvalidApplicationID
	}

	return nil
}

func (v *Validator) validateGetHealthCheckResultsRequest(req types.GetHealthCheckResultsRequest) error {
	if req.ApplicationID == "" {
		return types.ErrInvalidApplicationID
	}

	if _, err := uuid.Parse(req.ApplicationID); err != nil {
		return types.ErrInvalidApplicationID
	}

	if req.Limit < 0 || req.Limit > 1000 {
		req.Limit = 100
	}

	return nil
}

func (v *Validator) validateGetHealthCheckStatsRequest(req types.GetHealthCheckStatsRequest) error {
	if req.ApplicationID == "" {
		return types.ErrInvalidApplicationID
	}

	if _, err := uuid.Parse(req.ApplicationID); err != nil {
		return types.ErrInvalidApplicationID
	}

	return nil
}
