import os
import subprocess
from typing import Optional

from app.utils.directory_manager import path_exists_and_not_force, remove_directory
from app.utils.protocols import LoggerProtocol

from .messages import (
    debug_clone_completed,
    debug_cloning_repo,
    debug_directory_removal_failed,
    debug_executing_git_clone,
    debug_git_clone_failed,
    debug_git_clone_success,
    debug_path_exists_force_disabled,
    debug_removing_directory,
    debug_unexpected_error,
    failed_to_prepare_target_directory,
    invalid_path,
    invalid_repo,
    invalid_repository_url,
    path_already_exists_use_force,
    prerequisites_validation_failed,
)


def _is_valid_repo_format(repo: str) -> bool:
    """Validate repository URL format."""
    return (
        repo.startswith(("http://", "https://", "git://", "ssh://"))
        or (repo.endswith(".git") and not repo.startswith("github.com:"))
        or ("@" in repo and ":" in repo and repo.count("@") == 1)
    )


def _validate_repo(repo: str) -> str:
    """Validate and normalize repository URL."""
    stripped_repo = repo.strip()
    if not stripped_repo:
        raise ValueError(invalid_repo)
    if not _is_valid_repo_format(stripped_repo):
        raise ValueError(invalid_repository_url)
    return stripped_repo


def _validate_path(path: str) -> str:
    """Validate and normalize path."""
    stripped_path = path.strip()
    if not stripped_path:
        raise ValueError(invalid_path)
    return stripped_path


def _build_clone_command(repo: str, path: str, branch: Optional[str] = None) -> list[str]:
    """Build git clone command."""
    cmd = ["git", "clone", "--depth=1"]
    if branch:
        cmd.extend(["-b", branch])
    cmd.extend([repo, path])
    return cmd


def _execute_git_clone(
    repo: str, path: str, branch: Optional[str], logger: Optional[LoggerProtocol] = None
) -> tuple[bool, Optional[str]]:
    """Execute git clone command."""
    cmd = _build_clone_command(repo, path, branch)

    if logger:
        logger.debug(debug_executing_git_clone.format(command=" ".join(cmd)))

    try:
        subprocess.run(cmd, capture_output=True, text=True, check=True)
        if logger:
            logger.debug(debug_git_clone_success)
        return True, None
    except subprocess.CalledProcessError as e:
        error_msg = e.stderr or str(e)
        if logger:
            logger.debug(debug_git_clone_failed.format(code=e.returncode, error=error_msg))
        return False, error_msg
    except Exception as e:
        error_msg = str(e)
        if logger:
            logger.debug(debug_unexpected_error.format(error_type=type(e).__name__, error=error_msg))
        return False, error_msg


def _prepare_target_directory(path: str, force: bool, logger: Optional[LoggerProtocol] = None) -> bool:
    """Prepare target directory for cloning."""
    if force and os.path.exists(path):
        if logger:
            logger.debug(debug_removing_directory.format(path=path))
        success = remove_directory(path, logger)
        if not success and logger:
            logger.debug(debug_directory_removal_failed)
        return success
    return True


def _validate_prerequisites(path: str, force: bool, logger: Optional[LoggerProtocol] = None) -> bool:
    """Validate prerequisites for cloning."""
    if path_exists_and_not_force(path, force):
        if logger:
            logger.debug(debug_path_exists_force_disabled.format(path=path))
            logger.error(path_already_exists_use_force.format(path=path))
        return False
    return True


def clone_repository(
    repo: str,
    path: str,
    branch: Optional[str] = None,
    force: bool = False,
    logger: Optional[LoggerProtocol] = None,
) -> tuple[bool, Optional[str]]:
    """Clone a git repository."""
    import time

    start_time = time.time()

    repo = _validate_repo(repo)
    path = _validate_path(path)

    if logger:
        logger.debug(debug_cloning_repo.format(repo=repo, path=path, force=force))

    if not _validate_prerequisites(path, force, logger):
        return False, prerequisites_validation_failed

    if not _prepare_target_directory(path, force, logger):
        return False, failed_to_prepare_target_directory

    success, error = _execute_git_clone(repo, path, branch, logger)

    duration = time.time() - start_time
    if logger:
        logger.debug(debug_clone_completed.format(duration=f"{duration:.2f}", success=success))

    return success, error
