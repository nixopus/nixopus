import os
from typing import Dict, Optional, Tuple

from app.commands.proxy.proxy import load_config
from app.commands.service.service import cleanup_docker_resources, start_services
from app.utils.protocols import LoggerProtocol
from app.utils.timeout import timeout_wrapper

from .config_schema import build_port_env_vars
from .config_utils import copy_caddyfile_to_target, setup_proxy_config


def cleanup_docker_services(
    compose_file: str,
    dry_run: bool,
    logger: Optional[LoggerProtocol] = None,
) -> None:
    if dry_run:
        if logger:
            logger.info(
                f"[dry-run] Would run: docker compose -f {compose_file} down --rmi all --volumes --remove-orphans"
            )
        return

    if not compose_file or not os.path.exists(compose_file):
        if logger:
            logger.debug(f"Compose file not found at {compose_file}; skipping docker cleanup")
        return

    try:
        success, output = cleanup_docker_resources(
            compose_file=compose_file,
            logger=logger,
            remove_images="all",
            remove_volumes=True,
            remove_orphans=True,
        )
        if success:
            if logger:
                logger.info("Docker resources cleaned (images, volumes, orphans)")
        else:
            if logger:
                logger.warning("Docker cleanup did not fully succeed; continuing")
    except Exception as e:
        if logger:
            logger.warning(f"Failed to cleanup docker resources: {e}")


def setup_proxy_configuration(
    full_source_path: str,
    host_ip: str,
    view_domain: Optional[str],
    api_domain: Optional[str],
    view_port: str,
    api_port: str,
    config: dict,
    dry_run: bool,
    logger: Optional[LoggerProtocol] = None,
) -> None:
    if not dry_run:
        caddy_json_template = setup_proxy_config(
            full_source_path,
            host_ip,
            view_domain,
            api_domain,
            view_port,
            api_port,
        )
        copy_caddyfile_to_target(full_source_path, config, logger)
        if logger:
            logger.debug(f"Proxy config created: {caddy_json_template}")
    else:
        caddy_json_template = os.path.join(full_source_path, "helpers", "caddy.json")
        if logger:
            logger.debug(f"Proxy config created: {caddy_json_template}")


def build_service_env_vars(
    api_port: Optional[int],
    view_port: Optional[int],
    db_port: Optional[int],
    redis_port: Optional[int],
    caddy_admin_port: Optional[int],
    caddy_http_port: Optional[int],
    caddy_https_port: Optional[int],
    supertokens_port: Optional[int],
) -> Dict[str, str]:
    return build_port_env_vars(
        api_port=api_port,
        view_port=view_port,
        db_port=db_port,
        redis_port=redis_port,
        caddy_admin_port=caddy_admin_port,
        caddy_http_port=caddy_http_port,
        caddy_https_port=caddy_https_port,
        supertokens_port=supertokens_port,
    )


def start_docker_services(
    compose_file: str,
    env_vars: Dict[str, str],
    timeout: int,
    dry_run: bool,
    logger: Optional[LoggerProtocol] = None,
    profiles: Optional[list[str]] = None,
) -> Tuple[bool, Optional[str]]:
    if dry_run:
        if logger:
            logger.info(f"[DRY RUN] Would start services using {compose_file}")
        return True, None

    original_env = os.environ.copy()
    os.environ.update(env_vars)

    try:
        try:
            with timeout_wrapper(timeout):
                success, error = start_services(
                    name="all",
                    detach=True,
                    env_file=None,
                    compose_file=compose_file,
                    logger=logger,
                    profiles=profiles,
                )
        except TimeoutError:
            return False, "Operation timed out"
        
        return success, error
    finally:
        for key in env_vars:
            if key in original_env:
                os.environ[key] = original_env[key]
            else:
                os.environ.pop(key, None)


def load_proxy_config(
    caddy_json_config: str,
    proxy_port: int,
    timeout: int,
    dry_run: bool,
    logger: Optional[LoggerProtocol] = None,
) -> Tuple[bool, Optional[str]]:
    if dry_run:
        if logger:
            logger.info(f"[DRY RUN] Would load proxy config from {caddy_json_config}")
        return True, None

    try:
        with timeout_wrapper(timeout):
            success, error = load_config(caddy_json_config, proxy_port, logger)
    except TimeoutError:
        return False, "Operation timed out"

    if success:
        if logger and not dry_run:
            logger.success("Caddy proxy configuration loaded successfully")
    
    return success, error

