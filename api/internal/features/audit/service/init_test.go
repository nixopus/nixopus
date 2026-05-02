package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	_ "modernc.org/sqlite"
)

type recordingRepo struct {
	getErr    error
	createErr error
	logs      []*types.AuditLog
	total     int

	lastFilters  map[string]interface{}
	lastPage     int
	lastPageSize int
}

func (r *recordingRepo) CreateAuditLog(_ *types.AuditLog) error {
	return r.createErr
}

func (r *recordingRepo) GetAuditLogs(filters map[string]interface{}, page, pageSize int) ([]*types.AuditLog, int, error) {
	r.lastFilters = filters
	r.lastPage = page
	r.lastPageSize = pageSize
	if r.getErr != nil {
		return nil, 0, r.getErr
	}
	return r.logs, r.total, nil
}

func TestNewAuditServiceWithRepository(t *testing.T) {
	repo := &recordingRepo{}
	ctx := context.Background()
	s := NewAuditServiceWithRepository(repo, ctx, logger.NewLogger())
	require.NotNil(t, s)
}

func TestNewAuditService(t *testing.T) {
	t.Parallel()
	sqldb, err := sql.Open("sqlite", "file:newsrv?mode=memory")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	t.Cleanup(func() { _ = db.Close() })
	s := NewAuditService(db, context.Background(), logger.NewLogger())
	require.NotNil(t, s)
}

func TestLogAction(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	req := &AuditLogRequest{
		UserID:         uuid.New(),
		OrganizationID: uuid.New(),
		Action:         types.AuditActionCreate,
		ResourceType:   types.AuditResourceUser,
		ResourceID:     uuid.New(),
		RequestID:      uuid.New(),
	}

	t.Run("create_error", func(t *testing.T) {
		repo := &recordingRepo{createErr: errors.New("db down")}
		s := NewAuditServiceWithRepository(repo, ctx, logger.NewLogger())
		err := s.LogAction(req)
		require.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		repo := &recordingRepo{}
		s := NewAuditServiceWithRepository(repo, ctx, logger.NewLogger())
		err := s.LogAction(req)
		require.NoError(t, err)
	})
}

func TestGetAuditLogsDelegations(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := &recordingRepo{logs: []*types.AuditLog{{}}, total: 1}
	s := NewAuditServiceWithRepository(repo, ctx, logger.NewLogger())
	uid := uuid.New()
	rid := uuid.New()

	_, _, err := s.GetAuditLogs(map[string]interface{}{"a": 1}, 2, 20)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"a": 1}, repo.lastFilters)

	_, _, err = s.GetAuditLogsByResource(types.AuditResourceUser, rid, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, types.AuditResourceUser, repo.lastFilters["resource_type"])
	assert.Equal(t, rid, repo.lastFilters["resource_id"])

	_, _, err = s.GetAuditLogsByUser(uid, 3, 15)
	require.NoError(t, err)
	assert.Equal(t, uid, repo.lastFilters["user_id"])
	assert.Equal(t, 3, repo.lastPage)
	assert.Equal(t, 15, repo.lastPageSize)

	oid := uuid.New()
	_, _, err = s.GetAuditLogsByOrganization(oid, 1, 5)
	require.NoError(t, err)
	assert.Equal(t, oid, repo.lastFilters["organization_id"])
}

func TestGetActivities_errors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := &recordingRepo{getErr: errors.New("scan failed")}
	s := NewAuditServiceWithRepository(repo, ctx, logger.NewLogger())
	_, _, err := s.GetActivities(nil, 1, 10)
	require.Error(t, err)
}

