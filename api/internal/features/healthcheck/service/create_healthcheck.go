package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/healthcheck/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *HealthCheckService) CreateHealthCheck(userID uuid.UUID, organizationID uuid.UUID, req *types.CreateHealthCheckRequest) (*shared_types.HealthCheck, error) {
	s.logger.Log(logger.Info, "creating health check", "application_id: "+req.ApplicationID)

	applicationID, err := uuid.Parse(req.ApplicationID)
	if err != nil {
		return nil, types.ErrInvalidApplicationID
	}

	// Check if health check already exists
	existing, err := s.storage.GetHealthCheckByApplicationID(applicationID, organizationID)
	if err == nil && existing != nil {
		return nil, types.ErrHealthCheckAlreadyExists
	}

	endpoint := req.Endpoint
	if endpoint == "" {
		endpoint = "/"
	}

	method := req.Method
	if method == "" {
		method = "GET"
	}

	expectedStatus := req.ExpectedStatus
	if len(expectedStatus) == 0 {
		expectedStatus = []int{200}
	}

	timeoutSeconds := req.TimeoutSeconds
	if timeoutSeconds == 0 {
		timeoutSeconds = 30
	}

	intervalSeconds := req.IntervalSeconds
	if intervalSeconds == 0 {
		intervalSeconds = 60
	}

	failureThreshold := req.FailureThreshold
	if failureThreshold == 0 {
		failureThreshold = 3
	}

	successThreshold := req.SuccessThreshold
	if successThreshold == 0 {
		successThreshold = 1
	}

	retentionDays := req.RetentionDays
	if retentionDays == 0 {
		retentionDays = 30
	}

	now := time.Now()
	healthCheck := &shared_types.HealthCheck{
		ID:               uuid.New(),
		ApplicationID:    applicationID,
		OrganizationID:   organizationID,
		Enabled:          true,
		Endpoint:         endpoint,
		Method:           method,
		ExpectedStatus:   expectedStatus,
		TimeoutSeconds:   timeoutSeconds,
		IntervalSeconds:  intervalSeconds,
		FailureThreshold: failureThreshold,
		SuccessThreshold: successThreshold,
		Headers:          req.Headers,
		Body:             req.Body,
		ConsecutiveFails: 0,
		RetentionDays:    retentionDays,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.storage.CreateHealthCheck(healthCheck); err != nil {
		s.logger.Log(logger.Error, "failed to create health check", err.Error())
		return nil, err
	}

	return healthCheck, nil
}
