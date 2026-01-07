import os
import re
from typing import Any, Callable, List, Optional, Tuple

import typer
from rich.progress import BarColumn, Progress, SpinnerColumn, TaskProgressColumn, TextColumn

from app.commands.clone.clone import clone_repository
from app.commands.preflight.preflight import check_required_ports
from app.utils.config import (
    API_PORT,
    CADDY_ADMIN_PORT,
    CADDY_HTTP_PORT,
    CADDY_HTTPS_PORT,
    DEFAULT_BRANCH,
    DEFAULT_REPO,
    PORTS,
    PROXY_PORT,
    SSH_KEY_SIZE,
    SSH_KEY_TYPE,
    SUPERTOKENS_API_PORT,
    VIEW_PORT,
    get_active_config,
    get_config_value,
)
from app.utils.timeout import timeout_wrapper
from app.utils.installation_tracker import track_installation_failure, track_installation_success, track_staging_installation

from .admin_registration import register_admin_user_step
from .types import InstallParams
from .config_utils import (
    get_access_url,
    get_host_ip_or_default,
    is_custom_repo_or_branch,
)
from .deps import install_all_deps
from .environment import ConfigResolver, create_service_env_files
from .messages import (
    clone_failed,
    dependency_installation_timeout,
    installation_failed,
    installing_nixopus,
    operation_timed_out,
    proxy_load_failed,
    services_start_failed,
    ssh_setup_failed,
)
from .services import (
    build_service_env_vars,
    cleanup_docker_services,
    load_proxy_config,
    setup_proxy_configuration,
    start_docker_services,
)
from .rollback import perform_installation_rollback
from .ssh import SSHConfig, generate_ssh_key_with_config
from .validate import validate_domains, validate_host_ip, validate_repo


def validate_install_params(params: InstallParams) -> None:
    validate_domains(params.api_domain, params.view_domain)
    validate_repo(params.repo)
    validate_host_ip(params.host_ip)


def create_config_resolver(config: dict, params: InstallParams) -> ConfigResolver:
    return ConfigResolver(
        config,
        repo=params.repo,
        branch=params.branch,
        api_port=params.api_port,
        view_port=params.view_port,
        db_port=params.db_port,
        redis_port=params.redis_port,
        caddy_admin_port=params.caddy_admin_port,
        caddy_http_port=params.caddy_http_port,
        caddy_https_port=params.caddy_https_port,
        supertokens_port=params.supertokens_port,
        staging=params.staging,
    )


def _extract_default_port_from_env_var(env_var_value: Any) -> Optional[int]:
    if not isinstance(env_var_value, str):
        return None
    
    match = re.search(r':-(\d+)', env_var_value)
    if match:
        try:
            return int(match.group(1))
        except ValueError:
            pass
    
    cleaned = re.sub(r'\$\{[^}]+\}', '', env_var_value).strip()
    if cleaned and cleaned.isdigit():
        try:
            return int(cleaned)
        except ValueError:
            pass
    
    return None


def _get_config_port_value(config: dict, port_path: str) -> Optional[int]:
    try:
        keys = port_path.split(".")
        value = config
        for key in keys:
            if isinstance(value, dict) and key in value:
                value = value[key]
            else:
                return None
        
        if not isinstance(value, str):
            return None
        
        return _extract_default_port_from_env_var(value)
    except (KeyError, ValueError, TypeError, AttributeError):
        return None


def build_ports_to_check(config: dict, params: InstallParams) -> List[int]:
    config_ports = get_config_value(config, PORTS)
    config_ports = [int(port) for port in config_ports] if isinstance(config_ports, list) else [int(config_ports)]
    config_api_port = _get_config_port_value(config, API_PORT)
    config_view_port = _get_config_port_value(config, VIEW_PORT)
    config_db_port = _get_config_port_value(config, "services.db.env.DB_PORT")
    config_redis_port = _get_config_port_value(config, "services.redis.env.REDIS_PORT")
    config_caddy_admin_port = _get_config_port_value(config, CADDY_ADMIN_PORT)
    config_caddy_http_port = _get_config_port_value(config, CADDY_HTTP_PORT)
    config_caddy_https_port = _get_config_port_value(config, CADDY_HTTPS_PORT)
    config_supertokens_port = _get_config_port_value(config, SUPERTOKENS_API_PORT)
    
    port_overrides = {}
    if config_api_port is not None and params.api_port is not None:
        port_overrides[config_api_port] = params.api_port
    if config_view_port is not None and params.view_port is not None:
        port_overrides[config_view_port] = params.view_port
    if config_db_port is not None and params.db_port is not None:
        port_overrides[config_db_port] = params.db_port
    if config_redis_port is not None and params.redis_port is not None:
        port_overrides[config_redis_port] = params.redis_port
    if config_caddy_admin_port is not None and params.caddy_admin_port is not None:
        port_overrides[config_caddy_admin_port] = params.caddy_admin_port
    if config_caddy_http_port is not None and params.caddy_http_port is not None:
        port_overrides[config_caddy_http_port] = params.caddy_http_port
    if config_caddy_https_port is not None and params.caddy_https_port is not None:
        port_overrides[config_caddy_https_port] = params.caddy_https_port
    if config_supertokens_port is not None and params.supertokens_port is not None:
        port_overrides[config_supertokens_port] = params.supertokens_port
    
    ports = []
    for config_port in config_ports:
        user_port = port_overrides.get(config_port)
        ports.append(user_port if user_port is not None else config_port)
    
    return ports


