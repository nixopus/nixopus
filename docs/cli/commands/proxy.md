# proxy - Caddy Proxy Management

The `proxy` command controls the Caddy reverse proxy server that handles HTTP routing, SSL termination, and load balancing for Nixopus services. Manage proxy configuration, monitor status, and control the proxy lifecycle.

## Quick Start
```bash
# Load proxy configuration
nixopus proxy load

# Check proxy status
nixopus proxy status

# Stop proxy server
nixopus proxy stop
```

## Overview

The proxy command manages Caddy as the reverse proxy for Nixopus:
- HTTP/HTTPS routing to API and view services
- Configuration loading and management
- Proxy status monitoring
- Graceful proxy shutdown

## Subcommands

### `load` - Load Proxy Configuration

Load and apply Caddy proxy configuration from file with validation support.

```bash
nixopus proxy load [OPTIONS]
```

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--proxy-port` | `-p` | Caddy admin API port | `2019` |
| `--verbose` | `-v` | Show detailed logging | `false` |
| `--output` | `-o` | Output format (text, json) | `text` |
| `--dry-run` | | Validate configuration without applying | `false` |
| `--config-file` | `-c` | Path to Caddy configuration file | None |
| `--timeout` | `-t` | Operation timeout in seconds | `10` |

**Examples:**

```bash
# Load default proxy configuration
nixopus proxy load

# Load custom configuration file
nixopus proxy load --config-file /path/to/caddy.json

# Validate configuration without applying
nixopus proxy load --config-file caddy.json --dry-run

# Load with custom admin port
nixopus proxy load --proxy-port 2019 --verbose
```

### `status` - Check Proxy Status

Display status information about the Caddy proxy server.

```bash
nixopus proxy status [OPTIONS]
```

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--proxy-port` | `-p` | Caddy admin API port | `2019` |
| `--verbose` | `-v` | Show detailed status information | `false` |
| `--output` | `-o` | Output format (text, json) | `text` |
| `--dry-run` | | Preview operation without executing | `false` |
| `--timeout` | `-t` | Operation timeout in seconds | `10` |

**Examples:**

```bash
# Basic proxy status
nixopus proxy status

# Detailed status information
nixopus proxy status --verbose

# JSON output for monitoring
nixopus proxy status --output json

# Check with custom admin port
nixopus proxy status --proxy-port 2019
```

### `stop` - Stop Proxy Server

Gracefully stop the Caddy proxy server.

```bash
nixopus proxy stop [OPTIONS]
```

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--proxy-port` | `-p` | Caddy admin API port | `2019` |
| `--verbose` | `-v` | Show detailed logging | `false` |
| `--output` | `-o` | Output format (text, json) | `text` |
| `--dry-run` | | Preview stop operation | `false` |
| `--timeout` | `-t` | Operation timeout in seconds | `10` |

**Examples:**

```bash
# Graceful proxy shutdown
nixopus proxy stop

# Stop with detailed logging
nixopus proxy stop --verbose

# Preview stop operation
nixopus proxy stop --dry-run

# Stop with custom admin port
nixopus proxy stop --proxy-port 2019
```

## Configuration

The proxy command reads configuration values from the built-in `config.prod.yaml` file to determine the default Caddy admin port.

### Default Configuration Values

| Setting | Default Value | Configuration Path | Description |
|---------|---------------|-------------------|-------------|
| Proxy Port | `2019` | `services.caddy.env.PROXY_PORT` | Caddy admin API port |
| Timeout | `10` seconds | N/A | Operation timeout (hardcoded default) |

### Configuration Source

Configuration is loaded from the built-in `config.prod.yaml`:

```yaml
# Built-in configuration
services:
  caddy:
    env:
      PROXY_PORT: ${PROXY_PORT:-2019}
```

### Overriding Configuration

You can override defaults using command-line options:

```bash
# Use custom admin port
nixopus proxy status --proxy-port 8080

# Use custom config file
nixopus proxy load --config-file /custom/caddy.json

# Combine both
nixopus proxy load --proxy-port 8080 --config-file /custom/caddy.json
```

## Error Handling

Common error scenarios and solutions:

| Error | Cause | Solution |
|-------|-------|----------|
| Connection refused | Caddy admin API not running | Start Caddy or check admin port |
| Configuration file not found | Invalid config file path | Check file path and permissions |
| Invalid configuration | Malformed Caddy config | Validate JSON/config syntax |
| Permission denied | Insufficient network permissions | Use sudo or check port availability |
| Operation timeout | Network or server issues | Increase timeout with `--timeout` option |
