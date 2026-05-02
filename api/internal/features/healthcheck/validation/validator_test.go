package validation

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/healthcheck/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newValidator builds a Validator with nil storage — none of the validation
// methods perform storage lookups, so nil is safe for all unit tests here.
func newValidator() *Validator {
	return NewValidator(nil)
}

// ---------------------------------------------------------------------------
// NewValidator
// ---------------------------------------------------------------------------

func TestNewValidator(t *testing.T) {
	v := NewValidator(nil)
	assert.NotNil(t, v)
}

// ---------------------------------------------------------------------------
// ValidateRequest — dispatch
// ---------------------------------------------------------------------------

func TestValidateRequest_DispatchesCreateRequest(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Endpoint:      "/health",
		Method:        "GET",
	})
	assert.NoError(t, err)
}

func TestValidateRequest_DispatchesUpdateRequest(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
	})
	assert.NoError(t, err)
}

func TestValidateRequest_DispatchesToggleRequest(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.ToggleHealthCheckRequest{
		ApplicationID: uuid.New().String(),
	})
	assert.NoError(t, err)
}

func TestValidateRequest_DispatchesGetResultsRequest(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.GetHealthCheckResultsRequest{
		ApplicationID: uuid.New().String(),
	})
	assert.NoError(t, err)
}

func TestValidateRequest_DispatchesGetStatsRequest(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.GetHealthCheckStatsRequest{
		ApplicationID: uuid.New().String(),
	})
	assert.NoError(t, err)
}

func TestValidateRequest_UnknownTypeReturnsError(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest("unexpected string type")
	assert.ErrorIs(t, err, types.ErrInvalidRequestType)
}

// ---------------------------------------------------------------------------
// validateCreateHealthCheckRequest — ApplicationID
// ---------------------------------------------------------------------------

func TestValidateCreate_MissingApplicationID(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{})
	assert.ErrorIs(t, err, types.ErrInvalidApplicationID)
}

func TestValidateCreate_InvalidApplicationIDNotUUID(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: "not-a-uuid",
	})
	assert.ErrorIs(t, err, types.ErrInvalidApplicationID)
}

// ---------------------------------------------------------------------------
// validateCreateHealthCheckRequest — Endpoint
// ---------------------------------------------------------------------------

func TestValidateCreate_EmptyEndpointDefaultsToSlash(t *testing.T) {
	v := newValidator()
	// Empty endpoint should be treated as "/" and pass validation
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Endpoint:      "",
	})
	assert.NoError(t, err)
}

func TestValidateCreate_EndpointStartingWithSlash(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Endpoint:      "/health",
	})
	assert.NoError(t, err)
}

func TestValidateCreate_EndpointFullHTTPURL(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Endpoint:      "http://external.example.com/ping",
	})
	assert.NoError(t, err)
}

func TestValidateCreate_EndpointFullHTTPSURL(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Endpoint:      "https://external.example.com/ping",
	})
	assert.NoError(t, err)
}

func TestValidateCreate_InvalidEndpoint(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Endpoint:      "health",
	})
	assert.ErrorIs(t, err, types.ErrInvalidEndpoint)
}

// ---------------------------------------------------------------------------
// validateCreateHealthCheckRequest — Method
// ---------------------------------------------------------------------------

func TestValidateCreate_EmptyMethodDefaultsToGET(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Method:        "",
	})
	assert.NoError(t, err)
}

func TestValidateCreate_MethodGET(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Method:        "GET",
	})
	assert.NoError(t, err)
}

func TestValidateCreate_MethodPOST(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Method:        "POST",
	})
	assert.NoError(t, err)
}

func TestValidateCreate_MethodHEAD(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Method:        "HEAD",
	})
	assert.NoError(t, err)
}

func TestValidateCreate_MethodCaseInsensitive(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Method:        "get",
	})
	assert.NoError(t, err)
}

func TestValidateCreate_InvalidMethod(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Method:        "DELETE",
	})
	assert.ErrorIs(t, err, types.ErrInvalidMethod)
}

