package storage

import (
	"context"
	"fmt"
	"log"

	telemetrytypes "github.com/nixopus/nixopus/api/internal/features/telemetry/types"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

// ExtensionTemplateLoader syncs extension YAML from disk into the database and serves
// extension rows for deploy and other callers.
type ExtensionTemplateLoader interface {
	LoadExtensionsFromTemplates(ctx context.Context) error
	GetExtensionByID(ctx context.Context, extensionID string) (*types.Extension, error)
}

type Store struct {
	DB              *bun.DB
	ExtensionLoader ExtensionTemplateLoader
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
		DB: db,
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
	s.DB.RegisterModel((*telemetrytypes.CliInstallation)(nil))
	if err := s.CreateTable(ctx, (*telemetrytypes.CliInstallation)(nil)); err != nil {
		return fmt.Errorf("telemetry cli_installations: %w", err)
	}

	s.DB.RegisterModel((*types.OrganizationUsers)(nil))
	s.DB.RegisterModel((*types.ComposeService)(nil))
	s.DB.RegisterModel((*types.Extension)(nil))
	s.DB.RegisterModel((*types.ExtensionVariable)(nil))

	if s.ExtensionLoader != nil {
		if err := s.ExtensionLoader.LoadExtensionsFromTemplates(ctx); err != nil {
			log.Printf("Warning: Failed to load extensions from templates: %v", err)
		} else {
			log.Println("Extensions loaded successfully from templates")
		}
	}

	return nil
}

func (s *Store) DropAllTables(ctx context.Context) error {
	models := []interface{}{
		(*telemetrytypes.CliInstallation)(nil),
		(*types.ApplicationLogs)(nil),
		(*types.ApplicationDeploymentStatus)(nil),
		(*types.ApplicationDeployment)(nil),
		(*types.ApplicationDomain)(nil),
		(*types.ComposeService)(nil),
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
