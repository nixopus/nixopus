package storage

import (
	"context"
	"fmt"

	"github.com/raghavyuva/nixopus-api/internal/features/organization/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type Store struct {
	DB           *bun.DB
	Organization storage.OrganizationRepository
}

type App struct {
	Config *types.Config
	Store  *Store
	Ctx    context.Context
}

func NewApp(config *types.Config, store *Store, ctx context.Context) *App {
	return &App{Config: config, Store: store, Ctx: ctx}
}

func NewStore(db *bun.DB) *Store {
	return &Store{
		DB:           db,
		Organization: &storage.OrganizationStore{DB: db, Ctx: context.Background()},
	}
}

func (s *Store) CreateTable(ctx context.Context, model interface{}) error {
	_, err := s.DB.NewCreateTable().Model(model).IfNotExists().Exec(ctx)
	return err
}

func (s *Store) DropTable(ctx context.Context, model interface{}) error {
	_, err := s.DB.NewDropTable().Model(model).IfExists().Exec(ctx)
	return err
}

func (s *Store) Init(ctx context.Context) error {
	s.DB.RegisterModel((*types.OrganizationUsers)(nil))

	// tables := []interface{}{
	// 	(*types.User)(nil),
	// 	(*types.Role)(nil),
	// 	(*types.RefreshToken)(nil),
	// 	(*types.Permission)(nil),
	// 	(*types.RolePermissions)(nil),
	// 	(*types.Organization)(nil),
	// 	(*types.OrganizationUsers)(nil),
	// 	(*types.SMTPConfigs)(nil),
	// 	(*types.NotificationPreferences)(nil),
	// 	(*types.PreferenceItem)(nil),
	// 	(*types.Domain)(nil),
	// 	(*types.Server)(nil),
	// 	(*types.GithubConnector)(nil),
	// 	(*types.Application)(nil),
	// 	(*types.ApplicationStatus)(nil),
	// 	(*types.ApplicationLogs)(nil),
	// }

	// for _, model := range tables {
	// 	if err := s.CreateTable(ctx, model); err != nil {
	// 		return fmt.Errorf("failed to create table for %T: %w", model, err)
	// 	}
	// }

	return nil
}

func (s *Store) DropAllTables(ctx context.Context) error {
	models := []interface{}{
		(*types.ApplicationLogs)(nil),
		(*types.ApplicationDeploymentStatus)(nil),
		(*types.ApplicationDeployment)(nil),
		(*types.ApplicationStatus)(nil),
		(*types.Application)(nil),
		(*types.GithubConnector)(nil),
		(*types.Domain)(nil),
		(*types.PreferenceItem)(nil),
		(*types.NotificationPreferences)(nil),
		(*types.SMTPConfigs)(nil),
		(*types.OrganizationUsers)(nil),
		(*types.Organization)(nil),
		(*types.RefreshToken)(nil),
		&struct {
			bun.BaseModel `bun:"table:verification_tokens"`
		}{},
		(*types.User)(nil),
	}

	for _, model := range models {
		if err := s.DropTable(ctx, model); err != nil {
			return fmt.Errorf("dropping table for %T: %w", model, err)
		}
	}

	return nil
}

func (s *Store) TableExists(ctx context.Context, tableName string) (bool, error) {
	exists, err := s.DB.NewSelect().
		Table("information_schema.tables").
		Where("table_name = ?", tableName).
		Exists(ctx)
	return exists, err
}
