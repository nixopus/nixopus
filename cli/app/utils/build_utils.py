import platform
import re
import subprocess
from pathlib import Path
from typing import Optional, Tuple

from app.utils.build_constants import (
    ARCH_AMD64,
    ARCH_ARM64,
    DEFAULT_ARCH,
    DEFAULT_PACKAGE_TYPE,
    PACKAGE_MANAGER_DETECTION_ORDER,
    PACKAGE_TYPE_APK,
    PACKAGE_TYPE_DEB,
    PACKAGE_TYPE_RPM,
    PACKAGE_TYPE_TAR,
    VERSION_EXTRACTION_PATTERN,
    get_package_type_from_manager,
    get_rpm_architecture,
    normalize_architecture,
)


def detect_architecture() -> str:
    arch = platform.machine()
    return normalize_architecture(arch)


def detect_architecture_from_uname(uname_output: str) -> str:
    arch = uname_output.strip()
    return normalize_architecture(arch)


def detect_package_type() -> str:
    import shutil
    
    for manager, pkg_type in PACKAGE_MANAGER_DETECTION_ORDER:
        if shutil.which(manager):
            return pkg_type
    
    return DEFAULT_PACKAGE_TYPE


def detect_package_type_from_container(container_name: str, timeout: int = 10) -> Tuple[bool, Optional[str], Optional[str]]:
    detection_script = (
        "if command -v apt-get >/dev/null 2>&1; then echo deb; "
        "elif command -v dnf >/dev/null 2>&1; then echo rpm; "
        "elif command -v yum >/dev/null 2>&1; then echo rpm; "
        "elif command -v apk >/dev/null 2>&1; then echo apk; "
        "else echo tar; fi"
    )
    
    try:
        result = subprocess.run(
            ["lxc", "exec", container_name, "--", "bash", "-c", detection_script],
            capture_output=True,
            text=True,
            check=True,
            timeout=timeout,
        )
        pkg_type = result.stdout.strip()
        
        # Validate package type
        if pkg_type not in [PACKAGE_TYPE_DEB, PACKAGE_TYPE_RPM, PACKAGE_TYPE_APK, PACKAGE_TYPE_TAR]:
            return False, None, f"Unknown package type detected: {pkg_type}"
        
        return True, pkg_type, None
    except subprocess.CalledProcessError as e:
        error_msg = e.stderr if e.stderr else str(e)
        return False, None, f"Failed to detect package type: {error_msg}"
    except subprocess.TimeoutExpired:
        return False, None, f"Timeout detecting package type (exceeded {timeout}s)"
    except FileNotFoundError:
        return False, None, "lxc command not found. Is LXD installed?"


def extract_version_from_pyproject(pyproject_path: Path) -> Tuple[bool, Optional[str], Optional[str]]:
    if not pyproject_path.exists():
        return False, None, f"pyproject.toml not found at {pyproject_path}"
    
    try:
        content = pyproject_path.read_text(encoding="utf-8")
        pattern = re.compile(VERSION_EXTRACTION_PATTERN, re.MULTILINE)
        match = pattern.search(content)
        
        if match:
            version = match.group(1).strip()
            if version:
                return True, version, None
            return False, None, "Version found but empty"
        
        return False, None, "Version not found in pyproject.toml"
    except Exception as e:
        return False, None, f"Error reading pyproject.toml: {str(e)}"


def get_container_architecture(container_name: str, timeout: int = 10) -> Tuple[bool, Optional[str], Optional[str]]:
    try:
        result = subprocess.run(
            ["lxc", "exec", container_name, "--", "uname", "-m"],
            capture_output=True,
            text=True,
            check=True,
            timeout=timeout,
        )
        arch = detect_architecture_from_uname(result.stdout)
        return True, arch, None
    except subprocess.CalledProcessError as e:
        error_msg = e.stderr if e.stderr else str(e)
        return False, None, f"Failed to get container architecture: {error_msg}"
    except subprocess.TimeoutExpired:
        return False, None, f"Timeout getting container architecture (exceeded {timeout}s)"
    except FileNotFoundError:
        return False, None, "lxc command not found. Is LXD installed?"


def validate_file_exists(file_path: Path, description: str = "File") -> Tuple[bool, Optional[str]]:
    if not file_path.exists():
        return False, f"{description} not found: {file_path}"
    if not file_path.is_file():
        return False, f"{description} is not a file: {file_path}"
    return True, None


def validate_directory_exists(dir_path: Path, description: str = "Directory") -> Tuple[bool, Optional[str]]:
    if not dir_path.exists():
        return False, f"{description} not found: {dir_path}"
    if not dir_path.is_dir():
        return False, f"{description} is not a directory: {dir_path}"
    return True, None


def check_docker_available() -> Tuple[bool, Optional[str]]:
    try:
        result = subprocess.run(
            ["docker", "version"],
            capture_output=True,
            text=True,
            check=True,
            timeout=5,
        )
        return True, None
    except subprocess.CalledProcessError:
        return False, "Docker command failed. Is Docker installed and running?"
    except FileNotFoundError:
        return False, "Docker command not found. Is Docker installed?"
    except subprocess.TimeoutExpired:
        return False, "Docker command timed out. Is Docker daemon running?"


def check_docker_image_exists(image_name: str) -> Tuple[bool, Optional[str]]:   
    try:
        result = subprocess.run(
            ["docker", "image", "inspect", image_name],
            capture_output=True,
            text=True,
            check=True,
        )
        return True, None
    except subprocess.CalledProcessError:
        return False, f"Docker image not found: {image_name}"
    except FileNotFoundError:
        return False, "Docker command not found. Is Docker installed?"
    except Exception as e:
        return False, f"Error checking Docker image: {str(e)}"

