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

type GithubConnectorRepository interface {
	CreateConnector(connector *shared_types.GithubConnector) error
	UpdateConnector(ConnectorID, InstallationID string) error
	DeleteConnector(ConnectorID string, UserID string) error
	GetConnector(ConnectorID string) (*shared_types.GithubConnector, error)
	GetAllConnectors(UserID string) ([]shared_types.GithubConnector, error)
	GetConnectorByAppID(AppID string) (*shared_types.GithubConnector, error)
}

// CreateConnector creates a new GitHub connector for the given user.
//
// The connector is created with the InstallationID set to an empty string,
// indicating that it is not associated with any installation yet.
//
// The connector is also created with the CreatedAt, UpdatedAt, and DeletedAt
// fields set to the current time, indicating that the connector is newly
// created.
//
// If the connector cannot be created, an error is returned.
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

// UpdateConnector updates a GitHub connector with the given InstallationID.
//
// The method starts a transaction, and if any errors occur during the update
// process, the transaction is rolled back and the error is returned.
//
// If the connector cannot be updated, an error is returned.
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

// GetConnector retrieves a GitHub connector by its unique identifier.
//
// The method queries the storage to find a GitHub connector associated with
// the provided ConnectorID. If found, it returns the connector object; otherwise,
// it returns an error.
//
// Parameters:
//
//	ConnectorID - the unique identifier of the GitHub connector to retrieve.
//
// Returns:
//
//	*shared_types.GithubConnector - a pointer to the GitHub connector object if found.
//	error - an error if the connector cannot be retrieved or does not exist.
func (s *GithubConnectorStorage) GetConnector(ConnectorID string) (*shared_types.GithubConnector, error) {
	var connector shared_types.GithubConnector
	err := s.DB.NewSelect().Model(&connector).Where("id = ? AND deleted_at IS NULL", ConnectorID).Scan(s.Ctx)
	return &connector, err
}

// GetAllConnectors retrieves all GitHub connectors associated with the provided UserID.
//
// The method queries the storage to find all GitHub connectors associated with
// the provided UserID. If found, it returns a slice of connector objects;
// otherwise, it returns an error.
//
// Parameters:
//
//	UserID - the unique identifier of the user whose connectors to retrieve.
//
// Returns:
//
//	[]shared_types.GithubConnector - a slice of GitHub connector objects if found.
//	error - an error if the connectors cannot be retrieved or do not exist.
func (s *GithubConnectorStorage) GetAllConnectors(UserID string) ([]shared_types.GithubConnector, error) {
	var connectors []shared_types.GithubConnector
	err := s.DB.NewSelect().Model(&connectors).Where("user_id = ? AND deleted_at IS NULL", UserID).Scan(s.Ctx)
	return connectors, err
}

// DeleteConnector performs a soft delete on a GitHub connector.
//
// The method sets the DeletedAt field to the current time, effectively
// marking the connector as deleted without removing it from the database.
// It also verifies that the connector belongs to the provided UserID.
//
// Parameters:
//
//	ConnectorID - the unique identifier of the connector to delete.
//	UserID - the unique identifier of the user who owns the connector.
//
// Returns:
//
//	error - an error if the connector cannot be deleted or does not exist.
func (s *GithubConnectorStorage) DeleteConnector(ConnectorID string, UserID string) error {
	tx, err := s.DB.BeginTx(s.Ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var connector shared_types.GithubConnector
	_, err = tx.NewUpdate().Model(&connector).
		Set("deleted_at = NOW()").
		Set("updated_at = NOW()").
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", ConnectorID, UserID).
		Exec(s.Ctx)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// GetConnectorByAppID retrieves a GitHub connector by its GitHub app ID.
//
// The method queries the storage to find a GitHub connector associated with
// the provided AppID. If found, it returns the connector object; otherwise,
// it returns an error.
//
// Parameters:
//
//	AppID - the GitHub app ID of the GitHub connector to retrieve.
//
// Returns:
//
//	*shared_types.GithubConnector - a pointer to the GitHub connector object if found.
//	error - an error if the connector cannot be retrieved or does not exist.
func (s *GithubConnectorStorage) GetConnectorByAppID(AppID string) (*shared_types.GithubConnector, error) {
	var connector shared_types.GithubConnector
	err := s.DB.NewSelect().Model(&connector).Where("app_id = ?", AppID).Scan(s.Ctx)
	return &connector, err
}
