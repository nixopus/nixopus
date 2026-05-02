package service

// White-box tests for TrailService — covers pure functions and mock-injectable methods.

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	machine_storage "github.com/nixopus/nixopus/api/internal/features/machine/storage"
	machine_types "github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/queue"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type errReader struct{}

func (errReader) Read(_ []byte) (int, error) {
	return 0, fmt.Errorf("entropy unavailable")
}

// mockTrailRepo implements machine_storage.TrailRepository for unit tests.
type mockTrailRepo struct {
	getActiveProvisionFn             func(userID, orgID string) (*shared_types.UserProvisionDetails, error)
	countActiveProvisionsFn          func() (int, error)
	createActiveUserProvisionFn      func(details *shared_types.UserProvisionDetails) error
	getUserProvisionDetailsByIDFn    func(sessionID string) (*shared_types.UserProvisionDetails, error)
	getUserProvisionStatusFn         func(userID string) (machine_types.UserProvisionStatus, error)
	updateUserProvisionStatusFn      func(userID string, status machine_types.UserProvisionStatus) error
	updateUserProvisionWithErrorFn   func(sessionID string, errorMsg string) error
	updateUserProvisionDetailsStepFn func(sessionID string, step shared_types.ProvisionStep) error
	getUserByIDFn                    func(userID string) (*shared_types.User, error)
	isSubdomainTakenFn               func(subdomain string) (bool, error)
	getCompletedProvisionByUserIDFn  func(userID string) (*shared_types.UserProvisionDetails, error)
	selectBestServerFn               func(vcpus, memMB, diskGB int) (string, error)
}

func (m *mockTrailRepo) GetActiveProvisionByUserAndOrg(userID, orgID string) (*shared_types.UserProvisionDetails, error) {
	if m.getActiveProvisionFn != nil {
		return m.getActiveProvisionFn(userID, orgID)
	}
	return nil, nil
}
func (m *mockTrailRepo) CountActiveProvisions() (int, error) {
	if m.countActiveProvisionsFn != nil {
		return m.countActiveProvisionsFn()
	}
	return 0, nil
}
func (m *mockTrailRepo) CreateActiveUserProvision(details *shared_types.UserProvisionDetails) error {
	if m.createActiveUserProvisionFn != nil {
		return m.createActiveUserProvisionFn(details)
	}
	return nil
}
func (m *mockTrailRepo) GetUserProvisionDetailsByID(sessionID string) (*shared_types.UserProvisionDetails, error) {
	if m.getUserProvisionDetailsByIDFn != nil {
		return m.getUserProvisionDetailsByIDFn(sessionID)
	}
	return nil, nil
}
func (m *mockTrailRepo) GetUserProvisionStatus(userID string) (machine_types.UserProvisionStatus, error) {
	if m.getUserProvisionStatusFn != nil {
		return m.getUserProvisionStatusFn(userID)
	}
	return machine_types.UserProvisionStatusPending, nil
}
func (m *mockTrailRepo) UpdateUserProvisionStatus(userID string, status machine_types.UserProvisionStatus) error {
	if m.updateUserProvisionStatusFn != nil {
		return m.updateUserProvisionStatusFn(userID, status)
	}
	return nil
}
func (m *mockTrailRepo) UpdateUserProvisionDetailsWithError(sessionID string, errorMsg string) error {
	if m.updateUserProvisionWithErrorFn != nil {
		return m.updateUserProvisionWithErrorFn(sessionID, errorMsg)
	}
	return nil
}
func (m *mockTrailRepo) UpdateUserProvisionDetailsStep(sessionID string, step shared_types.ProvisionStep) error {
	if m.updateUserProvisionDetailsStepFn != nil {
		return m.updateUserProvisionDetailsStepFn(sessionID, step)
	}
	return nil
}
func (m *mockTrailRepo) GetUserByID(userID string) (*shared_types.User, error) {
	if m.getUserByIDFn != nil {
		return m.getUserByIDFn(userID)
	}
	return nil, nil
}
func (m *mockTrailRepo) IsSubdomainTaken(subdomain string) (bool, error) {
	if m.isSubdomainTakenFn != nil {
		return m.isSubdomainTakenFn(subdomain)
	}
	return false, nil
}
func (m *mockTrailRepo) GetCompletedProvisionByUserID(userID string) (*shared_types.UserProvisionDetails, error) {
	if m.getCompletedProvisionByUserIDFn != nil {
		return m.getCompletedProvisionByUserIDFn(userID)
	}
	return nil, nil
}
func (m *mockTrailRepo) SelectBestServer(vcpus, memMB, diskGB int) (string, error) {
	if m.selectBestServerFn != nil {
		return m.selectBestServerFn(vcpus, memMB, diskGB)
	}
	return "", nil
}

