import random
import string
from pathlib import Path
from typing import Optional, Tuple

from app.utils.logger import log_error, log_success 
from .utils import lxc_file_push_and_chmod
from app.utils.timeout import timeout_wrapper
from .container import cleanup_container, configure_proxy_ports, create_container, preinstall_dependencies, setup_networking
from .installation import run_installation_script
from .messages import (
    api_accessible,
    app_accessible,
    available_images,
    container_name_info,
    copying_script,
    dry_run_mode,
    dry_run_would_execute,
    end_dry_run,
    failed_preinstall_deps,
    failed_to_copy_files,
    failed_to_copy_script,
    failed_to_list_images,
    lxd_command_not_available,
    port_unavailable,
    script_not_found,
    test_failed,
    test_success,
    test_timed_out,
)
from .types import TestParams
from .utils import check_lxd_available, check_port_available, list_lxd_images


def copy_installation_script(params: TestParams, workspace_root: Path) -> Tuple[bool, Optional[str]]:
    if params.logger:
        params.logger.debug(copying_script)

    install_script_path = workspace_root / "scripts" / "install.sh"
    if not install_script_path.exists():
        return False, script_not_found.format(path=install_script_path)

    success, error = lxc_file_push_and_chmod(
        params.container_name,
        install_script_path,
        "/tmp/install.sh",
        error_prefix="Failed to copy installation script",
    )

    if not success:
        return False, failed_to_copy_script.format(error=error or "unknown")

    return True, None


def _validate_prerequisites(params: TestParams) -> Tuple[bool, Optional[str]]:
    available, error = check_lxd_available()
    if not available:
        return False, error or lxd_command_not_available

    success, error, images = list_lxd_images(params)
    if not success:
        return False, error or failed_to_list_images

    if params.verbose and params.logger and images:
        images_str = ", ".join(images[:10])
        params.logger.info(available_images.format(images=images_str))

    return True, None


def _validate_ports(params: TestParams) -> Tuple[bool, Optional[str]]:
    if params.app_port:
        available, error = check_port_available(params.app_port)
        if not available:
            return False, error or port_unavailable

    if params.api_port:
        available, error = check_port_available(params.api_port)
        if not available:
            return False, error or port_unavailable

    return True, None


def _generate_container_name(params: TestParams) -> None:
    if not params.container_name:
        random_suffix = "".join(random.choices(string.ascii_lowercase + string.digits, k=8))
        params.container_name = f"nixopus-test-{random_suffix}"


def _execute_test_steps(params: TestParams, workspace_root: Path) -> Tuple[bool, Optional[str]]:
    success, error = create_container(params)
    if not success:
        return False, error or test_failed

    success, error = setup_networking(params)
    if not success:
        cleanup_container(params)
        return False, error or test_failed

    success, error = preinstall_dependencies(params)
    if not success:
        cleanup_container(params)
        return False, error or failed_preinstall_deps

    success, error = copy_installation_script(params, workspace_root)
    if not success:
        cleanup_container(params)
        return False, error or failed_to_copy_files

    if params.app_port or params.api_port:
        success, error = configure_proxy_ports(params)
        if not success:
            cleanup_container(params)
            return False, error or test_failed

    success, error = run_installation_script(params)
    if not success:
        cleanup_container(params)
        return False, error or test_failed

    return True, None


def _log_success_info(params: TestParams) -> None:
    if not params.logger:
        return

    params.logger.info(container_name_info.format(name=params.container_name))
    if params.app_port:
        params.logger.info(app_accessible.format(port=params.app_port))
    if params.api_port:
        params.logger.info(api_accessible.format(port=params.api_port))


def run_test(params: TestParams) -> None:
    if params.dry_run and params.logger:
        params.logger.info(dry_run_mode)

    success, error = _validate_prerequisites(params)
    if not success:
        log_error(error, verbose=params.verbose)
        raise SystemExit(1)

    success, error = _validate_ports(params)
    if not success:
        log_error(error, verbose=params.verbose)
        raise SystemExit(1)

    _generate_container_name(params)

    try:
        with timeout_wrapper(params.timeout):
            workspace_root = Path(__file__).parent.parent.parent.parent.parent
            success, error = _execute_test_steps(params, workspace_root)
            if not success:
                log_error(error, verbose=params.verbose)
                raise SystemExit(1)

        if params.dry_run and params.logger:
            params.logger.info(end_dry_run)
            params.logger.info("Dry run completed - no actions taken")
        else:
            log_success(test_success, verbose=params.verbose)
            _log_success_info(params)

    except TimeoutError:
        log_error(test_timed_out.format(timeout=params.timeout), verbose=params.verbose)
        cleanup_container(params)
        raise SystemExit(1)
    except Exception as e:
        log_error(f"{test_failed}: {str(e)}", verbose=params.verbose)
        cleanup_container(params)
        raise SystemExit(1)
