import os
import shutil
from pathlib import Path
from typing import Callable, Dict, List, Optional, Tuple

from app.commands.uninstall.uninstall import remove_ssh_keys_step
from app.utils.config import API_ENV_FILE, CADDY_CONFIG_VOLUME, VIEW_ENV_FILE, get_config_value
from app.utils.protocols import LoggerProtocol

from .environment import ConfigResolver
from .services import cleanup_docker_services


def is_valid_path(path: Optional[str]) -> bool:
    return (
        path is not None
        and isinstance(path, str)
        and path.strip() != ""
        and os.path.exists(path)
    )


def handle_dry_run(
    message: str,
    dry_run: bool,
    logger: Optional[LoggerProtocol] = None,
) -> Tuple[bool, Optional[str]]:
    if dry_run:
        if logger:
            logger.info(f"[DRY RUN] {message}")
        return True, None
    return False, None


def safe_remove_file(
    file_path: str,
    logger: Optional[LoggerProtocol] = None,
    debug_msg: Optional[str] = None,
) -> bool:
    if not is_valid_path(file_path):
        return False
    try:
        os.remove(file_path)
        if logger and debug_msg:
            logger.debug(debug_msg)
        return True
    except Exception:
        return False


def safe_remove_directory(
    dir_path: str,
    logger: Optional[LoggerProtocol] = None,
    debug_msg: Optional[str] = None,
) -> bool:
    if not is_valid_path(dir_path):
        return False
    try:
        shutil.rmtree(dir_path)
        if logger and debug_msg:
            logger.debug(debug_msg)
        return True
    except Exception:
        return False


def get_config_value_safe(
    config: dict,
    key: str,
    logger: Optional[LoggerProtocol] = None,
) -> Optional[str]:
    try:
        return get_config_value(config, key)
    except KeyError:
        if logger:
            logger.debug(f"{key} not found in config, skipping")
        return None


def remove_caddy_json(
    full_source_path: str,
    logger: Optional[LoggerProtocol] = None,
) -> None:
    caddy_json_path = os.path.join(full_source_path, "helpers", "caddy.json")
    safe_remove_file(
        caddy_json_path,
        logger,
        f"Removed proxy config: {caddy_json_path}",
    )


def remove_caddyfile(
    config: dict,
    logger: Optional[LoggerProtocol] = None,
) -> None:
    target_dir = get_config_value_safe(config, CADDY_CONFIG_VOLUME, logger)
    if not target_dir or not isinstance(target_dir, str) or not target_dir.strip():
        return
    target_caddyfile = os.path.join(target_dir, "Caddyfile")
    safe_remove_file(
        target_caddyfile,
        logger,
        f"Removed Caddyfile: {target_caddyfile}",
    )


def rollback_proxy_config(
    full_source_path: Optional[str],
    config: dict,
    dry_run: bool,
    logger: Optional[LoggerProtocol] = None,
) -> Tuple[bool, Optional[str]]:
    dry_run_result = handle_dry_run("Would remove proxy config files", dry_run, logger)
    if dry_run_result[0]:
        return dry_run_result

    try:
        if full_source_path and is_valid_path(full_source_path):
            remove_caddy_json(full_source_path, logger)
        remove_caddyfile(config, logger)
        return True, None
    except Exception as e:
        error_msg = f"Failed to remove proxy config: {str(e)}"
        if logger:
            logger.warning(error_msg)
        return False, error_msg


def rollback_docker_services(
    compose_file: Optional[str],
    dry_run: bool,
    logger: Optional[LoggerProtocol] = None,
) -> Tuple[bool, Optional[str]]:
    dry_run_result = handle_dry_run("Would cleanup Docker services", dry_run, logger)
    if dry_run_result[0]:
        return dry_run_result

    if not compose_file or not isinstance(compose_file, str) or not compose_file.strip():
        if logger:
            logger.debug("Compose file path not available, skipping Docker cleanup")
        return True, None

    try:
        cleanup_docker_services(compose_file, dry_run, logger)
        return True, None
    except Exception as e:
        error_msg = f"Failed to cleanup Docker services: {str(e)}"
        if logger:
            logger.warning(error_msg)
        return False, error_msg