// Verify mockTrailRepo satisfies the interface at compile time.
var _ machine_storage.TrailRepository = (*mockTrailRepo)(nil)

func newTestTrailService(repo machine_storage.TrailRepository) *TrailService {
	return NewTrailService(&shared_storage.Store{}, context.Background(), logger.NewLogger(), repo)
}

// ---------- getStringValue ----------

func TestGetStringValue_Nil(t *testing.T) {
	assert.Equal(t, "", getStringValue(nil))
}

func TestGetStringValue_NonNil(t *testing.T) {
	s := "hello"
	assert.Equal(t, "hello", getStringValue(&s))
}

// ---------- generateRandomString ----------

func TestGenerateRandomString_Length(t *testing.T) {
	s, err := generateRandomString(6)
	require.NoError(t, err)
	assert.Len(t, s, 6)
}

func TestGenerateRandomString_Uniqueness(t *testing.T) {
	a, _ := generateRandomString(8)
	b, _ := generateRandomString(8)
	assert.NotEqual(t, a, b)
}

func TestGenerateRandomString_ReadError(t *testing.T) {
	prev := randomReader
	randomReader = errReader{}
	t.Cleanup(func() { randomReader = prev })

	_, err := generateRandomString(8)
	require.Error(t, err)
}

// ---------- generateRandomSubdomain ----------

func TestGenerateRandomSubdomain(t *testing.T) {
	s, err := generateRandomSubdomain()
	require.NoError(t, err)
	assert.Len(t, s, 8) // 4 bytes → 8 hex chars
}

func TestGenerateRandomSubdomain_ReadError(t *testing.T) {
	prev := randomReader
	randomReader = errReader{}
	t.Cleanup(func() { randomReader = prev })

	_, err := generateRandomSubdomain()
	require.Error(t, err)
}

func TestTrailService_GenerateSubdomain_GeneratorError(t *testing.T) {
	prev := randomReader
	randomReader = errReader{}
	t.Cleanup(func() { randomReader = prev })

	svc := newTestTrailService(&mockTrailRepo{})
	_, err := svc.GenerateSubdomain()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate subdomain")
}

// ---------- IsImageAllowed ----------

func TestTrailService_IsImageAllowed_EmptyList(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	// config.AppConfig.Trail.AllowedImages is typically empty in test env
	// When list is empty, all images are allowed
	assert.True(t, svc.IsImageAllowed("ubuntu:22.04"))
}

func TestTrailService_IsImageAllowed_WithList_Allowed(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	svc.config = &shared_types.TrailConfig{
		AllowedImages: []string{"ubuntu:22.04", "debian:11"},
	}
	assert.True(t, svc.IsImageAllowed("ubuntu:22.04"))
	assert.True(t, svc.IsImageAllowed("debian:11"))
}

func TestTrailService_IsImageAllowed_WithList_Denied(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	svc.config = &shared_types.TrailConfig{
		AllowedImages: []string{"ubuntu:22.04"},
	}
	assert.False(t, svc.IsImageAllowed("alpine:latest"))
}

// ---------- GenerateContainerName ----------

func TestGenerateContainerName_Basic(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	name := svc.GenerateContainerName("Alice Smith")
	assert.Contains(t, name, "alice-smith")
}

func TestGenerateContainerName_SpecialChars(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	name := svc.GenerateContainerName("user@example.com")
	assert.NotContains(t, name, "@")
	assert.NotContains(t, name, ".")
}

func TestGenerateContainerName_LongName(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	long := "averylongnamethatexceedstwentycharacters"
	name := svc.GenerateContainerName(long)
	parts := len(name) // should be ≤ 20 + 1 + 6 (prefix + dash + suffix)
	assert.LessOrEqual(t, parts, 27)
}

func TestGenerateContainerName_EmptyName(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	name := svc.GenerateContainerName("")
	assert.True(t, len(name) > 0)
	assert.Contains(t, name, "trail-") // falls back to "trail" prefix
}

func TestGenerateContainerName_OnlySpecialChars(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	name := svc.GenerateContainerName("!@#$%")
	assert.NotEmpty(t, name)
}

// ---------- calculateProgress ----------

func TestCalculateProgress_Completed(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	progress, msg := svc.calculateProgress(nil, machine_types.UserProvisionStatusCompleted, nil)
	assert.Equal(t, 100, progress)
	assert.Contains(t, msg, "completed")
}

func TestCalculateProgress_Failed_NoError(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	progress, msg := svc.calculateProgress(nil, machine_types.UserProvisionStatusFailed, nil)
	assert.Equal(t, 0, progress)
	assert.Contains(t, msg, "failed")
}