def run_preflight_checks(config: dict, params: InstallParams) -> None:
    ports = build_ports_to_check(config, params)
    check_required_ports(ports, logger=params.logger)


def install_dependencies(params: InstallParams) -> None:
    try:
        with timeout_wrapper(params.timeout):
            install_all_deps(verbose=params.verbose, output="json", dry_run=params.dry_run)
    except TimeoutError:
        raise Exception(dependency_installation_timeout)


def clone_repository_step(config_resolver: ConfigResolver, params: InstallParams) -> None:
    if params.dry_run:
        if params.logger:
            params.logger.info(
                f"[DRY RUN] Would clone {config_resolver.get(DEFAULT_REPO)} to {config_resolver.get('full_source_path')}"
            )
        return

    try:
        with timeout_wrapper(params.timeout):
            # Enable interactive mode if we're in a TTY environment
            import sys
            interactive = sys.stdin.isatty() and sys.stdout.isatty()
            
            success, error = clone_repository(
                repo=config_resolver.get(DEFAULT_REPO),
                path=config_resolver.get("full_source_path"),
                branch=config_resolver.get(DEFAULT_BRANCH),
                force=params.force,
                logger=params.logger,
                interactive=interactive,
            )
    except TimeoutError:
        raise Exception(f"{clone_failed}: {operation_timed_out}")
    
    if not success:
        raise Exception(f"{clone_failed}: {error}")


def cleanup_docker_step(config_resolver: ConfigResolver, params: InstallParams) -> None:
    compose_file = config_resolver.get("compose_file_path")
    cleanup_docker_services(compose_file, params.dry_run, params.logger)


def create_env_files_step(config: dict, config_resolver: ConfigResolver, params: InstallParams) -> None:
    success, error = create_service_env_files(
        config,
        config_resolver,
        get_host_ip_or_default(params.host_ip),
        params.api_domain,
        params.view_domain,
        external_db_url=params.external_db_url,
        logger=params.logger,
    )
    if not success:
        raise Exception(error)


def setup_proxy_config_step(config: dict, config_resolver: ConfigResolver, params: InstallParams) -> None:
    full_source_path = config_resolver.get("full_source_path")
    setup_proxy_configuration(
        full_source_path,
        get_host_ip_or_default(params.host_ip),
        params.view_domain,
        params.api_domain,
        config_resolver.get(VIEW_PORT),
        config_resolver.get(API_PORT),
        config,
        params.dry_run,
        params.logger,
    )


def setup_ssh_step(config: dict, config_resolver: ConfigResolver, params: InstallParams) -> None:
    ssh_config = SSHConfig(
        path=config_resolver.get("ssh_key_path"),
        key_type=get_config_value(config, SSH_KEY_TYPE),
        key_size=get_config_value(config, SSH_KEY_SIZE),
        passphrase=None,
        verbose=params.verbose,
        output="text",
        dry_run=params.dry_run,
        force=params.force,
        set_permissions=True,
        add_to_authorized_keys=True,
        create_ssh_directory=True,
    )
    try:
        with timeout_wrapper(params.timeout):
            result = generate_ssh_key_with_config(ssh_config, logger=params.logger)
    except TimeoutError:
        raise Exception(f"{ssh_setup_failed}: {operation_timed_out}")
    
    if not result.success:
        raise Exception(ssh_setup_failed)


def start_services_step(config_resolver: ConfigResolver, params: InstallParams) -> None:
    compose_file = config_resolver.get("compose_file_path")
    env_vars = build_service_env_vars(
        params.api_port,
        params.view_port,
        params.db_port,
        params.redis_port,
        params.caddy_admin_port,
        params.caddy_http_port,
        params.caddy_https_port,
        params.supertokens_port,
    )
    
    profiles = [] if params.external_db_url else ["local-db"]
    
    success, error = start_docker_services(
        compose_file,
        env_vars,
        params.timeout,
        params.dry_run,
        params.logger,
        profiles=profiles,
        verify_health=params.verify_health,
        health_check_timeout=params.health_check_timeout,
    )
    if not success:
        raise Exception(f"{services_start_failed}: {error}")


