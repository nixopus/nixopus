import os
import subprocess
from typing import Optional

from app.utils.protocols import LoggerProtocol

from .messages import (
    docker_command_completed,
    docker_command_executing,
    docker_command_failed,
    docker_command_stderr,
    docker_command_stdout,
    docker_unexpected_error,
    service_action_failed,
    service_action_info,
    service_action_unexpected_error,
)


def _build_command(action: str, name: str = "all", env_file: str = None, compose_file: str = None, **kwargs) -> list[str]:
    """Build docker compose command."""
    cmd = ["docker", "compose"]
    if compose_file:
        cmd.extend(["-f", compose_file])
    if env_file:
        cmd.extend(["--env-file", env_file])
    
    profiles = kwargs.get("profiles")
    if profiles:
        for profile in profiles:
            cmd.extend(["--profile", profile])
    
    cmd.append(action)

    if action == "up" and kwargs.get("detach", False):
        cmd.append("-d")

    if name != "all":
        cmd.append(name)

    return cmd


def _build_cleanup_command(
    compose_file: str,
    remove_images: str = "all",
    remove_volumes: bool = True,
    remove_orphans: bool = True,
    env_file: str = None,
) -> list[str]:
    """Build docker compose cleanup command (down with prune flags)."""
    cmd = ["docker", "compose"]
    if compose_file:
        cmd.extend(["-f", compose_file])
    if env_file:
        cmd.extend(["--env-file", env_file])
    cmd.append("down")

    if remove_images:
        cmd.extend(["--rmi", remove_images])
    if remove_volumes:
        cmd.append("--volumes")
    if remove_orphans:
        cmd.append("--remove-orphans")

    return cmd


def _log_command_info(cmd: list[str], action: str, name: str, logger: Optional[LoggerProtocol]) -> None:
    """Log command execution info."""
    if logger:
        logger.debug(docker_command_executing.format(command=" ".join(cmd)))
        logger.debug(service_action_info.format(action=action, name=name))


def _log_output(output: str, logger: Optional[LoggerProtocol], is_error: bool = False) -> None:
    """Log command output."""
    if not logger or not output.strip():
        return
    
    if is_error:
        logger.debug(docker_command_stderr.format(output=output.strip()))
    else:
        logger.debug(docker_command_stdout.format(output=output.strip()))


def _handle_process_error(return_code: int, output: str, action: str, logger: Optional[LoggerProtocol]) -> tuple[bool, str]:
    """Handle process execution error."""
    error_msg = output or f"Process exited with code {return_code}"
    
    if logger:
        logger.debug(docker_command_failed.format(return_code=return_code))
        _log_output(output, logger, is_error=True)
        logger.error(service_action_failed.format(action=action, error=error_msg))
    
    return False, error_msg


def _execute_streaming(cmd: list[str], action: str, logger: Optional[LoggerProtocol]) -> tuple[bool, str]:
    """Execute command with streaming output (for non-detached up)."""
    process = subprocess.Popen(
        cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, bufsize=1, universal_newlines=True
    )

    output_lines = []
    if logger:
        logger.debug("Docker container logs:")
        logger.debug("-" * 50)

    for line in process.stdout:
        if logger:
            logger.debug(line.rstrip())
        output_lines.append(line.rstrip())

    return_code = process.wait()
    full_output = "\n".join(output_lines)

    if return_code == 0:
        if logger:
            logger.debug(docker_command_completed.format(action=action))
            _log_output(full_output, logger)
        return True, full_output
    
    return _handle_process_error(return_code, full_output, action, logger)


def _execute_normal(cmd: list[str], action: str, logger: Optional[LoggerProtocol]) -> tuple[bool, str]:
    """Execute command with captured output."""
    result = subprocess.run(cmd, capture_output=True, text=True, check=True)

    if logger:
        logger.debug(docker_command_completed.format(action=action))
        _log_output(result.stdout, logger)
        _log_output(result.stderr, logger)

    return True, result.stdout or result.stderr


def execute_services(
    action: str,
    name: str = "all",
    env_file: str = None,
    compose_file: str = None,
    logger: Optional[LoggerProtocol] = None,
    **kwargs
) -> tuple[bool, str]:
    """Execute docker compose command for services."""
    cmd = _build_command(action, name, env_file, compose_file, **kwargs)
    _log_command_info(cmd, action, name, logger)

    try:
        if action == "up" and not kwargs.get("detach", False):
            return _execute_streaming(cmd, action, logger)
        
        return _execute_normal(cmd, action, logger)

    except subprocess.CalledProcessError as e:
        if logger:
            logger.debug(docker_command_failed.format(return_code=e.returncode))
            _log_output(e.stdout or "", logger)
            _log_output(e.stderr or "", logger, is_error=True)
            logger.error(service_action_failed.format(action=action, error=e.stderr or str(e)))
        return False, e.stderr or e.stdout or str(e)
    except Exception as e:
        if logger:
            logger.debug(docker_unexpected_error.format(action=action, error=str(e)))
            logger.error(service_action_unexpected_error.format(action=action, error=e))
        return False, str(e)


def cleanup_docker_resources(
    compose_file: str,
    logger: Optional[LoggerProtocol] = None,
    remove_images: str = "all",
    remove_volumes: bool = True,
    remove_orphans: bool = True,
    env_file: str = None,
) -> tuple[bool, str]:
    """Run docker compose down with prune flags."""
    cmd = _build_cleanup_command(compose_file, remove_images, remove_volumes, remove_orphans, env_file)

    if logger:
        logger.debug(docker_command_executing.format(command=" ".join(cmd)))
    
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, check=False)

        if logger:
            if result.stdout and result.stdout.strip():
                logger.debug(docker_command_stdout.format(output=result.stdout.strip()))
            if result.stderr and result.stderr.strip():
                logger.debug(docker_command_stderr.format(output=result.stderr.strip()))

        if result.returncode == 0:
            if logger:
                logger.debug(docker_command_completed.format(action="cleanup"))
            return True, result.stdout or result.stderr
        else:
            if logger:
                logger.debug(docker_command_failed.format(return_code=result.returncode))
            return False, result.stderr or result.stdout
    except Exception as e:
        if logger:
            logger.debug(docker_unexpected_error.format(action="cleanup", error=str(e)))
        return False, str(e)


def start_services(
    name: str = "all",
    detach: bool = True,
    env_file: str = None,
    compose_file: str = None,
    logger: Optional[LoggerProtocol] = None,
    profiles: Optional[list[str]] = None,
) -> tuple[bool, str]:
    """Start docker compose services."""
    return execute_services("up", name, env_file, compose_file, logger, detach=detach, profiles=profiles)


def stop_services(
    name: str = "all",
    env_file: str = None,
    compose_file: str = None,
    logger: Optional[LoggerProtocol] = None,
) -> tuple[bool, str]:
    """Stop docker compose services."""
    return execute_services("down", name, env_file, compose_file, logger)

