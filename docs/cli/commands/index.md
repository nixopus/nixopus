# Commands Quick Guide

Quick reference for all Nixopus CLI commands with essential examples.

## Command Reference

| Command | Usage | Key Options |
|---------|-------|-------------|
| **[preflight](./preflight.md)** | `nixopus preflight check` | `--verbose`, `--output json` |
| **[install](./install.md)** | `nixopus install --api-domain api.example.com` | `--force`, `--dry-run` |
| **[uninstall](./uninstall.md)** | `nixopus uninstall --dry-run` | `--force`, `--verbose` |
| **[service](./service.md)** | `nixopus service up --detach` | `--name api`, `--env-file` |
| **[conf](./conf.md)** | `nixopus conf set DATABASE_URL=...` | `--service api`, `--dry-run` |
| **[proxy](./proxy.md)** | `nixopus proxy load` | `--config-file`, `--proxy-port` |
| **[clone](./clone.md)** | `nixopus clone --branch develop` | `--path`, `--repo`, `--force` |
| **[version](./version.md)** | `nixopus version` | - |
| **[test](./test.md)** | `nixopus test` (dev only) | Requires `ENV=DEVELOPMENT` |

## Common Workflows

### Initial Setup
```bash
nixopus preflight check
nixopus install --api-domain api.example.com
nixopus service up --detach
nixopus proxy load
nixopus service ps
```

### Configuration Management
```bash
nixopus conf list --service api
nixopus conf set DATABASE_URL=postgresql://localhost:5432/nixopus
nixopus service restart --name api
```

### Development Setup
```bash
nixopus clone --branch develop
nixopus install --dry-run
nixopus service up --env-file .env.development
export ENV=DEVELOPMENT && nixopus test
```

### Troubleshooting
```bash
nixopus preflight check --verbose
nixopus preflight ports 80 443 8080
nixopus service ps --verbose
nixopus proxy status --verbose
```

## Global Options

| Option | Description | Example |
|--------|-------------|---------|
| `--verbose, -v` | Show detailed output | `nixopus install --verbose` |
| `--output, -o` | Format (text, json) | `nixopus service ps --output json` |
| `--dry-run, -d` | Preview without executing | `nixopus install --dry-run` |
| `--timeout, -t` | Operation timeout (seconds) | `nixopus service up --timeout 120` |
| `--help` | Show command help | `nixopus proxy --help` |

## Get Help

```bash
nixopus --help                  # General help
nixopus install --help          # Command help
nixopus service up --help       # Subcommand help
```