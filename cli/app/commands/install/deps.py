import subprocess
import json
from app.utils.config import Config
from app.utils.lib import HostInformation
from app.utils.logger import Logger
from app.commands.preflight.deps import Deps, DepsConfig
from app.utils.config import DEPS
from .messages import (
    unsupported_package_manager,
    no_supported_package_manager,
    failed_to_install,
    installing_dep,
    dry_run_update_cmd,
    dry_run_install_cmd,
)

def get_deps_from_config():
    config = Config()
    deps = config.get_yaml_value(DEPS)
    return list(deps.keys())

def get_installed_deps(dep_names, os_name, package_manager, timeout=2, verbose=False):
    config = DepsConfig(
        deps=list(dep_names),
        timeout=timeout,
        verbose=verbose,
        output="json",
        os=os_name,
        package_manager=package_manager,
    )
    deps_checker = Deps()
    results = deps_checker.check(config)
    return {r.dependency: r.is_available for r in results}

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
        logger.info(dry_run_update_cmd.format(cmd=' '.join(cmd)))
    else:
        subprocess.check_call(cmd)

def install_dep(dep, package_manager, logger, dry_run=False):
    try:
        if package_manager == "apt":
            cmd = ["sudo", "apt-get", "install", "-y", dep]
        elif package_manager == "brew":
            cmd = ["brew", "install", dep]
        elif package_manager == "apk":
            cmd = ["sudo", "apk", "add", dep]
        elif package_manager == "yum":
            cmd = ["sudo", "yum", "install", "-y", dep]
        elif package_manager == "dnf":
            cmd = ["sudo", "dnf", "install", "-y", dep]
        elif package_manager == "pacman":
            cmd = ["sudo", "pacman", "-S", "--noconfirm", dep]
        else:
            raise Exception(unsupported_package_manager.format(package_manager=package_manager))
        logger.info(installing_dep.format(dep=dep))
        if dry_run:
            logger.info(dry_run_install_cmd.format(cmd=' '.join(cmd)))
            return True
        subprocess.check_call(cmd)
        return True
    except Exception as e:
        logger.error(failed_to_install.format(dep=dep, error=e))
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
    to_install = [dep for dep in deps if not installed.get(dep)]
    results = []
    for dep in to_install:
        ok = install_dep(dep, package_manager, logger, dry_run=dry_run)
        results.append({"dependency": dep, "installed": ok})
    installed_after = get_installed_deps(deps, os_name, package_manager, verbose=verbose)
    failed = [dep for dep, ok in installed_after.items() if not ok]
    if failed and not dry_run:
        raise Exception(failed_to_install.format(dep=','.join(failed), error=''))
    if output == "json":
        return json.dumps({"installed": results, "failed": failed, "dry_run": dry_run})
    return True
