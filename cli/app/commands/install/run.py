import ipaddress
import json
import os
import re
import shutil
import subprocess
from urllib.parse import urlparse

import typer
import yaml
from rich.progress import BarColumn, Progress, SpinnerColumn, TaskProgressColumn, TextColumn

from app.commands.clone.clone import clone_repository
from app.commands.conf.conf import write_env_file
from app.commands.preflight.preflight import check_required_ports
from app.commands.proxy.proxy import load_config
from app.commands.service.service import cleanup_docker_resources, start_services
from app.utils.config import (
    API_ENV_FILE,
    API_PORT,
    CADDY_ADMIN_PORT,
    CADDY_CONFIG_VOLUME,
    CADDY_HTTP_PORT,
    CADDY_HTTPS_PORT,
    DEFAULT_BRANCH,
    DEFAULT_COMPOSE_FILE,
    DEFAULT_PATH,
    DEFAULT_REPO,
    DOCKER_PORT,
    NIXOPUS_CONFIG_DIR,
    PORTS,
    PROXY_PORT,
    SSH_FILE_PATH,
    SSH_KEY_SIZE,
    SSH_KEY_TYPE,
    SUPERTOKENS_API_PORT,
    VIEW_ENV_FILE,
    VIEW_PORT,
    Config,
)
from app.utils.directory_manager import create_directory
from app.utils.file_manager import get_directory_path, set_permissions
from app.utils.host_information import get_public_ip
from app.utils.protocols import LoggerProtocol
from app.utils.timeout import timeout_wrapper

from .deps import install_all_deps
from .messages import (
    clone_failed,
    configuration_key_has_no_default_value,
    created_env_file,
    dependency_installation_timeout,
    env_file_creation_failed,
    env_file_permissions_failed,
    installation_failed,
    installing_nixopus,
    operation_timed_out,
    proxy_config_created,
    proxy_load_failed,
    services_start_failed,
    ssh_setup_failed,
)
from .ssh import SSHConfig, generate_ssh_key_with_config

