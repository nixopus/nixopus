# Migration Templates

This file contains templates for common migration patterns in Nixopus.

## Basic Table Creation

### Up Migration Template

```sql
-- Description: Create [table_name] table for [purpose]
-- Seq Number: [XXX]
-- Module: [module_name]

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE [table_name] (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Add indexes
CREATE INDEX idx_[table_name]_name ON [table_name](name);
CREATE INDEX idx_[table_name]_created_at ON [table_name](created_at);

-- Add any constraints
-- ALTER TABLE [table_name] ADD CONSTRAINT chk_[table_name]_status
--     CHECK (status IN ('active', 'inactive'));
```

### Down Migration Template

```sql
-- Rollback: Drop [table_name] table
-- Seq Number: [XXX]
-- Module: [module_name]

DROP TABLE IF EXISTS [table_name];
```

## Table with Foreign Key

### Up Migration Template

```sql
-- Description: Create [table_name] table with foreign key to [parent_table]
-- Seq Number: [XXX]
-- Module: [module_name]

CREATE TABLE [table_name] (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    [parent_table]_id UUID NOT NULL REFERENCES [parent_table](id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    value TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes
CREATE INDEX idx_[table_name]_[parent_table]_id ON [table_name]([parent_table]_id);
CREATE INDEX idx_[table_name]_name ON [table_name](name);

-- Add unique constraint if needed
-- CREATE UNIQUE INDEX idx_[table_name]_unique_[parent_table]_name
--     ON [table_name]([parent_table]_id, name);
```

### Down Migration Template

```sql
-- Rollback: Drop [table_name] table
-- Seq Number: [XXX]
-- Module: [module_name]

DROP TABLE IF EXISTS [table_name];
```

## Add Column to Existing Table

### Up Migration Template

```sql
-- Description: Add [column_name] column to [table_name]
-- Seq Number: [XXX]
-- Module: [module_name]

-- Add the new column
ALTER TABLE [table_name] ADD COLUMN [column_name] VARCHAR(255);

-- Add default value if needed
-- ALTER TABLE [table_name] ALTER COLUMN [column_name] SET DEFAULT 'default_value';

-- Add constraint if needed
-- ALTER TABLE [table_name] ADD CONSTRAINT chk_[table_name]_[column_name]
--     CHECK ([column_name] IN ('value1', 'value2', 'value3'));

-- Add index if needed
-- CREATE INDEX idx_[table_name]_[column_name] ON [table_name]([column_name]);
```

### Down Migration Template

```sql
-- Rollback: Remove [column_name] column from [table_name]
-- Seq Number: [XXX]
-- Module: [module_name]

-- Drop index if created
-- DROP INDEX IF EXISTS idx_[table_name]_[column_name];

-- Drop constraint if created
-- ALTER TABLE [table_name] DROP CONSTRAINT IF EXISTS chk_[table_name]_[column_name];

-- Drop the column
ALTER TABLE [table_name] DROP COLUMN IF EXISTS [column_name];
```

## Add Foreign Key Constraint

### Up Migration Template

```sql
-- Description: Add foreign key constraint between [table_name] and [referenced_table]
-- Seq Number: [XXX]
-- Module: [module_name]

-- Add foreign key constraint
ALTER TABLE [table_name]
    ADD CONSTRAINT fk_[table_name]_[referenced_table]_id
    FOREIGN KEY ([referenced_table]_id) REFERENCES [referenced_table](id) ON DELETE CASCADE;

-- Add index for performance
CREATE INDEX idx_[table_name]_[referenced_table]_id ON [table_name]([referenced_table]_id);
```

### Down Migration Template

```sql
-- Rollback: Remove foreign key constraint between [table_name] and [referenced_table]
-- Seq Number: [XXX]
-- Module: [module_name]

-- Drop index
DROP INDEX IF EXISTS idx_[table_name]_[referenced_table]_id;

-- Drop foreign key constraint
ALTER TABLE [table_name] DROP CONSTRAINT IF EXISTS fk_[table_name]_[referenced_table]_id;
```

## Create Index

### Up Migration Template

