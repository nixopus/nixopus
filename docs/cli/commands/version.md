# version - CLI Version Information

The `version` command displays the current version of the Nixopus CLI. Essential for troubleshooting and support requests.

## Quick Start
```bash
# Display version information
nixopus version

# Short version flag
nixopus --version
```

## Overview

The version command provides basic version information about the Nixopus CLI installation using the package metadata.

## Command Syntax

```bash
nixopus version
```

**Alternative Forms:**
```bash
nixopus --version
nixopus -v
```

**Examples:**

```bash
# Display version information
nixopus version

# Alternative syntax
nixopus --version
nixopus -v
```

## Configuration

The version command does not use external configuration. It reads version information directly from the installed package metadata using Python's `importlib.metadata.version()`.

### Version Source

The version is determined from:
- **Package metadata** - Installed package version from `importlib.metadata`
- **Display formatting** - Rich console formatting for output

## Error Handling

Common error scenarios and solutions:

| Error | Cause | Solution |
|-------|-------|----------|
| Package not found | CLI not properly installed | Reinstall using `poetry install nixopus` |
| Import error | Python environment issues | Check Python installation and PATH |
| Permission denied | File system permissions | Check package installation permissions |