func TestGetActivities_skipsNilLog(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := &recordingRepo{
		logs:  []*types.AuditLog{nil, sampleLog(t)},
		total: 2,
	}
	s := NewAuditServiceWithRepository(repo, ctx, logger.NewLogger())
	acts, n, err := s.GetActivities(nil, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	require.Len(t, acts, 1)
}

func TestGetActivitiesByOrganization_filters(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := &recordingRepo{logs: nil, total: 0}
	s := NewAuditServiceWithRepository(repo, ctx, logger.NewLogger())
	orgID := uuid.New()
	_, _, err := s.GetActivitiesByOrganization(orgID, 1, 10, "needle", "user")
	require.NoError(t, err)
	require.Equal(t, orgID, repo.lastFilters["organization_id"])
	require.Equal(t, "needle", repo.lastFilters["search"])
	require.Equal(t, "user", repo.lastFilters["resource_type"])
}

func sampleLog(t *testing.T) *types.AuditLog {
	t.Helper()
	return &types.AuditLog{
		ID:             uuid.New(),
		UserID:         uuid.New(),
		OrganizationID: uuid.New(),
		Action:         types.AuditActionCreate,
		ResourceType:   types.AuditResourceUser,
		ResourceID:     uuid.New(),
		CreatedAt:      time.Unix(1700000000, 0).UTC(),
	}
}

func Test_convertToActivity(t *testing.T) {
	t.Parallel()
	s := &AuditService{}

	assert.Nil(t, s.convertToActivity(nil))

	log := sampleLog(t)
	log.User = nil
	act := s.convertToActivity(log)
	require.NotNil(t, act)
	assert.Equal(t, "Unknown user", act.Actor)

	log = sampleLog(t)
	log.User = &types.User{Email: "only@mail.test"}
	act = s.convertToActivity(log)
	assert.Equal(t, "only@mail.test", act.Actor)

	log = sampleLog(t)
	log.User = &types.User{Username: "sam", Email: "e@e.com"}
	log.Metadata = map[string]any{"k": "v"}
	act = s.convertToActivity(log)
	assert.Equal(t, "sam", act.Actor)
	assert.Equal(t, log.Metadata, act.Metadata)
}

func Test_getActionColor(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "green", getActionColor(types.AuditActionCreate))
	assert.Equal(t, "blue", getActionColor(types.AuditActionUpdate))
	assert.Equal(t, "red", getActionColor(types.AuditActionDelete))
	assert.Equal(t, "gray", getActionColor(types.AuditActionAccess))
	assert.Equal(t, "gray", getActionColor(types.AuditAction("other")))
}