// ---------------------------------------------------------------------------
// validateCreateHealthCheckRequest — TimeoutSeconds
// ---------------------------------------------------------------------------

func TestValidateCreate_ZeroTimeoutDefaultsTo30(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID:  uuid.New().String(),
		TimeoutSeconds: 0,
	})
	assert.NoError(t, err)
}

func TestValidateCreate_TimeoutTooLow(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID:  uuid.New().String(),
		TimeoutSeconds: 4,
	})
	assert.ErrorIs(t, err, types.ErrInvalidTimeout)
}

func TestValidateCreate_TimeoutTooHigh(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID:  uuid.New().String(),
		TimeoutSeconds: 121,
	})
	assert.ErrorIs(t, err, types.ErrInvalidTimeout)
}

func TestValidateCreate_TimeoutAtMinBoundary(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID:  uuid.New().String(),
		TimeoutSeconds: 5,
	})
	assert.NoError(t, err)
}

func TestValidateCreate_TimeoutAtMaxBoundary(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID:  uuid.New().String(),
		TimeoutSeconds: 120,
	})
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// validateCreateHealthCheckRequest — IntervalSeconds
// ---------------------------------------------------------------------------

func TestValidateCreate_ZeroIntervalDefaultsTo60(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
	})
	assert.NoError(t, err)
}

func TestValidateCreate_IntervalTooLow(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID:   uuid.New().String(),
		IntervalSeconds: 29,
	})
	assert.ErrorIs(t, err, types.ErrInvalidInterval)
}

func TestValidateCreate_IntervalTooHigh(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID:   uuid.New().String(),
		IntervalSeconds: 3601,
	})
	assert.ErrorIs(t, err, types.ErrInvalidInterval)
}

func TestValidateCreate_IntervalAtMinBoundary(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID:   uuid.New().String(),
		IntervalSeconds: 30,
	})
	assert.NoError(t, err)
}

func TestValidateCreate_IntervalAtMaxBoundary(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID:   uuid.New().String(),
		IntervalSeconds: 3600,
	})
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// validateCreateHealthCheckRequest — FailureThreshold
// ---------------------------------------------------------------------------

func TestValidateCreate_ZeroFailureThresholdDefaults(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
	})
	assert.NoError(t, err)
}

func TestValidateCreate_FailureThresholdTooLow(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID:    uuid.New().String(),
		FailureThreshold: -1,
	})
	assert.ErrorIs(t, err, types.ErrInvalidThreshold)
}

func TestValidateCreate_FailureThresholdTooHigh(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID:    uuid.New().String(),
		FailureThreshold: 11,
	})
	assert.ErrorIs(t, err, types.ErrInvalidThreshold)
}

func TestValidateCreate_FailureThresholdAtBoundaries(t *testing.T) {
	v := newValidator()
	for _, threshold := range []int{1, 10} {
		err := v.ValidateRequest(&types.CreateHealthCheckRequest{
			ApplicationID:    uuid.New().String(),
			FailureThreshold: threshold,
		})
		assert.NoError(t, err, "expected no error for FailureThreshold=%d", threshold)
	}
}

// ---------------------------------------------------------------------------
// validateCreateHealthCheckRequest — SuccessThreshold
// ---------------------------------------------------------------------------

func TestValidateCreate_ZeroSuccessThresholdDefaults(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
	})
	assert.NoError(t, err)
}

func TestValidateCreate_SuccessThresholdTooLow(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID:    uuid.New().String(),
		SuccessThreshold: -1,
	})
	assert.ErrorIs(t, err, types.ErrInvalidThreshold)
}

func TestValidateCreate_SuccessThresholdTooHigh(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID:    uuid.New().String(),
		SuccessThreshold: 11,
	})
	assert.ErrorIs(t, err, types.ErrInvalidThreshold)
}

