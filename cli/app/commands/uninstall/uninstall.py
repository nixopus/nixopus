import os
import shutil
import typer
from pathlib import Path
from typing import Optional

from app.commands.service.service import stop_services
from app.utils.config import (
    DEFAULT_COMPOSE_FILE,
    NIXOPUS_CONFIG_DIR,
    SSH_FILE_PATH,
    get_active_config,
    get_yaml_value,
)
from app.utils.logger import create_logger
from app.utils.protocols import LoggerProtocol
from app.utils.timeout import timeout_wrapper

from .messages import (
    authorized_keys_not_found,
    compose_file_not_found_skip,
    config_dir_not_exist_skip,
    config_directory_removal_failed,
    failed_at_step,
    operation_timed_out,
    removed_config_dir,
    removed_private_key,
    removed_public_key,
    removed_ssh_key_from,
    services_stop_failed,
    skipped_removal_config_dir,
    ssh_key_not_found_in_authorized_keys,
    ssh_keys_removal_failed,
    ssh_public_key_not_found_skip,
    uninstall_completed,
    uninstall_completed_info,
    uninstall_dry_run_mode,
    uninstall_failed,
    uninstall_thank_you,
)

_config = get_active_config()
_config_dir = get_yaml_value(_config, NIXOPUS_CONFIG_DIR)
_compose_file = get_yaml_value(_config, DEFAULT_COMPOSE_FILE)
_ssh_key_path = _config_dir + "/" + get_yaml_value(_config, SSH_FILE_PATH)


def _get_compose_file_path() -> str:
    """Get the full path to the compose file."""
    return os.path.join(_config_dir, _compose_file)


def _get_authorized_keys_path() -> Path:
    """Get the path to authorized_keys file."""
    return Path.home() / ".ssh" / "authorized_keys"


def _read_public_key(public_key_path: Path) -> Optional[str]:
    """Read public key content from file."""
    if not public_key_path.exists():
        return None
    try:
        with open(public_key_path, "r") as f:
            return f.read().strip()
    except Exception:
        return None


def _remove_key_from_authorized_keys(authorized_keys_path: Path, public_key_content: str, logger: Optional[LoggerProtocol] = None) -> bool:
    """Remove SSH key from authorized_keys file."""
    if not authorized_keys_path.exists():
        if logger:
            logger.debug(authorized_keys_not_found)
        return False

    try:
        with open(authorized_keys_path, "r") as f:
            lines = f.readlines()

        original_count = len(lines)
        filtered_lines = [line for line in lines if public_key_content not in line]

        if len(filtered_lines) < original_count:
            with open(authorized_keys_path, "w") as f:
                f.writelines(filtered_lines)
            if logger:
                logger.debug(removed_ssh_key_from.format(authorized_keys_path=authorized_keys_path))
            return True
        else:
            if logger:
                logger.debug(ssh_key_not_found_in_authorized_keys)
            return False
    except Exception:
        return False


def _remove_file_if_exists(file_path: Path, logger: Optional[LoggerProtocol] = None, debug_message: str = None):
    """Remove file if it exists."""
    if file_path.exists():
        file_path.unlink()
        if logger and debug_message:
            logger.debug(debug_message)


def stop_services_step(
    compose_file_path: str,
    timeout: int,
    logger: Optional[LoggerProtocol] = None,
) -> tuple[bool, Optional[str]]:
    """Stop Docker services."""
    if not os.path.exists(compose_file_path):
        if logger:
            logger.debug(compose_file_not_found_skip.format(compose_file_path=compose_file_path))
        return True, None

    try:
        with timeout_wrapper(timeout):
            success, error = stop_services(
                name="all",
                env_file=None,
                compose_file=compose_file_path,
                logger=logger,
            )
            if not success:
                return False, f"{services_stop_failed}: {error}"
            return True, None
    except TimeoutError:
        return False, f"{services_stop_failed}: {operation_timed_out}"


