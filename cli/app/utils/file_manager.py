import os
import stat
from typing import Optional, Tuple


def set_permissions(file_path: str, mode: int, logger=None) -> Tuple[bool, Optional[str]]:
    """Set file permissions"""
    try:
        if logger:
            logger.debug(f"Setting permissions {oct(mode)} on {file_path}")

        os.chmod(file_path, mode)

        if logger:
            logger.debug("File permissions set successfully")
        return True, None
    except Exception as e:
        error_msg = f"Failed to set permissions on {file_path}: {e}"
        if logger:
            logger.error(error_msg)
        return False, error_msg


def append_to_file(
    file_path: str, content: str, mode: int = stat.S_IRUSR | stat.S_IWUSR | stat.S_IRGRP | stat.S_IROTH, logger=None
) -> Tuple[bool, Optional[str]]:
    """Append content to a file and set permissions"""
    try:
        with open(file_path, "a") as f:
            f.write(f"\n{content}\n")

        set_permissions(file_path, mode, logger)

        if logger:
            logger.debug(f"Content appended to {file_path}")
        return True, None
    except Exception as e:
        error_msg = f"Failed to append to {file_path}: {e}"
        if logger:
            logger.error(error_msg)
        return False, error_msg


def read_file_content(file_path: str, logger=None) -> Tuple[bool, Optional[str], Optional[str]]:
    """Read file content"""
    try:
        with open(file_path, "r") as f:
            content = f.read().strip()
        return True, content, None
    except Exception as e:
        error_msg = f"Failed to read {file_path}: {e}"
        if logger:
            logger.error(error_msg)
        return False, None, error_msg


def expand_user_path(path: str) -> str:
    """Expand user home directory in path"""
    return os.path.expanduser(path)


def get_directory_path(file_path: str) -> str:
    """Get the directory path from a file path"""
    return os.path.dirname(file_path)


def get_public_key_path(private_key_path: str) -> str:
    """Get the public key path from a private key path"""
    return f"{private_key_path}.pub"