def rollback_ssh_keys(
    ssh_key_path: Optional[str],
    dry_run: bool,
    logger: Optional[LoggerProtocol] = None,
) -> Tuple[bool, Optional[str]]:
    dry_run_result = handle_dry_run("Would remove SSH keys", dry_run, logger)
    if dry_run_result[0]:
        return dry_run_result

    if not ssh_key_path or not isinstance(ssh_key_path, str) or not ssh_key_path.strip():
        if logger:
            logger.debug("SSH key path not available, skipping SSH key removal")
        return True, None

    success, error = remove_ssh_keys_step(ssh_key_path, logger)
    if not success:
        error_msg = f"Failed to remove SSH keys: {error}"
        if logger:
            logger.warning(error_msg)
        return False, error_msg
    return True, None


def get_env_file_paths(
    config_resolver: ConfigResolver,
    full_source_path: Optional[str],
    logger: Optional[LoggerProtocol] = None,
) -> List[str]:
    env_files = []

    try:
        api_env_file = config_resolver.get(API_ENV_FILE)
        if api_env_file:
            env_files.append(api_env_file)
    except (KeyError, AttributeError):
        if logger:
            logger.debug("API env file path not available, skipping")

    try:
        view_env_file = config_resolver.get(VIEW_ENV_FILE)
        if view_env_file:
            env_files.append(view_env_file)
    except (KeyError, AttributeError):
        if logger:
            logger.debug("View env file path not available, skipping")

    if full_source_path and isinstance(full_source_path, str) and full_source_path.strip():
        combined_env_file = os.path.join(full_source_path, ".env")
        env_files.append(combined_env_file)

    return env_files


def rollback_env_files(
    config_resolver: ConfigResolver,
    full_source_path: Optional[str],
    dry_run: bool,
    logger: Optional[LoggerProtocol] = None,
) -> Tuple[bool, Optional[str]]:
    dry_run_result = handle_dry_run("Would remove environment files", dry_run, logger)
    if dry_run_result[0]:
        return dry_run_result

    try:
        env_files = get_env_file_paths(config_resolver, full_source_path, logger)
        for env_file in env_files:
            if env_file and isinstance(env_file, str) and env_file.strip():
                safe_remove_file(
                    env_file,
                    logger,
                    f"Removed env file: {env_file}",
                )
        return True, None
    except Exception as e:
        error_msg = f"Failed to remove env files: {str(e)}"
        if logger:
            logger.warning(error_msg)
        return False, error_msg


def rollback_repository(
    full_source_path: Optional[str],
    dry_run: bool,
    logger: Optional[LoggerProtocol] = None,
) -> Tuple[bool, Optional[str]]:
    repo_path = full_source_path or "N/A"
    dry_run_result = handle_dry_run(f"Would remove repository: {repo_path}", dry_run, logger)
    if dry_run_result[0]:
        return dry_run_result

    if not full_source_path or not isinstance(full_source_path, str) or not full_source_path.strip():
        if logger:
            logger.debug("Repository path not available, skipping repository removal")
        return True, None

    try:
        safe_remove_directory(
            full_source_path,
            logger,
            f"Removed repository: {full_source_path}",
        )
        return True, None
    except Exception as e:
        error_msg = f"Failed to remove repository: {str(e)}"
        if logger:
            logger.warning(error_msg)
        return False, error_msg


def check_file_exists(path: Optional[str], issue_template: str) -> Optional[str]:
    if path and isinstance(path, str) and path.strip() and os.path.exists(path):
        try:
            return issue_template.format(path=path)
        except (KeyError, ValueError):
            return issue_template.replace("{path}", str(path))
    return None


def check_ssh_keys_exist(ssh_key_path: Optional[str]) -> List[str]:
    issues = []
    if not ssh_key_path or not isinstance(ssh_key_path, str) or not ssh_key_path.strip():
        return issues

    ssh_key_file = Path(ssh_key_path)
    if ssh_key_file.exists():
        issues.append(f"SSH key still exists: {ssh_key_path}")

    public_key_file = ssh_key_file.with_suffix(".pub")
    if public_key_file.exists():
        issues.append(f"SSH public key still exists: {public_key_file}")

    return issues


def verify_state_zero(
    compose_file: Optional[str],
    ssh_key_path: Optional[str],
    full_source_path: Optional[str],
    logger: Optional[LoggerProtocol] = None,
) -> Tuple[bool, List[str]]:
    issues = []

    compose_issue = check_file_exists(compose_file, "Compose file still exists: {path}")
    if compose_issue:
        issues.append(compose_issue)

    issues.extend(check_ssh_keys_exist(ssh_key_path))

    repo_issue = check_file_exists(full_source_path, "Repository directory still exists: {path}")
    if repo_issue:
        issues.append(repo_issue)

    if logger and issues:
        logger.warning("State verification found remaining resources:")
        for issue in issues:
            logger.warning(f"  - {issue}")

    return len(issues) == 0, issues