def remove_ssh_keys_step(
    ssh_key_path: str,
    logger: Optional[LoggerProtocol] = None,
) -> tuple[bool, Optional[str]]:
    """Remove SSH keys from system."""
    ssh_key_file = Path(ssh_key_path)
    public_key_file = ssh_key_file.with_suffix(".pub")

    if not public_key_file.exists():
        if logger:
            logger.debug(ssh_public_key_not_found_skip.format(public_key_path=public_key_file))
        return True, None

    try:
        public_key_content = _read_public_key(public_key_file)
        if public_key_content:
            authorized_keys_path = _get_authorized_keys_path()
            _remove_key_from_authorized_keys(authorized_keys_path, public_key_content, logger)

        _remove_file_if_exists(ssh_key_file, logger, removed_private_key.format(ssh_key_path=ssh_key_file))
        _remove_file_if_exists(public_key_file, logger, removed_public_key.format(public_key_path=public_key_file))

        return True, None
    except Exception as e:
        return False, f"{ssh_keys_removal_failed}: {str(e)}"


def remove_config_directory_step(
    config_dir_path: Path,
    force: bool,
    logger: Optional[LoggerProtocol] = None,
) -> tuple[bool, Optional[str]]:
    """Remove configuration directory."""
    if not config_dir_path.exists():
        if logger:
            logger.debug(config_dir_not_exist_skip.format(config_dir_path=config_dir_path))
        return True, None

    try:
        if force:
            shutil.rmtree(config_dir_path)
            if logger:
                logger.debug(removed_config_dir.format(config_dir_path=config_dir_path))
        else:
            if logger:
                logger.info(skipped_removal_config_dir.format(config_dir_path=config_dir_path))
        return True, None
    except Exception as e:
        return False, f"{config_directory_removal_failed}: {str(e)}"


def uninstall(
    logger: Optional[LoggerProtocol] = None,
    timeout: int = 300,
    dry_run: bool = False,
    force: bool = False,
) -> tuple[bool, Optional[str]]:
    """Uninstall Nixopus from the system."""
    if dry_run:
        if logger:
            logger.info(uninstall_dry_run_mode)
            logger.info("Would execute: Stopping services")
            logger.info("Would execute: Removing SSH keys")
            logger.info("Would execute: Removing configuration directory")
        return True, None

    steps = [
        ("Stopping services", lambda: stop_services_step(_get_compose_file_path(), timeout, logger)),
        ("Removing SSH keys", lambda: remove_ssh_keys_step(_ssh_key_path, logger)),
        ("Removing configuration directory", lambda: remove_config_directory_step(Path(_config_dir), force, logger)),
    ]

    for step_name, step_func in steps:
        success, error = step_func()
        if not success:
            error_msg = f"{failed_at_step.format(step_name=step_name)}: {error}" if error else failed_at_step.format(step_name=step_name)
            return False, error_msg

    if logger:
        logger.success(uninstall_completed)
        logger.info(uninstall_completed_info)
        logger.info(uninstall_thank_you)

    return True, None


uninstall_app = typer.Typer(help="Uninstall Nixopus", invoke_without_command=True)


@uninstall_app.callback()
def uninstall_callback(
    ctx: typer.Context,
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Show more details while uninstalling"),
    timeout: int = typer.Option(300, "--timeout", "-t", help="How long to wait for each step (in seconds)"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="See what would happen, but don't make changes"),
    force: bool = typer.Option(False, "--force", "-f", help="Remove files without confirmation prompts"),
):
    """Uninstall Nixopus completely from the system"""
    if ctx.invoked_subcommand is None:
        logger = create_logger(verbose=verbose)

        if dry_run:
            success, error = uninstall(logger=logger, timeout=timeout, dry_run=dry_run, force=force)
            if not success:
                logger.error(f"{uninstall_failed}: {error}")
                raise typer.Exit(1)
            return

        try:
            success, error = uninstall(logger=logger, timeout=timeout, dry_run=dry_run, force=force)
            if not success:
                logger.error(f"{uninstall_failed}: {error}")
                raise typer.Exit(1)
        except Exception as e:
            logger.error(f"{uninstall_failed}: {str(e)}")
            raise typer.Exit(1)