func TestValidateCreate_SuccessThresholdAtBoundaries(t *testing.T) {
	v := newValidator()
	for _, threshold := range []int{1, 10} {
		err := v.ValidateRequest(&types.CreateHealthCheckRequest{
			ApplicationID:    uuid.New().String(),
			SuccessThreshold: threshold,
		})
		assert.NoError(t, err, "expected no error for SuccessThreshold=%d", threshold)
	}
}

// ---------------------------------------------------------------------------
// validateCreateHealthCheckRequest — RetentionDays
// ---------------------------------------------------------------------------

func TestValidateCreate_ZeroRetentionDefaultsTo30(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
	})
	assert.NoError(t, err)
}

func TestValidateCreate_RetentionDaysTooLow(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		RetentionDays: -1,
	})
	assert.ErrorIs(t, err, types.ErrInvalidRetentionDays)
}

func TestValidateCreate_RetentionDaysTooHigh(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		RetentionDays: 366,
	})
	assert.ErrorIs(t, err, types.ErrInvalidRetentionDays)
}

func TestValidateCreate_RetentionDaysAtBoundaries(t *testing.T) {
	v := newValidator()
	for _, days := range []int{1, 365} {
		err := v.ValidateRequest(&types.CreateHealthCheckRequest{
			ApplicationID: uuid.New().String(),
			RetentionDays: days,
		})
		assert.NoError(t, err, "expected no error for RetentionDays=%d", days)
	}
}

// ---------------------------------------------------------------------------
// validateCreateHealthCheckRequest — ExpectedStatus defaults
// ---------------------------------------------------------------------------

func TestValidateCreate_EmptyExpectedStatusDefaultsTo200(t *testing.T) {
	v := newValidator()
	// The function defaults ExpectedStatus when empty — just verify no error
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
	})
	assert.NoError(t, err)
}

func TestValidateCreate_ExplicitExpectedStatus(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID:  uuid.New().String(),
		ExpectedStatus: []int{200, 201, 204},
	})
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// validateUpdateHealthCheckRequest — ApplicationID
// ---------------------------------------------------------------------------

func TestValidateUpdate_MissingApplicationID(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{})
	assert.ErrorIs(t, err, types.ErrInvalidApplicationID)
}

func TestValidateUpdate_InvalidApplicationIDNotUUID(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID: "not-a-uuid",
	})
	assert.ErrorIs(t, err, types.ErrInvalidApplicationID)
}

// ---------------------------------------------------------------------------
// validateUpdateHealthCheckRequest — Endpoint
// ---------------------------------------------------------------------------

func TestValidateUpdate_EmptyEndpointSkipsCheck(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Endpoint:      "",
	})
	assert.NoError(t, err)
}

func TestValidateUpdate_EndpointStartingWithSlash(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Endpoint:      "/new-path",
	})
	assert.NoError(t, err)
}

func TestValidateUpdate_EndpointFullHTTPURL(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Endpoint:      "http://external.example.com/ping",
	})
	assert.NoError(t, err)
}

func TestValidateUpdate_EndpointFullHTTPSURL(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Endpoint:      "https://external.example.com/ping",
	})
	assert.NoError(t, err)
}

func TestValidateUpdate_InvalidEndpoint(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Endpoint:      "no-leading-slash",
	})
	assert.ErrorIs(t, err, types.ErrInvalidEndpoint)
}

// ---------------------------------------------------------------------------
// validateUpdateHealthCheckRequest — Method
// ---------------------------------------------------------------------------

func TestValidateUpdate_EmptyMethodSkipsCheck(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Method:        "",
	})
	assert.NoError(t, err)
}

func TestValidateUpdate_ValidMethod(t *testing.T) {
	v := newValidator()
	for _, method := range []string{"GET", "POST", "HEAD", "get", "post", "head"} {
		err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
			ApplicationID: uuid.New().String(),
			Method:        method,
		})
		assert.NoError(t, err, "expected no error for method=%s", method)
	}
}

func TestValidateUpdate_InvalidMethod(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Method:        "PATCH",
	})
	assert.ErrorIs(t, err, types.ErrInvalidMethod)
}

