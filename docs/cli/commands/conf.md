# conf - Configuration Management

The `conf` command provides comprehensive configuration management for Nixopus services. Manage environment variables, service settings, and application configuration across API and view services with support for multiple environments.

## Quick Start
```bash
# List current configuration
nixopus conf list --service api

# Set configuration value
nixopus conf set DATABASE_URL=postgresql://user:pass@localhost:5432/nixopus

# Delete configuration key
nixopus conf delete OLD_CONFIG_KEY

# Set view service configuration
nixopus conf set --service view NODE_ENV=production
```

## Overview

The conf command handles all aspects of Nixopus configuration:
- Environment variable management for services
- Multi-service configuration support (API, view)
- Environment file management (.env files)

## Subcommands

### `list` - Display Configuration

Show all configuration values for specified services with optional filtering and formatting.

```bash
nixopus conf list [OPTIONS]
```

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--service` | `-s` | Target service (api, view) | `api` |
| `--verbose` | `-v` | Show detailed logging and metadata | `false` |
| `--output` | `-o` | Output format (text, json) | `text` |
| `--env-file` | `-e` | Custom environment file path | None |
| `--dry-run` | `-d` | Dry run mode | `false` |
| `--timeout` | `-t` | Operation timeout in seconds | `10` |

**Examples:**

```bash
# List API service configuration
nixopus conf list --service api

# Get JSON output
nixopus conf list --output json

# Use custom environment file
nixopus conf list --env-file /custom/path/.env
```


### `set` - Update Configuration

Set configuration values using KEY=VALUE format with service targeting.

```bash
nixopus conf set KEY=VALUE [OPTIONS]
```

**Arguments:**
- `KEY=VALUE` - Configuration pair (required, single value only)

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--service` | `-s` | Target service (api, view) | `api` |
| `--verbose` | `-v` | Show detailed logging | `false` |
| `--output` | `-o` | Output format (text, json) | `text` |
| `--dry-run` | `-d` | Preview configuration changes | `false` |
| `--env-file` | `-e` | Custom environment file path | None |
| `--timeout` | `-t` | Operation timeout in seconds | `10` |

**Examples:**

```bash
# Set configuration value
nixopus conf set DATABASE_URL=postgresql://user:pass@localhost:5432/nixopus

# Set for view service
nixopus conf set --service view NODE_ENV=production

# Preview changes
nixopus conf set DEBUG=true --dry-run
```

### `delete` - Remove Configuration

Remove configuration keys from service environments with safety checks.

```bash
nixopus conf delete KEY [OPTIONS]
```

**Arguments:**
- `KEY` - Configuration key to remove (required, single key only)

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--service` | `-s` | Target service (api, view) | `api` |
| `--verbose` | `-v` | Show detailed logging | `false` |
| `--output` | `-o` | Output format (text, json) | `text` |
| `--dry-run` | `-d` | Preview deletion without executing | `false` |
| `--env-file` | `-e` | Custom environment file path | None |
| `--timeout` | `-t` | Operation timeout in seconds | `10` |

**Examples:**

```bash
# Delete configuration key
nixopus conf delete OLD_CONFIG_KEY

# Preview deletion
nixopus conf delete TEMP_CONFIG --dry-run
```

## Configuration

The conf command manages environment variables stored in service-specific `.env` files. Configuration is loaded from the built-in `config.prod.yaml` file to determine default environment file locations.

### Default Environment File Locations

| Service | Default Environment File | Configuration Path |
|---------|-------------------------|-------------------|
| API | `/etc/nixopus/source/api/.env` | `services.api.env.API_ENV_FILE` |
| View | `/etc/nixopus/source/view/.env` | `services.view.env.VIEW_ENV_FILE` |

### Configuration Source

Environment file paths are determined from the built-in `config.prod.yaml`:

```yaml
# Built-in configuration
services:
  api:
    env:
      API_ENV_FILE: ${API_ENV_FILE:-/etc/nixopus/source/api/.env}
  view:
    env:
      VIEW_ENV_FILE: ${VIEW_ENV_FILE:-/etc/nixopus/source/view/.env}
```

### Overriding Environment Files

You can specify custom environment files using the `--env-file` option:

```bash
# Use custom environment file
nixopus conf list --env-file /custom/path/.env

# Set configuration in custom file
nixopus conf set DATABASE_URL=custom --env-file /custom/path/.env

# Delete from custom file
nixopus conf delete OLD_KEY --env-file /custom/path/.env
```

### Permission Requirements

Environment files require appropriate read/write permissions:

```bash
# Check current permissions
ls -la /etc/nixopus/source/api/.env

# Fix permissions if needed (may require sudo)
sudo chmod 644 /etc/nixopus/source/api/.env
sudo chown $(whoami) /etc/nixopus/source/api/.env

# Or use sudo for operations on system files
sudo nixopus conf set DATABASE_URL=value --service api
```


## Error Handling

Common error scenarios and solutions:

| Error | Cause | Solution |
|-------|-------|----------|
| File not found | Environment file doesn't exist | Create the file or use `--env-file` with existing file |
| Permission denied | Insufficient file permissions | Use `sudo` or fix file permissions with `chmod` |
| Invalid KEY=VALUE format | Missing equals sign in set command | Ensure format is `KEY=VALUE` |
| Service not found | Invalid service name | Use `api` or `view` for `--service` option |
| Operation timeout | Command taking too long | Increase `--timeout` value |

**Note**: When using `sudo`, ensure the environment files remain accessible to the services that need them.
