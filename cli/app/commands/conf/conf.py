import os
import shutil
import tempfile
from typing import Dict, Optional

from app.utils.protocols import LoggerProtocol

from .messages import (
    backup_created,
    backup_creation_failed,
    backup_file_not_found,
    backup_remove_failed,
    backup_removed,
    backup_restore_attempt,
    backup_restore_failed,
    backup_restore_success,
    file_write_failed,
)

def _create_backup(file_path: str) -> tuple[bool, Optional[str], Optional[str]]:
    if not os.path.exists(file_path):
        return True, None, None

    try:
        backup_path = f"{file_path}.backup"
        shutil.copy2(file_path, backup_path)
        return True, backup_path, None
    except Exception as e:
        return False, None, backup_creation_failed.format(error=e)


def _restore_backup(backup_path: str, file_path: str, logger: Optional[LoggerProtocol] = None) -> tuple[bool, Optional[str]]:
    if not os.path.exists(backup_path):
        return False, backup_file_not_found.format(path=backup_path)

    try:
        shutil.copy2(backup_path, file_path)
        os.remove(backup_path)
        if logger:
            logger.debug(backup_restore_success)
        return True, None
    except Exception as e:
        error_msg = backup_restore_failed.format(error=e)
        if logger:
            logger.error(error_msg)
        return False, error_msg


def _cleanup_backup(backup_path: Optional[str], logger: Optional[LoggerProtocol] = None) -> None:
    if not backup_path or not os.path.exists(backup_path):
        return

    try:
        os.remove(backup_path)
        if logger:
            logger.debug(backup_removed)
    except Exception as e:
        if logger:
            logger.warning(backup_remove_failed.format(error=e))


def _handle_write_error(
    file_path: str, backup_path: Optional[str], error: str, logger: Optional[LoggerProtocol] = None
) -> tuple[bool, str]:
    if not backup_path:
        return False, error

    if logger:
        logger.warning(backup_restore_attempt)
    restore_success, restore_error = _restore_backup(backup_path, file_path, logger)
    
    if not restore_success:
        return False, restore_error
    
    return False, error


def _atomic_write(file_path: str, config: Dict[str, str]) -> tuple[bool, Optional[str]]:
    temp_path = None
    try:
        os.makedirs(os.path.dirname(file_path), exist_ok=True)

        with tempfile.NamedTemporaryFile(mode="w", delete=False, dir=os.path.dirname(file_path)) as temp_file:
            for key, value in sorted(config.items()):
                temp_file.write(f"{key}={value}\n")
            temp_file.flush()
            try:
                os.fsync(temp_file.fileno())
            except (OSError, AttributeError):
                pass
            temp_path = temp_file.name

        os.replace(temp_path, file_path)
        return True, None
    except Exception as e:
        if temp_path and os.path.exists(temp_path):
            try:
                os.unlink(temp_path)
            except:
                pass
        return False, file_write_failed.format(error=e)


def write_env_file(file_path: str, config: Dict[str, str], logger: Optional[LoggerProtocol] = None) -> tuple[bool, Optional[str]]:
    try:
        success, backup_path, error = _create_backup(file_path)
        if not success:
            return False, error

        if backup_path and logger:
            logger.debug(backup_created.format(backup_path=backup_path))

        success, error = _atomic_write(file_path, config)
        if not success:
            return _handle_write_error(file_path, backup_path, error, logger)

        _cleanup_backup(backup_path, logger)
        return True, None

    except Exception as e:
        return False, file_write_failed.format(error=e)
