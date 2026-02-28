package live

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/sftp"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// IndexResult summarizes what was indexed.
type IndexResult struct {
	Indexed int `json:"indexed"`
	Skipped int `json:"skipped"`
}

// IndexFromStaging walks the staging directory via SFTP and indexes every
// indexable file by calling IndexFileChunks. Existing chunks for each file
// are replaced (IndexFileChunks handles upsert per path).
func IndexFromStaging(ctx context.Context, store *shared_storage.Store, stagingPath string, applicationID, organizationID uuid.UUID) (*IndexResult, error) {
	if store == nil || store.DB == nil {
		return nil, nil
	}

	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, organizationID.String())
	result := &IndexResult{}

	err := utils.WithSFTPClient(orgCtx, func(client *sftp.Client) error {
		walker := client.Walk(stagingPath)
		for walker.Step() {
			if err := walker.Err(); err != nil {
				return err
			}
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
				result.Skipped++
				continue
			}

			absPath := walker.Path()
			rel, err := filepath.Rel(stagingPath, absPath)
			if err != nil || strings.HasPrefix(rel, "..") {
				result.Skipped++
				continue
			}
			rel = filepath.ToSlash(filepath.Clean(rel))

			content, err := utils.ReadFileBytesFromClient(client, absPath)
			if err != nil || !isTextContent(content) {
				result.Skipped++
				continue
			}

			if err := IndexFileChunks(orgCtx, store, applicationID, rel, content); err != nil {
				return err
			}
			result.Indexed++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
