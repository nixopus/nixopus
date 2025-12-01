import json
import shutil
import subprocess

from app.utils.config import DEPS, Config
from app.utils.lib import HostInformation, ParallelProcessor
from app.utils.logger import Logger

from .messages import (
    dry_run_install_cmd,
    dry_run_update_cmd,
    failed_to_install,
    installing_dep,
    no_supported_package_manager,
    unsupported_package_manager,
)


def get_deps_from_config():
    config = Config()
    deps = config.get_yaml_value(DEPS)
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
    checker = DependencyChecker(Logger(verbose=verbose))
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
        if package_manager == "apt":
            cmd = ["sudo", "apt-get", "install", "-y", package]
        elif package_manager == "brew":
            cmd = ["brew", "install", package]
        elif package_manager == "apk":
            cmd = ["sudo", "apk", "add", package]
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
        subprocess.check_call(cmd, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
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

    def check_dependency(self, dep, package_manager):
        try:
            if dep["command"]:
                is_available = shutil.which(dep["command"]) is not None
                return is_available
            return True
        except Exception:
            return False


def install_all_deps(verbose=False, output="text", dry_run=False):
    logger = Logger(verbose=verbose)
    deps = get_deps_from_config()
    os_name = HostInformation.get_os_name()
    package_manager = HostInformation.get_package_manager()
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

    results = ParallelProcessor.process_items(
        to_install,
        install_wrapper,
        max_workers=min(len(to_install), 8),
        error_handler=error_handler,
    )

    installed_after = get_installed_deps(deps, os_name, package_manager, verbose=verbose)
    failed = [dep["name"] for dep in deps if not installed_after.get(dep["name"])]
    if failed and not dry_run:
        raise Exception(failed_to_install.format(dep=",".join(failed), error=""))
    if output == "json":
        return json.dumps({"installed": results, "failed": failed, "dry_run": dry_run})
    return True
