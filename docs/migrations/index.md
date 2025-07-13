# Database Migrations

Nixopus uses a migration system built with Go and PostgreSQL to manage database schema changes using [Bun ORM](https://bun.uptrace.dev/guide/migrations.html). This document provides a comprehensive outline and guide on how migrations work in the project, how to create new migrations, and how to manage the migration lifecycle.

## Overview

The migration system in Nixopus is designed to:

- Automatically apply schema changes when the server starts
- Support both forward (up) and backward (down) migrations
- Organize migrations by feature modules
- Provide transaction safety for all migration operations
- Track applied migrations in the database

## Migration Architecture

### Core Components

The migration system consists of several key components:

- **Migrator**: The main [migration engine](https://github.com/raghavyuva/nixopus/blob/master/api/internal/storage/migration.go) (`api/internal/storage/migration.go`)
- **Migration Files**: SQL files organized by [feature modules](https://github.com/raghavyuva/nixopus/tree/master/api/migrations)
- **Migration Table**: A database table that tracks applied migrations
- **Configuration**: Automatic migration execution during server startup

### Directory Structure

```text
api/migrations/
├── applications/         # Application-related schema
├── audit/                # Audit logging schema
├── auth/                 # Authentication and user management
├── containers/           # Container management schema
├── dashboard/            # Dashboard and analytics
├── domains/              # Domain management
├── feature-flags/        # Feature flag system
├── integrations/         # External integrations
├── notifications/        # Notification system
├── organizations/        # Organization management
├── rbac/                 # Role-based access control
├── terminal/             # Terminal and shell access
└── users/                # User profile management
```

### Migration File Naming Convention

Migration files follow a strict naming convention:

```text
{sequence_no}_{descriptive_name}_{up|down}.sql
```

Examples:

- `001_create_users_up.sql`
- `001_create_users_down.sql`
- `002_create_refresh_token_up.sql`
- `002_create_refresh_token_down.sql`

## How Migrations Work

### Automatic Execution

Migrations are automatically executed when the Nixopus server starts. The process follows these steps:

1. **Server Initialization**: During server startup, the migration system is initialized
2. **Migration Discovery**: The system scans the `./migrations` directory recursively
3. **Migration Loading**: All migration files are loaded and parsed
4. **State Check**: The system checks which migrations have already been applied
5. **Migration Execution**: Pending migrations are applied in order
6. **Transaction Safety**: Each migration runs within a database transaction

![image](https://media2.dev.to/dynamic/image/width=800%2Cheight=%2Cfit=scale-down%2Cgravity=auto%2Cformat=auto/https%3A%2F%2Fdev-to-uploads.s3.amazonaws.com%2Fuploads%2Farticles%2Fnu808jb4jnzs8vl5c5ko.png)

### Migration Tracking

Bun ORM (the Go ORM you're using) does support migration tracking. It provides migration tooling but leaves tracking and management up to you.

The system maintains a `migrations` table in the database to track applied migrations:

```sql
CREATE TABLE migrations (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    applied_at TIMESTAMP WITH TIME ZONE NOT NULL
);
```

![db-schema](https://raw.githubusercontent.com/shravan20/nixopus/refs/heads/master/assets/db-schema.png)

### Transaction Safety

Each migration operation is wrapped in a database transaction to ensure atomicity:

- If a migration fails, the transaction is rolled back
- The migration table is only updated after successful execution
- Multiple migrations can be applied in a single transaction for performance

## Creating New Migrations

### Step 1: Choose the Appropriate Module

Determine which feature module your migration belongs to:

- **auth**: User authentication, sessions, tokens
- **users**: User profiles, preferences
- **organizations**: Organization structure, membership
- **rbac**: Roles, permissions, access control
- **applications**: Application definitions, configurations
- **containers**: Container management, deployments
- **terminal**: Terminal access, shell configurations
- **notifications**: Alert systems, messaging
- **integrations**: External service connections
- **dashboard**: Analytics, reporting
- **domains**: DNS, domain management
- **feature-flags**: Feature toggle system
- **audit**: System logging, audit trails

### Step 2: Determine Version Number

Find the highest version number and increment by 1:

```bash
# Check existing migrations in the auth module
ls api/migrations/auth/
# Output: 001_create_users_up.sql, 001_create_users_down.sql, ...
# Next version would be 002
```

### Step 3: Create Migration Files

Create both up and down migration files:

**Up Migration Example** (`003_add_user_preferences_up.sql`):

```sql
-- Add user preferences table
CREATE TABLE user_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    theme VARCHAR(50) DEFAULT 'light',
    language VARCHAR(10) DEFAULT 'en',
    timezone VARCHAR(100) DEFAULT 'UTC',
    notifications_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for performance
CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);
CREATE UNIQUE INDEX idx_user_preferences_unique_user ON user_preferences(user_id);

-- Add trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_user_preferences_updated_at
    BEFORE UPDATE ON user_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

**Down Migration Example** (`003_add_user_preferences_down.sql`):

```sql
-- Remove user preferences table and related objects
DROP TRIGGER IF EXISTS update_user_preferences_updated_at ON user_preferences;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS user_preferences;
```

### Step 4: Test Your Migration

Before committing, test your migration:

1. **Apply the migration**: Start the server to apply new migrations
2. **Verify schema**: Check that tables and indexes are created correctly
3. **Test rollback**: Use the migration tools to test the down migration
4. **Data integrity**: Ensure existing data is preserved

## Migration Operations

### Running Migrations

Migrations run automatically when the server starts, but you can also run them manually for testing:

```bash
# The server automatically runs migrations on startup
cd api
air # development environment
```

### Rolling Back Migrations

To roll back the most recent migration:

```go
// This functionality is available in the codebase
// but typically used for development/testing
migrator := storage.NewMigrator(db)
migrator.LoadMigrationsFromFS("./migrations")
migrator.MigrateDown()
```

### Checking Migration Status

The migration system provides several utility functions:

```go
// Get all applied migrations
applied, err := migrator.GetAppliedMigrations()

// Check if specific migration is applied
for _, migration := range applied {
    if migration.Name == "001_create_users" {
        fmt.Println("Users table migration is applied")
    }
}
```

### Resetting Migrations (Development Only)

For development purposes, you can reset the migration state:

```go
// WARNING: This drops all tables - development only
storage.MigrateDownAll(db, "./migrations")

// Or reset just the migration tracking
storage.ResetMigrations(db)
```

## Standard Design Practices to be followed

### Migration Design

1. **Atomic Changes**: Each migration should represent a single, complete change
2. **Backward Compatibility**: Design migrations to be backward compatible when possible
3. **Data Preservation**: Always consider existing data when modifying schemas
4. **Dependencies**: Ensure migrations can run independently

### Naming Conventions

1. **Descriptive Names**: Use clear, descriptive names for migrations
2. **Consistent Format**: Follow the established naming pattern
3. **Version Ordering**: Use zero-padded numbers for proper ordering
4. **Module Organization**: Keep related migrations in the same module

### SQL Best Practices

1. **Use Transactions**: Wrap complex operations in explicit transactions
2. **Add Indexes**: Include necessary indexes for performance
3. **Constraints**: Define proper foreign keys and constraints
4. **Comments**: Add comments to explain complex operations
5. **Extensions**: Check for required PostgreSQL extensions

### Testing Guidelines

1. **Local Testing**: Always test migrations locally before committing
2. **Rollback Testing**: Verify that down migrations work correctly
3. **Data Integrity**: Ensure data is preserved during migrations

## Common Patterns

### Creating Tables

```sql
-- Standard table creation pattern
CREATE TABLE example_table (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Add indexes
CREATE INDEX idx_example_table_name ON example_table(name);
CREATE INDEX idx_example_table_created_at ON example_table(created_at);
```

### Adding Columns

```sql
-- Add new column with default value
ALTER TABLE users ADD COLUMN status VARCHAR(50) DEFAULT 'active';

-- Add constraint
ALTER TABLE users ADD CONSTRAINT chk_user_status
    CHECK (status IN ('active', 'inactive', 'suspended'));
```

### Modifying Columns

```sql
-- Change column type
ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(320);

-- Add NOT NULL constraint
ALTER TABLE users ALTER COLUMN email SET NOT NULL;

-- Add default value
ALTER TABLE users ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP;
```

### Foreign Key Relationships

```sql
-- Add foreign key constraint
ALTER TABLE user_preferences
    ADD CONSTRAINT fk_user_preferences_user_id
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
```

This migration system provides a robust foundation for managing database schema changes in Nixopus while maintaining data integrity and system reliability.