def load_proxy_step(config_resolver: ConfigResolver, params: InstallParams) -> None:    
    proxy_port = config_resolver.get(PROXY_PORT)
    try:
        proxy_port = int(proxy_port)
    except (ValueError, TypeError):
        proxy_port = 2019

    full_source_path = config_resolver.get("full_source_path")
    caddy_json_config = os.path.join(full_source_path, "helpers", "caddy.json")

    success, error = load_proxy_config(
        caddy_json_config,
        proxy_port,
        params.timeout,
        params.dry_run,
        params.logger,
    )
    if not success:
        raise Exception(f"{proxy_load_failed}: {error}")


def build_installation_steps(
    config: dict,
    config_resolver: ConfigResolver,
    params: InstallParams,
) -> List[Tuple[str, Callable[[], None]]]:
    services_step_desc = "Starting services" if not params.verify_health else "Starting and verifying services"
    
    steps = [
        ("Preflight checks", lambda: run_preflight_checks(config, params)),
        ("Installing dependencies", lambda: install_dependencies(params)),
        ("Cloning repository", lambda: clone_repository_step(config_resolver, params)),
        ("Setting up proxy config", lambda: setup_proxy_config_step(config, config_resolver, params)),
        ("Creating environment files", lambda: create_env_files_step(config, config_resolver, params)),
        ("Generating SSH keys", lambda: setup_ssh_step(config, config_resolver, params)),
        (services_step_desc, lambda: start_services_step(config_resolver, params)),
        ("Loading proxy configuration", lambda: load_proxy_step(config_resolver, params)),
    ]

    if params.force:
        steps.insert(2, ("Cleaning up Docker resources", lambda: cleanup_docker_step(config_resolver, params)))

    if (params.admin_email or params.admin_password) and params.verify_health:
        steps.append(("Registering admin user", lambda: register_admin_user_step(config_resolver, params)))

    return steps

def show_success_message(config_resolver: ConfigResolver, params: InstallParams) -> None:
    nixopus_accessible_at = get_access_url(
        params.view_domain,
        params.api_domain,
        get_host_ip_or_default(params.host_ip),
        config_resolver.get(VIEW_PORT),
    )

    if params.logger:
        params.logger.success("Installation Complete!")
        params.logger.info(f"Nixopus is accessible at: {nixopus_accessible_at}")
        params.logger.highlight("Thank you for installing Nixopus!")
        params.logger.info("Please visit the documentation at https://docs.nixopus.com for more information.")
        params.logger.info("If you have any questions, please visit the community forum at https://discord.gg/skdcq39Wpv")
        params.logger.highlight("See you in the community!")
    
    try:
        track_installation_success(staging=params.staging, logger=params.logger)
    except Exception:
        pass


def handle_installation_error(error: Exception, params: InstallParams, context: str = "") -> None:  
    if not params.logger:
        return
    
    context_msg = f" during {context}" if context else ""
    if params.verbose:
        params.logger.error(f"{installation_failed}{context_msg}: {str(error)}")
    else:
        params.logger.error(f"{installation_failed}{context_msg}")
    
    try:
        track_installation_failure(
            failed_step=context,
            staging=params.staging,
            error_message=str(error) if params.verbose else None,
            logger=params.logger,
        )
    except Exception:
        pass


def run_installation(params: InstallParams) -> None:
    config = get_active_config(user_config_file=params.config_file)
    
    validate_install_params(params)
    
    if is_custom_repo_or_branch(params.repo, params.branch):
        if params.logger:
            compose_file = "docker-compose-staging.yml" if params.staging else "docker-compose.yml"
            params.logger.info(f"Custom repository/branch detected - will use {compose_file}")
    
    if params.staging:
        try:
            track_staging_installation(logger=params.logger)
        except Exception:
            pass
    
    config_resolver = create_config_resolver(config, params)
    steps = build_installation_steps(config, config_resolver, params)
    completed_steps = []
    failed_step = None

    try:
        with Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]{task.description}"),
            BarColumn(),
            TaskProgressColumn(),
            transient=True,
            refresh_per_second=2,
        ) as progress:
            main_task = progress.add_task(installing_nixopus, total=len(steps))

            for i, (step_name, step_func) in enumerate(steps):
                progress.update(main_task, description=f"{installing_nixopus} - {step_name} ({i+1}/{len(steps)})")
                try:
                    step_func()
                    completed_steps.append(step_name)
                    progress.advance(main_task, 1)
                except Exception as e:
                    progress.update(main_task, description=f"Failed at {step_name}")
                    failed_step = step_name
                    raise

            progress.update(main_task, completed=True, description="Installation completed")

        show_success_message(config_resolver, params)

    except Exception as e:
        handle_installation_error(e, params, failed_step or "")
        
        if not params.no_rollback and completed_steps:
            perform_installation_rollback(
                completed_steps,
                config_resolver,
                config,
                params.dry_run,
                params.logger,
            )
        
        if params.logger:
            params.logger.error(f"{installation_failed}: {str(e)}")
        raise typer.Exit(1)
