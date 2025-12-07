import os
from dataclasses import dataclass
from typing import Callable, List, Optional, Tuple

import typer
from rich.progress import BarColumn, Progress, SpinnerColumn, TaskProgressColumn, TextColumn

from app.commands.clone.clone import clone_repository
from app.commands.preflight.preflight import check_required_ports
from app.utils.config import (
    API_PORT,
    DEFAULT_BRANCH,
    DEFAULT_REPO,
    PORTS,
    PROXY_PORT,
    SSH_KEY_SIZE,
    SSH_KEY_TYPE,
    VIEW_PORT,
    Config,
)
from app.utils.protocols import LoggerProtocol
from app.utils.timeout import timeout_wrapper

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
from .ssh import SSHConfig, generate_ssh_key_with_config
from .validate import validate_domains, validate_host_ip, validate_repo


@dataclass
class InstallParams:
    logger: Optional[LoggerProtocol] = None
    verbose: bool = False
    timeout: int = 300
    force: bool = False
    dry_run: bool = False
    config_file: Optional[str] = None
    api_domain: Optional[str] = None
    view_domain: Optional[str] = None
    host_ip: Optional[str] = None
    repo: Optional[str] = None
    branch: Optional[str] = None
    api_port: Optional[int] = None
    view_port: Optional[int] = None
    db_port: Optional[int] = None
    redis_port: Optional[int] = None
    caddy_admin_port: Optional[int] = None
    caddy_http_port: Optional[int] = None
    caddy_https_port: Optional[int] = None
    supertokens_port: Optional[int] = None


def validate_install_params(params: InstallParams) -> None:
    validate_domains(params.api_domain, params.view_domain)
    validate_repo(params.repo)
    validate_host_ip(params.host_ip)


def create_config_resolver(config: Config, params: InstallParams) -> ConfigResolver:
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
    )


def run_preflight_checks(config: Config, params: InstallParams) -> None:
    ports = config.get(PORTS)
    ports = [int(port) for port in ports] if isinstance(ports, list) else [int(ports)]
    check_required_ports(ports, logger=params.logger)


def install_dependencies(params: InstallParams) -> None:
    try:
        with timeout_wrapper(params.timeout):
            result = install_all_deps(verbose=params.verbose, output="json", dry_run=params.dry_run)
    except TimeoutError:
        raise Exception(dependency_installation_timeout)


def setup_clone_and_config(config_resolver: ConfigResolver, params: InstallParams) -> None:
    if params.dry_run:
        if params.logger:
            params.logger.info(
                f"[DRY RUN] Would clone {config_resolver.get(DEFAULT_REPO)} to {config_resolver.get('full_source_path')}"
            )
        return

    try:
        with timeout_wrapper(params.timeout):
            success, error = clone_repository(
                repo=config_resolver.get(DEFAULT_REPO),
                path=config_resolver.get("full_source_path"),
                branch=config_resolver.get(DEFAULT_BRANCH),
                force=params.force,
                logger=params.logger,
            )
    except TimeoutError:
        raise Exception(f"{clone_failed}: {operation_timed_out}")
    if not success:
        raise Exception(f"{clone_failed}: {error}")


def cleanup_docker_step(config_resolver: ConfigResolver, params: InstallParams) -> None:
    compose_file = config_resolver.get("compose_file_path")
    cleanup_docker_services(compose_file, params.dry_run, params.logger)


def create_env_files_step(config: Config, config_resolver: ConfigResolver, params: InstallParams) -> None:
    success, error = create_service_env_files(
        config,
        config_resolver,
        get_host_ip_or_default(params.host_ip),
        params.api_domain,
        params.view_domain,
        params.logger,
    )
    if not success:
        raise Exception(error)


def setup_proxy_config_step(config: Config, config_resolver: ConfigResolver, params: InstallParams) -> None:
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


