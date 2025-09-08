no_version_conflicts_message = "No version conflicts to check"  # Message constants for conflict command

# General messages
conflict_check_help = "Check for tool version conflicts"
error_checking_conflicts = "Error checking conflicts: {error}"
no_conflicts_found = "No version conflicts found"
conflict_checking_tool = "Checking tool: {tool}"
conflict_loading_config = "Loading configuration from {path}"
conflict_config_loaded = "Configuration loaded successfully"
conflict_config_not_found = "Configuration file not found at {path}"
conflict_invalid_config = "Invalid configuration file: {error}"

# Tool-specific messages
tool_not_found = "Tool not found"
tool_version_mismatch = "Version mismatch"
tool_version_compatible = "Version compatible"

# Error messages
error_checking_tool_version = "Error checking version for {tool}: {error}"
error_parsing_version = "Error parsing version for {tool}: {error}"
timeout_checking_tool = "Timeout checking tool: {tool}"

# Success/Info messages
conflicts_found_warning = "Found {count} version conflict(s)"
no_conflicts_info = "No version conflicts found"

# Status messages
checking_conflicts_info = "Checking for tool version conflicts..."

# Version specification help
supported_version_formats_info = """
Supported version formats in config files:
  - Exact version: "1.20.3"
  - Range operators: ">=1.20.0, <2.0.0"
  - Greater/less than: ">=1.20.0", "<2.0.0"
  - Compatible release: "~=1.20.0"
  - Major.minor only: "1.20" (treated as >=1.20.0, <1.21.0)
"""

unsupported_version_format_warning = "Unsupported version format '{format}' for {tool}. {help}"

# warning messages
no_deps_found_warning = "No dependencies found in configuration"
