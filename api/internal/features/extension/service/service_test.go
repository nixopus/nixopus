package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockExtensionRepo struct {
	extByID       *types.Extension
	extByEID      *types.Extension
	list          *types.ExtensionListResponse
	categories    []types.ExtensionCategory
	errGet        error
	errGetEID     error
	errList       error
	errCategories error
}

func (m *mockExtensionRepo) CreateExtension(*types.Extension) error { return nil }
func (m *mockExtensionRepo) CreateExtensionVariables([]types.ExtensionVariable) error {
	return nil
}
func (m *mockExtensionRepo) UpdateExtension(*types.Extension) error { return nil }
func (m *mockExtensionRepo) DeleteExtension(string) error           { return nil }

func (m *mockExtensionRepo) GetExtension(string) (*types.Extension, error) {
	return m.extByID, m.errGet
}

func (m *mockExtensionRepo) GetExtensionByID(string) (*types.Extension, error) {
	return m.extByEID, m.errGetEID
}

func (m *mockExtensionRepo) ListExtensions(types.ExtensionListParams) (*types.ExtensionListResponse, error) {
	return m.list, m.errList
}

func (m *mockExtensionRepo) ListCategories() ([]types.ExtensionCategory, error) {
	return m.categories, m.errCategories
}

type mockExtCache struct {
	getExt       *types.Extension
	getExtErr    error
	getEID       *types.Extension
	getEIDErr    error
	getCats      []types.ExtensionCategory
	getCatsErr   error
	setExtErr    error
	setCatsErr   error
	setExtCount  int
	setCatsCount int
}

func (m *mockExtCache) GetExtension(_ context.Context, _ string) (*types.Extension, error) {
	if m.getExtErr != nil {
		return nil, m.getExtErr
	}
	return m.getExt, nil
}

func (m *mockExtCache) GetExtensionByExtID(_ context.Context, _ string) (*types.Extension, error) {
	if m.getEIDErr != nil {
		return nil, m.getEIDErr
	}
	return m.getEID, nil
}

func (m *mockExtCache) GetExtensionCategories(_ context.Context) ([]types.ExtensionCategory, error) {
	if m.getCatsErr != nil {
		return nil, m.getCatsErr
	}
	return m.getCats, nil
}

func (m *mockExtCache) SetExtension(_ context.Context, _ *types.Extension) error {
	m.setExtCount++
	return m.setExtErr
}

func (m *mockExtCache) SetExtensionCategories(_ context.Context, _ []types.ExtensionCategory) error {
	m.setCatsCount++
	return m.setCatsErr
}

func sampleExtension() *types.Extension {
	return &types.Extension{
		ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		ExtensionID: "sample-ext",
		Name:        "Sample",
		Description: "d",
		Author:      "a",
		Icon:        "i",
		Category:    types.ExtensionCategoryUtility,
	}
}

func TestNewExtensionService(t *testing.T) {
	ctx := context.Background()
	repo := &mockExtensionRepo{}
	log := logger.NewLogger()
	svc := NewExtensionService(ctx, log, repo, nil)
	require.NotNil(t, svc)
	assert.Nil(t, svc.cache)
}

func TestExtensionService_GetExtension_cacheHit(t *testing.T) {
	cached := sampleExtension()
	cache := &mockExtCache{getExt: cached}
	repo := &mockExtensionRepo{extByID: &types.Extension{Name: "db"}}
	svc := NewExtensionService(context.Background(), logger.NewLogger(), repo, cache)

	out, err := svc.GetExtension(cached.ID.String())
	require.NoError(t, err)
	assert.Equal(t, cached.Name, out.Name)
	assert.Zero(t, cache.setExtCount)
}

func TestExtensionService_GetExtension_cacheErrFallsThrough(t *testing.T) {
	cache := &mockExtCache{getExtErr: errors.New("redis down")}
	ext := sampleExtension()
	repo := &mockExtensionRepo{extByID: ext}
	svc := NewExtensionService(context.Background(), logger.NewLogger(), repo, cache)

	out, err := svc.GetExtension(ext.ID.String())
	require.NoError(t, err)
	assert.Equal(t, ext.Name, out.Name)
	require.Equal(t, 1, cache.setExtCount)
}

func TestExtensionService_GetExtension_cacheMissSetsCache(t *testing.T) {
	cache := &mockExtCache{getExt: nil}
	ext := sampleExtension()
	repo := &mockExtensionRepo{extByID: ext}
	svc := NewExtensionService(context.Background(), logger.NewLogger(), repo, cache)

	out, err := svc.GetExtension(ext.ID.String())
	require.NoError(t, err)
	assert.Equal(t, ext, out)
	assert.Equal(t, 1, cache.setExtCount)
}

func TestExtensionService_GetExtension_cacheSetErrStillReturns(t *testing.T) {
	cache := &mockExtCache{getExt: nil, setExtErr: errors.New("set failed")}
	ext := sampleExtension()
	repo := &mockExtensionRepo{extByID: ext}
	svc := NewExtensionService(context.Background(), logger.NewLogger(), repo, cache)

	out, err := svc.GetExtension(ext.ID.String())
	require.NoError(t, err)
	assert.Equal(t, ext, out)
}

func TestExtensionService_GetExtension_noCache(t *testing.T) {
	ext := sampleExtension()
	repo := &mockExtensionRepo{extByID: ext}
	svc := NewExtensionService(context.Background(), logger.NewLogger(), repo, nil)

	out, err := svc.GetExtension(ext.ID.String())
	require.NoError(t, err)
	assert.Equal(t, ext, out)
}