func Test_generateMessage_exhaustive(t *testing.T) {
	t.Parallel()
	s := &AuditService{}
	actor := "actor"
	wAct := types.AuditAction("weird")

	cases := []struct {
		res types.AuditResourceType
		act types.AuditAction
		nv  map[string]any
	}{
		{types.AuditResourceUser, types.AuditActionCreate, map[string]any{"email": "u@x"}},
		{types.AuditResourceUser, types.AuditActionCreate, map[string]any{}},
		{types.AuditResourceUser, types.AuditActionUpdate, map[string]any{"role": "admin"}},
		{types.AuditResourceUser, types.AuditActionUpdate, map[string]any{}},
		{types.AuditResourceUser, types.AuditActionDelete, nil},
		{types.AuditResourceUser, wAct, nil},
		{types.AuditResourceOrganization, types.AuditActionCreate, map[string]any{"name": "Org"}},
		{types.AuditResourceOrganization, types.AuditActionCreate, map[string]any{}},
		{types.AuditResourceOrganization, types.AuditActionUpdate, map[string]any{"name": "N2"}},
		{types.AuditResourceOrganization, types.AuditActionUpdate, map[string]any{"description": "d"}},
		{types.AuditResourceOrganization, types.AuditActionUpdate, map[string]any{}},
		{types.AuditResourceOrganization, types.AuditActionDelete, nil},
		{types.AuditResourceOrganization, wAct, nil},
		{types.AuditResourceRole, types.AuditActionCreate, map[string]any{"name": "r"}},
		{types.AuditResourceRole, types.AuditActionCreate, map[string]any{}},
		{types.AuditResourceRole, types.AuditActionUpdate, nil},
		{types.AuditResourceRole, types.AuditActionDelete, nil},
		{types.AuditResourceRole, wAct, nil},
		{types.AuditResourcePermission, types.AuditActionCreate, nil},
		{types.AuditResourcePermission, types.AuditActionUpdate, nil},
		{types.AuditResourcePermission, types.AuditActionDelete, nil},
		{types.AuditResourcePermission, wAct, nil},
		{types.AuditResourceApplication, types.AuditActionCreate, map[string]any{"name": "app"}},
		{types.AuditResourceApplication, types.AuditActionCreate, map[string]any{}},
		{types.AuditResourceApplication, types.AuditActionUpdate, map[string]any{"name": "app2"}},
		{types.AuditResourceApplication, types.AuditActionUpdate, map[string]any{}},
		{types.AuditResourceApplication, types.AuditActionDelete, nil},
		{types.AuditResourceApplication, wAct, nil},
		{types.AuditResourceDeployment, types.AuditActionCreate, nil},
		{types.AuditResourceDeployment, types.AuditActionUpdate, nil},
		{types.AuditResourceDeployment, types.AuditActionDelete, nil},
		{types.AuditResourceDeployment, wAct, nil},
		{types.AuditResourceDomain, types.AuditActionCreate, map[string]any{"domain": "x.com"}},
		{types.AuditResourceDomain, types.AuditActionCreate, map[string]any{}},
		{types.AuditResourceDomain, types.AuditActionUpdate, nil},
		{types.AuditResourceDomain, types.AuditActionDelete, nil},
		{types.AuditResourceDomain, wAct, nil},
		{types.AuditResourceGithubConnector, types.AuditActionCreate, nil},
		{types.AuditResourceGithubConnector, types.AuditActionUpdate, nil},
		{types.AuditResourceGithubConnector, types.AuditActionDelete, nil},
		{types.AuditResourceGithubConnector, wAct, nil},
		{types.AuditResourceSmtpConfig, types.AuditActionCreate, nil},
		{types.AuditResourceSmtpConfig, types.AuditActionUpdate, nil},
		{types.AuditResourceSmtpConfig, types.AuditActionDelete, nil},
		{types.AuditResourceSmtpConfig, wAct, nil},
		{types.AuditResourceNotification, types.AuditActionCreate, nil},
		{types.AuditResourceNotification, types.AuditActionUpdate, nil},
		{types.AuditResourceNotification, types.AuditActionDelete, nil},
		{types.AuditResourceNotification, wAct, nil},
		{types.AuditResourceFeatureFlag, types.AuditActionCreate, map[string]any{"name": "ff"}},
		{types.AuditResourceFeatureFlag, types.AuditActionCreate, map[string]any{}},
		{types.AuditResourceFeatureFlag, types.AuditActionUpdate, map[string]any{"enabled": true}},
		{types.AuditResourceFeatureFlag, types.AuditActionUpdate, map[string]any{"enabled": false}},
		{types.AuditResourceFeatureFlag, types.AuditActionUpdate, map[string]any{}},
		{types.AuditResourceFeatureFlag, types.AuditActionDelete, nil},
		{types.AuditResourceFeatureFlag, wAct, nil},
		{types.AuditResourceFileManager, types.AuditActionCreate, map[string]any{"name": "f"}},
		{types.AuditResourceFileManager, types.AuditActionCreate, map[string]any{"path": "/p"}},
		{types.AuditResourceFileManager, types.AuditActionCreate, map[string]any{}},
		{types.AuditResourceFileManager, types.AuditActionUpdate, map[string]any{"name": "f2"}},
		{types.AuditResourceFileManager, types.AuditActionUpdate, map[string]any{}},
		{types.AuditResourceFileManager, types.AuditActionDelete, map[string]any{"name": "d"}},
		{types.AuditResourceFileManager, types.AuditActionDelete, map[string]any{}},
		{types.AuditResourceFileManager, wAct, nil},
		{types.AuditResourceContainer, types.AuditActionCreate, nil},
		{types.AuditResourceContainer, types.AuditActionUpdate, nil},
		{types.AuditResourceContainer, types.AuditActionDelete, nil},
		{types.AuditResourceContainer, wAct, nil},
		{types.AuditResourceAudit, types.AuditActionCreate, nil},
		{types.AuditResourceAudit, types.AuditActionUpdate, nil},
		{types.AuditResourceAudit, types.AuditActionDelete, nil},
		{types.AuditResourceTerminal, types.AuditActionCreate, nil},
		{types.AuditResourceTerminal, types.AuditActionUpdate, nil},
		{types.AuditResourceTerminal, types.AuditActionDelete, nil},
		{types.AuditResourceTerminal, wAct, nil},
		{types.AuditResourceIntegration, types.AuditActionCreate, nil},
		{types.AuditResourceIntegration, types.AuditActionUpdate, nil},
		{types.AuditResourceIntegration, types.AuditActionDelete, nil},
		{types.AuditResourceIntegration, wAct, nil},
		{types.AuditResourceType("other_kind"), types.AuditActionCreate, nil},
	}
	for _, c := range cases {
		out := s.generateMessage(actor, c.act, c.res, c.nv)
		require.NotEmpty(t, out, "empty for %#v", c)
	}
}
