# clone - Repository Cloning

The `clone` command clones the Nixopus repository with basic configuration options. By default, it clones the main Nixopus repository to a configured local path.

## Quick Start
```bash
# Clone with default settings (from config)
nixopus clone

# Clone specific branch
nixopus clone --branch develop

# Clone from custom repository
nixopus clone --repo https://github.com/yourfork/nixopus.git

# Preview clone operation
nixopus clone --dry-run
```

## Overview

The clone command provides basic Git repository cloning functionality with configuration-driven defaults for the Nixopus repository.

## Command Syntax

```bash
nixopus clone [OPTIONS]
```

## Options

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--repo` | `-r` | Repository URL to clone | `https://github.com/raghavyuva/nixopus` |
| `--branch` | `-b` | Branch to clone | `master` |
| `--path` | `-p` | Local path for cloning | `/etc/nixopus/source` |
| `--force` | `-f` | Force clone (overwrite existing directory) | `false` |
| `--verbose` | `-v` | Show detailed logging | `false` |
| `--output` | `-o` | Output format | `text` |
| `--dry-run` | `-d` | Preview clone operation without executing | `false` |
| `--timeout` | `-t` | Operation timeout in seconds | `10` |

**Examples:**

```bash
# Basic clone with default configuration
nixopus clone

# Clone specific branch
nixopus clone --branch develop

# Clone from custom repository
nixopus clone --repo https://github.com/yourfork/nixopus.git

# Clone to custom path with force overwrite
nixopus clone --path /opt/nixopus --force

# Preview operation without executing
nixopus clone --dry-run --verbose

# Clone with increased timeout
nixopus clone --timeout 30
```

## Configuration

The clone command reads configuration values from the built-in [`config.prod.yaml`](https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/helpers/config.prod.yaml) file. Command-line options override these defaults.

### Default Configuration Values

| Setting | Default Value | Configuration Path | Description |
|---------|---------------|-------------------|-------------|
| Repository URL | `https://github.com/raghavyuva/nixopus` | `clone.repo` | The Git repository to clone |
| Branch | `master` | `clone.branch` | The Git branch to clone |
| Clone Path | `{nixopus-config-dir}/source` | `clone.source-path` | Local directory for cloning (relative to config dir) |
| Config Directory | `/etc/nixopus` | `nixopus-config-dir` | Base configuration directory |
| Timeout | `10` seconds | N/A | Operation timeout (hardcoded default) |

### Configuration Source

The configuration is loaded from the built-in [`config.prod.yaml`](https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/master/helpers/config.prod.yaml) file packaged with the CLI. This file contains environment variable placeholders that can be overridden:

```yaml
# Built-in configuration (from config.prod.yaml)
nixopus-config-dir: /etc/nixopus
clone:
  repo: "https://github.com/raghavyuva/nixopus"
  branch: "master"
  source-path: source
```

### Overriding Configuration

You can override defaults using command-line options only:

```bash
# Override repository URL
nixopus clone --repo https://github.com/yourfork/nixopus.git

# Override branch
nixopus clone --branch develop

# Override clone path (absolute path)
nixopus clone --path /opt/nixopus

# Override multiple options
nixopus clone --repo https://github.com/yourfork/nixopus.git --branch develop --path /custom/path
```

**Note**: The clone command does not support user configuration files or environment variable overrides for these settings. Configuration is handled internally through the built-in config file.

## Behavior

1. **Validates** repository URL and accessibility
2. **Checks** if destination path exists
3. **Removes** existing directory if `--force` is used
4. **Clones** repository using Git
5. **Reports** success or failure

## Dry Run Mode

Use `--dry-run` to preview what the command would do without making changes:

```bash
nixopus clone --dry-run --repo custom-repo.git --branch develop
```

This shows the planned actions without executing them.

## Error Handling

Common error scenarios and solutions:

| Error | Cause | Solution |
|-------|-------|----------|
| Repository not accessible | Network issues, invalid URL, or authentication required | Check network connection, verify repository URL, or configure Git credentials |
| Destination path already exists | Directory exists at clone path | Use `--force` to overwrite or choose different `--path` |
| Invalid branch or repository URL | Branch doesn't exist or URL is malformed | Verify branch exists and URL is correct |
| Permission denied | Insufficient permissions for destination path | Use `sudo nixopus clone` or choose a path with write permissions |
| Timeout exceeded | Clone taking longer than specified timeout | Increase timeout with `--timeout` option or check network speed |

### Permission Issues

If you encounter permission errors when cloning to system directories:

```bash
# Use sudo for system-wide installation
sudo nixopus clone --path /opt/nixopus

# Or clone to user directory (recommended)
nixopus clone --path ~/nixopus
```

**Note**: When using `sudo`, the cloned repository will be owned by root. Consider using user directories unless system-wide installation is required.
