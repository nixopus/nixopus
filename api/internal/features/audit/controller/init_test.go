package controller

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/audit/service"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	_ "modernc.org/sqlite"
)

func TestNewAuditControllerWithService(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := service.NewAuditServiceWithRepository(&noopRepo{}, ctx, logger.NewLogger())
	c := NewAuditControllerWithService(svc, ctx, logger.NewLogger())
	require.NotNil(t, c)
}

func TestNewAuditController(t *testing.T) {
	t.Parallel()
	sqldb, err := sql.Open("sqlite", "file:newctl?mode=memory")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()
	c := NewAuditController(db, ctx, logger.NewLogger())
	require.NotNil(t, c)
}

type noopRepo struct{}

func (noopRepo) CreateAuditLog(_ *types.AuditLog) error { return nil }

func (noopRepo) GetAuditLogs(_ map[string]interface{}, _, _ int) ([]*types.AuditLog, int, error) {
	return nil, 0, nil
}

func TestGetRecentAuditLogs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	orgID := uuid.New()

	t.Run("unauthorized_no_user", func(t *testing.T) {
		svc := service.NewAuditServiceWithRepository(&noopRepo{}, ctx, logger.NewLogger())
		c := NewAuditControllerWithService(svc, ctx, logger.NewLogger())
		req := httptest.NewRequest(http.MethodGet, "/?page=1&page_size=10", nil)
		fctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
		_, err := c.GetRecentAuditLogs(fctx)
		var u fuego.UnauthorizedError
		require.ErrorAs(t, err, &u)
	})

	t.Run("missing_org_header", func(t *testing.T) {
		svc := service.NewAuditServiceWithRepository(&noopRepo{}, ctx, logger.NewLogger())
		c := NewAuditControllerWithService(svc, ctx, logger.NewLogger())
		user := &types.User{ID: uuid.New(), Email: "e@e", Name: "n"}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(context.WithValue(req.Context(), types.UserContextKey, user))
		fctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
		_, err := c.GetRecentAuditLogs(fctx)
		var b fuego.BadRequestError
		require.ErrorAs(t, err, &b)
		assert.Contains(t, b.Detail, "organization")
	})

	t.Run("invalid_org", func(t *testing.T) {
		svc := service.NewAuditServiceWithRepository(&noopRepo{}, ctx, logger.NewLogger())
		c := NewAuditControllerWithService(svc, ctx, logger.NewLogger())
		user := &types.User{ID: uuid.New(), Email: "e@e", Name: "n"}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-ORGANIZATION-ID", "not-a-uuid")
		req = req.WithContext(context.WithValue(req.Context(), types.UserContextKey, user))
		fctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
		_, err := c.GetRecentAuditLogs(fctx)
		var b fuego.BadRequestError
		require.ErrorAs(t, err, &b)
	})

	t.Run("service_error", func(t *testing.T) {
		repo := &errRepo{err: errors.New("boom")}
		svc := service.NewAuditServiceWithRepository(repo, ctx, logger.NewLogger())
		c := NewAuditControllerWithService(svc, ctx, logger.NewLogger())
		user := &types.User{ID: uuid.New(), Email: "e@e", Name: "n"}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-ORGANIZATION-ID", orgID.String())
		req = req.WithContext(context.WithValue(req.Context(), types.UserContextKey, user))
		fctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
		_, err := c.GetRecentAuditLogs(fctx)
		var h fuego.HTTPError
		require.ErrorAs(t, err, &h)
		assert.Equal(t, http.StatusInternalServerError, h.Status)
	})

	t.Run("success_pagination_and_pageSize_alias", func(t *testing.T) {
		lid := uuid.New()
		ts := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		repo := &fixedActivityRepo{id: lid, createdAt: ts}
		svc := service.NewAuditServiceWithRepository(repo, ctx, logger.NewLogger())
		c := NewAuditControllerWithService(svc, ctx, logger.NewLogger())
		user := &types.User{ID: uuid.New(), Email: "e@e", Name: "n"}
		rawURL := "/?page=bad&pageSize=200&search=x&resource_type=user"
		req := httptest.NewRequest(http.MethodGet, rawURL, nil)
		req.Header.Set("X-ORGANIZATION-ID", orgID.String())
		req = req.WithContext(context.WithValue(req.Context(), types.UserContextKey, user))
		fctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
		resp, err := c.GetRecentAuditLogs(fctx)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 1, resp.Data.Pagination.CurrentPage)
		assert.Equal(t, 100, resp.Data.Pagination.PageSize)
		assert.Equal(t, 25, resp.Data.Pagination.TotalCount)
		assert.Equal(t, 1, resp.Data.Pagination.TotalPages)
		assert.False(t, resp.Data.Pagination.HasNext)
		assert.False(t, resp.Data.Pagination.HasPrev)
		require.Len(t, repo.gotFilters, 3)
		assert.Equal(t, orgID, repo.gotFilters["organization_id"])
		assert.Equal(t, "x", repo.gotFilters["search"])
		assert.Equal(t, "user", repo.gotFilters["resource_type"])
		require.Len(t, resp.Data.Activities, 1)
		assert.Equal(t, lid.String(), resp.Data.Activities[0].ID)
	})

	t.Run("page_size_fallback_query_param", func(t *testing.T) {
		lid := uuid.New()
		repo := &fixedActivityRepo{id: lid, createdAt: time.Now()}
		svc := service.NewAuditServiceWithRepository(repo, ctx, logger.NewLogger())
		c := NewAuditControllerWithService(svc, ctx, logger.NewLogger())
		user := &types.User{ID: uuid.New(), Email: "e@e", Name: "n"}
		req := httptest.NewRequest(http.MethodGet, "/?pageSize=7", nil)
		req.Header.Set("X-ORGANIZATION-ID", orgID.String())
		req = req.WithContext(context.WithValue(req.Context(), types.UserContextKey, user))
		fctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
		_, err := c.GetRecentAuditLogs(fctx)
		require.NoError(t, err)
		assert.Equal(t, 7, repo.lastPageSize)
	})

	t.Run("pagination_has_next_prev", func(t *testing.T) {
		lid := uuid.New()
		repo := &fixedActivityRepo{id: lid, createdAt: time.Now(), totalOverride: 30}
		svc := service.NewAuditServiceWithRepository(repo, ctx, logger.NewLogger())
		c := NewAuditControllerWithService(svc, ctx, logger.NewLogger())
		user := &types.User{ID: uuid.New(), Email: "e@e", Name: "n"}
		req := httptest.NewRequest(http.MethodGet, "/?page=2&page_size=10", nil)
		req.Header.Set("X-ORGANIZATION-ID", orgID.String())
		req = req.WithContext(context.WithValue(req.Context(), types.UserContextKey, user))
		fctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
		resp, err := c.GetRecentAuditLogs(fctx)
		require.NoError(t, err)
		assert.True(t, resp.Data.Pagination.HasNext)
		assert.True(t, resp.Data.Pagination.HasPrev)
		assert.Equal(t, 3, resp.Data.Pagination.TotalPages)
	})
}

type errRepo struct{ err error }

func (e errRepo) CreateAuditLog(_ *types.AuditLog) error { return nil }

func (e errRepo) GetAuditLogs(_ map[string]interface{}, _, _ int) ([]*types.AuditLog, int, error) {
	return nil, 0, e.err
}

type fixedActivityRepo struct {
	id            uuid.UUID
	createdAt     time.Time
	gotFilters    map[string]interface{}
	lastPageSize  int
	totalOverride int
}

func (f *fixedActivityRepo) CreateAuditLog(_ *types.AuditLog) error { return nil }

func (f *fixedActivityRepo) GetAuditLogs(filters map[string]interface{}, page, pageSize int) ([]*types.AuditLog, int, error) {
	f.gotFilters = filters
	f.lastPageSize = pageSize
	total := 25
	if f.totalOverride > 0 {
		total = f.totalOverride
	}
	log := &types.AuditLog{
		ID:             f.id,
		UserID:         uuid.New(),
		OrganizationID: uuid.New(),
		Action:         types.AuditActionCreate,
		ResourceType:   types.AuditResourceUser,
		ResourceID:     uuid.New(),
		CreatedAt:      f.createdAt,
	}
	return []*types.AuditLog{log}, total, nil
}