// ---------------------------------------------------------------------------
// validateUpdateHealthCheckRequest — TimeoutSeconds
// ---------------------------------------------------------------------------

func TestValidateUpdate_ZeroTimeoutSkipsCheck(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID:  uuid.New().String(),
		TimeoutSeconds: 0,
	})
	assert.NoError(t, err)
}

func TestValidateUpdate_TimeoutTooLow(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID:  uuid.New().String(),
		TimeoutSeconds: 4,
	})
	assert.ErrorIs(t, err, types.ErrInvalidTimeout)
}

func TestValidateUpdate_TimeoutTooHigh(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID:  uuid.New().String(),
		TimeoutSeconds: 121,
	})
	assert.ErrorIs(t, err, types.ErrInvalidTimeout)
}

func TestValidateUpdate_ValidTimeout(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID:  uuid.New().String(),
		TimeoutSeconds: 30,
	})
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// validateUpdateHealthCheckRequest — IntervalSeconds
// ---------------------------------------------------------------------------

func TestValidateUpdate_ZeroIntervalSkipsCheck(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID:   uuid.New().String(),
		IntervalSeconds: 0,
	})
	assert.NoError(t, err)
}

func TestValidateUpdate_IntervalTooLow(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID:   uuid.New().String(),
		IntervalSeconds: 10,
	})
	assert.ErrorIs(t, err, types.ErrInvalidInterval)
}

func TestValidateUpdate_IntervalTooHigh(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID:   uuid.New().String(),
		IntervalSeconds: 4000,
	})
	assert.ErrorIs(t, err, types.ErrInvalidInterval)
}

func TestValidateUpdate_ValidInterval(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID:   uuid.New().String(),
		IntervalSeconds: 60,
	})
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// validateUpdateHealthCheckRequest — FailureThreshold
// ---------------------------------------------------------------------------

func TestValidateUpdate_ZeroFailureThresholdSkipsCheck(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID:    uuid.New().String(),
		FailureThreshold: 0,
	})
	assert.NoError(t, err)
}

func TestValidateUpdate_FailureThresholdTooLow(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID:    uuid.New().String(),
		FailureThreshold: -1,
	})
	assert.ErrorIs(t, err, types.ErrInvalidThreshold)
}

func TestValidateUpdate_FailureThresholdTooHigh(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID:    uuid.New().String(),
		FailureThreshold: 11,
	})
	assert.ErrorIs(t, err, types.ErrInvalidThreshold)
}

func TestValidateUpdate_ValidFailureThreshold(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID:    uuid.New().String(),
		FailureThreshold: 5,
	})
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// validateUpdateHealthCheckRequest — SuccessThreshold
// ---------------------------------------------------------------------------

func TestValidateUpdate_ZeroSuccessThresholdSkipsCheck(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID:    uuid.New().String(),
		SuccessThreshold: 0,
	})
	assert.NoError(t, err)
}

func TestValidateUpdate_SuccessThresholdTooLow(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID:    uuid.New().String(),
		SuccessThreshold: -1,
	})
	assert.ErrorIs(t, err, types.ErrInvalidThreshold)
}

func TestValidateUpdate_SuccessThresholdTooHigh(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID:    uuid.New().String(),
		SuccessThreshold: 11,
	})
	assert.ErrorIs(t, err, types.ErrInvalidThreshold)
}

func TestValidateUpdate_ValidSuccessThreshold(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID:    uuid.New().String(),
		SuccessThreshold: 2,
	})
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// validateUpdateHealthCheckRequest — RetentionDays
// ---------------------------------------------------------------------------

func TestValidateUpdate_ZeroRetentionSkipsCheck(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		RetentionDays: 0,
	})
	assert.NoError(t, err)
}

func TestValidateUpdate_RetentionDaysTooLow(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		RetentionDays: -1,
	})
	assert.ErrorIs(t, err, types.ErrInvalidRetentionDays)
}

func TestValidateUpdate_RetentionDaysTooHigh(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		RetentionDays: 366,
	})
	assert.ErrorIs(t, err, types.ErrInvalidRetentionDays)
}

