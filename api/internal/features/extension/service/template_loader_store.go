package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

var extensionTemplateUpdateColumns = []string{
	"name", "description", "author", "icon", "category", "extension_type", "version",
	"is_verified", "featured", "yaml_content", "parsed_content", "content_hash",
	"validation_status", "updated_at", "deleted_at",
}

// extensionTemplateTx is a transactional unit for batch extension/variable writes.
type extensionTemplateTx interface {
	insertExtensions(ctx context.Context, extensions []*types.Extension) error
	insertVariables(ctx context.Context, vars []types.ExtensionVariable) error
	deleteVariablesForExtensionUUIDs(ctx context.Context, extensionUUIDs []uuid.UUID) error
	updateExtensionRow(ctx context.Context, ext *types.Extension) error
	commit() error
	rollback() error
}

// extensionTemplateStore abstracts persistence for TemplateLoader (bun-backed in production, mock in tests).
type extensionTemplateStore interface {
	beginLoadTx(ctx context.Context) (extensionTemplateTx, error)
	fetchExtensionsByExtensionIDs(ctx context.Context, extensionIDs []string) ([]types.Extension, error)
	fetchNonDeletedExtensions(ctx context.Context) ([]types.Extension, error)
	softDeleteExtensionsByExtensionIDs(ctx context.Context, extensionIDs []string) error
	getExtensionWithVariables(ctx context.Context, extensionID string) (*types.Extension, error)
}

type bunExtensionTemplateStore struct {
	db *bun.DB
}

func newBunExtensionTemplateStore(db *bun.DB) *bunExtensionTemplateStore {
	return &bunExtensionTemplateStore{db: db}
}

func (s *bunExtensionTemplateStore) beginLoadTx(ctx context.Context) (extensionTemplateTx, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &bunExtensionTemplateTx{tx: tx}, nil
}

func (s *bunExtensionTemplateStore) fetchExtensionsByExtensionIDs(ctx context.Context, extensionIDs []string) ([]types.Extension, error) {
	var existingExtensions []types.Extension
	err := s.db.NewSelect().
		Model(&existingExtensions).
		Where("extension_id IN (?)", bun.In(extensionIDs)).
		Scan(ctx)
	return existingExtensions, err
}

func (s *bunExtensionTemplateStore) fetchNonDeletedExtensions(ctx context.Context) ([]types.Extension, error) {
	var allExtensions []types.Extension
	err := s.db.NewSelect().
		Model(&allExtensions).
		Where("deleted_at IS NULL").
		Scan(ctx)
	return allExtensions, err
}

func (s *bunExtensionTemplateStore) softDeleteExtensionsByExtensionIDs(ctx context.Context, extensionIDs []string) error {
	_, err := s.db.NewUpdate().
		Model((*types.Extension)(nil)).
		Set("deleted_at = NOW()").
		Where("extension_id IN (?) AND deleted_at IS NULL", bun.In(extensionIDs)).
		Exec(ctx)
	return err
}

func (s *bunExtensionTemplateStore) getExtensionWithVariables(ctx context.Context, extensionID string) (*types.Extension, error) {
	var extension types.Extension
	err := s.db.NewSelect().
		Model(&extension).
		Relation("Variables").
		Where("extension_id = ? AND deleted_at IS NULL", extensionID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &extension, nil
}

type bunExtensionTemplateTx struct {
	tx bun.Tx
}

func (b *bunExtensionTemplateTx) insertExtensions(ctx context.Context, extensions []*types.Extension) error {
	_, err := b.tx.NewInsert().Model(&extensions).Exec(ctx)
	return err
}

func (b *bunExtensionTemplateTx) insertVariables(ctx context.Context, vars []types.ExtensionVariable) error {
	if len(vars) == 0 {
		return nil
	}
	_, err := b.tx.NewInsert().Model(&vars).Exec(ctx)
	return err
}

func (b *bunExtensionTemplateTx) deleteVariablesForExtensionUUIDs(ctx context.Context, extensionUUIDs []uuid.UUID) error {
	_, err := b.tx.NewDelete().
		Model((*types.ExtensionVariable)(nil)).
		Where("extension_id IN (?)", bun.In(extensionUUIDs)).
		Exec(ctx)
	return err
}

func (b *bunExtensionTemplateTx) updateExtensionRow(ctx context.Context, ext *types.Extension) error {
	_, err := b.tx.NewUpdate().
		Model(ext).
		Column(extensionTemplateUpdateColumns...).
		Where("id = ?", ext.ID).
		Exec(ctx)
	return err
}

func (b *bunExtensionTemplateTx) commit() error {
	return b.tx.Commit()
}

func (b *bunExtensionTemplateTx) rollback() error {
	return b.tx.Rollback()
}
