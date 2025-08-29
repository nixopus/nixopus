import os
import subprocess
import re
from typing import Dict, List, Optional, Any, Tuple
from packaging import version
from packaging.specifiers import SpecifierSet
from packaging.version import Version

from app.utils.logger import Logger
from app.utils.protocols import LoggerProtocol
from app.utils.output_formatter import OutputFormatter
from app.utils.lib import ParallelProcessor
from app.utils.config import Config, DEPS
from .models import ConflictCheckResult, ConflictConfig
from .messages import *


class VersionParser:

    @staticmethod
    def is_major_minor_format(requirement: str) -> bool:
        """Check if the requirement is in major.minor format (e.g., '1.20')."""
        return bool(re.match(r"^\d+\.\d+$", requirement))

    @staticmethod
    def _search_version(pattern: str, output: str, flags: int = re.IGNORECASE) -> Optional[str]:
        """Helper to search for a version pattern and return group(1) if found."""
        if match := re.search(pattern, output, flags):
            return match.group(1)
        return None

    """Utility class for parsing and comparing versions."""

    # Version pattern mappings for different tools
    VERSION_PATTERNS = [
        r"version\s+(\d+\.\d+\.\d+)",  # "version 1.20.3", "version 2.1.0"
        r"v(\d+\.\d+\.\d+)",  # "v1.20.3", "v2.1.0"
        r"(\d+\.\d+\.\d+)",  # "1.20.3", "2.1.0" (standalone)
        r"Version\s+(\d+\.\d+\.\d+)",  # "Version 1.20.3", "Version 2.1.0"
        r"(\d+\.\d+)",  # "1.20", "2.1" (major.minor only)
    ]

    # Version operators for requirement specifications
    VERSION_OPERATORS = [">=", "<=", ">", "<", "==", "!=", "~", "^"]

    # Supported version specification formats in config files
    # Format: "description": "example"
    # SUPPORTED_VERSION_FORMATS
    #     "exact_version": "1.20.3"
    #     "range_operators": ">=1.20.0, <2.0.0"
    #     "greater_than_equal": ">=1.20.0"
    #     "less_than": "<2.0.0"
    #     "compatible_range": "~=1.20.0"  # Python-style compatible release
    #     "major_minor_only": "1.20"  # Implies >=1.20.0, <1.21.0

    @staticmethod
    def parse_version_output(tool: str, output: str) -> Optional[str]:
        """Parse version from tool output."""
        try:
            # Common version patterns
            for pattern in VersionParser.VERSION_PATTERNS:
                if version := VersionParser._search_version(pattern, output):
                    return version

            # Tool-specific parsing for unique output formats
            if tool == "go":
                # "go version go1.20.3 darwin/amd64" -> "1.20.3"
                if version := VersionParser._search_version(r"go(\d+\.\d+\.\d+)", output, 0):
                    return version

            elif tool == "curl":
                # "curl 7.53.1 (x86_64-apple-darwin14.5.0)..." -> "7.53.1"
                if version := VersionParser._search_version(r"curl\s+(\d+\.\d+\.\d+)", output, 0):
                    return version

            elif tool == "ssh" or tool == "open-ssh" or tool == "openssh-server":
                # "OpenSSH_9.8p1, LibreSSL 3.3.6" -> "9.8.1"
                if match := re.search(r"OpenSSH_(\d+\.\d+)(?:p(\d+))?", output):
                    major_minor = match.group(1)
                    patch = match.group(2) or "0"
                    return f"{major_minor}.{patch}"

            elif tool == "redis":
                # "Redis server v=7.0.11 sha=00000000:0..." -> "7.0.11"
                if version := VersionParser._search_version(r"v=(\d+\.\d+\.\d+)", output, 0):
                    return version

            elif tool == "postgresql" or tool == "psql":
                # "psql (PostgreSQL) 14.9" -> "14.9"
                if version := VersionParser._search_version(r"PostgreSQL\)\s+(\d+\.\d+)", output, 0):
                    return version

            elif tool == "air":
                # Air might have specific format, keeping flexible for now
                if version := VersionParser._search_version(r"(\d+\.\d+\.\d+)", output, 0):
                    return version

            return None
        except Exception as e:
            raise ValueError(error_parsing_version.format(tool=tool, error=str(e)))

    @staticmethod
    def compare_versions(current: str, expected: str) -> bool:
        """Compare version against requirement specification."""
        try:
            # Handle simple version comparisons (backwards compatibility)
            if not any(op in expected for op in VersionParser.VERSION_OPERATORS):
                # Default to >= for simple version strings
                return version.parse(current) >= version.parse(expected)

            # Handle version ranges and specifiers
            spec_set = SpecifierSet(expected)
            return Version(current) in spec_set

        except Exception:
            # Fallback to string comparison
            return current == expected

    @staticmethod
    def normalize_version_requirement(requirement: str) -> str:
        """
        Parse version requirement and return a normalized specifier.
        """

        if not requirement:
            return requirement

        requirement = requirement.strip()

        # If it already contains operators, return as-is
        if any(op in requirement for op in VersionParser.VERSION_OPERATORS):
            return requirement

        # Handle major.minor format (e.g., "1.20" -> ">=1.20.0, <1.21.0")
        if VersionParser.is_major_minor_format(requirement):
            try:
                parts = requirement.split(".")
                major, minor = int(parts[0]), int(parts[1])
                return f">={requirement}.0, <{major}.{minor + 1}.0"
            except (ValueError, IndexError):
                return f">={requirement}"

        # Handle exact version format (e.g., "1.20.3" -> "==1.20.3")
        if re.match(r"^\d+\.\d+\.\d+$", requirement):
            return f"=={requirement}"

        # If none of the above, treat as exact match
        return f"=={requirement}"

    @staticmethod
    def validate_version_format(requirement: str) -> bool:
        """
        Validate if the version requirement follows supported formats.
        Returns True if the format is supported, False otherwise.
        """

        if not requirement:
            return True

        requirement = requirement.strip()

        # Check if it contains supported operators
        if any(op in requirement for op in VersionParser.VERSION_OPERATORS):
            return True

        # Check for major.minor format
        if VersionParser.is_major_minor_format(requirement):
            return True

        # Check for exact version format
        if re.match(r"^\d+\.\d+\.\d+$", requirement):
            return True

        # If none match, it's unsupported
        return False


