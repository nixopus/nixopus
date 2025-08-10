import typer
import os
import yaml
import json
import shutil
from rich.progress import Progress, SpinnerColumn, TextColumn, BarColumn, TaskProgressColumn
from app.utils.protocols import LoggerProtocol
from app.utils.config import Config, VIEW_ENV_FILE, API_ENV_FILE, DEFAULT_REPO, DEFAULT_BRANCH, DEFAULT_PATH, NIXOPUS_CONFIG_DIR, PORTS, DEFAULT_COMPOSE_FILE, PROXY_PORT, SSH_KEY_TYPE, SSH_KEY_SIZE, SSH_FILE_PATH, VIEW_PORT, API_PORT, DOCKER_PORT, CADDY_CONFIG_VOLUME
from app.utils.timeout import TimeoutWrapper
from app.commands.preflight.run import PreflightRunner
from app.commands.clone.clone import Clone, CloneConfig
from app.utils.lib import HostInformation, FileManager
from app.commands.conf.base import BaseEnvironmentManager
import re
from app.commands.service.up import Up, UpConfig
from app.commands.proxy.load import Load, LoadConfig
from .ssh import SSH, SSHConfig
from .messages import (
    installation_failed, installing_nixopus,
    dependency_installation_timeout,
    clone_failed, env_file_creation_failed, env_file_permissions_failed, 
    proxy_config_created, ssh_setup_failed, services_start_failed, proxy_load_failed,
    operation_timed_out, created_env_file, configuration_key_has_no_default_value
)
from .deps import install_all_deps

_config = Config()
_config_dir = _config.get_yaml_value(NIXOPUS_CONFIG_DIR)
_source_path = _config.get_yaml_value(DEFAULT_PATH)

DEFAULTS = {
    'proxy_port': _config.get_yaml_value(PROXY_PORT),
    'ssh_key_type': _config.get_yaml_value(SSH_KEY_TYPE),   
    'ssh_key_size': _config.get_yaml_value(SSH_KEY_SIZE),
    'ssh_passphrase': None,
    'service_name': 'all',
    'service_detach': True,
    'required_ports': [int(port) for port in _config.get_yaml_value(PORTS)],
    'repo_url': _config.get_yaml_value(DEFAULT_REPO),
    'branch_name': _config.get_yaml_value(DEFAULT_BRANCH),
    'source_path': _source_path,
    'config_dir': _config_dir,
    'api_env_file_path': _config.get_yaml_value(API_ENV_FILE),
    'view_env_file_path': _config.get_yaml_value(VIEW_ENV_FILE),
    'compose_file': _config.get_yaml_value(DEFAULT_COMPOSE_FILE),
    'full_source_path': os.path.join(_config_dir, _source_path),
    'ssh_key_path': _config_dir + "/" + _config.get_yaml_value(SSH_FILE_PATH),
    'compose_file_path': _config_dir + "/" + _config.get_yaml_value(DEFAULT_COMPOSE_FILE),
    'host_os': HostInformation.get_os_name(),
    'package_manager': HostInformation.get_package_manager(),
    'view_port': _config.get_yaml_value(VIEW_PORT),
    'api_port': _config.get_yaml_value(API_PORT),   
    'docker_port': _config.get_yaml_value(DOCKER_PORT),
}


