import json
import shutil
import subprocess

from app.utils.config import DEPS, get_active_config, get_yaml_value
from app.utils.host_information import get_os_name, get_package_manager
from app.utils.parallel_processor import process_parallel
from app.utils.logger import create_logger

from .messages import (
    dry_run_install_cmd,
    dry_run_update_cmd,
    failed_to_install,
    installing_dep,
    no_supported_package_manager,
    unsupported_package_manager,
)


def get_deps_from_config():
    config = get_active_config()
    deps = get_yaml_value(config, DEPS)
    return [
        {
            "name": name,
            "package": dep.get("package", name),
            "command": dep.get("command", ""),
            "install_command": dep.get("install_command", ""),
        }
        for name, dep in deps.items()
    ]


def get_installed_deps(deps, os_name, package_manager, timeout=2, verbose=False):
    checker = DependencyChecker(create_logger(verbose=verbose))
    return {dep["name"]: checker.check_dependency(dep, package_manager) for dep in deps}

def update_system_packages(package_manager, logger, dry_run=False):
    if package_manager == "apt":
        cmd = ["sudo", "apt-get", "update"]
    elif package_manager == "brew":
        cmd = ["brew", "update"]
    elif package_manager == "apk":
        cmd = ["sudo", "apk", "update"]
    elif package_manager == "yum":
        cmd = ["sudo", "yum", "update"]
    elif package_manager == "dnf":
        cmd = ["sudo", "dnf", "update"]
    elif package_manager == "pacman":
        cmd = ["sudo", "pacman", "-Sy"]
    else:
        raise Exception(unsupported_package_manager.format(package_manager=package_manager))
    if dry_run:
        logger.info(dry_run_update_cmd.format(cmd=" ".join(cmd)))
    else:
        subprocess.check_call(cmd, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)


def install_dep(dep, package_manager, logger, dry_run=False):
    package = dep["package"]
    install_command = dep.get("install_command", "")
    try:
        if install_command:
            if dry_run:
                logger.info(f"[DRY RUN] Would run: {install_command}")
                return True
            result = subprocess.run(install_command, shell=True, capture_output=True, text=True)
            if result.returncode != 0:
                error_output = result.stderr.strip() or result.stdout.strip()
                if error_output:
                    logger.error(f"Installation command output: {error_output}")
                raise subprocess.CalledProcessError(result.returncode, install_command, result.stdout, result.stderr)
            return True
        
        # Check if package is already installed before attempting installation
        if package:
            checker = DependencyChecker(logger)
            if checker._check_package_installed(package, package_manager):
                if logger:
                    logger.info(f"Package {package} is already installed, skipping installation")
                return True
        
        if package_manager == "apt":
            cmd = ["sudo", "apt-get", "install", "-y", package]
        elif package_manager == "brew":
            cmd = ["brew", "install", package]
        elif package_manager == "apk":
            cmd = ["sudo", "apk", "add", "--no-cache", package]
        elif package_manager == "yum":
            cmd = ["sudo", "yum", "install", "-y", package]
        elif package_manager == "dnf":
            cmd = ["sudo", "dnf", "install", "-y", package]
        elif package_manager == "pacman":
            cmd = ["sudo", "pacman", "-S", "--noconfirm", package]
        else:
            raise Exception(unsupported_package_manager.format(package_manager=package_manager))
        logger.info(installing_dep.format(dep=package))
        if dry_run:
            logger.info(dry_run_install_cmd.format(cmd=" ".join(cmd)))
            return True
        # Allow non-zero exit codes for already-installed packages
        result = subprocess.run(cmd, stdout=subprocess.DEVNULL, stderr=subprocess.PIPE, text=True)
        if result.returncode != 0:
            # Check if the error is due to package already being installed
            error_msg = result.stderr.lower()
            if any(phrase in error_msg for phrase in ["already installed", "is already the newest", "nothing to do"]):
                if logger:
                    logger.info(f"Package {package} is already installed")
                return True
            # For other errors, raise the exception
            raise subprocess.CalledProcessError(result.returncode, cmd, "", result.stderr)
        return True
    except subprocess.CalledProcessError as e:
        error_msg = str(e)
        if "docker" in package.lower() and (e.returncode == 100 or "permission" in error_msg.lower()):
            logger.error(failed_to_install.format(dep=package, error=f"Exit code {e.returncode}"))
            logger.error("Docker installation requires root privileges.")
            logger.error("Please run: sudo nixopus install")
        else:
            logger.error(failed_to_install.format(dep=package, error=e))
        return False
    except Exception as e:
        logger.error(failed_to_install.format(dep=package, error=e))
        return False