class ToolVersionChecker:
    """Handles version checking for different tools."""

    # Tool name mappings for command execution
    TOOL_MAPPING = {"open-ssh": "ssh", "open-sshserver": "sshd", "python3-venv": "python3"}  # TODO: @shravan20 Fix this issue

    def __init__(self, logger: LoggerProtocol, deps_config: Optional[Dict[str, Any]] = None, timeout: int = 10):
        self.timeout = timeout  # Default timeout for individual subprocess calls
        self.logger = logger
        self.deps_config = deps_config or {}

    def get_tool_version(self, tool: str) -> Optional[str]:
        """Get version of a tool."""
        try:
            # get version-command from deps config
            cmd = None
            if tool in self.deps_config:
                tool_cfg = self.deps_config[tool]
                cmd = tool_cfg.get("version-command")
            # Fallback to default if not found
            if not cmd:
                cmd = [tool, "--version"]

            result = subprocess.run(cmd, capture_output=True, text=True, timeout=self.timeout)

            if result.returncode == 0:
                return VersionParser.parse_version_output(tool, result.stdout)
            else:
                # fallback to alternative command if available
                alt_cmd = [tool, "-v"]
                result = subprocess.run(alt_cmd, capture_output=True, text=True, timeout=self.timeout)
                if result.returncode == 0:
                    return VersionParser.parse_version_output(tool, result.stdout)

        except subprocess.TimeoutExpired:
            self.logger.error(timeout_checking_tool.format(tool=tool))
            return None
        except Exception as e:
            self.logger.error(error_checking_tool_version.format(tool=tool, error=str(e)))
            return None

        return None

    def check_tool_version(self, tool: str, expected_version: Optional[str]) -> ConflictCheckResult:
        """Check a single tool's version against expected version."""
        command_name = self.TOOL_MAPPING.get(tool, tool)
        current_version = self.get_tool_version(command_name)

        if current_version is None:
            return ConflictCheckResult(
                tool=tool, expected=expected_version, current=None, status=tool_not_found, conflict=True
            )

        if expected_version is None or expected_version == "":
            # Just check existence
            return ConflictCheckResult(
                tool=tool, expected="present", current=current_version, status=tool_version_compatible, conflict=False
            )

        # Parse version requirement to handle ranges
        normalized_expected = VersionParser.normalize_version_requirement(expected_version)

        # Check version compatibility
        is_compatible = VersionParser.compare_versions(current_version, normalized_expected)

        return ConflictCheckResult(
            tool=tool,
            expected=normalized_expected,
            current=current_version,
            status=tool_version_compatible if is_compatible else tool_version_mismatch,
            conflict=not is_compatible,
        )


