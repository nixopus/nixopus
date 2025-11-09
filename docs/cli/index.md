# Nixopus CLI

Command line interface for managing Nixopus applications and services. Built with Python and Typer, providing an intuitive terminal experience for deployment and management.

## Quick Start

```bash
# Install CLI
curl -sSL https://raw.githubusercontent.com/raghavyuva/nixopus/master/cli/install.sh | bash

# Verify installation
nixopus --help

# Check system requirements
nixopus preflight check

# Install Nixopus
nixopus install
```

## Available Commands

For detailed command reference, see the [CLI Reference](./cli-reference.md) page.

| Command | Description | Key Subcommands |
|---------|-------------|-----------------|
| **preflight** | System readiness checks | check, ports, deps |
| **conflict** | Tool version conflict detection | - |
| **install** | Complete Nixopus installation | ssh, deps |
| **uninstall** | Remove Nixopus from system | - |
| **service** | Control Docker services | up, down, ps, restart |
| **conf** | Manage application settings | list, set, delete |
| **proxy** | Caddy proxy management | load, status, stop |
| **clone** | Repository cloning with Git | - |
| **version** | Display CLI version information | - |
| **test** | Run CLI tests (development only) | - |

## Common Workflows

### Initial Setup
```bash
# 1. Check system requirements
nixopus preflight check

# 2. Verify tool versions
nixopus conflict

# 3. Install with custom domains
nixopus install --api-domain api.example.com --view-domain app.example.com

# 4. Start services
nixopus service up --detach

# 5. Load proxy configuration
nixopus proxy load

# 6. Verify everything is running
nixopus service ps
```

### Configuration Management
```bash
# View current configuration
nixopus conf list --service api

# Update settings
nixopus conf set DATABASE_URL=postgresql://user:pass@localhost:5432/nixopus

# Restart services to apply changes
nixopus service restart
```

### Development Setup
```bash
# Clone repository
nixopus clone --branch develop

# Check for version conflicts
nixopus conflict --config-file config.dev.yaml

# Preview installation
nixopus install --dry-run

# Start development environment
nixopus service up --env-file .env.development

# Run tests
export ENV=DEVELOPMENT
nixopus test
```

## Global Options

Most commands support these options:

| Option | Shorthand | Description |
|--------|-----------|-------------|
| `--verbose` | `-v` | Show detailed output |
| `--output` | `-o` | Output format (text, json) |
| `--dry-run` | `-d` | Preview without executing |
| `--timeout` | `-t` | Operation timeout in seconds |
| `--help` | | Show command help |

## Getting Help

```bash
# General help
nixopus --help

# Command-specific help
nixopus install --help
nixopus service --help

# Subcommand help
nixopus service up --help
```

## Reference

For complete command reference and detailed documentation, see the [CLI Reference](./cli-reference.md) page.