def execute_single_rollback(
    step_name: str,
    rollback_func: Callable,
    logger: Optional[LoggerProtocol] = None,
) -> Tuple[bool, Optional[str]]:
    if logger:
        logger.info(f"Rolling back: {step_name}")

    try:
        success, error = rollback_func()
        if success:
            if logger:
                logger.debug(f"Successfully rolled back: {step_name}")
            return True, None
        else:
            error_msg = f"{step_name}: {error}"
            if logger:
                logger.warning(f"Failed to rollback {step_name}: {error}")
            return False, error_msg
    except Exception as e:
        error_msg = f"{step_name}: {str(e)}"
        if logger:
            logger.warning(f"Exception during rollback of {step_name}: {error_msg}")
        return False, error_msg


def execute_rollback(
    completed_steps: List[str],
    rollback_functions: Dict[str, Callable],
    logger: Optional[LoggerProtocol] = None,
) -> Tuple[bool, List[str]]:
    failed_rollbacks = []

    for step_name in reversed(completed_steps):
        if step_name not in rollback_functions:
            if logger:
                logger.warning(f"No rollback function for step: {step_name}")
            continue

        success, error = execute_single_rollback(
            step_name,
            rollback_functions[step_name],
            logger,
        )
        if not success:
            failed_rollbacks.append(error or step_name)

    return len(failed_rollbacks) == 0, failed_rollbacks


def get_config_value_safe_from_resolver(
    config_resolver: ConfigResolver,
    key: str,
    logger: Optional[LoggerProtocol] = None,
) -> Optional[str]:
    try:
        return config_resolver.get(key)
    except Exception:
        if logger:
            logger.debug(f"Could not get {key} for rollback")
        return None


def build_rollback_functions(
    full_source_path: Optional[str],
    compose_file: Optional[str],
    ssh_key_path: Optional[str],
    config_resolver: ConfigResolver,
    config: dict,
    dry_run: bool,
    logger: Optional[LoggerProtocol] = None,
) -> Dict[str, Callable]:
    return {
        "Loading proxy configuration": lambda: rollback_proxy_config(
            full_source_path, config, dry_run, logger
        ),
        "Starting services": lambda: rollback_docker_services(
            compose_file, dry_run, logger
        ),
        "Generating SSH keys": lambda: rollback_ssh_keys(
            ssh_key_path, dry_run, logger
        ),
        "Creating environment files": lambda: rollback_env_files(
            config_resolver, full_source_path, dry_run, logger
        ),
        "Setting up proxy config": lambda: rollback_proxy_config(
            full_source_path, config, dry_run, logger
        ),
        "Cloning repository": lambda: rollback_repository(
            full_source_path, dry_run, logger
        ),
    }


def log_rollback_results(
    rollback_success: bool,
    failed_rollbacks: List[str],
    logger: Optional[LoggerProtocol] = None,
) -> None:
    from .messages import rollback_completed, rollback_failed

    if rollback_success:
        if logger:
            logger.info(rollback_completed)
    else:
        if logger:
            logger.warning(rollback_failed)
            for failure in failed_rollbacks:
                logger.warning(f"  - {failure}")


def perform_installation_rollback(
    completed_steps: List[str],
    config_resolver: ConfigResolver,
    config: dict,
    dry_run: bool,
    logger: Optional[LoggerProtocol] = None,
) -> None:
    from .messages import rollback_starting

    if logger:
        logger.info(rollback_starting)

    full_source_path = get_config_value_safe_from_resolver(
        config_resolver, "full_source_path", logger
    )
    compose_file = get_config_value_safe_from_resolver(
        config_resolver, "compose_file_path", logger
    )
    ssh_key_path = get_config_value_safe_from_resolver(
        config_resolver, "ssh_key_path", logger
    )

    rollback_functions = build_rollback_functions(
        full_source_path,
        compose_file,
        ssh_key_path,
        config_resolver,
        config,
        dry_run,
        logger,
    )

    rollback_success, failed_rollbacks = execute_rollback(
        completed_steps, rollback_functions, logger
    )

    log_rollback_results(rollback_success, failed_rollbacks, logger)

    is_clean, issues = verify_state_zero(
        compose_file, ssh_key_path, full_source_path, logger
    )

    if not is_clean and logger:
        logger.warning("Some resources may still remain. Please clean up manually if needed.")
