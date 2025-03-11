package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

// Migration represents a database migration with an ID and SQL content.
type Migration struct {
	ID   int64
	Name string
	Up   string
	Down string
}

// MigrationTable represents the database table that tracks applied migrations.
type MigrationTable struct {
	bun.BaseModel `bun:"table:migrations"`

	ID        int64     `bun:"id,pk,autoincrement"`
	Name      string    `bun:"name,notnull"`
	AppliedAt time.Time `bun:"applied_at,notnull"`
}

// Migrator handles database migrations using the bun.DB connection.
type Migrator struct {
	db         *bun.DB
	migrations []*Migration
	ctx        context.Context
}

// NewMigrator creates a new Migrator instance.
func NewMigrator(db *bun.DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: make([]*Migration, 0),
		ctx:        context.Background(),
	}
}

// LoadMigrationsFromFS loads migration files from the file system.
func (m *Migrator) LoadMigrationsFromFS(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	upFiles := make(map[string]string)
	downFiles := make(map[string]string)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		parts := strings.Split(name, "_")
		if len(parts) < 2 {
			continue
		}

		var baseName string
		var isUp bool

		if strings.HasSuffix(name, "_up.sql") {
			baseName = strings.TrimSuffix(name, "_up.sql")
			isUp = true
		} else if strings.HasSuffix(name, "_down.sql") {
			baseName = strings.TrimSuffix(name, "_down.sql")
			isUp = false
		} else {
			continue
		}

		content, err := os.ReadFile(fmt.Sprintf("%s/%s", path, name))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", name, err)
		}

		if isUp {
			upFiles[baseName] = string(content)
		} else {
			downFiles[baseName] = string(content)
		}
	}

	for baseName, upSQL := range upFiles {
		downSQL := downFiles[baseName]

		idParts := strings.Split(baseName, "_")
		if len(idParts) < 1 {
			continue
		}

		var id int64
		_, err := fmt.Sscanf(idParts[0], "%d", &id)
		if err != nil {
			return fmt.Errorf("invalid migration ID format in %s: %w", baseName, err)
		}

		m.migrations = append(m.migrations, &Migration{
			ID:   id,
			Name: baseName,
			Up:   upSQL,
			Down: downSQL,
		})
	}

	// Sort migrations by ID
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].ID < m.migrations[j].ID
	})

	return nil
}

// InitMigrationTable creates the migrations table if it doesn't exist.
func (m *Migrator) InitMigrationTable() error {
	_, err := m.db.NewCreateTable().
		Model((*MigrationTable)(nil)).
		IfNotExists().
		Exec(m.ctx)

	return err
}

// GetAppliedMigrations returns a list of all applied migrations.
func (m *Migrator) GetAppliedMigrations() ([]MigrationTable, error) {
	var migrations []MigrationTable
	err := m.db.NewSelect().
		Model(&migrations).
		Order("id ASC").
		Scan(m.ctx)

	return migrations, err
}

// MigrateUp applies all pending migrations.
func (m *Migrator) MigrateUp() error {
	err := m.InitMigrationTable()
	if err != nil {
		return fmt.Errorf("failed to initialize migration table: %w", err)
	}

	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedMap := make(map[string]bool)
	for _, migration := range applied {
		appliedMap[migration.Name] = true
	}

	tx, err := m.db.BeginTx(m.ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	for _, migration := range m.migrations {
		if appliedMap[migration.Name] {
			continue
		}

		log.Printf("Applying migration: %s", migration.Name)

		_, err = tx.ExecContext(m.ctx, migration.Up)
		if err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Name, err)
		}

		_, err = tx.NewInsert().
			Model(&MigrationTable{
				Name:      migration.Name,
				AppliedAt: time.Now(),
			}).
			Exec(m.ctx)

		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.Name, err)
		}
	}

	return tx.Commit()
}

