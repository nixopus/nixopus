# uninstall - Complete Nixopus Removal

The `uninstall` command completely removes Nixopus from your system. This is a destructive operation that permanently removes all Nixopus components.

## Quick Start
```bash
# Standard uninstallation
nixopus uninstall

# Preview what will be removed
nixopus uninstall --dry-run --verbose

# Force uninstallation without prompts
nixopus uninstall --force
```

## Overview

The uninstall command completely removes Nixopus from your system including services, configuration files, and data.

## Command Syntax

```bash
nixopus uninstall [OPTIONS]
```

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--verbose` | `-v` | Show detailed uninstallation progress | `false` |
| `--timeout` | `-t` | Operation timeout in seconds | `300` |
| `--dry-run` | `-d` | Preview what would be removed without executing | `false` |
| `--force` | `-f` | Skip confirmation prompts and force removal | `false` |

**Examples:**

```bash
# Interactive uninstallation
nixopus uninstall

# Preview uninstallation
nixopus uninstall --dry-run --verbose

# Force uninstallation without prompts
nixopus uninstall --force

# Custom timeout
nixopus uninstall --timeout 600 --verbose
```

## Configuration

The uninstall command does not use external configuration files. It operates with hardcoded default values.

### Default Configuration Values

| Setting | Default Value | Description |
|---------|---------------|-------------|
| Timeout | `300` seconds | Maximum time to wait for each uninstallation step |
| Verbose | `false` | Show detailed logging during uninstallation |
| Dry Run | `false` | Preview mode without making actual changes |
| Force | `false` | Skip confirmation prompts |

### Overriding Configuration

You can override defaults using command-line options:

```bash
# Use custom timeout
nixopus uninstall --timeout 600

# Enable verbose logging
nixopus uninstall --verbose

# Preview without changes
nixopus uninstall --dry-run

# Force uninstall without prompts
nixopus uninstall --force
```

## Error Handling

Common error scenarios and solutions:

| Error | Cause | Solution |
|-------|-------|----------|
| Permission denied | Insufficient file system permissions | Use `sudo nixopus uninstall` |
| Services still running | Docker containers won't stop | Force stop with `docker stop` command |
| Files in use | Configuration files locked | Close applications using Nixopus files |
| Timeout exceeded | Uninstall taking too long | Increase timeout with `--timeout` option |

If permission issues occur, use sudo:
```bash
sudo nixopus uninstall --verbose
```

## Related Commands

- **[service](./service.md)** - Stop services before uninstalling
- **[conf](./conf.md)** - Backup configuration before removal