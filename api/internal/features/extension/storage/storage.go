package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

type ExtensionStorage struct {
	DB     *bun.DB
	Ctx    context.Context
	Logger *logger.Logger // optional; nil disables storage logs
}

func (s *ExtensionStorage) storageLog(sev logger.Severity, msg, data string) {
	if s.Logger == nil {
		return
	}
	s.Logger.Log(sev, msg, data)
}

type ExtensionStorageInterface interface {
	CreateExtension(extension *types.Extension) error
	CreateExtensionVariables(vars []types.ExtensionVariable) error
	GetExtension(id string) (*types.Extension, error)
	GetExtensionByID(extensionID string) (*types.Extension, error)
	UpdateExtension(extension *types.Extension) error
	DeleteExtension(id string) error
	ListExtensions(params types.ExtensionListParams) (*types.ExtensionListResponse, error)
	ListCategories() ([]types.ExtensionCategory, error)
}

func (s *ExtensionStorage) CreateExtension(extension *types.Extension) error {
	ctxStr := fmt.Sprintf("extension_id=%s", extension.ExtensionID)
	s.storageLog(logger.Debug, "storage: CreateExtension", ctxStr)
	_, err := s.DB.NewInsert().Model(extension).Exec(s.Ctx)
	if err != nil {
		s.storageLog(logger.Error, fmt.Sprintf("storage: CreateExtension: %v", err), ctxStr)
		return err
	}
	return nil
}

func (s *ExtensionStorage) CreateExtensionVariables(vars []types.ExtensionVariable) error {
	if len(vars) == 0 {
		return nil
	}
	s.storageLog(logger.Debug, "storage: CreateExtensionVariables", fmt.Sprintf("count=%d", len(vars)))
	_, err := s.DB.NewInsert().Model(&vars).Exec(s.Ctx)
	if err != nil {
		s.storageLog(logger.Error, fmt.Sprintf("storage: CreateExtensionVariables: %v", err), "")
		return err
	}
	return nil
}

func (s *ExtensionStorage) GetExtension(id string) (*types.Extension, error) {
	ctxStr := fmt.Sprintf("id=%s", id)
	s.storageLog(logger.Debug, "storage: GetExtension", ctxStr)
	var extension types.Extension
	err := s.DB.NewSelect().
		Model(&extension).
		Relation("Variables").
		Where("id = ? AND deleted_at IS NULL", id).
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.storageLog(logger.Debug, "storage: GetExtension not found", ctxStr)
			return nil, errors.New("extension not found")
		}
		s.storageLog(logger.Error, fmt.Sprintf("storage: GetExtension: %v", err), ctxStr)
		return nil, err
	}
	s.storageLog(logger.Debug, "storage: GetExtension ok", ctxStr)
	return &extension, nil
}

func (s *ExtensionStorage) GetExtensionByID(extensionID string) (*types.Extension, error) {
	ctxStr := fmt.Sprintf("extension_id=%s", extensionID)
	s.storageLog(logger.Debug, "storage: GetExtensionByID", ctxStr)
	var extension types.Extension
	err := s.DB.NewSelect().
		Model(&extension).
		Relation("Variables").
		Where("extension_id = ? AND deleted_at IS NULL", extensionID).
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.storageLog(logger.Debug, "storage: GetExtensionByID not found", ctxStr)
			return nil, errors.New("extension not found")
		}
		s.storageLog(logger.Error, fmt.Sprintf("storage: GetExtensionByID: %v", err), ctxStr)
		return nil, err
	}
	s.storageLog(logger.Debug, "storage: GetExtensionByID ok", ctxStr)
	return &extension, nil
}

func (s *ExtensionStorage) UpdateExtension(extension *types.Extension) error {
	ctxStr := fmt.Sprintf("id=%s extension_id=%s", extension.ID, extension.ExtensionID)
	s.storageLog(logger.Debug, "storage: UpdateExtension", ctxStr)
	_, err := s.DB.NewUpdate().
		Model(extension).
		Where("id = ?", extension.ID).
		Exec(s.Ctx)
	if err != nil {
		s.storageLog(logger.Error, fmt.Sprintf("storage: UpdateExtension: %v", err), ctxStr)
		return err
	}
	return nil
}

