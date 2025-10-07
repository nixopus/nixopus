package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type ExtensionStorage struct {
	DB  *bun.DB
	Ctx context.Context
	tx  *bun.Tx
}

type ExtensionStorageInterface interface {
	CreateExtension(extension *types.Extension) error
	CreateExtensionVariables(vars []types.ExtensionVariable) error
	GetExtension(id string) (*types.Extension, error)
	GetExtensionByID(extensionID string) (*types.Extension, error)
	UpdateExtension(extension *types.Extension) error
	DeleteExtension(id string) error
	ListExtensions(params types.ExtensionListParams) (*types.ExtensionListResponse, error)
	CreateExecution(exec *types.ExtensionExecution) error
	CreateExecutionSteps(steps []types.ExecutionStep) error
	ListExecutionSteps(executionID string) ([]types.ExecutionStep, error)
	UpdateExecutionStep(step *types.ExecutionStep) error
	UpdateExecution(exec *types.ExtensionExecution) error
	GetExecutionByID(id string) (*types.ExtensionExecution, error)
	BeginTx() (bun.Tx, error)
	WithTx(tx bun.Tx) ExtensionStorageInterface
}

func (s *ExtensionStorage) BeginTx() (bun.Tx, error) {
	return s.DB.BeginTx(s.Ctx, nil)
}

func (s *ExtensionStorage) WithTx(tx bun.Tx) ExtensionStorageInterface {
	return &ExtensionStorage{
		DB:  s.DB,
		Ctx: s.Ctx,
		tx:  &tx,
	}
}

func (s *ExtensionStorage) getDB() bun.IDB {
	if s.tx != nil {
		return *s.tx
	}
	return s.DB
}

func (s *ExtensionStorage) CreateExtension(extension *types.Extension) error {
	_, err := s.getDB().NewInsert().Model(extension).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *ExtensionStorage) CreateExtensionVariables(vars []types.ExtensionVariable) error {
	if len(vars) == 0 {
		return nil
	}
	_, err := s.getDB().NewInsert().Model(&vars).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *ExtensionStorage) GetExtension(id string) (*types.Extension, error) {
	var extension types.Extension
	err := s.getDB().NewSelect().
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
	err := s.getDB().NewSelect().
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
	_, err := s.getDB().NewUpdate().
		Model(extension).
		Where("id = ?", extension.ID).
		Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *ExtensionStorage) DeleteExtension(id string) error {
	_, err := s.getDB().NewUpdate().
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

	query := s.getDB().NewSelect().
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
				WhereOr("category ILIKE ?", searchPattern)
		})
	}

	sortColumn := string(params.SortBy)
	if params.SortDir == types.SortDirectionDesc {
		query = query.Order(sortColumn + " DESC")
	} else {
		query = query.Order(sortColumn + " ASC")
	}

	var total int
	countQuery := s.getDB().NewSelect().
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
				WhereOr("category ILIKE ?", searchPattern)
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

	return &types.ExtensionListResponse{
		Extensions: extensions,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *ExtensionStorage) CreateExecution(exec *types.ExtensionExecution) error {
	_, err := s.getDB().NewInsert().Model(exec).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *ExtensionStorage) CreateExecutionSteps(steps []types.ExecutionStep) error {
	if len(steps) == 0 {
		return nil
	}
	_, err := s.getDB().NewInsert().Model(&steps).Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *ExtensionStorage) ListExecutionSteps(executionID string) ([]types.ExecutionStep, error) {
	var steps []types.ExecutionStep
	err := s.getDB().NewSelect().
		Model(&steps).
		Where("execution_id = ?", executionID).
		Order("step_order ASC").
		Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return steps, nil
}

func (s *ExtensionStorage) UpdateExecutionStep(step *types.ExecutionStep) error {
	_, err := s.getDB().NewUpdate().
		Model(step).
		Where("id = ?", step.ID).
		Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *ExtensionStorage) UpdateExecution(exec *types.ExtensionExecution) error {
	_, err := s.getDB().NewUpdate().
		Model(exec).
		Where("id = ?", exec.ID).
		Exec(s.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *ExtensionStorage) GetExecutionByID(id string) (*types.ExtensionExecution, error) {
	var exec types.ExtensionExecution
	err := s.getDB().NewSelect().
		Model(&exec).
		Where("id = ?", id).
		Scan(s.Ctx)
	if err != nil {
		return nil, err
	}
	return &exec, nil
}