func TestExtensionService_GetExtension_storageErr(t *testing.T) {
	repo := &mockExtensionRepo{errGet: errors.New("not found")}
	svc := NewExtensionService(context.Background(), logger.NewLogger(), repo, nil)

	_, err := svc.GetExtension(uuid.New().String())
	require.Error(t, err)
}

func TestExtensionService_GetExtensionByID_cacheHit(t *testing.T) {
	cached := sampleExtension()
	cache := &mockExtCache{getEID: cached}
	repo := &mockExtensionRepo{}
	svc := NewExtensionService(context.Background(), logger.NewLogger(), repo, cache)

	out, err := svc.GetExtensionByID(cached.ExtensionID)
	require.NoError(t, err)
	assert.Equal(t, cached, out)
}

func TestExtensionService_GetExtensionByID_cacheErrFallsThrough(t *testing.T) {
	cache := &mockExtCache{getEIDErr: errors.New("fail")}
	ext := sampleExtension()
	repo := &mockExtensionRepo{extByEID: ext}
	svc := NewExtensionService(context.Background(), logger.NewLogger(), repo, cache)

	out, err := svc.GetExtensionByID(ext.ExtensionID)
	require.NoError(t, err)
	assert.Equal(t, ext, out)
}

func TestExtensionService_GetExtensionByID_cacheSetErr(t *testing.T) {
	cache := &mockExtCache{setExtErr: errors.New("set err")}
	ext := sampleExtension()
	repo := &mockExtensionRepo{extByEID: ext}
	svc := NewExtensionService(context.Background(), logger.NewLogger(), repo, cache)

	out, err := svc.GetExtensionByID(ext.ExtensionID)
	require.NoError(t, err)
	assert.Equal(t, ext, out)
}

func TestExtensionService_GetExtensionByID_storageErr(t *testing.T) {
	repo := &mockExtensionRepo{errGetEID: errors.New("missing")}
	svc := NewExtensionService(context.Background(), logger.NewLogger(), repo, nil)

	_, err := svc.GetExtensionByID("x")
	require.Error(t, err)
}

func TestExtensionService_ListExtensions(t *testing.T) {
	resp := &types.ExtensionListResponse{Total: 1, Page: 1, PageSize: 12}
	repo := &mockExtensionRepo{list: resp}
	svc := NewExtensionService(context.Background(), logger.NewLogger(), repo, nil)

	out, err := svc.ListExtensions(types.ExtensionListParams{})
	require.NoError(t, err)
	assert.Equal(t, resp, out)
}

func TestExtensionService_ListExtensions_err(t *testing.T) {
	repo := &mockExtensionRepo{errList: errors.New("db")}
	svc := NewExtensionService(context.Background(), logger.NewLogger(), repo, nil)

	_, err := svc.ListExtensions(types.ExtensionListParams{})
	require.Error(t, err)
}

func TestExtensionService_ListCategories_cacheHit(t *testing.T) {
	cats := []types.ExtensionCategory{types.ExtensionCategoryUtility}
	cache := &mockExtCache{getCats: cats}
	repo := &mockExtensionRepo{}
	svc := NewExtensionService(context.Background(), logger.NewLogger(), repo, cache)

	out, err := svc.ListCategories()
	require.NoError(t, err)
	assert.Equal(t, cats, out)
}

func TestExtensionService_ListCategories_cacheErrFallsThrough(t *testing.T) {
	cache := &mockExtCache{getCatsErr: errors.New("cache")}
	cats := []types.ExtensionCategory{types.ExtensionCategoryDatabase}
	repo := &mockExtensionRepo{categories: cats}
	svc := NewExtensionService(context.Background(), logger.NewLogger(), repo, cache)

	out, err := svc.ListCategories()
	require.NoError(t, err)
	assert.Equal(t, cats, out)
	assert.Equal(t, 1, cache.setCatsCount)
}

func TestExtensionService_ListCategories_cacheMissSetCache(t *testing.T) {
	cache := &mockExtCache{}
	cats := []types.ExtensionCategory{types.ExtensionCategoryOther}
	repo := &mockExtensionRepo{categories: cats}
	svc := NewExtensionService(context.Background(), logger.NewLogger(), repo, cache)

	out, err := svc.ListCategories()
	require.NoError(t, err)
	assert.Equal(t, cats, out)
	assert.Equal(t, 1, cache.setCatsCount)
}

func TestExtensionService_ListCategories_setCatsErr(t *testing.T) {
	cache := &mockExtCache{setCatsErr: errors.New("set cat")}
	repo := &mockExtensionRepo{categories: []types.ExtensionCategory{types.ExtensionCategoryGame}}
	svc := NewExtensionService(context.Background(), logger.NewLogger(), repo, cache)

	out, err := svc.ListCategories()
	require.NoError(t, err)
	assert.Len(t, out, 1)
}

func TestExtensionService_ListCategories_noCache(t *testing.T) {
	cats := []types.ExtensionCategory{types.ExtensionCategorySecurity}
	repo := &mockExtensionRepo{categories: cats}
	svc := NewExtensionService(context.Background(), logger.NewLogger(), repo, nil)

	out, err := svc.ListCategories()
	require.NoError(t, err)
	assert.Equal(t, cats, out)
}

func TestExtensionService_ListCategories_storageErr(t *testing.T) {
	repo := &mockExtensionRepo{errCategories: errors.New("db")}
	svc := NewExtensionService(context.Background(), logger.NewLogger(), repo, nil)

	_, err := svc.ListCategories()
	require.Error(t, err)
}
