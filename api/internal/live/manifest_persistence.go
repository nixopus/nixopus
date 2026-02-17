package live

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/mover"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

// AddPath adds a path and its checksum to the manifest in DB.
func AddPath(ctx context.Context, store *shared_storage.Store, applicationID uuid.UUID, path, checksum string) error {
	if store == nil || store.DB == nil {
		return nil
	}
	paths, _, err := LoadManifest(ctx, store, applicationID)
	if err != nil {
		return err
	}
	if paths == nil {
		paths = make(map[string]string)
	}
	paths[normalizeManifestPath(path)] = checksum
	return PersistManifest(ctx, store, applicationID, paths)
}

// RemovePath removes a path from the manifest in DB.
func RemovePath(ctx context.Context, store *shared_storage.Store, applicationID uuid.UUID, path string) error {
	if store == nil || store.DB == nil {
		return nil
	}
	paths, _, err := LoadManifest(ctx, store, applicationID)
	if err != nil {
		return err
	}
	if paths != nil {
		delete(paths, normalizeManifestPath(path))
	} else {
		paths = make(map[string]string)
	}
	return PersistManifest(ctx, store, applicationID, paths)
}

func normalizeManifestPath(p string) string {
	return filepath.ToSlash(filepath.Clean(p))
}

// PersistManifest saves paths and cached root_hash to application_context.
func PersistManifest(ctx context.Context, store *shared_storage.Store, applicationID uuid.UUID, paths map[string]string) error {
	if store == nil || store.DB == nil {
		return nil
	}
	pathMap := types.PathChecksumMap(paths)
	if pathMap == nil {
		pathMap = make(types.PathChecksumMap)
	}
	now := time.Now().UTC()
	tree := mover.BuildFromPaths(paths)
	simhash := mover.ComputeSimhash(paths)

	ac := &types.ApplicationContext{
		ApplicationID: applicationID,
		RootHash:      tree.RootHash,
		Simhash:       simhash,
		Paths:         pathMap,
		UpdatedAt:     now,
	}
	_, err := store.DB.NewInsert().
		Model(ac).
		On("CONFLICT (application_id) DO UPDATE").
		Set("root_hash = EXCLUDED.root_hash").
		Set("simhash = EXCLUDED.simhash").
		Set("paths = EXCLUDED.paths").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	return err
}

// LoadManifest loads paths, root_hash, and simhash from application_context.
// Returns nil paths if not found.
func LoadManifest(ctx context.Context, store *shared_storage.Store, applicationID uuid.UUID) (paths map[string]string, rootHash string, err error) {
	if store == nil || store.DB == nil {
		return nil, "", nil
	}
	var ac types.ApplicationContext
	err = store.DB.NewSelect().
		Model(&ac).
		Where("application_id = ?", applicationID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", nil
		}
		return nil, "", err
	}
	if ac.Paths == nil {
		return nil, ac.RootHash, nil
	}
	return map[string]string(ac.Paths), ac.RootHash, nil
}
