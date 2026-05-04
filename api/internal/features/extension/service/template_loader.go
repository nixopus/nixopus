package service

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/google/uuid"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

// TemplateLoader reads extension metadata.yaml trees from disk and syncs them into Postgres.
type TemplateLoader struct {
	store extensionTemplateStore
}

var _ shared_storage.ExtensionTemplateLoader = (*TemplateLoader)(nil)

// NewTemplateLoader constructs a loader backed by the application database.
func NewTemplateLoader(db *bun.DB) *TemplateLoader {
	return newTemplateLoader(newBunExtensionTemplateStore(db))
}

func newTemplateLoader(store extensionTemplateStore) *TemplateLoader {
	return &TemplateLoader{store: store}
}

func (l *TemplateLoader) LoadExtensionsFromDirectory(ctx context.Context, dirPath string) error {
	p := newExtensionYAMLParser()

	extensions, allVariables, err := p.loadExtensionsFromDirectory(dirPath)
	if err != nil {
		return fmt.Errorf("failed to load extensions from directory: %w", err)
	}

	log.Printf("Found %d extension files in %s", len(extensions), dirPath)

	if len(extensions) == 0 {
		return nil
	}

	foundExtensionIDs := make(map[string]bool)
	for _, extension := range extensions {
		foundExtensionIDs[extension.ExtensionID] = true
	}

	extensionIDs := make([]string, 0, len(extensions))
	for _, ext := range extensions {
		extensionIDs = append(extensionIDs, ext.ExtensionID)
	}

	existingExtensions, err := l.store.fetchExtensionsByExtensionIDs(ctx, extensionIDs)
	if err != nil {
		return fmt.Errorf("failed to fetch existing extensions: %w", err)
	}

	existingMap := make(map[string]*types.Extension)
	for i := range existingExtensions {
		existingMap[existingExtensions[i].ExtensionID] = &existingExtensions[i]
	}

	var toInsert []*types.Extension
	var insertVariables [][]types.ExtensionVariable
	var toUpdate []*types.Extension
	var updateVariables [][]types.ExtensionVariable
	skippedCount := 0

	for i, extension := range extensions {
		variables := allVariables[i]
		existing, exists := existingMap[extension.ExtensionID]

		if exists {
			if existing.DeletedAt != nil {
				extension.ID = existing.ID
				extension.CreatedAt = existing.CreatedAt
				toUpdate = append(toUpdate, extension)
				updateVariables = append(updateVariables, variables)
				continue
			}
			if existing.ContentHash == extension.ContentHash {
				skippedCount++
				continue
			}
			extension.ID = existing.ID
			extension.CreatedAt = existing.CreatedAt
			toUpdate = append(toUpdate, extension)
			updateVariables = append(updateVariables, variables)
		} else {
			extension.ID = uuid.New()
			toInsert = append(toInsert, extension)
			insertVariables = append(insertVariables, variables)
		}
	}

	log.Printf("Processing: %d new, %d updated, %d unchanged", len(toInsert), len(toUpdate), skippedCount)

	if len(toInsert) > 0 {
		if err := l.batchInsertExtensions(ctx, toInsert, insertVariables); err != nil {
			return fmt.Errorf("failed to batch insert extensions: %w", err)
		}
	}

	if len(toUpdate) > 0 {
		if err := l.batchUpdateExtensions(ctx, toUpdate, updateVariables); err != nil {
			return fmt.Errorf("failed to batch update extensions: %w", err)
		}
	}

	if err := l.removeDeletedExtensions(ctx, foundExtensionIDs); err != nil {
		log.Printf("Warning: Failed to remove deleted extensions: %v", err)
	}

	return nil
}

func (l *TemplateLoader) batchInsertExtensions(ctx context.Context, extensions []*types.Extension, allVariables [][]types.ExtensionVariable) error {
	if len(extensions) == 0 {
		return nil
	}

	tx, err := l.store.beginLoadTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.rollback()

	if err := tx.insertExtensions(ctx, extensions); err != nil {
		return fmt.Errorf("failed to batch insert extensions: %w", err)
	}

	var allVars []types.ExtensionVariable
	for i, extension := range extensions {
		variables := allVariables[i]
		for j := range variables {
			variables[j].ID = uuid.New()
			variables[j].ExtensionID = extension.ID
			allVars = append(allVars, variables[j])
		}
	}

	if len(allVars) > 0 {
		if err := tx.insertVariables(ctx, allVars); err != nil {
			return fmt.Errorf("failed to batch insert variables: %w", err)
		}
	}

	if err := tx.commit(); err != nil {
		return err
	}
	return nil
}

func (l *TemplateLoader) batchUpdateExtensions(ctx context.Context, extensions []*types.Extension, allVariables [][]types.ExtensionVariable) error {
	if len(extensions) == 0 {
		return nil
	}

	tx, err := l.store.beginLoadTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.rollback()

	extensionIDs := make([]uuid.UUID, len(extensions))
	for i, ext := range extensions {
		extensionIDs[i] = ext.ID
	}

	if err := tx.deleteVariablesForExtensionUUIDs(ctx, extensionIDs); err != nil {
		return fmt.Errorf("failed to bulk delete variables: %w", err)
	}

	for _, extension := range extensions {
		if err := tx.updateExtensionRow(ctx, extension); err != nil {
			return fmt.Errorf("failed to update extension %s: %w", extension.ExtensionID, err)
		}
	}

	var allVars []types.ExtensionVariable
	for i, extension := range extensions {
		variables := allVariables[i]
		for j := range variables {
			variables[j].ID = uuid.New()
			variables[j].ExtensionID = extension.ID
			allVars = append(allVars, variables[j])
		}
	}

	if len(allVars) > 0 {
		if err := tx.insertVariables(ctx, allVars); err != nil {
			return fmt.Errorf("failed to batch insert variables: %w", err)
		}
	}

	if err := tx.commit(); err != nil {
		return err
	}
	return nil
}

func (l *TemplateLoader) removeDeletedExtensions(ctx context.Context, foundExtensionIDs map[string]bool) error {
	allExtensions, err := l.store.fetchNonDeletedExtensions(ctx)
	if err != nil {
		return fmt.Errorf("failed to query extensions: %w", err)
	}

	var extensionsToDelete []string
	for _, ext := range allExtensions {
		if !foundExtensionIDs[ext.ExtensionID] {
			extensionsToDelete = append(extensionsToDelete, ext.ExtensionID)
		}
	}

	if len(extensionsToDelete) == 0 {
		return nil
	}

	log.Printf("Removing %d extensions that are no longer in templates directory", len(extensionsToDelete))

	if err := l.store.softDeleteExtensionsByExtensionIDs(ctx, extensionsToDelete); err != nil {
		return fmt.Errorf("failed to delete removed extensions: %w", err)
	}

	return nil
}

// LoadExtensionsFromTemplates loads from ./templates relative to the process working directory.
func (l *TemplateLoader) LoadExtensionsFromTemplates(ctx context.Context) error {
	templatesPath := filepath.Join(".", "templates")
	return l.LoadExtensionsFromDirectory(ctx, templatesPath)
}

func (l *TemplateLoader) GetExtensionByID(ctx context.Context, extensionID string) (*types.Extension, error) {
	ext, err := l.store.getExtensionWithVariables(ctx, extensionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get extension: %w", err)
	}
	return ext, nil
}
