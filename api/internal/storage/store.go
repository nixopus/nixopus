package storage

import (
	"context"
	"fmt"

	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/uptrace/bun"
)

type Store struct {
	DB *bun.DB
}

type App struct {
	Config *types.Config
	Store  *Store
	Ctx    context.Context
}

func NewApp(config *types.Config, store *Store, ctx context.Context) *App {
	return &App{
		Config: config,
		Store:  store,
		Ctx:    ctx,
	}
}

func NewStore(db *bun.DB) *Store {
	return &Store{
		DB: db,
	}
}

func (s *Store) CreateUserTable(ctx context.Context) error {
	_, err := s.DB.NewCreateTable().Model((*types.User)(nil)).IfNotExists().Exec(ctx)
	return err
}

func (s *Store) DropUserTable(ctx context.Context) error {
	_, err := s.DB.NewDropTable().Model((*types.User)(nil)).IfExists().Exec(ctx)
	return err
}

func (s *Store) CreateRefreshTokenTable(ctx context.Context) error {
	_, err := s.DB.NewCreateTable().Model((*types.RefreshToken)(nil)).IfNotExists().Exec(ctx)
	return err
}

func (s *Store) DropRefreshTokenTable(ctx context.Context) error {
	_, err := s.DB.NewDropTable().Model((*types.RefreshToken)(nil)).IfExists().Exec(ctx)
	return err
}

// Init initializes the store by ensuring the user table exists in the database.
// It checks if the user table exists, and if not, it creates the user table.
// Returns an error if the table existence check or table creation fails.
func (s *Store) Init(ctx context.Context) error {
	is_user_table_exist, err := s.TableExists(ctx, "users")
	if err != nil {
		return fmt.Errorf("failed to check if user table exists: %w", err)
	}

	if !is_user_table_exist {
		if err := s.CreateUserTable(ctx); err != nil {
			return fmt.Errorf("failed to create user table: %w", err)
		}
	}

	is_refresh_token_table_exist, err := s.TableExists(ctx, "refresh_tokens")
	if err != nil {
		return fmt.Errorf("failed to check if refresh token table exists: %w", err)
	}

	if !is_refresh_token_table_exist {
		if err := s.CreateRefreshTokenTable(ctx); err != nil {
			return fmt.Errorf("failed to create refresh token table: %w", err)
		}
	}

	return nil
}

func (s *Store) TableExists(ctx context.Context, model interface{}) (bool, error) {
	exists, err := s.DB.NewSelect().
		Table("information_schema.tables").
		Where("table_name = ?", model).
		Exists(ctx)
	return exists, err
}