def setup_ssh_step(config: Config, config_resolver: ConfigResolver, params: InstallParams) -> None:
    ssh_config = SSHConfig(
        path=config_resolver.get("ssh_key_path"),
        key_type=config.get(SSH_KEY_TYPE),
        key_size=config.get(SSH_KEY_SIZE),
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
    success, error = start_docker_services(
        compose_file,
        env_vars,
        params.timeout,
        params.dry_run,
        params.logger,
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


def handle_installation_error(error: Exception, params: InstallParams, context: str = "") -> None:
    if not params.logger:
        return
    
    context_msg = f" during {context}" if context else ""
    if params.verbose:
        params.logger.error(f"{installation_failed}{context_msg}: {str(error)}")
    else:
        params.logger.error(f"{installation_failed}{context_msg}")


def build_installation_steps(
    config: Config,
    config_resolver: ConfigResolver,
    params: InstallParams,
) -> List[Tuple[str, Callable[[], None]]]:
    steps = [
        ("Preflight checks", lambda: run_preflight_checks(config, params)),
        ("Installing dependencies", lambda: install_dependencies(params)),
        ("Cloning repository", lambda: setup_clone_and_config(config_resolver, params)),
        ("Setting up proxy config", lambda: setup_proxy_config_step(config, config_resolver, params)),
        ("Creating environment files", lambda: create_env_files_step(config, config_resolver, params)),
        ("Generating SSH keys", lambda: setup_ssh_step(config, config_resolver, params)),
        ("Starting services", lambda: start_services_step(config_resolver, params)),
    ]

    if params.force:
        steps.insert(2, ("Cleaning up Docker resources", lambda: cleanup_docker_step(config_resolver, params)))

    if params.api_domain and params.view_domain:
        steps.append(("Loading proxy configuration", lambda: load_proxy_step(config_resolver, params)))

    return steps


def run_installation(params: InstallParams) -> None:
    config = Config()
    config.load_user_config(params.config_file)
    
    validate_install_params(params)
    
    if is_custom_repo_or_branch(params.repo, params.branch):
        if params.logger:
            params.logger.info("Custom repository/branch detected - will use docker-compose-staging.yml")
    
    config_resolver = create_config_resolver(config, params)
    steps = build_installation_steps(config, config_resolver, params)

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
                    progress.advance(main_task, 1)
                except Exception as e:
                    progress.update(main_task, description=f"Failed at {step_name}")
                    raise

            progress.update(main_task, completed=True, description="Installation completed")

        show_success_message(config_resolver, params)

    except Exception as e:
        handle_installation_error(e, params)
        if params.logger:
            params.logger.error(f"{installation_failed}: {str(e)}")
        raise typer.Exit(1)


class Install:
    def __init__(
        self,
        logger: LoggerProtocol = None,
        verbose: bool = False,
        timeout: int = 300,
        force: bool = False,
        dry_run: bool = False,
        config_file: str = None,
        api_domain: str = None,
        view_domain: str = None,
        host_ip: str = None,
        repo: str = None,
        branch: str = None,
        api_port: int = None,
        view_port: int = None,
        db_port: int = None,
        redis_port: int = None,
        caddy_admin_port: int = None,
        caddy_http_port: int = None,
        caddy_https_port: int = None,
        supertokens_port: int = None,
    ):
        self.params = InstallParams(
            logger=logger,
            verbose=verbose,
            timeout=timeout,
            force=force,
            dry_run=dry_run,
            config_file=config_file,
            api_domain=api_domain,
            view_domain=view_domain,
            host_ip=host_ip,
            repo=repo,
            branch=branch,
            api_port=api_port,
            view_port=view_port,
            db_port=db_port,
            redis_port=redis_port,
            caddy_admin_port=caddy_admin_port,
            caddy_http_port=caddy_http_port,
            caddy_https_port=caddy_https_port,
            supertokens_port=supertokens_port,
        )

    def run(self):
        run_installation(self.params)
