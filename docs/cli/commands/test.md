# test - CLI Testing Utilities

The `test` command runs tests for the Nixopus CLI in development environments. This command is restricted to development environments only.

## Quick Start
```bash
# Set development environment (required)
export ENV=DEVELOPMENT

# Run all tests
nixopus test

# Run specific test target
nixopus test version
```

## Overview

The test command provides basic testing capabilities for the Nixopus CLI. It requires the `ENV=DEVELOPMENT` environment variable to prevent accidental execution in production.

## Command Syntax

```bash
nixopus test [TARGET]
```

| Argument | Description | Required |
|----------|-------------|----------|
| `TARGET` | Specific test target (e.g., version) | No |

**Examples:**

```bash
# Set development environment first (required)
export ENV=DEVELOPMENT

# Run all tests
nixopus test

# Run specific command tests
nixopus test version
```

## Configuration

The test command does not use external configuration files. It operates with environment variable requirements.

### Environment Requirements

| Setting | Required Value | Description |
|---------|---------------|-------------|
| ENV | `DEVELOPMENT` | Must be set to enable testing |

### Configuration Source

The test command requires the `ENV=DEVELOPMENT` environment variable to be set:

```bash
# Required environment setup
export ENV=DEVELOPMENT
```

### Overriding Configuration

You can specify different test targets using the command argument:

```bash
# Test specific command
nixopus test version

# Test different command
nixopus test install
```

## Error Handling

Common error scenarios and solutions:

| Error | Cause | Solution |
|-------|-------|----------|
| Environment not development | ENV not set to DEVELOPMENT | Set `export ENV=DEVELOPMENT` |
| Test dependencies missing | Development packages not installed | Install with `poetry install --with dev` |
| Permission denied | File system permissions | Use `sudo` if necessary |

If permission issues occur, use sudo:
```bash
sudo nixopus test
```

## Related Commands

- **[version](./version.md)** - Check CLI version before running tests
- **[preflight](./preflight.md)** - Validate test environment setup