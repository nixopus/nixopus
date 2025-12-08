import os
import shutil
import stat
from typing import Optional

from app.utils.message import FAILED_TO_REMOVE_DIRECTORY_MESSAGE, REMOVED_DIRECTORY_MESSAGE


def path_exists(path: str) -> bool:
    """Check if a path exists"""
    return os.path.exists(path)


def path_exists_and_not_force(path: str, force: bool) -> bool:
    """Check if path exists and force is not enabled"""
    return os.path.exists(path) and not force


def remove_directory(path: str, logger=None) -> bool:
    """Remove a directory and all its contents"""
    if logger:
        logger.debug(f"Attempting to remove directory: {path}")
        logger.debug(f"Directory exists: {os.path.exists(path)}")
        logger.debug(f"Directory is directory: {os.path.isdir(path) if os.path.exists(path) else 'N/A'}")

    try:
        shutil.rmtree(path)
        if logger:
            logger.debug(REMOVED_DIRECTORY_MESSAGE.format(path=path))
            logger.debug(f"Directory {path} removed successfully")
        return True
    except Exception as e:
        if logger:
            logger.debug(f"Exception during directory removal: {type(e).__name__}: {str(e)}")
            logger.error(FAILED_TO_REMOVE_DIRECTORY_MESSAGE.format(path=path, error=e))
        return False


def create_directory(
    path: str, mode: int = stat.S_IRUSR | stat.S_IWUSR | stat.S_IXUSR, logger=None
) -> tuple[bool, Optional[str]]:
    """Create a directory with the specified mode"""
    try:
        if not os.path.exists(path):
            os.makedirs(path, mode=mode)
            if logger:
                logger.debug(f"Created directory: {path}")
        return True, None
    except Exception as e:
        error_msg = f"Failed to create directory {path}: {e}"
        if logger:
            logger.error(error_msg)
        return False, error_msg

