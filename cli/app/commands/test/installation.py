import sys
import subprocess
import shlex
import threading
import time
from pathlib import Path
from typing import List, Optional, Tuple

from .messages import (
    building_commands,
    dry_run_would_execute,
    installation_complete,
    installation_failed,
    installation_in_progress,
    installation_timed_out,
    process_completed,
    running_installation,
    streaming_output,
)
from .types import TestParams


def _build_cli_install_command(params: TestParams) -> str:
    cmd = "/tmp/install.sh"
    if params.repo:
        cmd += f" --repo {shlex.quote(params.repo)}"
    if params.branch:
        cmd += f" --branch {shlex.quote(params.branch)}"
    cmd += " --skip-nixopus-install"
    return cmd


def _build_nixopus_install_command(params: TestParams) -> str:
    cmd = "nixopus install --verbose"
    if params.repo:
        cmd += f" --repo {shlex.quote(params.repo)}"
    if params.branch:
        cmd += f" --branch {shlex.quote(params.branch)}"
    return cmd


def _build_install_command(params: TestParams, workspace_root: Path) -> str:
    env_vars = "export PYTHONUNBUFFERED=1 TERM=xterm-256color NO_COLOR=1; "
    cli_cmd = _build_cli_install_command(params)
    nixopus_cmd = _build_nixopus_install_command(params)
    return f"{env_vars}{cli_cmd} && {nixopus_cmd}"


def _read_process_output(
    process: subprocess.Popen[str], output_lines: List[str], output_complete: threading.Event
) -> None:
    try:
        while True:
            chunk = process.stdout.read(1024)
            if not chunk:
                break
            sys.stdout.write(chunk)
            sys.stdout.flush()
            output_lines.append(chunk)
    except Exception:
        pass
    finally:
        output_complete.set()


def _run_verbose_installation(
    params: TestParams, install_cmd: str
) -> Tuple[int, str]:
    if params.logger:
        params.logger.debug(streaming_output)

    process = subprocess.Popen(
        ["lxc", "exec", params.container_name, "--", "sh", "-c", install_cmd],
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
        bufsize=1,
    )

    output_lines: List[str] = []
    output_complete = threading.Event()
    output_thread = threading.Thread(
        target=_read_process_output, args=(process, output_lines, output_complete), daemon=True
    )
    output_thread.start()

    start_time = time.time()
    last_log_time = start_time
    while process.poll() is None:
        elapsed = time.time() - start_time
        if time.time() - last_log_time > 30:
            if params.logger:
                params.logger.debug(installation_in_progress.format(elapsed=int(elapsed)))
            last_log_time = time.time()

        if elapsed > params.timeout:
            if params.logger:
                params.logger.error(installation_timed_out.format(timeout=params.timeout))
            process.kill()
            returncode = process.wait()
            output_complete.wait(timeout=2)
            return returncode, "".join(output_lines)

        time.sleep(0.1)

    output_complete.wait(timeout=2)
    returncode = process.returncode

    if params.logger:
        params.logger.debug(process_completed.format(exit_code=returncode))

    return returncode, "".join(output_lines)


def _run_non_verbose_installation(
    params: TestParams, install_cmd: str
) -> Tuple[int, str]:
    result = subprocess.run(
        ["lxc", "exec", params.container_name, "--", "sh", "-c", install_cmd],
        capture_output=True,
        text=True,
        timeout=params.timeout,
    )
    return result.returncode, result.stdout + result.stderr


def _filter_error_output(output: str) -> str:
    error_lines = []
    for line in output.split("\n"):
        if "%" in line and ("Total" in line or "Xferd" in line or "Dload" in line):
            continue
        if line.strip():
            error_lines.append(line)
    return "\n".join(error_lines) if error_lines else output or "Unknown error"


def run_installation_script(params: TestParams) -> Tuple[bool, Optional[str]]:
    if params.logger:
        params.logger.info(running_installation.format(name=params.container_name))
        params.logger.debug(f"Container name: {params.container_name}")
        params.logger.debug(f"Health check timeout: {params.health_check_timeout}s")
        params.logger.debug(f"Overall timeout: {params.timeout}s")

    if params.dry_run:
        if params.logger:
            params.logger.info(
                dry_run_would_execute.format(
                    action=f"Run installation script in {params.container_name}"
                )
            )
        return True, None

    try:
        if params.logger:
            params.logger.debug(building_commands)

        workspace_root = Path(__file__).parent.parent.parent.parent.parent
        install_cmd = _build_install_command(params, workspace_root)

        if params.logger:
            cli_cmd = _build_cli_install_command(params)
            nixopus_cmd = _build_nixopus_install_command(params)
            params.logger.debug(f"CLI install command: {cli_cmd[:100]}...")
            params.logger.debug(f"Nixopus install command: {nixopus_cmd}")

        if params.verbose:
            returncode, output = _run_verbose_installation(params, install_cmd)
        else:
            returncode, output = _run_non_verbose_installation(params, install_cmd)

        if returncode != 0:
            error_msg = _filter_error_output(output)
            return False, installation_failed.format(exit_code=returncode, error=error_msg)

        if params.logger:
            params.logger.success(installation_complete.format(name=params.container_name))
        return True, None

    except subprocess.TimeoutExpired:
        return False, installation_timed_out.format(timeout=params.timeout)
    except subprocess.CalledProcessError as e:
        error_msg = e.stderr if hasattr(e, 'stderr') and e.stderr else str(e)
        if not error_msg:
            error_msg = _filter_error_output(e.stdout if hasattr(e, 'stdout') and e.stdout else "")
        return False, installation_failed.format(exit_code=e.returncode, error=error_msg)

