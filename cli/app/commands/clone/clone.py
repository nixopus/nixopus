import os
import subprocess
from typing import Optional

import typer

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
    installation_cancelled_by_user,
    invalid_path,
    invalid_repo,
    invalid_repository_url,
    path_already_exists_use_force,
    path_exists_prompt,
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


def _log_path_exists(path: str, logger: Optional[LoggerProtocol]) -> None:
    if logger:
        logger.debug(debug_path_exists_force_disabled.format(path=path))


def _handle_interactive_prompt(path: str, logger: Optional[LoggerProtocol]) -> bool:
    if logger:
        logger.warning(path_exists_prompt.format(path=path))
    
    # Show what files exist to make user aware of what will be deleted
    try:
        if os.path.exists(path) and os.path.isdir(path):
            contents = os.listdir(path)
            if contents and logger:
                preview = contents[:5]  # Show first 5 items
                logger.warning(f"Existing contents: {', '.join(preview)}{' ...' if len(contents) > 5 else ''}")
    except Exception:
        pass  # Silently fail, user will still see the prompt
    
    user_response = typer.confirm("⚠️  Continue with force flag? This will DELETE all existing data", default=False)
    
    if not user_response and logger:
        logger.info(installation_cancelled_by_user)
    
    return user_response


def _handle_non_interactive_error(path: str, logger: Optional[LoggerProtocol]) -> None:
    if logger:
        logger.error(path_already_exists_use_force.format(path=path))


# if path exists and force is not enabled, prompt user to continue with force flag
def _validate_prerequisites(path: str, force: bool, logger: Optional[LoggerProtocol] = None, interactive: bool = True) -> tuple[bool, bool]:
    # Path doesn't exist or force already enabled - all good
    if not path_exists_and_not_force(path, force):
        return True, force
    
    # Path exists and force is not enabled
    _log_path_exists(path, logger)
    
    # Handle interactive mode - prompt user
    if interactive:
        user_wants_force = _handle_interactive_prompt(path, logger)
        if user_wants_force:
            return True, True  # User confirmed, enable force
        return False, False  # User cancelled
    
    # Handle non-interactive mode - fail with error
    _handle_non_interactive_error(path, logger)
    return False, False


def clone_repository(
    repo: str,
    path: str,
    branch: Optional[str] = None,
    force: bool = False,
    logger: Optional[LoggerProtocol] = None,
    interactive: bool = True,
) -> tuple[bool, Optional[str]]:
    """Clone a git repository."""
    import time

    start_time = time.time()

    repo = _validate_repo(repo)
    path = _validate_path(path)

    if logger:
        logger.debug(debug_cloning_repo.format(repo=repo, path=path, force=force))

    is_valid, should_force = _validate_prerequisites(path, force, logger, interactive)
    if not is_valid:
        return False, prerequisites_validation_failed
    
    # Update force flag if user confirmed
    force = should_force

    if not _prepare_target_directory(path, force, logger):
        return False, failed_to_prepare_target_directory

    success, error = _execute_git_clone(repo, path, branch, logger)

    duration = time.time() - start_time
    if logger:
        logger.debug(debug_clone_completed.format(duration=f"{duration:.2f}", success=success))

    return success, error
