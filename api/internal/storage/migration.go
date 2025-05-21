package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	upFiles := make(map[string]string)
	downFiles := make(map[string]string)

	// Walk through all directories recursively
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".sql") {
			return nil
		}

		// Get relative path from migrations root
		relPath, err := filepath.Rel(path, filePath)

		if err != nil {
			fmt.Println("Error getting relative path:", relPath)
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filePath, err)
		}

		// Extract base name and determine if it's up or down migration
		fileName := info.Name()
		var baseName string
		var isUp bool

		if strings.HasSuffix(fileName, "_up.sql") {
			baseName = strings.TrimSuffix(fileName, "_up.sql")
			isUp = true
		} else if strings.HasSuffix(fileName, "_down.sql") {
			baseName = strings.TrimSuffix(fileName, "_down.sql")
			isUp = false
		} else {
			return nil
		}

		// Extract migration ID from the base name
		parts := strings.Split(baseName, "_")
		if len(parts) < 1 {
			return fmt.Errorf("invalid migration filename format: %s", fileName)
		}

		if isUp {
			upFiles[baseName] = string(content)
		} else {
			downFiles[baseName] = string(content)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk migrations directory: %w", err)
	}

	// Process migrations
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

		m.migrations = append(m.migrations,
			&Migration{
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

// MigrateUp applies all pending migrations in ID order.
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

	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].ID < m.migrations[j].ID
	})

	tx, err := m.db.BeginTx(m.ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	for _, migration := range m.migrations {
		if appliedMap[migration.Name] {
			continue
		}

		log.Printf("Applying migration: %s (ID: %d)", migration.Name, migration.ID)

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

// MigrateDownAll drops all tables in the database, effectively rolling back all migrations
func MigrateDownAll(db *bun.DB, migrationsPath string) error {
	log.Println("Dropping all tables from the database")

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	var schema string
	err = tx.QueryRow("SELECT current_schema()").Scan(&schema)
	if err != nil {
		return fmt.Errorf("failed to get current schema: %w", err)
	}
	_, err = tx.Exec("SET session_replication_role = 'replica';")
	if err != nil {
		return fmt.Errorf("failed to disable triggers: %w", err)
	}

	rows, err := tx.Query(fmt.Sprintf(`
		SELECT tablename FROM pg_tables 
		WHERE schemaname = '%s' AND 
		tablename != 'schema_migrations' AND 
		tablename != 'goose_db_version' AND 
		tablename != 'schema_version'
	`, schema))
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating over table rows: %w", err)
	}

	if len(tables) == 0 {
		log.Println("No tables to drop")
		return nil
	}

	dropStatement := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", strings.Join(tables, ", "))
	log.Printf("Executing: %s", dropStatement)

	_, err = tx.Exec(dropStatement)
	if err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	_, err = tx.Exec("DROP TABLE IF EXISTS migrations CASCADE")
	if err != nil {
		return fmt.Errorf("failed to drop migrations table: %w", err)
	}

	_, err = tx.Exec("SET session_replication_role = 'origin';")
	if err != nil {
		return fmt.Errorf("failed to re-enable triggers: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Println("All tables dropped successfully")
	return nil
}

// MigrateDown rolls back only the most recent migration
func MigrateDown(db *bun.DB, migrationsPath string) error {
	migrator := NewMigrator(db)

	err := migrator.LoadMigrationsFromFS(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	err = migrator.MigrateDown()
	if err != nil {
		return fmt.Errorf("failed to roll back migration: %w", err)
	}

	return nil
}

// ResetMigrations resets the migration state while keeping database objects
func ResetMigrations(db *bun.DB) error {
	log.Println("Resetting migration state (dropping migrations table)")
	_, err := db.NewDropTable().
		Model((*MigrationTable)(nil)).
		IfExists().
		Exec(context.Background())

	if err != nil {
		return fmt.Errorf("failed to drop migrations table: %w", err)
	}

	log.Println("Migration state reset successfully")
	return nil
}
