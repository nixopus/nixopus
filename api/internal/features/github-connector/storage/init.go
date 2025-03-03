package storage

import (
	"context"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type GithubConnectorStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

func (s *GithubConnectorStorage) CreateConnector(connector *shared_types.GithubConnector) error {
	tx, err := s.DB.BeginTx(s.Ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.NewInsert().Model(connector).Exec(s.Ctx)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *GithubConnectorStorage) UpdateConnector(ConnectorID, InstallationID string) error {
	tx, err := s.DB.BeginTx(s.Ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var connector shared_types.GithubConnector
	_, err = tx.NewUpdate().Model(&connector).
		SetColumn("installation_id", InstallationID).
		Where("id = ?", ConnectorID).Exec(s.Ctx)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *GithubConnectorStorage) GetConnector(ConnectorID string) (*shared_types.GithubConnector, error) {
	var connector shared_types.GithubConnector
	err := s.DB.NewSelect().Model(&connector).Where("id = ?", ConnectorID).Scan(s.Ctx)
	return &connector, err
}

func (s *GithubConnectorStorage) GetAllConnectors(UserID string) ([]shared_types.GithubConnector, error) {
	var connectors []shared_types.GithubConnector
	err := s.DB.NewSelect().Model(&connectors).Where("user_id = ?", UserID).Scan(s.Ctx)
	return connectors, err
}

func (s *GithubConnectorStorage) GetConnectorByAppID(AppID string) (*shared_types.GithubConnector, error) {
	var connector shared_types.GithubConnector
	err := s.DB.NewSelect().Model(&connector).Where("app_id = ?", AppID).Scan(s.Ctx)
	return &connector, err
}
