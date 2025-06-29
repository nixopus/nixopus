# Development Fixtures Guide

This guide explains how to use and extend the Nixopus development fixtures system.

## Overview

The fixtures system provides sample data for development and testing environments. It includes pre-configured users, organizations, roles, permissions, and feature flags to help you get started quickly.

## Quick Start

```bash
cd api

# Load all fixtures (recommended for new development setup)
make fixtures-load

# Get help on all available commands
make fixtures-help
```

## Available Commands

| Command | Description | Use Case |
|---------|-------------|----------|
| `make fixtures-load` | Load fixtures without affecting existing data | Normal development workflow |
| `make fixtures-recreate` | Drop and recreate all tables, then load fixtures | Clean slate development |
| `make fixtures-clean` | Truncate all tables, then load fixtures | Reset data while keeping schema |
| `make fixtures-help` | Display help information | Get command details |

## Fixture Files Structure

```
api/fixtures/development/
├── complete.yml              # Main entry point (imports all others)
├── users.yml                 # User accounts and profiles
├── organizations.yml         # Organization data
├── roles.yml                 # Role definitions
├── permissions.yml           # Permission definitions
├── role_permissions.yml      # Role-permission mappings
├── feature_flags.yml         # Feature flag configurations
└── organization_users.yml    # User-organization relationships
```

## Fixture File Details

### complete.yml
The main entry point that imports all other fixture files. This is the file used by the default make commands.

### users.yml
Contains sample user accounts with various roles and permissions:
- Admin users with full access
- Regular users with limited permissions
- Test users for different scenarios

### organizations.yml
Sample organizations for testing multi-tenant scenarios:
- Default organization
- Test organizations with different configurations

### roles.yml
Pre-defined roles in the system:
- Super Admin
- Organization Admin
- User
- Read-only roles

### permissions.yml
Comprehensive permission definitions for all system features:
- User management permissions
- Organization management permissions
- Container and deployment permissions
- Feature flag permissions

### role_permissions.yml
Mappings between roles and permissions, defining what each role can do in the system.

### feature_flags.yml
Feature flag configurations for testing different feature states:
- Enabled/disabled features
- Beta features
- Experimental functionality

### organization_users.yml
User-organization relationships for testing multi-tenant scenarios.

## Adding New Fixtures

### 1. Create a New Fixture File

Create a new YAML file in `api/fixtures/development/`:

```yaml
# my_feature.yml
- table: my_feature
  data:
    - id: "550e8400-e29b-41d4-a716-446655440001"
      name: "Sample Feature 1"
      description: "A sample feature for testing"
      created_at: "2024-01-01T00:00:00Z"
      updated_at: "2024-01-01T00:00:00Z"
    - id: "550e8400-e29b-41d4-a716-446655440002"
      name: "Sample Feature 2"
      description: "Another sample feature"
      created_at: "2024-01-01T00:00:00Z"
      updated_at: "2024-01-01T00:00:00Z"
```

### 2. Add to complete.yml

Update `api/fixtures/development/complete.yml` to include your new fixture:

```yaml
- import: users.yml
- import: organizations.yml
- import: roles.yml
- import: permissions.yml
- import: role_permissions.yml
- import: feature_flags.yml
- import: organization_users.yml
- import: my_feature.yml  # Add your new fixture here
```

### 3. Test Your Fixtures

```bash
cd api

# Test loading your new fixtures
make fixtures-load

# Or test with a clean slate
make fixtures-recreate
```

## Fixture File Format

Each fixture file follows this structure:

```yaml
- table: table_name
  data:
    - column1: value1
      column2: value2
      # ... more columns
    - column1: value3
      column2: value4
      # ... more columns
```

### Supported Data Types

- **Strings**: Regular text values
- **UUIDs**: Use string format with UUID values
- **Timestamps**: Use ISO 8601 format (e.g., "2024-01-01T00:00:00Z")
- **Booleans**: true/false
- **Numbers**: Integers and floats
- **JSON**: Use string format for JSON data
- **Null**: null for empty values

### Best Practices

1. **Use UUIDs for IDs**: Generate proper UUIDs for primary keys
2. **Consistent Timestamps**: Use realistic timestamps for created_at/updated_at
3. **Meaningful Data**: Use realistic, meaningful data for testing
4. **Dependencies**: Ensure referenced foreign keys exist in other fixtures
5. **File Organization**: Keep related data in separate files for maintainability

## Troubleshooting

### Common Issues

1. **Foreign Key Violations**: Ensure referenced data exists before loading dependent fixtures
2. **Duplicate Keys**: Check for duplicate primary keys in your fixture data
3. **Invalid Data Types**: Verify data types match your database schema
4. **Missing Tables**: Ensure database migrations have been run before loading fixtures

### Debug Commands

```bash
cd api
# Load specific fixture file only
go run cmd/fixtures/main.go -fixture=fixtures/development/users.yml
```

## Integration with Testing

The fixtures system is also used for integration tests. Test-specific fixtures can be created in separate directories:

```
api/fixtures/
├── development/     # Development environment fixtures
├── test/           # Test-specific fixtures
└── staging/        # Staging environment fixtures
```

## Contributing to Fixtures

When contributing new fixtures:

1. Follow the existing naming conventions
2. Include realistic, diverse test data
3. Document any special requirements or dependencies
4. Test your fixtures with different loading strategies
5. Update this documentation if adding new fixture types

## Need Help?

If you encounter issues with the fixtures system:

1. Check the help command: `make fixtures-help`
2. Review the fixture file syntax and data types
3. Verify your database schema matches the fixture data
4. Create an issue on GitHub with details about the problem 