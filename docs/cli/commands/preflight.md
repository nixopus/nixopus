# preflight - System Readiness Checks

The `preflight` command runs comprehensive system checks to ensure your environment is properly configured for Nixopus installation and operation.

## Quick Start
```bash
# Run full system check
nixopus preflight check

# Check specific ports
nixopus preflight ports 80 443 8080

# Verify dependencies
nixopus preflight deps docker git
```

## Overview

The preflight command performs system readiness checks including port availability and dependency verification.

## Subcommands

### `check` - Comprehensive System Check

Runs port availability checks based on configuration. This is the default command when running `preflight` without subcommands.

```bash
nixopus preflight check [OPTIONS]
nixopus preflight [OPTIONS]  # same as check
```

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--verbose` | `-v` | Show detailed logging information | `false` |
| `--output` | `-o` | Output format (text, json) | `text` |
| `--timeout` | `-t` | Operation timeout in seconds | `10` |

**Examples:**

```bash
# Basic system check
nixopus preflight check

# Detailed check with verbose output
nixopus preflight check --verbose
```

**What it does:**
- Reads required ports from the configuration file
- Checks if those ports are available on localhost
- Reports success if all configured ports are free

### `ports` - Port Availability Check

Verify specific ports are available for Nixopus services.

```bash
nixopus preflight ports [PORT...] [OPTIONS]
```

**Arguments:**
- `PORT...` - List of ports to check (required)

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--host` | `-h` | Host to check | `localhost` |
| `--verbose` | `-v` | Show detailed logging | `false` |
| `--output` | `-o` | Output format (text, json) | `text` |
| `--timeout` | `-t` | Operation timeout in seconds | `10` |

**Examples:**

```bash
# Check standard web ports
nixopus preflight ports 80 443

# Check ports on remote host
nixopus preflight ports 80 443 --host production.server.com

# Get JSON output
nixopus preflight ports 80 443 8080 --output json
```

**Output:**
The command outputs a formatted table or JSON showing port availability status for each port checked.

### `deps` - Dependency Verification

Check if required system dependencies are installed and accessible.

```bash
nixopus preflight deps [DEPENDENCY...] [OPTIONS]
```

**Arguments:**
- `DEPENDENCY...` - List of dependencies to check (required)

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--verbose` | `-v` | Show detailed logging | `false` |
| `--output` | `-o` | Output format (text, json) | `text` |
| `--timeout` | `-t` | Operation timeout in seconds | `10` |

**Common Dependencies:**
- `docker` - Docker container runtime
- `docker-compose` - Docker Compose orchestration
- `git` - Git version control
- `curl` - HTTP client utility
- `ssh` - SSH client

**Examples:**

```bash
# Check core dependencies
nixopus preflight deps docker git

# Check with verbose output
nixopus preflight deps docker git --verbose

# Get JSON output
nixopus preflight deps docker git --output json
```

**Output:**
The command outputs a formatted table showing dependency availability. Uses `shutil.which()` to check if commands are available in the system PATH.

## Configuration

The preflight command reads configuration values from the built-in `config.prod.yaml` file to determine which ports and dependencies to check.

### Default Configuration Values

| Setting | Default Value | Configuration Path | Description |
|---------|---------------|-------------------|-------------|
| Ports | `[2019, 80, 443, 7443, 8443, 6379, 5432]` | `ports` | Ports checked by default |
| Timeout | `10` seconds | N/A | Operation timeout (hardcoded default) |

### Configuration Source

Configuration is loaded from the built-in `config.prod.yaml`:

```yaml
# Built-in configuration
ports: [2019, 80, 443, 7443, 8443, 6379, 5432]

deps:
  curl:           { package: "curl",           command: "curl" }
  python3:        { package: "python3",        command: "python3" }
  python3-venv:   { package: "python3-venv",   command: "" }
  git:            { package: "git",            command: "git" }
  docker.io:      { package: "docker.io",      command: "docker" }
  openssl:        { package: "openssl",        command: "openssl" }
  openssh-client: { package: "openssh-client", command: "ssh" }
  openssh-server: { package: "openssh-server", command: "sshd" }
```

### Port Descriptions

| Port | Service | Purpose |
|------|---------|---------|
| `2019` | Caddy | Admin API port |
| `80` | HTTP | Web traffic |
| `443` | HTTPS | Secure web traffic |
| `7443` | View | Frontend service |
| `8443` | API | Backend service |
| `6379` | Redis | Cache/session store |
| `5432` | PostgreSQL | Database |

### Available Dependencies

| Command | Package | Purpose |
|---------|---------|---------|
| `curl` | curl | HTTP client utility |
| `python3` | python3 | Python runtime |
| `git` | git | Version control |
| `docker` | docker.io | Container runtime |
| `openssl` | openssl | SSL/TLS toolkit |
| `ssh` | openssh-client | SSH client |
| `sshd` | openssh-server | SSH server |

## Error Handling

Common error scenarios and solutions:

| Error | Cause | Solution |
|-------|-------|----------|
| Port already in use | Service running on checked port | Stop conflicting service or use different ports |
| Command not found | Dependency not installed | Install required package using system package manager |
| Permission denied | Docker requires elevated privileges | Add user to docker group or use sudo |
| Connection timeout | Network issues or slow system | Increase timeout with `--timeout` option |
| Invalid port number | Port outside valid range | Use port numbers between 1-65535 |

### Permission Issues

If you encounter permission errors, especially with Docker:

```bash
# Check if docker requires sudo
docker ps
# If this fails with permission denied:

# Add user to docker group
sudo usermod -aG docker $USER

# Restart shell session
newgrp docker

# Test without sudo
docker ps

# Or use sudo for preflight checks
sudo nixopus preflight deps docker
```

**Note**: When using `sudo`, ensure the command can access the same configuration files.