class ConflictChecker:
    """Main class for checking version conflicts."""

    def __init__(self, config: ConflictConfig, logger: LoggerProtocol):
        self.config = config
        self.logger = logger
        self.yaml_config = Config()
        # Load deps config for version-command lookup
        config_data = self._load_user_config(self.config.config_file)
        deps_config = config_data.get("deps", {})
        self.version_checker = ToolVersionChecker(logger, deps_config)

    def check_conflicts(self) -> List[ConflictCheckResult]:
        """Check for version conflicts."""
        results = []

        try:
            # Load configuration using standardized Config class
            config_data = self._load_user_config(self.config.config_file)

            # Extract version requirements from deps section
            deps = config_data.get("deps", {})

            if not deps:
                self.logger.warning(no_deps_found_warning)
                return results

            # Check version conflicts
            results.extend(self._check_version_conflicts(deps))

        except Exception as e:
            self.logger.error(f"Error loading configuration: {str(e)}")
            results.append(ConflictCheckResult(tool="configuration", status="error", conflict=True, error=str(e)))

        return results

    def _load_user_config(self, config_path: Optional[str]) -> Dict[str, Any]:
        """Load configuration.
        - If config_path is provided, load it as user config (overrides only what it contains).
        - If None, fall back to the built-in config used by Config (same as install command).
        """
        try:
            if not config_path:
                # Built-in config (nested dict) as default
                self.logger.debug("Loading built-in configuration (no --config-file provided)")
                return self.yaml_config.load_yaml_config()

            # Load user config and unflatten to nested structure
            self.logger.debug(conflict_loading_config.format(path=config_path))
            flattened_config = self.yaml_config.load_user_config(config_path)
            self.logger.debug(conflict_config_loaded)
            return self.yaml_config.unflatten_config(flattened_config)

        except FileNotFoundError:
            raise FileNotFoundError(conflict_config_not_found.format(path=config_path))
        except Exception as e:
            raise Exception(conflict_invalid_config.format(error=str(e)))

    def _check_version_conflicts(self, deps: Dict[str, Any]) -> List[ConflictCheckResult]:
        """Check for tool version conflicts from deps configuration."""
        # Extract version requirements from deps
        version_requirements = self._extract_version_requirements(deps)

        if not version_requirements:
            return []

        # Check versions in parallel
        results = ParallelProcessor.process_items(
            items=list(version_requirements.items()),
            processor_func=self._check_tool_version,
            max_workers=min(len(version_requirements), 10),
            error_handler=self._handle_check_error,
        )

        return results

    def _extract_version_requirements(self, deps: Dict[str, Any]) -> Dict[str, Optional[str]]:
        """Extract version requirements from deps configuration."""
        version_requirements = {}

        for tool, config in deps.items():
            if isinstance(config, dict):
                # Only check tools that have a version key (even if empty)
                if "version" in config:
                    version_req = config.get("version", "")
                    version_requirements[tool] = version_req if version_req else None

        return version_requirements

    def _check_tool_version(self, tool_requirement: Tuple[str, Optional[str]]) -> ConflictCheckResult:
        """Check version for a single tool."""
        tool, expected_version = tool_requirement
        return self.version_checker.check_tool_version(tool, expected_version)

    def _handle_check_error(self, tool_requirement: Tuple[str, Optional[str]], error: Exception) -> ConflictCheckResult:
        """Handle errors during version checking."""
        tool, expected_version = tool_requirement
        return ConflictCheckResult(
            tool=tool, expected=expected_version, current=None, status="error", conflict=True, error=str(error)
        )


class ConflictFormatter:
    """Handles formatting of conflict check results."""

    def __init__(self):
        self.output_formatter = OutputFormatter()

    def format_output(self, data: List[ConflictCheckResult], output_type: str) -> str:
        """Format conflict check results."""
        if not data:
            message = self.output_formatter.create_success_message(no_version_conflicts_message)
            return self.output_formatter.format_output(message, output_type)

        messages = []
        for result in data:
            data_dict = result.model_dump()
            message = self._format_single_result(result)

            if result.conflict:
                messages.append(self.output_formatter.create_error_message(message, data_dict))
            else:
                messages.append(self.output_formatter.create_success_message(message, data_dict))

        return self.output_formatter.format_output(messages, output_type)

    def _format_single_result(self, result: ConflictCheckResult) -> str:
        """Format a single conflict check result."""
        if result.conflict:
            if result.current is None:
                return f"{result.tool}: {result.status}"
            else:
                return f"{result.tool}: Expected {result.expected}, Found {result.current}"
        else:
            return f"{result.tool}: Version compatible ({result.current})"


class ConflictService:
    """Main service class for conflict checking functionality."""

    def __init__(self, config: ConflictConfig, logger: Optional[LoggerProtocol] = None):
        self.config = config
        self.logger = logger or Logger(verbose=config.verbose)
        self.checker = ConflictChecker(config, self.logger)
        self.formatter = ConflictFormatter()

    def check_conflicts(self) -> List[ConflictCheckResult]:
        """Check for conflicts and return results."""
        self.logger.debug("Starting version conflict checks")
        return self.checker.check_conflicts()

    def check_and_format(self, output_type: Optional[str] = None) -> str:
        """Check conflicts and return formatted output."""
        results = self.check_conflicts()
        output_format = output_type or self.config.output
        return self.formatter.format_output(results, output_format)
