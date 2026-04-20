package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

type ExtensionStorage struct {
	DB  *bun.DB
	Ctx context.Context
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
	_, err := s.DB.NewInsert().Model(extension).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *ExtensionStorage) CreateExtensionVariables(vars []types.ExtensionVariable) error {
	if len(vars) == 0 {
		return nil
	}
	_, err := s.DB.NewInsert().Model(&vars).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *ExtensionStorage) GetExtension(id string) (*types.Extension, error) {
	var extension types.Extension
	err := s.DB.NewSelect().
		Model(&extension).
		Relation("Variables").
		Where("id = ? AND deleted_at IS NULL", id).
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("extension not found")
		}
		return nil, err
	}
	return &extension, nil
}

func (s *ExtensionStorage) GetExtensionByID(extensionID string) (*types.Extension, error) {
	var extension types.Extension
	err := s.DB.NewSelect().
		Model(&extension).
		Relation("Variables").
		Where("extension_id = ? AND deleted_at IS NULL", extensionID).
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("extension not found")
		}
		return nil, err
	}
	return &extension, nil
}

func (s *ExtensionStorage) UpdateExtension(extension *types.Extension) error {
	_, err := s.DB.NewUpdate().
		Model(extension).
		Where("id = ?", extension.ID).
		Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *ExtensionStorage) DeleteExtension(id string) error {
	_, err := s.DB.NewUpdate().
		Model((*types.Extension)(nil)).
		Set("deleted_at = NOW()").
		Where("id = ?", id).
		Exec(s.Ctx)
	if err != nil {
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
		return nil, err
	}

	offset := (params.Page - 1) * params.PageSize
	query = query.Limit(params.PageSize).Offset(offset)

	err = query.Scan(s.Ctx)
	if err != nil {
		return nil, err
	}

	totalPages := (total + params.PageSize - 1) / params.PageSize

	if len(extensions) == 0 {
		extensions = make([]types.Extension, 0)
	}

	return &types.ExtensionListResponse{
		Extensions: extensions,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *ExtensionStorage) ListCategories() ([]types.ExtensionCategory, error) {
	var categories []types.ExtensionCategory
	err := s.DB.NewSelect().
		TableExpr("extensions").
		ColumnExpr("DISTINCT category").
		Where("deleted_at IS NULL").
		Scan(s.Ctx, &categories)
	if err != nil {
		return nil, err
	}
	return categories, nil
}
