import json
import os
import platform
import subprocess
import tempfile
import typer
import urllib.request
from typing import Optional

from app.commands.service.service import execute_services, start_services
from app.utils.config import DEFAULT_COMPOSE_FILE, NIXOPUS_CONFIG_DIR, get_active_config, get_yaml_value
from app.utils.logger import create_logger
from app.utils.protocols import LoggerProtocol

from .messages import (
    failed_to_pull_images,
    failed_to_start_services,
    images_pulled_successfully,
    nixopus_updated_successfully,
    pulling_latest_images,
    starting_services,
    updating_nixopus,
)

PACKAGE_JSON_URL = "https://raw.githubusercontent.com/raghavyuva/nixopus/master/package.json"
RELEASES_BASE_URL = "https://github.com/raghavyuva/nixopus/releases/download"


def _get_compose_file_path() -> str:
    """Get the full path to the compose file."""
    config = get_active_config()
    compose_file = get_yaml_value(config, DEFAULT_COMPOSE_FILE)
    config_dir = get_yaml_value(config, NIXOPUS_CONFIG_DIR)
    return f"{config_dir}/{compose_file}"


def _detect_arch() -> str:
    """Detect system architecture."""
    arch = platform.machine().lower()
    if arch in ["x86_64", "amd64"]:
        return "amd64"
    elif arch in ["aarch64", "arm64"]:
        return "arm64"
    else:
        raise Exception(f"Unsupported architecture: {arch}")


def _detect_os() -> str:
    """Detect operating system and package type."""
    system = platform.system().lower()
    if system == "darwin":
        return "tar"
    elif system == "linux":
        try:
            subprocess.run(["apt", "--version"], capture_output=True, check=True)
            return "deb"
        except:
            try:
                subprocess.run(["yum", "--version"], capture_output=True, check=True)
                return "rpm"
            except:
                try:
                    subprocess.run(["apk", "--version"], capture_output=True, check=True)
                    return "apk"
                except:
                    return "tar"
    else:
        return "tar"


def _build_package_name(version: str, arch: str, pkg_type: str) -> str:
    """Build package name based on version, architecture, and package type."""
    if pkg_type == "deb":
        return f"nixopus_{version}_{arch}.deb"
    elif pkg_type == "rpm":
        arch_name = "x86_64" if arch == "amd64" else "aarch64"
        return f"nixopus-{version}-1.{arch_name}.rpm"
    elif pkg_type == "apk":
        return f"nixopus_{version}_{arch}.apk"
    elif pkg_type == "tar":
        return f"nixopus-{version}.tar"
    else:
        raise Exception(f"Unknown package type: {pkg_type}")


def _install_tar_package(package_path: str) -> None:
    """Install tar package."""
    subprocess.run(["tar", "-xf", package_path, "-C", "/tmp"], check=True)
    
    try:
        subprocess.run(["cp", "/tmp/usr/local/bin/nixopus", "/usr/local/bin/"], check=True)
        subprocess.run(["chmod", "+x", "/usr/local/bin/nixopus"], check=True)
    except subprocess.CalledProcessError:
        subprocess.run(["sudo", "mkdir", "-p", "/usr/local/bin"], check=True)
        subprocess.run(["sudo", "cp", "/tmp/usr/local/bin/nixopus", "/usr/local/bin/"], check=True)
        subprocess.run(["sudo", "chmod", "+x", "/usr/local/bin/nixopus"], check=True)


def _install_system_package(package_path: str, pkg_type: str) -> None:
    """Install system package (deb, rpm, or apk)."""
    if pkg_type == "deb":
        subprocess.run(["sudo", "dpkg", "-i", package_path], check=True)
        subprocess.run(["sudo", "apt-get", "install", "-f", "-y"], check=True)
    elif pkg_type == "rpm":
        try:
            subprocess.run(["sudo", "dnf", "install", "-y", package_path], check=True)
        except subprocess.CalledProcessError:
            subprocess.run(["sudo", "yum", "install", "-y", package_path], check=True)
    elif pkg_type == "apk":
        subprocess.run(["sudo", "apk", "add", "--allow-untrusted", package_path], check=True)


def update(
    logger: Optional[LoggerProtocol] = None,
) -> tuple[bool, Optional[str]]:
    """Update Nixopus services."""
    compose_file_path = _get_compose_file_path()
    
    if logger:
        logger.info(updating_nixopus)
        logger.debug(pulling_latest_images)
    
    success, output = execute_services("pull", compose_file=compose_file_path, logger=logger)
    
    if not success:
        error_msg = failed_to_pull_images.format(error=output)
        if logger:
            logger.error(error_msg)
        return False, error_msg
    
    if logger:
        logger.debug(images_pulled_successfully)
        logger.debug(starting_services)
    
    success, output = start_services(compose_file=compose_file_path, detach=True, logger=logger)
    
    if not success:
        error_msg = failed_to_start_services.format(error=output)
        if logger:
            logger.error(error_msg)
        return False, error_msg
    
    if logger:
        logger.info(nixopus_updated_successfully)
    
    return True, None


def update_cli(logger: Optional[LoggerProtocol] = None) -> tuple[bool, Optional[str]]:
    """Update CLI tool."""
    if logger:
        logger.info("Updating CLI tool...")
    
    try:
        arch = _detect_arch()
        pkg_type = _detect_os()
        
        with urllib.request.urlopen(PACKAGE_JSON_URL) as response:
            package_json = json.loads(response.read().decode())
        
        cli_version = package_json.get("cli-version")
        if not cli_version:
            error_msg = "Could not find cli-version in package.json"
            if logger:
                logger.error(error_msg)
            return False, error_msg
        
        cli_packages = package_json.get("cli-packages", [])
        package_name = _build_package_name(cli_version, arch, pkg_type)
        
        if package_name not in cli_packages:
            error_msg = f"Package {package_name} not found in available packages"
            if logger:
                logger.error(error_msg)
            return False, error_msg
        
        download_url = f"{RELEASES_BASE_URL}/nixopus-{cli_version}/{package_name}"
        
        with tempfile.NamedTemporaryFile(delete=False) as temp_file:
            urllib.request.urlretrieve(download_url, temp_file.name)
            
            if pkg_type == "tar":
                _install_tar_package(temp_file.name)
            else:
                _install_system_package(temp_file.name, pkg_type)
            
            os.unlink(temp_file.name)
        
        if logger:
            logger.info("CLI tool updated successfully")
        
        return True, None
        
    except Exception as e:
        error_msg = f"Failed to update CLI tool: {str(e)}"
        if logger:
            logger.error(error_msg)
        return False, error_msg


update_app = typer.Typer(help="Update Nixopus", invoke_without_command=True)


@update_app.callback()
def update_callback(
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Show more details while updating"),
):
    """Update Nixopus"""
    logger = create_logger(verbose=verbose)
    success, error = update(logger=logger)
    if not success:
        raise typer.Exit(1)


@update_app.command()
def cli(
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Show more details while updating"),
):
    """Update CLI tool"""
    logger = create_logger(verbose=verbose)
    success, error = update_cli(logger=logger)
    if not success:
        raise typer.Exit(1)

