package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// ExecuteHealthCheck performs an HTTP health check for a given health check configuration
func (s *HealthCheckService) ExecuteHealthCheck(healthCheck *shared_types.HealthCheck) (*shared_types.HealthCheckResult, error) {
	startTime := time.Now()

	// Get application to construct URL
	deployStorage := &storage.DeployStorage{DB: s.store.DB, Ctx: s.ctx}
	application, err := deployStorage.GetApplicationById(healthCheck.ApplicationID.String(), healthCheck.OrganizationID)
	if err != nil {
		s.logger.Log(logger.Error, "failed to get application for health check", err.Error())
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	var url string
	// If endpoint is a full URL, use it directly; otherwise construct from application domain
	if strings.HasPrefix(healthCheck.Endpoint, "http://") || strings.HasPrefix(healthCheck.Endpoint, "https://") {
		url = healthCheck.Endpoint
	} else {
		protocol := "http"
		if strings.Contains(application.Domain, "localhost") || strings.Contains(application.Domain, "127.0.0.1") {
			protocol = "http"
		} else {
			protocol = "https"
		}
		url = fmt.Sprintf("%s://%s%s", protocol, application.Domain, healthCheck.Endpoint)
	}

	var req *http.Request
	var reqErr error

	if healthCheck.Method == "POST" && healthCheck.Body != "" {
		req, reqErr = http.NewRequest(healthCheck.Method, url, bytes.NewBufferString(healthCheck.Body))
	} else {
		req, reqErr = http.NewRequest(healthCheck.Method, url, nil)
	}

	if reqErr != nil {
		s.logger.Log(logger.Error, "failed to create HTTP request", reqErr.Error())
		result := &shared_types.HealthCheckResult{
			ID:            uuid.New(),
			HealthCheckID: healthCheck.ID,
			Status:        string(shared_types.HealthCheckStatusError),
			ErrorMessage:  reqErr.Error(),
			CheckedAt:     time.Now(),
		}
		return result, nil
	}

	if healthCheck.Headers != nil {
		for key, value := range healthCheck.Headers {
			req.Header.Set(key, value)
		}
	}

	// Set default headers if not provided
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Nixopus-HealthCheck/1.0")
	}

	client := &http.Client{
		Timeout: time.Duration(healthCheck.TimeoutSeconds) * time.Second,
	}

	resp, err := client.Do(req)
	responseTime := time.Since(startTime)
	responseTimeMs := int(responseTime.Milliseconds())

	result := &shared_types.HealthCheckResult{
		ID:             uuid.New(),
		HealthCheckID:  healthCheck.ID,
		ResponseTimeMs: responseTimeMs,
		CheckedAt:      time.Now(),
	}

	if err != nil {
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded") {
			result.Status = string(shared_types.HealthCheckStatusTimeout)
			result.ErrorMessage = fmt.Sprintf("Request timeout after %d seconds", healthCheck.TimeoutSeconds)
		} else {
			result.Status = string(shared_types.HealthCheckStatusError)
			result.ErrorMessage = err.Error()
		}
		return result, nil
	}
	defer resp.Body.Close()

	// Read response body (limit to 1KB to avoid memory issues)
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))

	result.StatusCode = resp.StatusCode

	isExpected := false
	for _, expectedCode := range healthCheck.ExpectedStatus {
		if resp.StatusCode == expectedCode {
			isExpected = true
			break
		}
	}

	if isExpected {
		result.Status = string(shared_types.HealthCheckStatusHealthy)
	} else {
		result.Status = string(shared_types.HealthCheckStatusUnhealthy)
		result.ErrorMessage = fmt.Sprintf("Expected status codes %v, got %d. Response: %s", healthCheck.ExpectedStatus, resp.StatusCode, string(bodyBytes))
	}

	return result, nil
}

// ProcessHealthCheckResult processes a health check result and updates the health check status
func (s *HealthCheckService) ProcessHealthCheckResult(healthCheck *shared_types.HealthCheck, result *shared_types.HealthCheckResult) error {
	if err := s.storage.AddHealthCheckResult(result); err != nil {
		s.logger.Log(logger.Error, "failed to save health check result", err.Error())
		return err
	}

	// Update consecutive fails counter
	consecutiveFails := healthCheck.ConsecutiveFails
	if result.Status == string(shared_types.HealthCheckStatusHealthy) {
		// Reset counter on success if we've reached success threshold
		if consecutiveFails >= healthCheck.SuccessThreshold {
			consecutiveFails = 0
		} else {
			consecutiveFails = 0
		}
	} else {
		consecutiveFails++
	}

	if err := s.storage.UpdateHealthCheckStatus(healthCheck.ID, consecutiveFails, result.CheckedAt); err != nil {
		s.logger.Log(logger.Error, "failed to update health check status", err.Error())
		return err
	}

	return nil
}