func TestValidateUpdate_ValidRetentionDays(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		RetentionDays: 90,
	})
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// validateToggleHealthCheckRequest
// ---------------------------------------------------------------------------

func TestValidateToggle_MissingApplicationID(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.ToggleHealthCheckRequest{})
	assert.ErrorIs(t, err, types.ErrInvalidApplicationID)
}

func TestValidateToggle_InvalidApplicationIDNotUUID(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.ToggleHealthCheckRequest{
		ApplicationID: "not-a-uuid",
	})
	assert.ErrorIs(t, err, types.ErrInvalidApplicationID)
}

func TestValidateToggle_ValidEnable(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.ToggleHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Enabled:       true,
	})
	assert.NoError(t, err)
}

func TestValidateToggle_ValidDisable(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.ToggleHealthCheckRequest{
		ApplicationID: uuid.New().String(),
		Enabled:       false,
	})
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// validateGetHealthCheckResultsRequest
// ---------------------------------------------------------------------------

func TestValidateGetResults_MissingApplicationID(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.GetHealthCheckResultsRequest{})
	assert.ErrorIs(t, err, types.ErrInvalidApplicationID)
}

func TestValidateGetResults_InvalidApplicationIDNotUUID(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.GetHealthCheckResultsRequest{
		ApplicationID: "not-a-uuid",
	})
	assert.ErrorIs(t, err, types.ErrInvalidApplicationID)
}

func TestValidateGetResults_OutOfRangeLimitIsReset(t *testing.T) {
	v := newValidator()
	// Negative limit — the validator resets it to 100, does NOT return an error
	err := v.ValidateRequest(&types.GetHealthCheckResultsRequest{
		ApplicationID: uuid.New().String(),
		Limit:         -5,
	})
	assert.NoError(t, err)
}

func TestValidateGetResults_LimitOverMaxIsReset(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.GetHealthCheckResultsRequest{
		ApplicationID: uuid.New().String(),
		Limit:         1001,
	})
	assert.NoError(t, err)
}

func TestValidateGetResults_ValidLimit(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.GetHealthCheckResultsRequest{
		ApplicationID: uuid.New().String(),
		Limit:         50,
	})
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// validateGetHealthCheckStatsRequest
// ---------------------------------------------------------------------------

func TestValidateGetStats_MissingApplicationID(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.GetHealthCheckStatsRequest{})
	assert.ErrorIs(t, err, types.ErrInvalidApplicationID)
}

func TestValidateGetStats_InvalidApplicationIDNotUUID(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.GetHealthCheckStatsRequest{
		ApplicationID: "not-a-uuid",
	})
	assert.ErrorIs(t, err, types.ErrInvalidApplicationID)
}

func TestValidateGetStats_Valid(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.GetHealthCheckStatsRequest{
		ApplicationID: uuid.New().String(),
		Period:        "24h",
	})
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Full happy path — all fields explicitly set
// ---------------------------------------------------------------------------

func TestValidateCreate_FullValidRequest(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.CreateHealthCheckRequest{
		ApplicationID:    uuid.New().String(),
		Endpoint:         "/health",
		Method:           "POST",
		ExpectedStatus:   []int{200, 201},
		TimeoutSeconds:   10,
		IntervalSeconds:  120,
		FailureThreshold: 5,
		SuccessThreshold: 2,
		RetentionDays:    14,
	})
	require.NoError(t, err)
}

func TestValidateUpdate_FullValidRequest(t *testing.T) {
	v := newValidator()
	err := v.ValidateRequest(&types.UpdateHealthCheckRequest{
		ApplicationID:    uuid.New().String(),
		Endpoint:         "https://newhost.example.com/health",
		Method:           "HEAD",
		ExpectedStatus:   []int{200},
		TimeoutSeconds:   60,
		IntervalSeconds:  300,
		FailureThreshold: 3,
		SuccessThreshold: 1,
		RetentionDays:    30,
	})
	require.NoError(t, err)
}
