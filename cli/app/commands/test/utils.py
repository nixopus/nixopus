import shlex
import subprocess
from pathlib import Path
from typing import List, Optional, Tuple, Union

from .messages import dry_run_would_execute, lxd_command_not_available, listing_images, port_in_use
from .types import TestParams   

def check_lxd_available() -> Tuple[bool, Optional[str]]:
    try:
        result = subprocess.run(
            ["lxc", "version"],
            capture_output=True,
            text=True,
            check=True,
        )
        return True, None
    except (subprocess.CalledProcessError, FileNotFoundError):
        return False, lxd_command_not_available


def list_lxd_images(params: TestParams) -> Tuple[bool, Optional[str], list]:
    if params.logger:
        params.logger.info(listing_images)

    if params.dry_run:
        if params.logger:
            params.logger.info(dry_run_would_execute.format(action="lxc image list"))
        return True, None, []

    try:
        result = subprocess.run(
            ["lxc", "image", "list"],
            capture_output=True,
            text=True,
            check=True,
        )
        images = []
        lines = result.stdout.strip().split("\n")
        for line in lines[1:]:
            if line.strip():
                parts = line.split()
                if len(parts) > 0:
                    images.append(parts[0])
        return True, None, images
    except subprocess.CalledProcessError as e:
        return False, f"Failed to list images: {e.stderr}", []


def check_port_available(port: int) -> Tuple[bool, Optional[str]]:
    try:
        result = subprocess.run(
            ["lsof", "-i", f":{port}"],
            capture_output=True,
            text=True,
        )
        if result.returncode == 0:
            return False, port_in_use.format(port=port)
        return True, None
    except FileNotFoundError:
        return True, None


def run_subprocess(
    cmd: Union[List[str], str],
    cwd: Optional[Union[str, Path]] = None,
    timeout: Optional[int] = None,
    check: bool = False,
    capture_output: bool = True,
    text: bool = True,
    error_prefix: Optional[str] = None,
) -> Tuple[bool, Optional[str], Optional[str]]:
    if isinstance(cmd, str):
        cmd = shlex.split(cmd)

    try:
        result = subprocess.run(
            cmd,
            cwd=str(cwd) if cwd else None,
            capture_output=capture_output,
            text=text,
            check=check,
            timeout=timeout,
        )

        stdout = result.stdout if capture_output else None
        if result.returncode == 0:
            return True, stdout, None

        error_output = (
            result.stderr.decode() if result.stderr and not text else result.stderr
        ) or (
            result.stdout.decode() if result.stdout and not text else result.stdout
        ) or "unknown error"

        error_msg = f"{error_prefix}: {error_output}" if error_prefix else error_output
        return False, stdout, error_msg

    except subprocess.CalledProcessError as e:
        error_msg = e.stderr if e.stderr else str(e)
        if error_prefix:
            error_msg = f"{error_prefix}: {error_msg}"
        stdout = e.stdout if capture_output else None
        return False, stdout, error_msg

    except subprocess.TimeoutExpired as e:
        timeout_str = f"{e.timeout}s" if hasattr(e, "timeout") and e.timeout else (f"{timeout}s" if timeout else "timeout period")
        error_msg = f"Command timed out after {timeout_str}"
        if error_prefix:
            error_msg = f"{error_prefix}: {error_msg}"
        return False, None, error_msg

    except FileNotFoundError:
        cmd_str = " ".join(cmd) if isinstance(cmd, list) else cmd
        error_msg = f"Command not found: {cmd_str.split()[0] if cmd_str else 'unknown'}"
        if error_prefix:
            error_msg = f"{error_prefix}: {error_msg}"
        return False, None, error_msg

    except Exception as e:
        error_msg = str(e)
        if error_prefix:
            error_msg = f"{error_prefix}: {error_msg}"
        return False, None, error_msg


def run_subprocess_simple(
    cmd: Union[List[str], str],
    cwd: Optional[Union[str, Path]] = None,
    timeout: Optional[int] = None,
    error_prefix: Optional[str] = None,
) -> Tuple[bool, Optional[str]]:
    success, _, error = run_subprocess(
        cmd=cmd,
        cwd=cwd,
        timeout=timeout,
        check=False,
        capture_output=True,
        text=True,
        error_prefix=error_prefix,
    )
    return success, error


def lxc_file_push(
    container_name: str,
    source: Union[str, Path],
    dest: str,
    error_prefix: Optional[str] = None,
) -> Tuple[bool, Optional[str]]:
    cmd = ["lxc", "file", "push", str(source), f"{container_name}/{dest}"]
    return run_subprocess_simple(cmd, error_prefix=error_prefix or "Failed to copy file")


def lxc_exec_chmod(
    container_name: str,
    file_path: str,
    mode: str = "+x",
    error_prefix: Optional[str] = None,
) -> Tuple[bool, Optional[str]]:
    if not file_path.startswith("/"):
        file_path = "/" + file_path.lstrip("/")
    cmd = ["lxc", "exec", container_name, "--", "chmod", mode, file_path]
    return run_subprocess_simple(cmd, error_prefix=error_prefix or "Failed to make executable")


def lxc_file_push_and_chmod(
    container_name: str,
    source: Union[str, Path],
    dest: str,
    chmod_mode: str = "+x",
    error_prefix: Optional[str] = None,
) -> Tuple[bool, Optional[str]]:
    success, error = lxc_file_push(container_name, source, dest, error_prefix)
    if not success:
        return False, error
    
    chmod_success, chmod_error = lxc_exec_chmod(container_name, dest, chmod_mode, error_prefix)
    if not chmod_success:
        return False, chmod_error
    
    return True, None
