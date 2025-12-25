"""Constants and shared utilities for build and package operations."""

from pathlib import Path
from typing import Dict, Optional, Tuple

# Package type constants
PACKAGE_TYPE_DEB = "deb"
PACKAGE_TYPE_RPM = "rpm"
PACKAGE_TYPE_APK = "apk"
PACKAGE_TYPE_TAR = "tar"

# Architecture mappings
ARCH_X86_64 = "x86_64"
ARCH_AMD64 = "amd64"
ARCH_AARCH64 = "aarch64"
ARCH_ARM64 = "arm64"

# Architecture normalization mapping (canonical -> normalized)
ARCH_NORMALIZATION: Dict[str, str] = {
    ARCH_X86_64: ARCH_AMD64,
    ARCH_AMD64: ARCH_AMD64,
    ARCH_AARCH64: ARCH_ARM64,
    ARCH_ARM64: ARCH_ARM64,
}

# RPM architecture mapping (normalized -> RPM)
RPM_ARCH_MAPPING: Dict[str, str] = {
    ARCH_AMD64: ARCH_X86_64,
    ARCH_ARM64: ARCH_AARCH64,
}

# Default values
DEFAULT_ARCH = ARCH_AMD64
DEFAULT_PACKAGE_TYPE = PACKAGE_TYPE_DEB
DEFAULT_DOCKER_IMAGE = "nixopus-cli-builder"
DEFAULT_POETRY_INSTALL_URL = "https://install.python-poetry.org"
DEFAULT_POETRY_HOME = "$HOME/.local/bin"

# Path constants (relative to workspace root)
CLI_DIR_NAME = "cli"
SCRIPTS_DIR_NAME = "scripts"
DIST_DIR_NAME = "dist"
PACKAGING_DIR_NAME = "packaging"
PYPROJECT_TOML_NAME = "pyproject.toml"
BUILD_SCRIPT_NAME = "build.sh"
INSTALL_SCRIPT_NAME = "install.sh"

# File paths within CLI directory
CLI_BINARY_NAME = "nixopus"
CLI_DIST_BINARY_PATH = f"{DIST_DIR_NAME}/{CLI_BINARY_NAME}"
CLI_PACKAGING_BIN_PATH = f"{PACKAGING_DIR_NAME}/usr/local/bin"

# Package naming patterns
PACKAGE_NAME_PATTERNS: Dict[str, str] = {
    PACKAGE_TYPE_DEB: "nixopus_{version}_{arch}.deb",
    PACKAGE_TYPE_RPM: "nixopus-{version}-1.{rpm_arch}.rpm",
    PACKAGE_TYPE_APK: "nixopus-{version}.apk",
    PACKAGE_TYPE_TAR: "nixopus-{version}.tar",
}

# Package file search patterns (for globbing)
PACKAGE_SEARCH_PATTERNS: Dict[str, str] = {
    PACKAGE_TYPE_DEB: "nixopus_*_{arch}.deb",
    PACKAGE_TYPE_RPM: "nixopus-*-1.{rpm_arch}.rpm",
    PACKAGE_TYPE_APK: "nixopus-*.apk",
    PACKAGE_TYPE_TAR: "nixopus-*.tar",
}

# Timeouts (in seconds)
DEFAULT_PACKAGE_MANAGER_WAIT_TIMEOUT = 10
DEFAULT_SUBPROCESS_TIMEOUT = 10
DEFAULT_OUTPUT_WAIT_TIMEOUT = 2

# Version extraction regex pattern
VERSION_EXTRACTION_PATTERN = r'^version\s*=\s*"([^"]+)"'

# Minimum Python version
MIN_PYTHON_VERSION = (3, 9)

# Package manager detection order
PACKAGE_MANAGER_DETECTION_ORDER = [
    ("apt-get", PACKAGE_TYPE_DEB),
    ("apt", PACKAGE_TYPE_DEB),
    ("dnf", PACKAGE_TYPE_RPM),
    ("yum", PACKAGE_TYPE_RPM),
    ("apk", PACKAGE_TYPE_APK),
    ("pacman", PACKAGE_TYPE_TAR),
]

# Package manager to package type mapping
PACKAGE_MANAGER_TO_TYPE: Dict[str, str] = {
    "apt-get": PACKAGE_TYPE_DEB,
    "apt": PACKAGE_TYPE_DEB,
    "dnf": PACKAGE_TYPE_RPM,
    "yum": PACKAGE_TYPE_RPM,
    "apk": PACKAGE_TYPE_APK,
    "pacman": PACKAGE_TYPE_TAR,
}

def normalize_architecture(arch: str) -> str:
    arch_lower = arch.lower().strip()
    return ARCH_NORMALIZATION.get(arch_lower, arch_lower)

def get_rpm_architecture(normalized_arch: str) -> str:
    return RPM_ARCH_MAPPING.get(normalized_arch, normalized_arch)

def get_package_type_from_manager(package_manager: str) -> Optional[str]:
    return PACKAGE_MANAGER_TO_TYPE.get(package_manager.lower())

def get_workspace_root(start_path: Optional[Path] = None) -> Path:
    if start_path is None:
        start_path = Path(__file__).resolve()
    
    current = start_path.resolve()
    
    if current.name == CLI_DIR_NAME or (current.parent.name == CLI_DIR_NAME):
        return current.parent.parent if current.name == CLI_DIR_NAME else current.parent
    
    if current.name == SCRIPTS_DIR_NAME or (current.parent.name == SCRIPTS_DIR_NAME):
        return current.parent.parent if current.name == SCRIPTS_DIR_NAME else current.parent
    
    for parent in current.parents:
        if (parent / CLI_DIR_NAME).exists() and (parent / SCRIPTS_DIR_NAME).exists():
            return parent
    
    return current.parent.parent