class Install:
    def __init__(self, logger: LoggerProtocol = None, verbose: bool = False, timeout: int = 300, force: bool = False, dry_run: bool = False, config_file: str = None, api_domain: str = None, view_domain: str = None):
        self.logger = logger
        self.verbose = verbose
        self.timeout = timeout
        self.force = force
        self.dry_run = dry_run
        self.config_file = config_file
        self.api_domain = api_domain
        self.view_domain = view_domain
        self._user_config = _config.load_user_config(self.config_file)
        self.progress = None
        self.main_task = None
        self._validate_domains()
    
    def _get_config(self, key: str):
        try:
            return _config.get_config_value(key, self._user_config, DEFAULTS)
        except ValueError:
            raise ValueError(configuration_key_has_no_default_value.format(key=key))
    
    def _validate_domains(self):
        if (self.api_domain is None) != (self.view_domain is None):
            raise ValueError("Both api_domain and view_domain must be provided together, or neither should be provided")
        
        if self.api_domain and self.view_domain:
            domain_pattern = re.compile(r'^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?))*$')
            if not domain_pattern.match(self.api_domain) or not domain_pattern.match(self.view_domain):
                raise ValueError("Invalid domain format. Domains must be valid hostnames")


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

    def _handle_installation_error(self, error, context=""):
        context_msg = f" during {context}" if context else ""
        if self.verbose:
            self.logger.error(f"{installation_failed}{context_msg}: {str(error)}")
        else:
            self.logger.error(f"{installation_failed}{context_msg}")

    def _run_preflight_checks(self):
        preflight_runner = PreflightRunner(logger=self.logger, verbose=self.verbose)
        preflight_runner.check_ports_from_config(
            config_key='required_ports', 
            user_config=self._user_config, 
            defaults=DEFAULTS
        )

    def _install_dependencies(self):
        try:
            with TimeoutWrapper(self.timeout):
                result = install_all_deps(verbose=self.verbose, output="json", dry_run=self.dry_run)
        except TimeoutError:
            raise Exception(dependency_installation_timeout)

    def _setup_clone_and_config(self):        
        clone_config = CloneConfig(
            repo=self._get_config('repo_url'),
            branch=self._get_config('branch_name'),
            path=self._get_config('full_source_path'),
            force=self.force,
            verbose=self.verbose,
            output="text",
            dry_run=self.dry_run
        )
        clone_service = Clone(logger=self.logger)
        try:
            with TimeoutWrapper(self.timeout):
                result = clone_service.clone(clone_config)
        except TimeoutError:
            raise Exception(f"{clone_failed}: {operation_timed_out}")
        if not result.success:
            raise Exception(f"{clone_failed}: {result.error}")

    def _create_env_files(self):        
        api_env_file = self._get_config('api_env_file_path')
        view_env_file = self._get_config('view_env_file_path')
        FileManager.create_directory(FileManager.get_directory_path(api_env_file), logger=self.logger)
        FileManager.create_directory(FileManager.get_directory_path(view_env_file), logger=self.logger)
        services = [
            ("api", "services.api.env", api_env_file),
            ("view", "services.view.env", view_env_file),
        ]
        env_manager = BaseEnvironmentManager(self.logger)
        
        for i, (service_name, service_key, env_file) in enumerate(services):            
            env_values = _config.get_service_env_values(service_key)
            updated_env_values = self._update_environment_variables(env_values)
            success, error = env_manager.write_env_file(env_file, updated_env_values)
            if not success:
                raise Exception(f"{env_file_creation_failed} {service_name}: {error}")            
            file_perm_success, file_perm_error = FileManager.set_permissions(env_file, 0o644)
            if not file_perm_success:
                raise Exception(f"{env_file_permissions_failed} {service_name}: {file_perm_error}")            
            self.logger.debug(created_env_file.format(service_name=service_name, env_file=env_file))

    def _setup_proxy_config(self):
        full_source_path = self._get_config('full_source_path')
        caddy_json_template = os.path.join(full_source_path, 'helpers', 'caddy.json')
        
        if not self.dry_run:
            with open(caddy_json_template, 'r') as f:
                config_str = f.read()
            
            host_ip = HostInformation.get_public_ip()
            view_port = self._get_config('view_port')
            api_port = self._get_config('api_port')

            view_domain = self.view_domain if self.view_domain is not None else host_ip
            api_domain = self.api_domain if self.api_domain is not None else host_ip

            config_str = config_str.replace('{env.APP_DOMAIN}', view_domain)
            config_str = config_str.replace('{env.API_DOMAIN}', api_domain)
            
            app_reverse_proxy_url = f"{host_ip}:{view_port}"
            api_reverse_proxy_url = f"{host_ip}:{api_port}"
            config_str = config_str.replace('{env.APP_REVERSE_PROXY_URL}', app_reverse_proxy_url)
            config_str = config_str.replace('{env.API_REVERSE_PROXY_URL}', api_reverse_proxy_url)
            
            caddy_config = json.loads(config_str)
            with open(caddy_json_template, 'w') as f:
                json.dump(caddy_config, f, indent=2)
            self._copy_caddyfile_to_target(full_source_path)
        
        self.logger.debug(f"{proxy_config_created}: {caddy_json_template}")

    def _setup_ssh(self):
        config = SSHConfig(
            path=self._get_config('ssh_key_path'),
            key_type=self._get_config('ssh_key_type'),
            key_size=self._get_config('ssh_key_size'),
            passphrase=self._get_config('ssh_passphrase'),
            verbose=self.verbose,
            output="text",
            dry_run=self.dry_run,
            force=self.force,
            set_permissions=True,
            add_to_authorized_keys=True,
            create_ssh_directory=True,
        )
        ssh_operation = SSH(logger=self.logger)
        try:
            with TimeoutWrapper(self.timeout):
                result = ssh_operation.generate(config)
        except TimeoutError:
            raise Exception(f"{ssh_setup_failed}: {operation_timed_out}")
        if not result.success:
            raise Exception(ssh_setup_failed)

    def _start_services(self):
        config = UpConfig(
            name=self._get_config('service_name'),
            detach=self._get_config('service_detach'),
            env_file=None,
            verbose=self.verbose,
            output="text",
            dry_run=self.dry_run,
            compose_file=self._get_config('compose_file_path')
        )

        up_service = Up(logger=self.logger)
        try:
            with TimeoutWrapper(self.timeout):
                result = up_service.up(config)
        except TimeoutError:
            raise Exception(f"{services_start_failed}: {operation_timed_out}")
        if not result.success:
            raise Exception(services_start_failed)

    def _load_proxy(self):
        proxy_port = self._get_config('proxy_port')
        full_source_path = self._get_config('full_source_path')
        caddy_json_config = os.path.join(full_source_path, 'helpers', 'caddy.json')
        config = LoadConfig(proxy_port=proxy_port, verbose=self.verbose, output="text", dry_run=self.dry_run, config_file=caddy_json_config)

        load_service = Load(logger=self.logger)
        try:
            with TimeoutWrapper(self.timeout):
                result = load_service.load(config)
        except TimeoutError:
            raise Exception(f"{proxy_load_failed}: {operation_timed_out}")

        if result.success:
            if not self.dry_run:
                self.logger.success(load_service.format_output(result, "text"))
        else:
            self.logger.error(result.error)
            raise Exception(proxy_load_failed)

    def _show_success_message(self):
        nixopus_accessible_at = self._get_access_url()
        
        self.logger.success("Installation Complete!")
        self.logger.info(f"Nixopus is accessible at: {nixopus_accessible_at}")
        self.logger.highlight("Thank you for installing Nixopus!")
        self.logger.info("Please visit the documentation at https://docs.nixopus.com for more information.")
        self.logger.info("If you have any questions, please visit the community forum at https://discord.gg/skdcq39Wpv")
        self.logger.highlight("See you in the community!")
    
    def _update_environment_variables(self, env_values: dict) -> dict:
        updated_env = env_values.copy()
        host_ip = HostInformation.get_public_ip()
        secure = self.api_domain is not None and self.view_domain is not None

        api_host = self.api_domain if secure else f"{host_ip}:{self._get_config('api_port')}"
        view_host = self.view_domain if secure else f"{host_ip}:{self._get_config('view_port')}"
        protocol = "https" if secure else "http"
        ws_protocol = "wss" if secure else "ws"
        key_map = {
            'ALLOWED_ORIGIN': f"{protocol}://{view_host}",
            'SSH_HOST': host_ip,
            'SSH_PRIVATE_KEY': self._get_config('ssh_key_path'),
            'WEBSOCKET_URL': f"{ws_protocol}://{api_host}/ws",
            'API_URL': f"{protocol}://{api_host}/api",
            'WEBHOOK_URL': f"{protocol}://{api_host}/api/v1/webhook",
        }

        for key, value in key_map.items():
            if key in updated_env:
                updated_env[key] = value

        return updated_env

    def _copy_caddyfile_to_target(self, full_source_path: str):
        try:
            source_caddyfile = os.path.join(full_source_path, 'helpers', 'Caddyfile')
            target_dir = _config.get_yaml_value(CADDY_CONFIG_VOLUME)
            target_caddyfile = os.path.join(target_dir, 'Caddyfile')
            FileManager.create_directory(target_dir, logger=self.logger)
            if os.path.exists(source_caddyfile):
                shutil.copy2(source_caddyfile, target_caddyfile)
                FileManager.set_permissions(target_caddyfile, 0o644, logger=self.logger)
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
            view_port = self._get_config('view_port')
            host_ip = HostInformation.get_public_ip()
            return f"http://{host_ip}:{view_port}"