class DependencyChecker:
    def __init__(self, logger=None):
        self.logger = logger

    def _check_package_installed(self, package, package_manager):
        """Check if a package is installed via the package manager."""
        try:
            if package_manager == "apt":
                # Use dpkg-query for more reliable checking
                result = subprocess.run(
                    ["dpkg-query", "-W", "-f=${Status}", package],
                    capture_output=True,
                    text=True,
                    timeout=5
                )
                # Check if package is installed (status contains "install ok installed")
                return result.returncode == 0 and "install ok installed" in result.stdout
            elif package_manager == "yum" or package_manager == "dnf":
                result = subprocess.run(
                    ["rpm", "-q", package],
                    capture_output=True,
                    text=True,
                    timeout=5
                )
                return result.returncode == 0
            elif package_manager == "apk":
                result = subprocess.run(
                    ["apk", "info", "-e", package],
                    capture_output=True,
                    text=True,
                    timeout=5
                )
                return result.returncode == 0
            elif package_manager == "pacman":
                result = subprocess.run(
                    ["pacman", "-Q", package],
                    capture_output=True,
                    text=True,
                    timeout=5
                )
                return result.returncode == 0
            elif package_manager == "brew":
                result = subprocess.run(
                    ["brew", "list", package],
                    capture_output=True,
                    text=True,
                    timeout=5
                )
                return result.returncode == 0
            return False
        except (subprocess.TimeoutExpired, subprocess.CalledProcessError, FileNotFoundError):
            return False

    def check_dependency(self, dep, package_manager):
        """Check if a dependency is installed by checking both command and package."""
        try:
            # First check if command exists (if specified)
            if dep.get("command"):
                is_available = shutil.which(dep["command"]) is not None
                if is_available:
                    return True
            
            # If command check failed or no command specified, check package installation
            if dep.get("package"):
                package_installed = self._check_package_installed(dep["package"], package_manager)
                if package_installed:
                    return True
            
            # If no command or package specified, assume it's available
            if not dep.get("command") and not dep.get("package"):
                return True
            
            return False
        except Exception:
            return False


def install_all_deps(verbose=False, output="text", dry_run=False):
    logger = create_logger(verbose=verbose)
    deps = get_deps_from_config()
    os_name = get_os_name()
    package_manager = get_package_manager()
    if not package_manager:
        raise Exception(no_supported_package_manager)
    installed = get_installed_deps(deps, os_name, package_manager, verbose=verbose)
    update_system_packages(package_manager, logger, dry_run=dry_run)
    to_install = [dep for dep in deps if not installed.get(dep["name"])]

    def install_wrapper(dep):
        ok = install_dep(dep, package_manager, logger, dry_run=dry_run)
        return {"dependency": dep["name"], "installed": ok}

    def error_handler(dep, exc):
        logger.error(f"Failed to install {dep['name']}: {exc}")
        return {"dependency": dep["name"], "installed": False}

    # For package managers that use locks (apt, yum, dnf, pacman), install sequentially
    # to avoid lock conflicts. This is especially important in containers.
    # Brew and apk can run in parallel.
    max_workers = 1 if package_manager in ["apt", "yum", "dnf", "pacman"] else min(len(to_install), 8)
    
    results = process_parallel(
        to_install,
        install_wrapper,
        max_workers=max_workers,
        error_handler=error_handler,
    )

    installed_after = get_installed_deps(deps, os_name, package_manager, verbose=verbose)
    failed = [dep["name"] for dep in deps if not installed_after.get(dep["name"])]
    if failed and not dry_run:
        raise Exception(failed_to_install.format(dep=",".join(failed), error=""))
    if output == "json":
        return json.dumps({"installed": results, "failed": failed, "dry_run": dry_run})
    return True