_config = Config()


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
        self.logger = logger
        self.verbose = verbose
        self.timeout = timeout
        self.force = force
        self.dry_run = dry_run
        self.config_file = config_file
        self.api_domain = api_domain
        self.view_domain = view_domain
        self.host_ip = host_ip
        self.repo = repo
        self.branch = branch
        self.api_port = api_port
        self.view_port = view_port
        self.db_port = db_port
        self.redis_port = redis_port
        self.caddy_admin_port = caddy_admin_port
        self.caddy_http_port = caddy_http_port
        self.caddy_https_port = caddy_https_port
        self.supertokens_port = supertokens_port
        _config.load_user_config(self.config_file)
        self.progress = None
        self.main_task = None
        self._validate_domains()
        self._validate_repo()
        self._validate_host_ip()
        # Log when using custom repository/branch and staging compose file
        if self._is_custom_repo_or_branch():
            if self.logger:
                self.logger.info("Custom repository/branch detected - will use docker-compose-staging.yml")

    def _get_config(self, path: str):
        if path == DEFAULT_REPO and self.repo is not None:
            return self.repo
        if path == DEFAULT_BRANCH and self.branch is not None:
            return self.branch

        if path == "full_source_path":
            return os.path.join(_config.get(NIXOPUS_CONFIG_DIR), _config.get(DEFAULT_PATH))

        if path == "ssh_key_path":
            return os.path.join(_config.get(NIXOPUS_CONFIG_DIR), _config.get(SSH_FILE_PATH))

        if path == "compose_file_path":
            compose_path = os.path.join(_config.get(NIXOPUS_CONFIG_DIR), _config.get(DEFAULT_COMPOSE_FILE))
            if self._is_custom_repo_or_branch():
                return compose_path.replace("docker-compose.yml", "docker-compose-staging.yml")
            return compose_path

        if path == API_PORT and self.api_port is not None:
            return str(self.api_port)
        if path == VIEW_PORT and self.view_port is not None:
            return str(self.view_port)
        if path == "services.db.env.DB_PORT" and self.db_port is not None:
            return str(self.db_port)
        if path == "services.redis.env.REDIS_PORT" and self.redis_port is not None:
            return str(self.redis_port)
        
        # Handle PROXY_PORT (which is an alias for CADDY_ADMIN_PORT)
        if path == PROXY_PORT:
            if self.caddy_admin_port is not None:
                return str(self.caddy_admin_port)
            # Fall back to CADDY_ADMIN_PORT from config, default to 2019
            try:
                return str(_config.get(CADDY_ADMIN_PORT))
            except (KeyError, ValueError):
                return "2019"
        
        if path == CADDY_ADMIN_PORT and self.caddy_admin_port is not None:
            return str(self.caddy_admin_port)
        if path == CADDY_HTTP_PORT and self.caddy_http_port is not None:
            return str(self.caddy_http_port)
        if path == CADDY_HTTPS_PORT and self.caddy_https_port is not None:
            return str(self.caddy_https_port)
        if path == SUPERTOKENS_API_PORT and self.supertokens_port is not None:
            return str(self.supertokens_port)

        return _config.get(path)

    def _validate_domains(self):
        if (self.api_domain is None) != (self.view_domain is None):
            raise ValueError("Both api_domain and view_domain must be provided together, or neither should be provided")

        if self.api_domain and self.view_domain:
            domain_pattern = re.compile(
                r"^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?))*$"
            )
            if not domain_pattern.match(self.api_domain) or not domain_pattern.match(self.view_domain):
                raise ValueError("Invalid domain format. Domains must be valid hostnames")

    def _validate_repo(self):
        if self.repo:
            # Basic validation for repository URL format
            if not (
                self.repo.startswith(("http://", "https://", "git://", "ssh://"))
                or (self.repo.endswith(".git") and not self.repo.startswith("github.com:"))
                or ("@" in self.repo and ":" in self.repo and self.repo.count("@") == 1)
            ):
                raise ValueError("Invalid repository URL format")

    def _validate_host_ip(self):
        if self.host_ip:
            try:
                ipaddress.ip_address(self.host_ip)
            except ValueError:
                raise ValueError(f"Invalid IP address format: {self.host_ip}")

    def _is_custom_repo_or_branch(self):
        """Check if custom repository or branch is provided (different from defaults)"""
        temp_config = Config()
        default_repo = temp_config.get(DEFAULT_REPO)
        default_branch = temp_config.get(DEFAULT_BRANCH)

        # Check if either repo or branch differs from defaults
        repo_differs = self.repo is not None and self.repo != default_repo
        branch_differs = self.branch is not None and self.branch != default_branch

        return repo_differs or branch_differs

    def _get_host_ip(self) -> str:
        if self.host_ip:
            return self.host_ip
        return get_public_ip()

    def run(self):
        steps = [
            ("Preflight checks", self._run_preflight_checks),
            ("Installing dependencies", self._install_dependencies),
            ("Cloning repository", self._setup_clone_and_config),
            ("Setting up proxy config", self._setup_proxy_config),
            ("Creating environment files", self._create_env_files),
            ("Generating SSH keys", self._setup_ssh),
            ("Starting services", self._start_services),
        ]

        # If force is enabled, add a Docker cleanup step before cloning to ensure fresh images/containers
        if self.force:
            # Insert cleanup as the third step (index 2), right before cloning the repo
            steps.insert(2, ("Cleaning up Docker resources", self._cleanup_docker))

        # Only add proxy steps if both api_domain and view_domain are provided
        if self.api_domain and self.view_domain:
            steps.append(("Loading proxy configuration", self._load_proxy))

        try:
            with Progress(
                SpinnerColumn(),
                TextColumn("[progress.description]{task.description}"),
                BarColumn(),
                TaskProgressColumn(),
                transient=True,
                refresh_per_second=2,
            ) as progress:
                self.progress = progress
                self.main_task = progress.add_task(installing_nixopus, total=len(steps))

                for i, (step_name, step_func) in enumerate(steps):
                    progress.update(self.main_task, description=f"{installing_nixopus} - {step_name} ({i+1}/{len(steps)})")
                    try:
                        step_func()
                        progress.advance(self.main_task, 1)
                    except Exception as e:
                        progress.update(self.main_task, description=f"Failed at {step_name}")
                        raise

                progress.update(self.main_task, completed=True, description="Installation completed")

            self._show_success_message()

        except Exception as e:
            self._handle_installation_error(e)
            self.logger.error(f"{installation_failed}: {str(e)}")
            raise typer.Exit(1)

    def _cleanup_docker(self):
        compose_file = self._get_config("compose_file_path")

        if self.dry_run:
            self.logger.info(
                f"[dry-run] Would run: docker compose -f {compose_file} down --rmi all --volumes --remove-orphans"
            )
            return

        if not compose_file or not os.path.exists(compose_file):
            # Nothing to clean specific to this deployment
            self.logger.debug(f"Compose file not found at {compose_file}; skipping docker cleanup")
            return

        # Use shared cleanup helper to keep behavior consistent across commands
        try:
            success, output = cleanup_docker_resources(
                compose_file=compose_file,
                logger=self.logger,
                remove_images="all",
                remove_volumes=True,
                remove_orphans=True,
            )
            if success:
                self.logger.info("Docker resources cleaned (images, volumes, orphans)")
            else:
                self.logger.warning("Docker cleanup did not fully succeed; continuing")
        except Exception as e:
            # best effort cleanup, hence safely ignore errors
            self.logger.warning(f"Failed to cleanup docker resources: {e}")

    def _handle_installation_error(self, error, context=""):
        context_msg = f" during {context}" if context else ""
        if self.verbose:
            self.logger.error(f"{installation_failed}{context_msg}: {str(error)}")
        else:
            self.logger.error(f"{installation_failed}{context_msg}")

    def _run_preflight_checks(self):
        ports = _config.get(PORTS)
        ports = [int(port) for port in ports] if isinstance(ports, list) else [int(ports)]
        check_required_ports(ports, logger=self.logger)

    def _install_dependencies(self):
        try:
            with timeout_wrapper(self.timeout):
                result = install_all_deps(verbose=self.verbose, output="json", dry_run=self.dry_run)
        except TimeoutError:
            raise Exception(dependency_installation_timeout)

    def _setup_clone_and_config(self):
        if self.dry_run:
            self.logger.info(f"[DRY RUN] Would clone {self._get_config(DEFAULT_REPO)} to {self._get_config('full_source_path')}")
            return

        try:
            with timeout_wrapper(self.timeout):
                success, error = clone_repository(
                    repo=self._get_config(DEFAULT_REPO),
                    path=self._get_config("full_source_path"),
                    branch=self._get_config(DEFAULT_BRANCH),
                    force=self.force,
                    logger=self.logger,
                )
        except TimeoutError:
            raise Exception(f"{clone_failed}: {operation_timed_out}")
        if not success:
            raise Exception(f"{clone_failed}: {error}")

    def _create_env_files(self):
        api_env_file = self._get_config(API_ENV_FILE)
        view_env_file = self._get_config(VIEW_ENV_FILE)

        full_source_path = self._get_config("full_source_path")
        combined_env_file = os.path.join(full_source_path, ".env")
        create_directory(get_directory_path(api_env_file), logger=self.logger)
        create_directory(get_directory_path(view_env_file), logger=self.logger)
        create_directory(get_directory_path(combined_env_file), logger=self.logger)

        services = [
            ("api", "services.api.env", api_env_file),
            ("view", "services.view.env", view_env_file),
        ]
        for service_name, service_key, env_file in services:
            env_values = _config.get_service_env_values(service_key)
            updated_env_values = self._update_environment_variables(env_values)
            success, error = write_env_file(env_file, updated_env_values, self.logger)
            if not success:
                raise Exception(f"{env_file_creation_failed} {service_name}: {error}")
            file_perm_success, file_perm_error = set_permissions(env_file, 0o644)
            if not file_perm_success:
                raise Exception(f"{env_file_permissions_failed} {service_name}: {file_perm_error}")
            self.logger.debug(created_env_file.format(service_name=service_name, env_file=env_file))

        # combined env file with both API and view variables
        api_env_values = _config.get_service_env_values("services.api.env")
        view_env_values = _config.get_service_env_values("services.view.env")

        combined_env_values = {}
        combined_env_values.update(self._update_environment_variables(api_env_values))
        combined_env_values.update(self._update_environment_variables(view_env_values))
        success, error = write_env_file(combined_env_file, combined_env_values, self.logger)

        if not success:
            raise Exception(f"{env_file_creation_failed} combined: {error}")
        file_perm_success, file_perm_error = set_permissions(combined_env_file, 0o644)
        if not file_perm_success:
            raise Exception(f"{env_file_permissions_failed} combined: {file_perm_error}")
        self.logger.debug(created_env_file.format(service_name="combined", env_file=combined_env_file))

    def _setup_proxy_config(self):
        full_source_path = self._get_config("full_source_path")
        caddy_json_template = os.path.join(full_source_path, "helpers", "caddy.json")

        if not self.dry_run:
            with open(caddy_json_template, "r") as f:
                config_str = f.read()

            host_ip = self._get_host_ip()
            view_port = self._get_config(VIEW_PORT)
            api_port = self._get_config(API_PORT)

            view_domain = self.view_domain if self.view_domain is not None else host_ip
            api_domain = self.api_domain if self.api_domain is not None else host_ip

            config_str = config_str.replace("{env.APP_DOMAIN}", view_domain)
            config_str = config_str.replace("{env.API_DOMAIN}", api_domain)

            app_reverse_proxy_url = f"{host_ip}:{view_port}"
            api_reverse_proxy_url = f"{host_ip}:{api_port}"
            config_str = config_str.replace("{env.APP_REVERSE_PROXY_URL}", app_reverse_proxy_url)
            config_str = config_str.replace("{env.API_REVERSE_PROXY_URL}", api_reverse_proxy_url)

            caddy_config = json.loads(config_str)
            with open(caddy_json_template, "w") as f:
                json.dump(caddy_config, f, indent=2)
            self._copy_caddyfile_to_target(full_source_path)

        self.logger.debug(f"{proxy_config_created}: {caddy_json_template}")

    def _setup_ssh(self):
        config = SSHConfig(
            path=self._get_config("ssh_key_path"),
            key_type=_config.get(SSH_KEY_TYPE),
            key_size=_config.get(SSH_KEY_SIZE),
            passphrase=None,
            verbose=self.verbose,
            output="text",
            dry_run=self.dry_run,
            force=self.force,
            set_permissions=True,
            add_to_authorized_keys=True,
            create_ssh_directory=True,
        )
        try:
            with timeout_wrapper(self.timeout):
                result = generate_ssh_key_with_config(config, logger=self.logger)
        except TimeoutError:
            raise Exception(f"{ssh_setup_failed}: {operation_timed_out}")
        if not result.success:
            raise Exception(ssh_setup_failed)

    def _start_services(self):
        compose_file = self._get_config("compose_file_path")
        
        if self.dry_run:
            self.logger.info(f"[DRY RUN] Would start services using {compose_file}")
            return
        
        env_vars = {}
        if self.api_port is not None:
            env_vars["API_PORT"] = str(self.api_port)
        if self.view_port is not None:
            env_vars["NEXT_PUBLIC_PORT"] = str(self.view_port)
            env_vars["VIEW_PORT"] = str(self.view_port)
        if self.db_port is not None:
            env_vars["DB_PORT"] = str(self.db_port)
        if self.redis_port is not None:
            env_vars["REDIS_PORT"] = str(self.redis_port)
        if self.caddy_admin_port is not None:
            env_vars["CADDY_ADMIN_PORT"] = str(self.caddy_admin_port)
        if self.caddy_http_port is not None:
            env_vars["CADDY_HTTP_PORT"] = str(self.caddy_http_port)
        if self.caddy_https_port is not None:
            env_vars["CADDY_HTTPS_PORT"] = str(self.caddy_https_port)
        if self.supertokens_port is not None:
            env_vars["SUPERTOKENS_PORT"] = str(self.supertokens_port)

        original_env = os.environ.copy()
        os.environ.update(env_vars)

        try:
            try:
                with timeout_wrapper(self.timeout):
                    success, error = start_services(
                        name="all",
                        detach=True,
                        env_file=None,
                        compose_file=compose_file,
                        logger=self.logger,
                    )
            except TimeoutError:
                raise Exception(f"{services_start_failed}: {operation_timed_out}")
            if not success:
                raise Exception(f"{services_start_failed}: {error}")
        finally:
            for key in env_vars:
                if key in original_env:
                    os.environ[key] = original_env[key]
                else:
                    os.environ.pop(key, None)

    def _load_proxy(self):
        proxy_port = self._get_config(PROXY_PORT)
        # Ensure proxy_port is an integer
        try:
            proxy_port = int(proxy_port)
        except (ValueError, TypeError):
            proxy_port = 2019  # Default fallback
        
        full_source_path = self._get_config("full_source_path")
        caddy_json_config = os.path.join(full_source_path, "helpers", "caddy.json")
        
        if self.dry_run:
            self.logger.info(f"[DRY RUN] Would load proxy config from {caddy_json_config}")
            return

        try:
            with timeout_wrapper(self.timeout):
                success, error = load_config(caddy_json_config, proxy_port, self.logger)
        except TimeoutError:
            raise Exception(f"{proxy_load_failed}: {operation_timed_out}")

        if success:
            if not self.dry_run:
                self.logger.success("Caddy proxy configuration loaded successfully")
        else:
            self.logger.error(error)
            raise Exception(f"{proxy_load_failed}: {error}")

    def _show_success_message(self):
        nixopus_accessible_at = self._get_access_url()

        self.logger.success("Installation Complete!")
        self.logger.info(f"Nixopus is accessible at: {nixopus_accessible_at}")
        self.logger.highlight("Thank you for installing Nixopus!")
        self.logger.info("Please visit the documentation at https://docs.nixopus.com for more information.")
        self.logger.info("If you have any questions, please visit the community forum at https://discord.gg/skdcq39Wpv")
        self.logger.highlight("See you in the community!")

    def _get_supertokens_connection_uri(self, protocol: str, api_host: str, supertokens_api_port: int, host_ip: str):
        protocol = protocol.replace("https", "http")
        
        host_without_port = api_host
        
        if api_host.startswith(("http://", "https://")):
            parsed = urlparse(api_host)
            host_without_port = parsed.hostname or api_host
        elif ":" in api_host:
            parts = api_host.rsplit(":", 1)
            if len(parts) == 2 and parts[1].isdigit():
                host_without_port = parts[0]
        
        try:
            ipaddress.ip_address(host_without_port)
            return f"{protocol}://{host_ip}:{supertokens_api_port}"
        except ValueError:
            return f"{protocol}://{host_without_port}:{supertokens_api_port}"

    def _update_environment_variables(self, env_values: dict) -> dict:
        updated_env = env_values.copy()
        host_ip = self._get_host_ip()
        secure = self.api_domain is not None and self.view_domain is not None

        api_host = self.api_domain if secure else f"{host_ip}:{self._get_config(API_PORT)}"
        view_host = self.view_domain if secure else f"{host_ip}:{self._get_config(VIEW_PORT)}"
        protocol = "https" if secure else "http"
        ws_protocol = "wss" if secure else "ws"
        supertokens_api_port = self._get_config(SUPERTOKENS_API_PORT) or 3567
        key_map = {
            "ALLOWED_ORIGIN": f"{protocol}://{view_host}",
            "SSH_HOST": host_ip,
            "SSH_PRIVATE_KEY": self._get_config("ssh_key_path"),
            "WEBSOCKET_URL": f"{ws_protocol}://{api_host}/ws",
            "API_URL": f"{protocol}://{api_host}/api",
            "WEBHOOK_URL": f"{protocol}://{api_host}/api/v1/webhook",
            "VIEW_DOMAIN": f"{protocol}://{view_host}",
            "SUPERTOKENS_API_KEY": "NixopusSuperTokensAPIKey",
            "SUPERTOKENS_API_DOMAIN": f"{protocol}://{api_host}/api",
            "SUPERTOKENS_WEBSITE_DOMAIN": f"{protocol}://{view_host}",
            # TODO: temp fix, remove this once we have a secure connection
            "SUPERTOKENS_CONNECTION_URI": self._get_supertokens_connection_uri(
                protocol, api_host, supertokens_api_port, host_ip
            ),
        }

        for key, value in key_map.items():
            if key in updated_env:
                updated_env[key] = value

        return updated_env

    def _copy_caddyfile_to_target(self, full_source_path: str):
        try:
            source_caddyfile = os.path.join(full_source_path, "helpers", "Caddyfile")
            target_dir = _config.get(CADDY_CONFIG_VOLUME)
            target_caddyfile = os.path.join(target_dir, "Caddyfile")
            create_directory(target_dir, logger=self.logger)
            if os.path.exists(source_caddyfile):
                shutil.copy2(source_caddyfile, target_caddyfile)
                set_permissions(target_caddyfile, 0o644, logger=self.logger)
                self.logger.debug(f"Copied Caddyfile from {source_caddyfile} to {target_caddyfile}")
            else:
                self.logger.warning(f"Source Caddyfile not found at {source_caddyfile}")

        except Exception as e:
            self.logger.error(f"Failed to copy Caddyfile: {str(e)}")

    def _get_access_url(self):
        if self.view_domain:
            return f"https://{self.view_domain}"
        elif self.api_domain:
            return f"https://{self.api_domain}"
        else:
            view_port = self._get_config(VIEW_PORT)
            host_ip = self._get_host_ip()
            return f"http://{host_ip}:{view_port}"