func TestCalculateProgress_Failed_WithError(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	errMsg := "out of memory"
	progress, msg := svc.calculateProgress(nil, machine_types.UserProvisionStatusFailed, &errMsg)
	assert.Equal(t, 0, progress)
	assert.Contains(t, msg, "out of memory")
}

func TestCalculateProgress_Pending(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	progress, msg := svc.calculateProgress(nil, machine_types.UserProvisionStatusPending, nil)
	assert.Equal(t, 0, progress)
	assert.Contains(t, msg, "Waiting")
}

func TestCalculateProgress_NilStep(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	progress, _ := svc.calculateProgress(nil, machine_types.UserProvisionStatusProvisioning, nil)
	assert.Equal(t, 5, progress)
}

func TestCalculateProgress_StepInitializing(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	step := shared_types.ProvisionStepInitializing
	progress, _ := svc.calculateProgress(&step, machine_types.UserProvisionStatusProvisioning, nil)
	assert.Equal(t, 5, progress)
}

func TestCalculateProgress_StepCreatingContainer(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	step := shared_types.ProvisionStepCreatingContainer
	progress, _ := svc.calculateProgress(&step, machine_types.UserProvisionStatusProvisioning, nil)
	assert.Equal(t, 15, progress)
}

func TestCalculateProgress_StepSetupNetworking(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	step := shared_types.ProvisionStepSetupNetworking
	progress, _ := svc.calculateProgress(&step, machine_types.UserProvisionStatusProvisioning, nil)
	assert.Equal(t, 25, progress)
}

func TestCalculateProgress_StepInstallingDeps(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	step := shared_types.ProvisionStepInstallingDeps
	progress, _ := svc.calculateProgress(&step, machine_types.UserProvisionStatusProvisioning, nil)
	assert.Equal(t, 45, progress)
}

func TestCalculateProgress_StepConfiguringSSH(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	step := shared_types.ProvisionStepConfiguringSSH
	progress, _ := svc.calculateProgress(&step, machine_types.UserProvisionStatusProvisioning, nil)
	assert.Equal(t, 65, progress)
}

func TestCalculateProgress_StepSetupSSHForwarding(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	step := shared_types.ProvisionStepSetupSSHForwarding
	progress, _ := svc.calculateProgress(&step, machine_types.UserProvisionStatusProvisioning, nil)
	assert.Equal(t, 75, progress)
}

func TestCalculateProgress_StepVerifyingSSH(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	step := shared_types.ProvisionStepVerifyingSSH
	progress, _ := svc.calculateProgress(&step, machine_types.UserProvisionStatusProvisioning, nil)
	assert.Equal(t, 85, progress)
}

func TestCalculateProgress_StepCompleted(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	step := shared_types.ProvisionStepCompleted
	progress, _ := svc.calculateProgress(&step, machine_types.UserProvisionStatusProvisioning, nil)
	assert.Equal(t, 100, progress)
}

func TestCalculateProgress_StepDefault(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	step := shared_types.ProvisionStep("unknown_step")
	progress, _ := svc.calculateProgress(&step, machine_types.UserProvisionStatusProvisioning, nil)
	assert.Equal(t, 50, progress)
}

// ---------- GenerateSubdomain ----------

func TestTrailService_GenerateSubdomain_Success(t *testing.T) {
	repo := &mockTrailRepo{
		isSubdomainTakenFn: func(_ string) (bool, error) { return false, nil },
	}
	svc := newTestTrailService(repo)
	sub, err := svc.GenerateSubdomain()
	require.NoError(t, err)
	assert.NotEmpty(t, sub)
}

func TestTrailService_GenerateSubdomain_SubdomainTakenOnce(t *testing.T) {
	calls := 0
	repo := &mockTrailRepo{
		isSubdomainTakenFn: func(_ string) (bool, error) {
			calls++
			return calls < 3, nil // first 2 attempts are taken
		},
	}
	svc := newTestTrailService(repo)
	sub, err := svc.GenerateSubdomain()
	require.NoError(t, err)
	assert.NotEmpty(t, sub)
	assert.Equal(t, 3, calls)
}

func TestTrailService_GenerateSubdomain_AlwaysTaken(t *testing.T) {
	repo := &mockTrailRepo{
		isSubdomainTakenFn: func(_ string) (bool, error) { return true, nil },
	}
	svc := newTestTrailService(repo)
	_, err := svc.GenerateSubdomain()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate unique subdomain")
}

