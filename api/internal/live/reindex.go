package live

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"github.com/raghavyuva/nixopus-api/internal/mover"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// Skip dirs when walking
var skipDirs = map[string]bool{
	"node_modules": true, ".git": true, "__pycache__": true, ".venv": true, "venv": true,
	"dist": true, "build": true, ".next": true, ".turbo": true, ".cache": true,
	"coverage": true, ".output": true, ".nuxt": true, ".svelte-kit": true,
	"target": true, ".gradle": true, ".parcel-cache": true, ".docusaurus": true,
	".expo": true, ".serverless": true, ".idea": true, ".vscode": true,
	".terraform": true, ".tox": true, "vendor": true, "bower_components": true,
	".sass-cache": true, ".pytest_cache": true, ".mypy_cache": true,
	".angular": true, "out": true, ".vercel": true, ".netlify": true, ".amplify": true,
}

var keepDotDirs = map[string]bool{
	".github": true, ".circleci": true, ".gitlab": true, ".husky": true,
}

var binaryExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
	".ico": true, ".svg": true, ".bmp": true, ".tiff": true,
	".woff": true, ".woff2": true, ".ttf": true, ".eot": true, ".otf": true,
	".mp3": true, ".mp4": true, ".wav": true, ".avi": true, ".mov": true,
	".webm": true, ".flac": true, ".ogg": true,
	".zip": true, ".tar": true, ".gz": true, ".bz2": true, ".xz": true,
	".rar": true, ".7z": true, ".zst": true,
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".ppt": true, ".pptx": true,
	".exe": true, ".dll": true, ".so": true, ".dylib": true, ".o": true,
	".a": true, ".lib": true, ".bin": true, ".lock": true, ".lockb": true,
	".map": true, ".wasm": true, ".pyc": true, ".pyo": true, ".class": true,
	".jar": true, ".db": true, ".sqlite": true, ".sqlite3": true, ".DS_Store": true,
}

func isSkippedDir(name string) bool {
	if skipDirs[name] {
		return true
	}
	return strings.HasPrefix(name, ".") && !keepDotDirs[name]
}

func isBinaryExt(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return binaryExts[ext]
}

// ReindexResult summarizes what was reindexed.
type ReindexResult struct {
	Added     int
	Modified  int
	Deleted   int
	Unchanged int
}

// ReindexFromStaging walks the staging directory via SFTP, builds path→hash, diffs against
// stored manifest, and applies incremental index updates. Uses a single SFTP connection.
// organizationID is required to resolve the SSH manager for the tenant.
func ReindexFromStaging(ctx context.Context, store *shared_storage.Store, stagingPath string, applicationID, organizationID uuid.UUID) (*ReindexResult, error) {
	if store == nil || store.DB == nil {
		return nil, nil
	}
	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, organizationID.String())

	currentPaths := make(map[string]string)
	var toSync, toDelete []string
	var storedPaths map[string]string

	err := utils.WithSFTPClient(orgCtx, func(client *sftp.Client) error {
		walker := client.Walk(stagingPath)
		for walker.Step() {
			if err := walker.Err(); err != nil {
				return err
			}
			path := walker.Path()
			info := walker.Stat()
			if info == nil {
				continue
			}
			if info.IsDir() {
				if isSkippedDir(info.Name()) {
					walker.SkipDir()
				}
				continue
			}
			if isBinaryExt(info.Name()) || info.Size() > int64(maxIndexableSize()) {
				continue
			}
			rel, err := filepath.Rel(stagingPath, path)
			if err != nil || strings.HasPrefix(rel, "..") {
				continue
			}
			rel = filepath.ToSlash(filepath.Clean(rel))
			content, err := utils.ReadFileBytesFromClient(client, path)
			if err != nil || !isTextContent(content) {
				continue
			}
			h := sha256.Sum256(content)
			currentPaths[rel] = hex.EncodeToString(h[:])
		}

		var loadErr error
		storedPaths, _, loadErr = LoadManifest(orgCtx, store, applicationID)
		if loadErr != nil {
			return loadErr
		}
		if storedPaths == nil {
			storedPaths = make(map[string]string)
		}

		tree := mover.BuildFromPaths(currentPaths)
		toSync, toDelete := mover.DiffAgainst(tree, storedPaths)

		for _, p := range toDelete {
			if err := DeleteFileChunks(orgCtx, store, applicationID, p); err != nil {
				return err
			}
		}
		for _, p := range toSync {
			fullPath := filepath.Join(stagingPath, filepath.FromSlash(p))
			content, err := utils.ReadFileBytesFromClient(client, fullPath)
			if err != nil {
				continue
			}
			if err := IndexFileChunks(orgCtx, store, applicationID, p, content); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	result := &ReindexResult{Deleted: len(toDelete)}
	for _, p := range toSync {
		if _, wasStored := storedPaths[p]; wasStored {
			result.Modified++
		} else {
			result.Added++
		}
	}

	for p := range storedPaths {
		if _, inCurrent := currentPaths[p]; inCurrent && !sliceContains(toSync, p) {
			result.Unchanged++
		}
	}

	// Persist final manifest (path→hash) in one write
	if err := PersistManifest(orgCtx, store, applicationID, currentPaths); err != nil {
		return nil, err
	}
	return result, nil
}

func sliceContains(s []string, x string) bool {
	for _, v := range s {
		if v == x {
			return true
		}
	}
	return false
}
