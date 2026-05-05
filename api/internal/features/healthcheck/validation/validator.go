package validation

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/healthcheck/storage"
	"github.com/nixopus/nixopus/api/internal/features/healthcheck/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

type Validator struct {
	storage storage.HealthCheckRepository
	Logger  *logger.Logger // optional; nil disables validation logs
}

func NewValidator(repository storage.HealthCheckRepository) *Validator {
	return NewValidatorWithLogger(repository, nil)
}

// NewValidatorWithLogger is like NewValidator but attaches a logger for Debug detail on rule failures.
func NewValidatorWithLogger(repository storage.HealthCheckRepository, log *logger.Logger) *Validator {
	return &Validator{
		storage: repository,
		Logger:  log,
	}
}

func (v *Validator) valog(sev logger.Severity, msg, data string) {
	if v == nil || v.Logger == nil {
		return
	}
	v.Logger.Log(sev, msg, data)
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
		v.valog(logger.Debug, "healthcheck validation: invalid request type", fmt.Sprintf("%T", req))
		return types.ErrInvalidRequestType
	}
}

func (v *Validator) validateCreateHealthCheckRequest(req types.CreateHealthCheckRequest) error {
	if req.ApplicationID == "" {
		v.valog(logger.Debug, "healthcheck validation: create rejected", "reason=invalid_application_id")
		return types.ErrInvalidApplicationID
	}

	if _, err := uuid.Parse(req.ApplicationID); err != nil {
		v.valog(logger.Debug, "healthcheck validation: create rejected", "reason=invalid_application_id_parse")
		return types.ErrInvalidApplicationID
	}

	if req.Endpoint == "" {
		req.Endpoint = "/"
	}

	// Accept either a path (starting with "/") or a full URL (starting with "http://" or "https://")
	if !strings.HasPrefix(req.Endpoint, "/") && !strings.HasPrefix(req.Endpoint, "http://") && !strings.HasPrefix(req.Endpoint, "https://") {
		v.valog(logger.Debug, "healthcheck validation: create rejected", "reason=invalid_endpoint")
		return types.ErrInvalidEndpoint
	}

	if req.Method == "" {
		req.Method = "GET"
	}

	validMethods := map[string]bool{"GET": true, "POST": true, "HEAD": true}
	if !validMethods[strings.ToUpper(req.Method)] {
		v.valog(logger.Debug, "healthcheck validation: create rejected", "reason=invalid_method")
		return types.ErrInvalidMethod
	}

	if req.TimeoutSeconds == 0 {
		req.TimeoutSeconds = 30
	}
	if req.TimeoutSeconds < 5 || req.TimeoutSeconds > 120 {
		v.valog(logger.Debug, "healthcheck validation: create rejected", "reason=invalid_timeout")
		return types.ErrInvalidTimeout
	}

	if req.IntervalSeconds == 0 {
		req.IntervalSeconds = 60
	}
	if req.IntervalSeconds < 30 || req.IntervalSeconds > 3600 {
		v.valog(logger.Debug, "healthcheck validation: create rejected", "reason=invalid_interval")
		return types.ErrInvalidInterval
	}

	if req.FailureThreshold == 0 {
		req.FailureThreshold = 3
	}
	if req.FailureThreshold < 1 || req.FailureThreshold > 10 {
		v.valog(logger.Debug, "healthcheck validation: create rejected", "reason=invalid_threshold_failure")
		return types.ErrInvalidThreshold
	}

	if req.SuccessThreshold == 0 {
		req.SuccessThreshold = 1
	}
	if req.SuccessThreshold < 1 || req.SuccessThreshold > 10 {
		v.valog(logger.Debug, "healthcheck validation: create rejected", "reason=invalid_threshold_success")
		return types.ErrInvalidThreshold
	}

	if req.RetentionDays == 0 {
		req.RetentionDays = 30
	}
	if req.RetentionDays < 1 || req.RetentionDays > 365 {
		v.valog(logger.Debug, "healthcheck validation: create rejected", "reason=invalid_retention_days")
		return types.ErrInvalidRetentionDays
	}

	if len(req.ExpectedStatus) == 0 {
		req.ExpectedStatus = []int{200}
	}

	return nil
}

