# service - Docker Compose Service Management

The `service` command provides comprehensive control over Nixopus services using Docker Compose integration. Manage the lifecycle of all Nixopus components including API, web interface, database, and proxy services.

## Quick Start
```bash
# Start all services
nixopus service up --detach

# Check service status
nixopus service ps

# Restart specific service
nixopus service restart --name api

# Stop all services
nixopus service down
```

## Overview

The service command acts as a Docker Compose wrapper with Nixopus-specific enhancements:
- Service lifecycle management (start, stop, restart, status)
- Environment-specific configuration loading
- Custom Docker Compose file support
- Service-specific targeting

## Subcommands

### `up` - Start Services

Start Nixopus services with dependency orchestration.

```bash
nixopus service up [OPTIONS]
```

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--name` | `-n` | Specific service name | `all` |
| `--detach` | `-d` | Run services in background | `false` |
| `--verbose` | `-v` | Show detailed logging | `false` |
| `--output` | `-o` | Output format (text, json) | `text` |
| `--dry-run` | | Preview operation without executing | `false` |
| `--env-file` | `-e` | Custom environment file path | None |
| `--compose-file` | `-f` | Custom Docker Compose file path | `/etc/nixopus/source/docker-compose.yml` |
| `--timeout` | `-t` | Operation timeout in seconds | `10` |

**Examples:**

```bash
# Start all services in foreground
nixopus service up

# Start all services in background
nixopus service up --detach

# Start specific service
nixopus service up --name api

# Preview operation
nixopus service up --dry-run
```

### `down` - Stop Services

Stop Nixopus services with graceful shutdown.

```bash
nixopus service down [OPTIONS]
```

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--name` | `-n` | Specific service name | `all` |
| `--verbose` | `-v` | Show detailed logging | `false` |
| `--output` | `-o` | Output format (text, json) | `text` |
| `--dry-run` | | Preview operation without executing | `false` |
| `--env-file` | `-e` | Custom environment file path | None |
| `--compose-file` | `-f` | Custom Docker Compose file path | `/etc/nixopus/source/docker-compose.yml` |
| `--timeout` | `-t` | Operation timeout in seconds | `10` |

**Examples:**

```bash
# Stop all services
nixopus service down

# Stop specific service
nixopus service down --name api

# Preview operation
nixopus service down --dry-run
```

### `ps` - Show Service Status

Display status information for Nixopus services.

```bash
nixopus service ps [OPTIONS]
```

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--name` | `-n` | Filter by specific service name | `all` |
| `--verbose` | `-v` | Show detailed service information | `false` |
| `--output` | `-o` | Output format (text, json) | `text` |
| `--dry-run` | `-d` | Preview operation without executing | `false` |
| `--env-file` | `-e` | Custom environment file path | None |
| `--compose-file` | `-f` | Custom Docker Compose file path | `/etc/nixopus/source/docker-compose.yml` |
| `--timeout` | `-t` | Operation timeout in seconds | `10` |

**Examples:**

```bash
# Show all services
nixopus service ps

# Show specific service
nixopus service ps --name api

# Get JSON output
nixopus service ps --output json
```

### `restart` - Restart Services

Restart services with configurable restart strategies.

```bash
nixopus service restart [OPTIONS]
```

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--name` | `-n` | Specific service name | `all` |
| `--verbose` | `-v` | Show detailed logging | `false` |
| `--output` | `-o` | Output format (text, json) | `text` |
| `--dry-run` | `-d` | Preview operation without executing | `false` |
| `--env-file` | `-e` | Custom environment file path | None |
| `--compose-file` | `-f` | Custom Docker Compose file path | `/etc/nixopus/source/docker-compose.yml` |
| `--timeout` | `-t` | Operation timeout in seconds | `10` |

**Examples:**

```bash
# Restart all services
nixopus service restart

# Restart specific service
nixopus service restart --name api

# Preview operation
nixopus service restart --dry-run
```

## Configuration

The service command reads configuration values from the built-in `config.prod.yaml` file to determine default compose file location.

### Default Configuration Values

| Setting | Default Value | Configuration Path | Description |
|---------|---------------|-------------------|-------------|
| Compose File | `source/docker-compose.yml` | `compose-file-path` | Docker Compose file path (relative to config dir) |
| Config Directory | `/etc/nixopus` | `nixopus-config-dir` | Base configuration directory |
| Timeout | `10` seconds | N/A | Operation timeout (hardcoded default) |

### Configuration Source

Configuration is loaded from the built-in `config.prod.yaml`:

```yaml
# Built-in configuration
nixopus-config-dir: /etc/nixopus
compose-file-path: source/docker-compose.yml
```

### Overriding Configuration

You can override defaults using command-line options:

```bash
# Use custom compose file
nixopus service up --compose-file /custom/docker-compose.yml

# Use custom environment file
nixopus service up --env-file /custom/.env

# Combine both
nixopus service up --compose-file /custom/compose.yml --env-file /custom/.env
```

## Error Handling

Common error scenarios and solutions:

| Error | Cause | Solution |
|-------|-------|----------|
| Compose file not found | Docker Compose file missing | Check file path or use `--compose-file` option |
| Docker daemon not running | Docker service stopped | Start Docker service: `sudo systemctl start docker` |
| Port already in use | Service running on required port | Stop conflicting service or change port configuration |
| Permission denied | Insufficient Docker permissions | Add user to docker group or use `sudo` |
| Service startup timeout | Service taking too long to start | Increase timeout with `--timeout` option |
