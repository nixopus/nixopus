# conflict - Tool Version Conflict Detection

The `conflict` command checks for version conflicts between required tools and their installed versions, helping ensure compatibility before deployment.

## Quick Start
```bash
# Check for version conflicts
nixopus conflict

# Check with custom config file
nixopus conflict --config-file /path/to/config.yaml

# Get JSON output
nixopus conflict --output json
```

## Overview

The conflict command analyzes your system's installed tool versions against the requirements specified in your configuration file. It helps identify potential compatibility issues by:
- Verifying tool availability
- Comparing installed versions with expected version ranges
- Supporting various version specification formats
- Providing detailed conflict reports

## Usage

```bash
nixopus conflict [OPTIONS]
```

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--config-file` | `-c` | Path to configuration file | `helpers/config.prod.yaml` |
| `--timeout` | `-t` | Timeout for tool checks in seconds | `5` |
| `--verbose` | `-v` | Show detailed logging | `false` |
| `--output` | `-o` | Output format (text/json) | `text` |

**Examples:**

```bash
# Basic conflict check
nixopus conflict

# Check with verbose output
nixopus conflict --verbose

# Use custom configuration file
nixopus conflict --config-file /custom/config.yaml

# Get JSON formatted results
nixopus conflict --output json

# Increase timeout for slower systems
nixopus conflict --timeout 10
```

## Configuration

The conflict command reads tool version requirements from your configuration file's `deps` section.

### Configuration Format

```yaml
deps:
  docker:
    version: ">=20.0.0"
  git:
    version: ">=2.30.0"
  go:
    version: "1.20"
  postgresql:
    version: ">=14.0.0, <16.0.0"
```

### Supported Version Formats

The conflict checker supports multiple version specification formats:

| Format | Example | Description |
|--------|---------|-------------|
| Exact version | `"1.20.3"` | Must match exactly |
| Range operators | `">=1.20.0, <2.0.0"` | Version ranges with multiple conditions |
| Greater/less than | `">=1.20.0"`, `"<2.0.0"` | Single comparison operators |
| Compatible release | `"~=1.20.0"` | Compatible release (Python-style) |
| Major.minor only | `"1.20"` | Treated as `>=1.20.0, <1.21.0` |

### Version Command Configuration

You can specify custom version commands for tools in the configuration:

```yaml
deps:
  custom-tool:
    version: ">=2.0.0"
    version-command: ["custom-tool", "--show-version"]
```

## Tool-Specific Support

The conflict checker includes built-in support for common tools:

- **Docker**: Checks container runtime version
- **Git**: Verifies source control version
- **Go**: Checks Go compiler version
- **PostgreSQL**: Validates database version
- **SSH/OpenSSH**: Checks SSH client/server versions
- **Redis**: Verifies Redis server version
- **Python**: Checks Python interpreter version