func (v *Validator) validateUpdateHealthCheckRequest(req types.UpdateHealthCheckRequest) error {
	if req.ApplicationID == "" {
		v.valog(logger.Debug, "healthcheck validation: update rejected", "reason=invalid_application_id")
		return types.ErrInvalidApplicationID
	}

	if _, err := uuid.Parse(req.ApplicationID); err != nil {
		v.valog(logger.Debug, "healthcheck validation: update rejected", "reason=invalid_application_id_parse")
		return types.ErrInvalidApplicationID
	}

	// Accept either a path (starting with "/") or a full URL (starting with "http://" or "https://")
	if req.Endpoint != "" && !strings.HasPrefix(req.Endpoint, "/") && !strings.HasPrefix(req.Endpoint, "http://") && !strings.HasPrefix(req.Endpoint, "https://") {
		v.valog(logger.Debug, "healthcheck validation: update rejected", "reason=invalid_endpoint")
		return types.ErrInvalidEndpoint
	}

	if req.Method != "" {
		validMethods := map[string]bool{"GET": true, "POST": true, "HEAD": true}
		if !validMethods[strings.ToUpper(req.Method)] {
			v.valog(logger.Debug, "healthcheck validation: update rejected", "reason=invalid_method")
			return types.ErrInvalidMethod
		}
	}

	if req.TimeoutSeconds != 0 && (req.TimeoutSeconds < 5 || req.TimeoutSeconds > 120) {
		v.valog(logger.Debug, "healthcheck validation: update rejected", "reason=invalid_timeout")
		return types.ErrInvalidTimeout
	}

	if req.IntervalSeconds != 0 && (req.IntervalSeconds < 30 || req.IntervalSeconds > 3600) {
		v.valog(logger.Debug, "healthcheck validation: update rejected", "reason=invalid_interval")
		return types.ErrInvalidInterval
	}

	if req.FailureThreshold != 0 && (req.FailureThreshold < 1 || req.FailureThreshold > 10) {
		v.valog(logger.Debug, "healthcheck validation: update rejected", "reason=invalid_threshold_failure")
		return types.ErrInvalidThreshold
	}

	if req.SuccessThreshold != 0 && (req.SuccessThreshold < 1 || req.SuccessThreshold > 10) {
		v.valog(logger.Debug, "healthcheck validation: update rejected", "reason=invalid_threshold_success")
		return types.ErrInvalidThreshold
	}

	if req.RetentionDays != 0 && (req.RetentionDays < 1 || req.RetentionDays > 365) {
		v.valog(logger.Debug, "healthcheck validation: update rejected", "reason=invalid_retention_days")
		return types.ErrInvalidRetentionDays
	}

	return nil
}

func (v *Validator) validateToggleHealthCheckRequest(req types.ToggleHealthCheckRequest) error {
	if req.ApplicationID == "" {
		v.valog(logger.Debug, "healthcheck validation: toggle rejected", "reason=invalid_application_id")
		return types.ErrInvalidApplicationID
	}

	if _, err := uuid.Parse(req.ApplicationID); err != nil {
		v.valog(logger.Debug, "healthcheck validation: toggle rejected", "reason=invalid_application_id_parse")
		return types.ErrInvalidApplicationID
	}

	return nil
}

func (v *Validator) validateGetHealthCheckResultsRequest(req types.GetHealthCheckResultsRequest) error {
	if req.ApplicationID == "" {
		v.valog(logger.Debug, "healthcheck validation: results rejected", "reason=invalid_application_id")
		return types.ErrInvalidApplicationID
	}

	if _, err := uuid.Parse(req.ApplicationID); err != nil {
		v.valog(logger.Debug, "healthcheck validation: results rejected", "reason=invalid_application_id_parse")
		return types.ErrInvalidApplicationID
	}

	if req.Limit < 0 || req.Limit > 1000 {
		req.Limit = 100
	}

	return nil
}

func (v *Validator) validateGetHealthCheckStatsRequest(req types.GetHealthCheckStatsRequest) error {
	if req.ApplicationID == "" {
		v.valog(logger.Debug, "healthcheck validation: stats rejected", "reason=invalid_application_id")
		return types.ErrInvalidApplicationID
	}

	if _, err := uuid.Parse(req.ApplicationID); err != nil {
		v.valog(logger.Debug, "healthcheck validation: stats rejected", "reason=invalid_application_id_parse")
		return types.ErrInvalidApplicationID
	}

	return nil
}
