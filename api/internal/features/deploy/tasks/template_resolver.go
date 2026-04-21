package tasks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/internal/utils"
	"github.com/pkg/sftp"
)

type TemplateSourceResolver struct{}

func (r *TemplateSourceResolver) Resolve(ctx context.Context, config SourceResolveConfig) (string, error) {
	app := config.Application
	templateID := app.TemplateID
	if templateID == "" {
		return "", fmt.Errorf("template_id is required for template source")
	}

	composeFile := filepath.Join(".", "templates", templateID, "docker-compose.yml")
	composeContent, err := os.ReadFile(composeFile)
	if err != nil {
		return "", fmt.Errorf("failed to read template compose file %s: %w", composeFile, err)
	}

	stagingPath := filepath.Join(
		"/var/nixopus/repos",
		app.UserID.String(),
		string(app.Environment),
		app.ID.String(),
	)

	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, app.OrganizationID.String())

	err = utils.WithSFTPClientFromPool(orgCtx, func(sftpClient *sftp.Client) error {
		if err := sftpClient.MkdirAll(stagingPath); err != nil {
			return fmt.Errorf("failed to create staging directory: %w", err)
		}

		destPath := filepath.Join(stagingPath, "docker-compose.yml")
		f, err := sftpClient.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create compose file on server: %w", err)
		}
		defer f.Close()

		if _, err := f.Write(composeContent); err != nil {
			return fmt.Errorf("failed to write compose file to server: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return stagingPath, nil
}
