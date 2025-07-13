# Migration Quick Reference

A quick reference guide for common migration operations in Nixopus.

## Quick Start

### Create a New Migration

1. **Choose the module** (auth, users, organizations, etc.)
2. **Find the next version number**:
   ```bash
   ls api/migrations/auth/ | grep "_up.sql" | tail -1
   # Example output: 002_create_refresh_token_up.sql
   # Next version: 003
   ```
3. **Create migration files**:

   ```bash
   # Create up migration
   touch api/migrations/auth/003_add_user_roles_up.sql

   # Create down migration
   touch api/migrations/auth/003_add_user_roles_down.sql
   ```

### Migration File Templates

#### Basic Table Creation

**Up Migration**:

```sql
CREATE TABLE table_name (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_table_name_name ON table_name(name);
```

**Down Migration**:

```sql
DROP TABLE IF EXISTS table_name;
```

#### Add Column

**Up Migration**:

```sql
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
CREATE INDEX idx_users_phone ON users(phone);
```

**Down Migration**:

```sql
DROP INDEX IF EXISTS idx_users_phone;
ALTER TABLE users DROP COLUMN IF EXISTS phone;
```

#### Add Foreign Key

**Up Migration**:

```sql
ALTER TABLE user_profiles
    ADD CONSTRAINT fk_user_profiles_user_id
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
```

**Down Migration**:

```sql
ALTER TABLE user_profiles DROP CONSTRAINT IF EXISTS fk_user_profiles_user_id;
```

## Module Guide

### Auth Module (`api/migrations/auth/`)

- User authentication tables
- Session management
- Token storage
- Password reset functionality

### Users Module (`api/migrations/users/`)

- User profile information
- User preferences
- User settings

### Organizations Module (`api/migrations/organizations/`)

- Organization structure
- Membership management
- Organization settings

### RBAC Module (`api/migrations/rbac/`)

- Role definitions
- Permission management
- Access control

### Applications Module (`api/migrations/applications/`)

- Application configurations
- Deployment settings
- Application metadata

### Containers Module (`api/migrations/containers/`)

- Container management
- Container configurations
- Runtime settings

## Common Operations

### Check Migration Status

```bash
# Migrations run automatically on server start
cd api && air
```

### Manual Migration Testing

```go
// Load and apply migrations
migrator := storage.NewMigrator(db)
err := migrator.LoadMigrationsFromFS("./migrations")
if err != nil {
    log.Fatal(err)
}
err = migrator.MigrateUp()
if err != nil {
    log.Fatal(err)
}
```

### Rollback Last Migration (Development)

```go
migrator := storage.NewMigrator(db)
migrator.LoadMigrationsFromFS("./migrations")
migrator.MigrateDown()
```

## Naming Conventions

### File Naming

```
{sequence_no}_{description}_{direction}.sql

Examples:
- 001_create_users_up.sql
- 001_create_users_down.sql
- 015_add_user_avatar_up.sql
- 015_add_user_avatar_down.sql
```

### Version Numbers

- Use zero-padded 3-digit numbers: `001`, `002`, `003`
- Increment sequentially

### Descriptions

- Use lowercase with underscores
- Be descriptive but concise
- Use verbs: `create`, `add`, `modify`, `remove`

## Common Patterns

### UUID Primary Keys

```sql
id UUID PRIMARY KEY DEFAULT uuid_generate_v4()
```

### Timestamps

```sql
created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
```

### Soft Deletes

```sql
deleted_at TIMESTAMP WITH TIME ZONE
```

### Foreign Key with Cascade

```sql
user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
```

### Indexes for Performance

```sql
CREATE INDEX idx_table_column ON table_name(column_name);
CREATE INDEX idx_table_created_at ON table_name(created_at);
```