// MigrateDown rolls back the most recent migration.
func (m *Migrator) MigrateDown() error {
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		return errors.New("no migrations to roll back")
	}

	lastApplied := applied[len(applied)-1]

	var migrationToRollback *Migration
	for _, migration := range m.migrations {
		if migration.Name == lastApplied.Name {
			migrationToRollback = migration
			break
		}
	}

	if migrationToRollback == nil {
		return fmt.Errorf("could not find migration file for %s", lastApplied.Name)
	}

	if migrationToRollback.Down == "" {
		return fmt.Errorf("no down migration available for %s", lastApplied.Name)
	}

	tx, err := m.db.BeginTx(m.ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	log.Printf("Rolling back migration: %s", migrationToRollback.Name)

	_, err = tx.ExecContext(m.ctx, migrationToRollback.Down)
	if err != nil {
		return fmt.Errorf("failed to roll back migration %s: %w", migrationToRollback.Name, err)
	}

	_, err = tx.NewDelete().
		Model((*MigrationTable)(nil)).
		Where("name = ?", migrationToRollback.Name).
		Exec(m.ctx)

	if err != nil {
		return fmt.Errorf("failed to remove migration record %s: %w", migrationToRollback.Name, err)
	}

	return tx.Commit()
}

func (m *Migrator) MigrateTo(targetVersion string) error {
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedMap := make(map[string]bool)
	for _, migration := range applied {
		appliedMap[migration.Name] = true
	}

	targetFound := false
	for _, migration := range m.migrations {
		if migration.Name == targetVersion {
			targetFound = true
			break
		}
	}

	if !targetFound {
		return fmt.Errorf("target migration %s not found", targetVersion)
	}

	needsUp := true
	for _, migration := range applied {
		if migration.Name == targetVersion {
			needsUp = false
			break
		}
	}

	if needsUp {
		tx, err := m.db.BeginTx(m.ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}
		defer tx.Rollback()

		for _, migration := range m.migrations {
			if migration.Name == targetVersion {
				if !appliedMap[migration.Name] {
					log.Printf("Applying migration: %s", migration.Name)
					_, err = tx.ExecContext(m.ctx, migration.Up)
					if err != nil {
						return fmt.Errorf("failed to apply migration %s: %w", migration.Name, err)
					}

					_, err = tx.NewInsert().
						Model(&MigrationTable{
							Name:      migration.Name,
							AppliedAt: time.Now(),
						}).
						Exec(m.ctx)

					if err != nil {
						return fmt.Errorf("failed to record migration %s: %w", migration.Name, err)
					}
				}
				break
			}

			if !appliedMap[migration.Name] {
				log.Printf("Applying migration: %s", migration.Name)
				_, err = tx.ExecContext(m.ctx, migration.Up)
				if err != nil {
					return fmt.Errorf("failed to apply migration %s: %w", migration.Name, err)
				}

				_, err = tx.NewInsert().
					Model(&MigrationTable{
						Name:      migration.Name,
						AppliedAt: time.Now(),
					}).
					Exec(m.ctx)

				if err != nil {
					return fmt.Errorf("failed to record migration %s: %w", migration.Name, err)
				}
			}
		}

		return tx.Commit()
	} else {
		for i := len(applied) - 1; i >= 0; i-- {
			appliedMigration := applied[i]

			if appliedMigration.Name == targetVersion {
				break
			}

			var migrationToRollback *Migration
			for _, migration := range m.migrations {
				if migration.Name == appliedMigration.Name {
					migrationToRollback = migration
					break
				}
			}

			if migrationToRollback == nil {
				return fmt.Errorf("could not find migration file for %s", appliedMigration.Name)
			}

			if migrationToRollback.Down == "" {
				return fmt.Errorf("no down migration available for %s", appliedMigration.Name)
			}

			tx, err := m.db.BeginTx(m.ctx, nil)
			if err != nil {
				return fmt.Errorf("failed to start transaction: %w", err)
			}

			log.Printf("Rolling back migration: %s", migrationToRollback.Name)

			_, err = tx.ExecContext(m.ctx, migrationToRollback.Down)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to roll back migration %s: %w", migrationToRollback.Name, err)
			}

			_, err = tx.NewDelete().
				Model((*MigrationTable)(nil)).
				Where("name = ?", migrationToRollback.Name).
				Exec(m.ctx)

			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to remove migration record %s: %w", migrationToRollback.Name, err)
			}

			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("failed to commit transaction: %w", err)
			}
		}

		return nil
	}
}

// RunMigrations is a helper function to initialize the migrator and run migrations
func RunMigrations(db *bun.DB, migrationsPath string) error {
	migrator := NewMigrator(db)

	err := migrator.LoadMigrationsFromFS(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	err = migrator.MigrateUp()
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