func (s *ExtensionStorage) DeleteExtension(id string) error {
	ctxStr := fmt.Sprintf("id=%s", id)
	s.storageLog(logger.Debug, "storage: DeleteExtension", ctxStr)
	_, err := s.DB.NewUpdate().
		Model((*types.Extension)(nil)).
		Set("deleted_at = NOW()").
		Where("id = ?", id).
		Exec(s.Ctx)
	if err != nil {
		s.storageLog(logger.Error, fmt.Sprintf("storage: DeleteExtension: %v", err), ctxStr)
		return err
	}
	return nil
}

func (s *ExtensionStorage) ListExtensions(params types.ExtensionListParams) (*types.ExtensionListResponse, error) {
	var extensions []types.Extension

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 12
	}
	if params.SortBy == "" {
		params.SortBy = types.ExtensionSortFieldName
	}
	if params.SortDir == "" {
		params.SortDir = types.SortDirectionAsc
	}

	ctxStr := fmt.Sprintf("page=%d page_size=%d sort_by=%s sort_dir=%s", params.Page, params.PageSize, params.SortBy, params.SortDir)
	if params.Category != nil {
		ctxStr += fmt.Sprintf(" category=%s", *params.Category)
	}
	if params.Type != nil {
		ctxStr += fmt.Sprintf(" type=%s", *params.Type)
	}
	if params.Search != "" {
		ctxStr += fmt.Sprintf(" search=%q", params.Search)
	}
	s.storageLog(logger.Debug, "storage: ListExtensions", ctxStr)

	query := s.DB.NewSelect().
		Model(&extensions).
		Relation("Variables").
		Where("deleted_at IS NULL")

	if params.Category != nil {
		query = query.Where("category = ?", *params.Category)
	}

	if params.Type != nil {
		query = query.Where("extension_type = ?", *params.Type)
	}

	if params.Search != "" {
		searchPattern := "%" + params.Search + "%"
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("name ILIKE ?", searchPattern).
				WhereOr("description ILIKE ?", searchPattern).
				WhereOr("author ILIKE ?", searchPattern).
				WhereOr("category::text ILIKE ?", searchPattern)
		})
	}

	sortColumn := string(params.SortBy)
	if params.SortDir == types.SortDirectionDesc {
		query = query.Order("featured DESC").Order(sortColumn + " DESC")
	} else {
		query = query.Order("featured DESC").Order(sortColumn + " ASC")
	}

	var total int
	countQuery := s.DB.NewSelect().
		Model((*types.Extension)(nil)).
		Where("deleted_at IS NULL")

	if params.Category != nil {
		countQuery = countQuery.Where("category = ?", *params.Category)
	}

	if params.Type != nil {
		countQuery = countQuery.Where("extension_type = ?", *params.Type)
	}

	if params.Search != "" {
		searchPattern := "%" + params.Search + "%"
		countQuery = countQuery.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("name ILIKE ?", searchPattern).
				WhereOr("description ILIKE ?", searchPattern).
				WhereOr("author ILIKE ?", searchPattern).
				WhereOr("category::text ILIKE ?", searchPattern)
		})
	}

	total, err := countQuery.Count(s.Ctx)
	if err != nil {
		s.storageLog(logger.Error, fmt.Sprintf("storage: ListExtensions count: %v", err), ctxStr)
		return nil, err
	}

	offset := (params.Page - 1) * params.PageSize
	query = query.Limit(params.PageSize).Offset(offset)

	err = query.Scan(s.Ctx)
	if err != nil {
		s.storageLog(logger.Error, fmt.Sprintf("storage: ListExtensions scan: %v", err), ctxStr)
		return nil, err
	}

	totalPages := (total + params.PageSize - 1) / params.PageSize

	if len(extensions) == 0 {
		extensions = make([]types.Extension, 0)
	}

	s.storageLog(logger.Debug, "storage: ListExtensions ok", fmt.Sprintf("%s total=%d rows=%d", ctxStr, total, len(extensions)))
	return &types.ExtensionListResponse{
		Extensions: extensions,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *ExtensionStorage) ListCategories() ([]types.ExtensionCategory, error) {
	s.storageLog(logger.Debug, "storage: ListCategories", "")
	var categories []types.ExtensionCategory
	err := s.DB.NewSelect().
		TableExpr("extensions").
		ColumnExpr("DISTINCT category").
		Where("deleted_at IS NULL").
		Scan(s.Ctx, &categories)
	if err != nil {
		s.storageLog(logger.Error, fmt.Sprintf("storage: ListCategories: %v", err), "")
		return nil, err
	}
	s.storageLog(logger.Debug, "storage: ListCategories ok", fmt.Sprintf("count=%d", len(categories)))
	return categories, nil
}
