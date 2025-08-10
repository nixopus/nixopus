# install - Nixopus Installation

The `install` command installs Nixopus with all required components and configuration. Provides comprehensive setup including dependencies and SSH key generation.

## Quick Start
```bash
# Basic installation
nixopus install

# Install with custom domains
nixopus install --api-domain api.example.com --view-domain app.example.com

# Preview installation changes
nixopus install --dry-run --verbose
```

## Overview

The install command provides a comprehensive setup process including system validation, dependency installation, and service configuration.

## Command Syntax

```bash
nixopus install [OPTIONS]
```

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--verbose` | `-v` | Show detailed installation progress | `false` |
| `--timeout` | `-t` | Installation timeout in seconds | `300` |
| `--force` | `-f` | Replace existing files without prompting | `false` |
| `--dry-run` | `-d` | Preview installation without making changes | `false` |
| `--config-file` | `-c` | Path to custom configuration file | None |
| `--api-domain` | `-ad` | Domain for API access | None |
| `--view-domain` | `-vd` | Domain for web interface | None |

**Examples:**

```bash
# Standard installation
nixopus install

# Production installation with custom domains
nixopus install --api-domain api.production.com --view-domain app.production.com --timeout 600

# Preview installation with verbose output
nixopus install --dry-run --verbose

# Force installation (overwrite existing files)
nixopus install --force
```

## Subcommands

### `ssh` - SSH Key Generation

Generate SSH key pairs with proper permissions and optional authorized_keys integration.

```bash
nixopus install ssh [OPTIONS]
```

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--path` | `-p` | SSH key file path | `~/.ssh/nixopus_rsa` |
| `--key-type` | `-t` | Key type (rsa, ed25519, ecdsa) | `rsa` |
| `--key-size` | `-s` | Key size in bits | `4096` |
| `--passphrase` | `-P` | Passphrase for key encryption | None |
| `--force` | `-f` | Overwrite existing SSH keys | `false` |
| `--set-permissions` | `-S` | Set proper file permissions | `true` |
| `--add-to-authorized-keys` | `-a` | Add public key to authorized_keys | `false` |
| `--create-ssh-directory` | `-c` | Create .ssh directory if needed | `true` |
| `--verbose` | `-v` | Show detailed output | `false` |
| `--output` | `-o` | Output format (text, json) | `text` |
| `--dry-run` | `-d` | Preview operation | `false` |
| `--timeout` | `-T` | Operation timeout in seconds | `10` |

**Examples:**

```bash
# Generate default RSA key
nixopus install ssh

# Generate RSA key with custom path and size
nixopus install ssh --path ~/.ssh/nixopus_rsa --key-type rsa --key-size 4096

# Generate encrypted key for production
nixopus install ssh --passphrase "secure-passphrase" --add-to-authorized-keys
```

### `deps` - Dependency Installation

Install and configure system dependencies required for Nixopus operation.

```bash
nixopus install deps [OPTIONS]
```

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--verbose` | `-v` | Show detailed installation progress | `false` |
| `--output` | `-o` | Output format (text, json) | `text` |
| `--dry-run` | `-d` | Preview dependency installation | `false` |
| `--timeout` | `-t` | Installation timeout in seconds | `10` |

**Examples:**

```bash
# Install all required dependencies
nixopus install deps

# Preview dependency installation
nixopus install deps --dry-run --verbose

# Get JSON output for automation
nixopus install deps --output json
```

## Configuration

The install command reads configuration values from the built-in `config.prod.yaml` file and accepts command-line overrides.

### Default Configuration Values

| Setting | Default Value | Description |
|---------|---------------|-------------|
| Timeout | `300` seconds | Maximum time to wait for installation steps |
| SSH Key Path | `~/.ssh/nixopus_rsa` | Default SSH key location |
| SSH Key Type | `rsa` | Default SSH key algorithm |
| SSH Key Size | `4096` bits | Default key size for RSA keys |

### Configuration Source

Configuration is loaded from the built-in `config.prod.yaml` and command-line options.

### Overriding Configuration

You can override defaults using command-line options:

```bash
# Use custom domains
nixopus install --api-domain api.example.com --view-domain app.example.com

# Use custom config file
nixopus install --config-file /path/to/config.yaml

# Custom timeout and force mode
nixopus install --timeout 600 --force
```

## Error Handling

Common error scenarios and solutions:

| Error | Cause | Solution |
|-------|-------|----------|
| Permission denied | Insufficient file system permissions | Use `sudo nixopus install` |
| Docker not available | Docker daemon not running | Start Docker service |
| Port conflicts | Ports already in use | Stop conflicting services |
| SSH key generation fails | SSH directory permissions | Fix SSH directory permissions |
| Installation timeout | Network or system issues | Increase timeout with `--timeout` option |

If permission issues occur, use sudo:
```bash
sudo nixopus install --verbose
```

## Related Commands

- **[preflight](./preflight.md)** - Run system checks before installation
- **[service](./service.md)** - Manage installed services
- **[conf](./conf.md)** - Configure installed services
- **[uninstall](./uninstall.md)** - Remove Nixopus installation