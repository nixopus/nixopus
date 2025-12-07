package loader

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/extension/parser"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type ExtensionLoader struct {
	db *bun.DB
}

func NewExtensionLoader(db *bun.DB) *ExtensionLoader {
	return &ExtensionLoader{
		db: db,
	}
}

func (l *ExtensionLoader) LoadExtensionsFromDirectory(ctx context.Context, dirPath string) error {
	parser := parser.NewParser()

	extensions, allVariables, err := parser.LoadExtensionsFromDirectory(dirPath)
	if err != nil {
		return fmt.Errorf("failed to load extensions from directory: %w", err)
	}

	log.Printf("Found %d extension files in %s", len(extensions), dirPath)

	for i, extension := range extensions {
		variables := allVariables[i]

		if err := l.upsertExtension(ctx, extension, variables); err != nil {
			log.Printf("Failed to upsert extension %s: %v", extension.ExtensionID, err)
			continue
		}

		if (i+1)%10 == 0 || i == len(extensions)-1 {
			log.Printf("Processed %d/%d extensions", i+1, len(extensions))
		}
	}

	return nil
}

func (l *ExtensionLoader) upsertExtension(ctx context.Context, extension *types.Extension, variables []types.ExtensionVariable) error {
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var existingExtension types.Extension
	err = tx.NewSelect().
		Model(&existingExtension).
		Where("extension_id = ?", extension.ExtensionID).
		Scan(ctx)

	if err != nil && err.Error() != "sql: no rows in result set" {
		return fmt.Errorf("failed to check existing extension: %w", err)
	}

	if err == nil {
		if existingExtension.ContentHash == extension.ContentHash {
			// log.Printf("Extension %s unchanged, skipping", extension.ExtensionID)
			return tx.Commit()
		}

		extension.ID = existingExtension.ID
		extension.CreatedAt = existingExtension.CreatedAt

		if _, err := tx.NewUpdate().
			Model(extension).
			Where("id = ?", extension.ID).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to update extension: %w", err)
		}

		if err := l.deleteExtensionVariables(ctx, tx, extension.ID); err != nil {
			return fmt.Errorf("failed to delete old variables: %w", err)
		}
	} else {
		extension.ID = uuid.New()

		if _, err := tx.NewInsert().
			Model(extension).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to insert extension: %w", err)
		}
	}

	for i := range variables {
		variables[i].ID = uuid.New()
		variables[i].ExtensionID = extension.ID
	}

	if len(variables) > 0 {
		if _, err := tx.NewInsert().
			Model(&variables).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to insert variables: %w", err)
		}
	}

	return tx.Commit()
}

func (l *ExtensionLoader) deleteExtensionVariables(ctx context.Context, tx bun.Tx, extensionID uuid.UUID) error {
	_, err := tx.NewDelete().
		Model((*types.ExtensionVariable)(nil)).
		Where("extension_id = ?", extensionID).
		Exec(ctx)
	return err
}

func (l *ExtensionLoader) LoadExtensionsFromTemplates(ctx context.Context) error {
	templatesPath := filepath.Join(".", "templates")
	return l.LoadExtensionsFromDirectory(ctx, templatesPath)
}

func (l *ExtensionLoader) GetExtensionByID(ctx context.Context, extensionID string) (*types.Extension, error) {
	var extension types.Extension

	err := l.db.NewSelect().
		Model(&extension).
		Relation("Variables").
		Where("extension_id = ? AND deleted_at IS NULL", extensionID).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get extension: %w", err)
	}

	return &extension, nil
}

func (l *ExtensionLoader) ListExtensions(ctx context.Context, category *types.ExtensionCategory) ([]types.Extension, error) {
	var extensions []types.Extension

	query := l.db.NewSelect().
		Model(&extensions).
		Relation("Variables").
		Where("deleted_at IS NULL")

	if category != nil {
		query = query.Where("category = ?", *category)
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list extensions: %w", err)
	}

	return extensions, nil
}
