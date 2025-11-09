# Nixopus CLI

Command line interface for managing Nixopus applications and services.

## Quick Start

```bash
# Install CLI
curl -sSL https://install.nixopus.com | bash -s -- --skip-nixopus-install

# Install Nixopus
nixopus install

```

## Commands

| Command | Description |
|---------|-------------|
| `preflight` | Check system requirements |
| `install` | Install Nixopus |
| `uninstall` | Remove Nixopus |
| `conf` | Manage configuration (list, set, delete) |
| `proxy` | Manage Caddy proxy (load, status, stop) |
| `version` | Show CLI version |

For detailed command documentation, see the [CLI Reference](./cli-reference.md).

## Getting Help

```bash
nixopus --help
nixopus <command> --help
```

## More Information

- [Installation Guide](./installation.md)
- [Configuration Guide](./config.md)
- [CLI Reference](./cli-reference.md)