func TestTrailService_GenerateSubdomain_StorageError(t *testing.T) {
	repo := &mockTrailRepo{
		isSubdomainTakenFn: func(_ string) (bool, error) { return false, fmt.Errorf("db error") },
	}
	svc := newTestTrailService(repo)
	_, err := svc.GenerateSubdomain()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check subdomain availability")
}

// ---------- GetStatus ----------

func TestTrailService_GetStatus_StorageError(t *testing.T) {
	repo := &mockTrailRepo{
		getUserProvisionDetailsByIDFn: func(_ string) (*shared_types.UserProvisionDetails, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	svc := newTestTrailService(repo)
	_, err := svc.GetStatus(uuid.New().String(), uuid.New().String())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to retrieve status")
}

func TestTrailService_GetStatus_NotFound(t *testing.T) {
	repo := &mockTrailRepo{
		getUserProvisionDetailsByIDFn: func(_ string) (*shared_types.UserProvisionDetails, error) {
			return nil, nil
		},
	}
	svc := newTestTrailService(repo)
	_, err := svc.GetStatus(uuid.New().String(), uuid.New().String())
	require.Error(t, err)
	assert.ErrorIs(t, err, machine_types.ErrProvisionNotFound)
}

func TestTrailService_GetStatus_WrongUser(t *testing.T) {
	ownerID := uuid.New()
	requestorID := uuid.New()
	repo := &mockTrailRepo{
		getUserProvisionDetailsByIDFn: func(_ string) (*shared_types.UserProvisionDetails, error) {
			return &shared_types.UserProvisionDetails{UserID: ownerID}, nil
		},
	}
	svc := newTestTrailService(repo)
	_, err := svc.GetStatus(requestorID.String(), uuid.New().String())
	require.Error(t, err)
	assert.ErrorIs(t, err, machine_types.ErrProvisionNotFound)
}

func TestTrailService_GetStatus_Success(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	subdomain := "abc123"
	domain := "abc123.trail.example.com"
	step := shared_types.ProvisionStepCompleted

	repo := &mockTrailRepo{
		getUserProvisionDetailsByIDFn: func(_ string) (*shared_types.UserProvisionDetails, error) {
			return &shared_types.UserProvisionDetails{
				UserID:    userID,
				ID:        sessionID,
				Subdomain: &subdomain,
				Domain:    &domain,
				Step:      &step,
			}, nil
		},
		getUserProvisionStatusFn: func(_ string) (machine_types.UserProvisionStatus, error) {
			return machine_types.UserProvisionStatusCompleted, nil
		},
	}
	svc := newTestTrailService(repo)
	resp, err := svc.GetStatus(userID.String(), sessionID.String())
	require.NoError(t, err)
	assert.Equal(t, string(machine_types.UserProvisionStatusCompleted), resp.Status)
	assert.Equal(t, 100, resp.Progress)
	assert.Equal(t, subdomain, resp.Subdomain)
	assert.Contains(t, resp.TrailURL, domain)
}

func TestTrailService_GetStatus_StatusError_FallsBackToPending(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()

	repo := &mockTrailRepo{
		getUserProvisionDetailsByIDFn: func(_ string) (*shared_types.UserProvisionDetails, error) {
			return &shared_types.UserProvisionDetails{UserID: userID, ID: sessionID}, nil
		},
		getUserProvisionStatusFn: func(_ string) (machine_types.UserProvisionStatus, error) {
			return "", fmt.Errorf("db error") // triggers fallback to pending
		},
	}
	svc := newTestTrailService(repo)
	resp, err := svc.GetStatus(userID.String(), sessionID.String())
	require.NoError(t, err)
	assert.Equal(t, string(machine_types.UserProvisionStatusPending), resp.Status)
}

// ---------- UpgradeResources ----------

func TestTrailService_UpgradeResources_StorageError(t *testing.T) {
	repo := &mockTrailRepo{
		getCompletedProvisionByUserIDFn: func(_ string) (*shared_types.UserProvisionDetails, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	svc := newTestTrailService(repo)
	err := svc.UpgradeResources(uuid.New().String(), uuid.New().String(), 2, 2048)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to look up provision")
}

func TestTrailService_UpgradeResources_NotFound(t *testing.T) {
	repo := &mockTrailRepo{
		getCompletedProvisionByUserIDFn: func(_ string) (*shared_types.UserProvisionDetails, error) {
			return nil, nil
		},
	}
	svc := newTestTrailService(repo)
	err := svc.UpgradeResources(uuid.New().String(), uuid.New().String(), 2, 2048)
	require.Error(t, err)
	assert.ErrorIs(t, err, machine_types.ErrProvisionNotFound)
}

func TestTrailService_UpgradeResources_NoContainerName(t *testing.T) {
	repo := &mockTrailRepo{
		getCompletedProvisionByUserIDFn: func(_ string) (*shared_types.UserProvisionDetails, error) {
			return &shared_types.UserProvisionDetails{}, nil // nil LXDContainerName
		},
	}
	svc := newTestTrailService(repo)
	err := svc.UpgradeResources(uuid.New().String(), uuid.New().String(), 2, 2048)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing container name")
}

func TestTrailService_UpgradeResources_EmptyContainerName(t *testing.T) {
	empty := ""
	repo := &mockTrailRepo{
		getCompletedProvisionByUserIDFn: func(_ string) (*shared_types.UserProvisionDetails, error) {
			return &shared_types.UserProvisionDetails{LXDContainerName: &empty}, nil
		},
	}
	svc := newTestTrailService(repo)
	err := svc.UpgradeResources(uuid.New().String(), uuid.New().String(), 2, 2048)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing container name")
}

func TestTrailService_UpgradeResources_EnqueueError(t *testing.T) {
	containerName := "trail-abc"
	repo := &mockTrailRepo{
		getCompletedProvisionByUserIDFn: func(_ string) (*shared_types.UserProvisionDetails, error) {
			return &shared_types.UserProvisionDetails{LXDContainerName: &containerName}, nil
		},
	}
	svc := newTestTrailService(repo)
	svc.enqueueResourceFn = func(_ context.Context, _ queue.ResourceUpdatePayload) error {
		return fmt.Errorf("redis unavailable")
	}
	err := svc.UpgradeResources(uuid.New().String(), uuid.New().String(), 2, 2048)
	require.Error(t, err)
	assert.ErrorIs(t, err, machine_types.ErrFailedToEnqueueTask)
}

func TestTrailService_UpgradeResources_Success_WithServerID(t *testing.T) {
	containerName := "trail-abc"
	serverID := uuid.New()
	repo := &mockTrailRepo{
		getCompletedProvisionByUserIDFn: func(_ string) (*shared_types.UserProvisionDetails, error) {
			return &shared_types.UserProvisionDetails{
				LXDContainerName: &containerName,
				ServerID:         &serverID,
			}, nil
		},
	}
	svc := newTestTrailService(repo)
	svc.enqueueResourceFn = func(_ context.Context, p queue.ResourceUpdatePayload) error {
		assert.Equal(t, serverID.String(), p.ServerID)
		return nil
	}
	err := svc.UpgradeResources(uuid.New().String(), uuid.New().String(), 2, 2048)
	require.NoError(t, err)
}

func TestTrailService_UpgradeResources_Success_NoServerID(t *testing.T) {
	containerName := "trail-xyz"
	repo := &mockTrailRepo{
		getCompletedProvisionByUserIDFn: func(_ string) (*shared_types.UserProvisionDetails, error) {
			return &shared_types.UserProvisionDetails{LXDContainerName: &containerName}, nil
		},
	}
	svc := newTestTrailService(repo)
	svc.enqueueResourceFn = func(_ context.Context, _ queue.ResourceUpdatePayload) error { return nil }
	err := svc.UpgradeResources(uuid.New().String(), uuid.New().String(), 4, 4096)
	require.NoError(t, err)
}

// ---------- EnqueueProvisionTask ----------

func TestTrailService_EnqueueProvisionTask_Injected(t *testing.T) {
	var called bool
	svc := newTestTrailService(&mockTrailRepo{})
	svc.enqueueProvisionFn = func(_ context.Context, _ machine_types.ProvisionPayload) error {
		called = true
		return nil
	}
	err := svc.EnqueueProvisionTask(context.Background(), machine_types.ProvisionPayload{})
	require.NoError(t, err)
	assert.True(t, called)
}

func TestTrailService_EnqueueProvisionTask_DefaultQueueError(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	svc.enqueueProvisionFn = nil
	err := svc.EnqueueProvisionTask(context.Background(), machine_types.ProvisionPayload{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "provision queue not initialized")
}

// ---------- ProvisionTrail ----------

func TestTrailService_ProvisionTrail_ImageNotAllowed(t *testing.T) {
	svc := newTestTrailService(&mockTrailRepo{})
	svc.config = &shared_types.TrailConfig{
		AllowedImages: []string{"ubuntu:22.04"},
	}
	_, err := svc.ProvisionTrail(uuid.New().String(), uuid.New().String(), machine_types.ProvisionRequest{
		Image: "alpine:latest",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, machine_types.ErrImageNotAllowed)
}

func TestTrailService_ProvisionTrail_GetActiveProvisionError(t *testing.T) {
	repo := &mockTrailRepo{
		getActiveProvisionFn: func(_, _ string) (*shared_types.UserProvisionDetails, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	svc := newTestTrailService(repo)
	svc.config = &shared_types.TrailConfig{DefaultImage: "ubuntu:22.04"}
	_, err := svc.ProvisionTrail(uuid.New().String(), uuid.New().String(), machine_types.ProvisionRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check active provisions")
}

func TestTrailService_ProvisionTrail_ActiveProvisionExists(t *testing.T) {
	existing := &shared_types.UserProvisionDetails{}
	repo := &mockTrailRepo{
		getActiveProvisionFn: func(_, _ string) (*shared_types.UserProvisionDetails, error) {
			return existing, nil
		},
	}
	svc := newTestTrailService(repo)
	svc.config = &shared_types.TrailConfig{DefaultImage: "ubuntu:22.04"}
	_, err := svc.ProvisionTrail(uuid.New().String(), uuid.New().String(), machine_types.ProvisionRequest{})
	require.Error(t, err)
	assert.ErrorIs(t, err, machine_types.ErrActiveProvisionExists)
}

func TestTrailService_ProvisionTrail_CountError(t *testing.T) {
	repo := &mockTrailRepo{
		getActiveProvisionFn: func(_, _ string) (*shared_types.UserProvisionDetails, error) {
			return nil, nil
		},
		countActiveProvisionsFn: func() (int, error) {
			return 0, fmt.Errorf("db error")
		},
	}
	svc := newTestTrailService(repo)
	svc.config = &shared_types.TrailConfig{DefaultImage: "ubuntu:22.04"}
	_, err := svc.ProvisionTrail(uuid.New().String(), uuid.New().String(), machine_types.ProvisionRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check system capacity")
}

func TestTrailService_ProvisionTrail_AtCapacity(t *testing.T) {
	repo := &mockTrailRepo{
		getActiveProvisionFn: func(_, _ string) (*shared_types.UserProvisionDetails, error) {
			return nil, nil
		},
		countActiveProvisionsFn: func() (int, error) {
			return 100, nil
		},
	}
	svc := newTestTrailService(repo)
	svc.config = &shared_types.TrailConfig{
		DefaultImage:        "ubuntu:22.04",
		MaxConcurrentTrails: 10,
	}
	_, err := svc.ProvisionTrail(uuid.New().String(), uuid.New().String(), machine_types.ProvisionRequest{})
	require.Error(t, err)
	assert.ErrorIs(t, err, machine_types.ErrSystemAtCapacity)
}

func TestTrailService_ProvisionTrail_SubdomainError(t *testing.T) {
	repo := &mockTrailRepo{
		getActiveProvisionFn:    func(_, _ string) (*shared_types.UserProvisionDetails, error) { return nil, nil },
		countActiveProvisionsFn: func() (int, error) { return 0, nil },
		isSubdomainTakenFn: func(_ string) (bool, error) {
			return false, fmt.Errorf("db error")
		},
	}
	svc := newTestTrailService(repo)
	svc.config = &shared_types.TrailConfig{DefaultImage: "ubuntu:22.04", MaxConcurrentTrails: 100}
	_, err := svc.ProvisionTrail(uuid.New().String(), uuid.New().String(), machine_types.ProvisionRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate subdomain")
}

func TestTrailService_ProvisionTrail_GetUserError(t *testing.T) {
	repo := &mockTrailRepo{
		getActiveProvisionFn:    func(_, _ string) (*shared_types.UserProvisionDetails, error) { return nil, nil },
		countActiveProvisionsFn: func() (int, error) { return 0, nil },
		isSubdomainTakenFn:      func(_ string) (bool, error) { return false, nil },
		getUserByIDFn: func(_ string) (*shared_types.User, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	svc := newTestTrailService(repo)
	svc.config = &shared_types.TrailConfig{DefaultImage: "ubuntu:22.04", MaxConcurrentTrails: 100}
	_, err := svc.ProvisionTrail(uuid.New().String(), uuid.New().String(), machine_types.ProvisionRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user")
}

func TestTrailService_ProvisionTrail_InvalidUserID(t *testing.T) {
	repo := &mockTrailRepo{
		getActiveProvisionFn:    func(_, _ string) (*shared_types.UserProvisionDetails, error) { return nil, nil },
		countActiveProvisionsFn: func() (int, error) { return 0, nil },
		isSubdomainTakenFn:      func(_ string) (bool, error) { return false, nil },
		getUserByIDFn:           func(_ string) (*shared_types.User, error) { return nil, nil },
	}
	svc := newTestTrailService(repo)
	svc.config = &shared_types.TrailConfig{DefaultImage: "ubuntu:22.04", MaxConcurrentTrails: 100, TrailDomain: "example.com"}
	_, err := svc.ProvisionTrail("not-a-uuid", uuid.New().String(), machine_types.ProvisionRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user ID")
}

func TestTrailService_ProvisionTrail_InvalidOrgID(t *testing.T) {
	repo := &mockTrailRepo{
		getActiveProvisionFn:    func(_, _ string) (*shared_types.UserProvisionDetails, error) { return nil, nil },
		countActiveProvisionsFn: func() (int, error) { return 0, nil },
		isSubdomainTakenFn:      func(_ string) (bool, error) { return false, nil },
		getUserByIDFn:           func(_ string) (*shared_types.User, error) { return nil, nil },
	}
	svc := newTestTrailService(repo)
	svc.config = &shared_types.TrailConfig{DefaultImage: "ubuntu:22.04", MaxConcurrentTrails: 100, TrailDomain: "example.com"}
	_, err := svc.ProvisionTrail(uuid.New().String(), "not-a-uuid", machine_types.ProvisionRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid organization ID")
}

func TestTrailService_ProvisionTrail_CreateProvisionError_Duplicate(t *testing.T) {
	repo := &mockTrailRepo{
		getActiveProvisionFn:    func(_, _ string) (*shared_types.UserProvisionDetails, error) { return nil, nil },
		countActiveProvisionsFn: func() (int, error) { return 0, nil },
		isSubdomainTakenFn:      func(_ string) (bool, error) { return false, nil },
		getUserByIDFn:           func(_ string) (*shared_types.User, error) { return nil, nil },
		createActiveUserProvisionFn: func(_ *shared_types.UserProvisionDetails) error {
			return fmt.Errorf("duplicate key: active_provision_per_user_org")
		},
	}
	svc := newTestTrailService(repo)
	svc.config = &shared_types.TrailConfig{DefaultImage: "ubuntu:22.04", MaxConcurrentTrails: 100, TrailDomain: "example.com"}
	_, err := svc.ProvisionTrail(uuid.New().String(), uuid.New().String(), machine_types.ProvisionRequest{})
	require.Error(t, err)
	assert.ErrorIs(t, err, machine_types.ErrActiveProvisionExists)
}

func TestTrailService_ProvisionTrail_CreateProvisionError_OtherError(t *testing.T) {
	repo := &mockTrailRepo{
		getActiveProvisionFn:    func(_, _ string) (*shared_types.UserProvisionDetails, error) { return nil, nil },
		countActiveProvisionsFn: func() (int, error) { return 0, nil },
		isSubdomainTakenFn:      func(_ string) (bool, error) { return false, nil },
		getUserByIDFn:           func(_ string) (*shared_types.User, error) { return nil, nil },
		createActiveUserProvisionFn: func(_ *shared_types.UserProvisionDetails) error {
			return fmt.Errorf("internal db error")
		},
	}
	svc := newTestTrailService(repo)
	svc.config = &shared_types.TrailConfig{DefaultImage: "ubuntu:22.04", MaxConcurrentTrails: 100, TrailDomain: "example.com"}
	_, err := svc.ProvisionTrail(uuid.New().String(), uuid.New().String(), machine_types.ProvisionRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create provision record")
}

func TestTrailService_ProvisionTrail_EnqueueError(t *testing.T) {
	repo := &mockTrailRepo{
		getActiveProvisionFn:           func(_, _ string) (*shared_types.UserProvisionDetails, error) { return nil, nil },
		countActiveProvisionsFn:        func() (int, error) { return 0, nil },
		isSubdomainTakenFn:             func(_ string) (bool, error) { return false, nil },
		getUserByIDFn:                  func(_ string) (*shared_types.User, error) { return nil, nil },
		createActiveUserProvisionFn:    func(_ *shared_types.UserProvisionDetails) error { return nil },
		updateUserProvisionStatusFn:    func(_ string, _ machine_types.UserProvisionStatus) error { return nil },
		selectBestServerFn:             func(_, _, _ int) (string, error) { return "srv-1", nil },
		updateUserProvisionWithErrorFn: func(_, _ string) error { return nil },
	}
	svc := newTestTrailService(repo)
	svc.config = &shared_types.TrailConfig{DefaultImage: "ubuntu:22.04", MaxConcurrentTrails: 100, TrailDomain: "example.com"}
	svc.enqueueProvisionFn = func(_ context.Context, _ machine_types.ProvisionPayload) error {
		return fmt.Errorf("redis unavailable")
	}
	_, err := svc.ProvisionTrail(uuid.New().String(), uuid.New().String(), machine_types.ProvisionRequest{})
	require.Error(t, err)
	assert.ErrorIs(t, err, machine_types.ErrFailedToEnqueueTask)
}

func TestTrailService_ProvisionTrail_EnqueueError_UpdatesFail(t *testing.T) {
	// Test the inner error paths when enqueue fails AND the subsequent updates also fail
	// (exercises the warning log branches for UpdateUserProvisionDetailsWithError and UpdateUserProvisionStatus)
	repo := &mockTrailRepo{
		getActiveProvisionFn:        func(_, _ string) (*shared_types.UserProvisionDetails, error) { return nil, nil },
		countActiveProvisionsFn:     func() (int, error) { return 0, nil },
		isSubdomainTakenFn:          func(_ string) (bool, error) { return false, nil },
		getUserByIDFn:               func(_ string) (*shared_types.User, error) { return nil, nil },
		createActiveUserProvisionFn: func(_ *shared_types.UserProvisionDetails) error { return nil },
		updateUserProvisionStatusFn: func(_ string, _ machine_types.UserProvisionStatus) error {
			return fmt.Errorf("status update failed")
		},
		selectBestServerFn: func(_, _, _ int) (string, error) { return "", nil },
		updateUserProvisionWithErrorFn: func(_, _ string) error {
			return fmt.Errorf("error update failed")
		},
	}
	svc := newTestTrailService(repo)
	svc.config = &shared_types.TrailConfig{DefaultImage: "ubuntu:22.04", MaxConcurrentTrails: 100, TrailDomain: "example.com"}
	svc.enqueueProvisionFn = func(_ context.Context, _ machine_types.ProvisionPayload) error {
		return fmt.Errorf("redis unavailable")
	}
	_, err := svc.ProvisionTrail(uuid.New().String(), uuid.New().String(), machine_types.ProvisionRequest{})
	require.Error(t, err)
	assert.ErrorIs(t, err, machine_types.ErrFailedToEnqueueTask)
}

func TestTrailService_ProvisionTrail_Success_WithUserName(t *testing.T) {
	repo := &mockTrailRepo{
		getActiveProvisionFn:        func(_, _ string) (*shared_types.UserProvisionDetails, error) { return nil, nil },
		countActiveProvisionsFn:     func() (int, error) { return 0, nil },
		isSubdomainTakenFn:          func(_ string) (bool, error) { return false, nil },
		getUserByIDFn:               func(_ string) (*shared_types.User, error) { return &shared_types.User{Name: "Alice"}, nil },
		createActiveUserProvisionFn: func(_ *shared_types.UserProvisionDetails) error { return nil },
		updateUserProvisionStatusFn: func(_ string, _ machine_types.UserProvisionStatus) error { return nil },
		selectBestServerFn:          func(_, _, _ int) (string, error) { return "", nil },
	}
	svc := newTestTrailService(repo)
	svc.config = &shared_types.TrailConfig{DefaultImage: "ubuntu:22.04", MaxConcurrentTrails: 100, TrailDomain: "example.com"}
	svc.enqueueProvisionFn = func(_ context.Context, _ machine_types.ProvisionPayload) error { return nil }

	resp, err := svc.ProvisionTrail(uuid.New().String(), uuid.New().String(), machine_types.ProvisionRequest{})
	require.NoError(t, err)
	assert.Equal(t, "provisioning", resp.Status)
}

func TestTrailService_ProvisionTrail_Success_WithEmail(t *testing.T) {
	repo := &mockTrailRepo{
		getActiveProvisionFn:        func(_, _ string) (*shared_types.UserProvisionDetails, error) { return nil, nil },
		countActiveProvisionsFn:     func() (int, error) { return 0, nil },
		isSubdomainTakenFn:          func(_ string) (bool, error) { return false, nil },
		getUserByIDFn:               func(_ string) (*shared_types.User, error) { return &shared_types.User{Email: "alice@example.com"}, nil },
		createActiveUserProvisionFn: func(_ *shared_types.UserProvisionDetails) error { return nil },
		updateUserProvisionStatusFn: func(_ string, _ machine_types.UserProvisionStatus) error { return nil },
		selectBestServerFn:          func(_, _, _ int) (string, error) { return "", nil },
	}
	svc := newTestTrailService(repo)
	svc.config = &shared_types.TrailConfig{DefaultImage: "ubuntu:22.04", MaxConcurrentTrails: 100, TrailDomain: "example.com"}
	svc.enqueueProvisionFn = func(_ context.Context, _ machine_types.ProvisionPayload) error { return nil }

	resp, err := svc.ProvisionTrail(uuid.New().String(), uuid.New().String(), machine_types.ProvisionRequest{})
	require.NoError(t, err)
	assert.Equal(t, "provisioning", resp.Status)
}