```sql
-- Description: Add index on [table_name].[column_name] for performance
-- Seq Number: [XXX]
-- Module: [module_name]

-- Create index
CREATE INDEX idx_[table_name]_[column_name] ON [table_name]([column_name]);

-- For composite index
-- CREATE INDEX idx_[table_name]_[column1]_[column2] ON [table_name]([column1], [column2]);

-- For unique index
-- CREATE UNIQUE INDEX idx_[table_name]_unique_[column_name] ON [table_name]([column_name]);
```

### Down Migration Template

```sql
-- Rollback: Remove index on [table_name].[column_name]
-- Seq Number: [XXX]
-- Module: [module_name]

DROP INDEX IF EXISTS idx_[table_name]_[column_name];
```

## Modify Column Type

### Up Migration Template

```sql
-- Description: Change [column_name] type in [table_name] from [old_type] to [new_type]
-- Seq Number: [XXX]
-- Module: [module_name]

-- Modify column type
ALTER TABLE [table_name] ALTER COLUMN [column_name] TYPE [new_type];

-- If changing to NOT NULL
-- ALTER TABLE [table_name] ALTER COLUMN [column_name] SET NOT NULL;

-- If adding default value
-- ALTER TABLE [table_name] ALTER COLUMN [column_name] SET DEFAULT 'default_value';
```

### Down Migration Template

```sql
-- Rollback: Change [column_name] type in [table_name] back to [old_type]
-- Seq Number: [XXX]
-- Module: [module_name]

-- Remove default if added
-- ALTER TABLE [table_name] ALTER COLUMN [column_name] DROP DEFAULT;

-- Remove NOT NULL if added
-- ALTER TABLE [table_name] ALTER COLUMN [column_name] DROP NOT NULL;

-- Change back to original type
ALTER TABLE [table_name] ALTER COLUMN [column_name] TYPE [old_type];
```

## Data Migration Template

### Up Migration Template

```sql
-- Description: Migrate data for [specific_purpose]
-- Seq Number: [XXX]
-- Module: [module_name]

-- Begin transaction for data safety
BEGIN;

-- Update existing data
UPDATE [table_name]
SET [column_name] = [new_value]
WHERE [condition];

-- Insert new data if needed
INSERT INTO [table_name] ([column1], [column2])
VALUES ([value1], [value2]);

-- Verify the changes
-- SELECT COUNT(*) FROM [table_name] WHERE [condition];

COMMIT;
```

### Down Migration Template

```sql
-- Rollback: Reverse data migration for [specific_purpose]
-- Seq Number: [XXX]
-- Module: [module_name]

-- Begin transaction for data safety
BEGIN;

-- Reverse the data changes
UPDATE [table_name]
SET [column_name] = [original_value]
WHERE [condition];

-- Remove inserted data if any
DELETE FROM [table_name] WHERE [condition];

COMMIT;
```

## Usage Instructions

1. **Copy the appropriate template** for your migration type
2. **Replace all placeholders** in square brackets `[placeholder]` with actual values:

   - `[table_name]` - The actual table name
   - `[column_name]` - The actual column name
   - `[XXX]` - The migration version number (e.g., 001, 002)
   - `[module_name]` - The migration module (auth, users, etc.)
   - `[purpose]` - Brief description of what the migration does

3. **Update the file header** with proper version and description
4. **Test the migration** locally before committing
5. **Create both up and down migrations** using the templates

## Naming Examples

### File Names

- `001_create_users_up.sql` / `001_create_users_down.sql`
- `002_add_user_avatar_up.sql` / `002_add_user_avatar_down.sql`
- `003_add_user_preferences_table_up.sql` / `003_add_user_preferences_table_down.sql`

### Table Names

- `users`, `user_preferences`, `user_sessions`
- `organizations`, `organization_members`, `organization_settings`
- `applications`, `application_configs`, `application_deployments`

### Index Names

- `idx_users_email`, `idx_users_created_at`
- `idx_user_preferences_user_id`
- `idx_organizations_name`

### Constraint Names

- `fk_user_preferences_user_id`
- `chk_users_status`
- `uq_users_email`

Remember to always test your migrations thoroughly before applying them to production!
