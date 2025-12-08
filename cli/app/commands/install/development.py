import os
import platform
import subprocess
import time
import urllib.request

import typer
from rich.progress import BarColumn, Progress, SpinnerColumn, TaskProgressColumn, TextColumn

from app.commands.clone.clone import clone_repository
from app.commands.conf.conf import write_env_file
from app.commands.preflight.preflight import check_ports_from_config
from app.commands.proxy.proxy import load_config
from app.commands.service.service import cleanup_docker_resources, start_services
from app.utils.config import (
    API_ENV_FILE,
    API_PORT,
    DEFAULT_BRANCH,
    DEFAULT_COMPOSE_FILE,
    DEFAULT_PATH,
    DEFAULT_REPO,
    NIXOPUS_CONFIG_DIR,
    PORTS,
    PROXY_PORT,
    SSH_FILE_PATH,
    SSH_KEY_SIZE,
    SSH_KEY_TYPE,
    VIEW_ENV_FILE,
    VIEW_PORT,
    get_active_config,
    get_config_value,
    get_service_env_values,
    get_yaml_value,
)
from app.utils.directory_manager import create_directory
from app.utils.file_manager import get_directory_path, set_permissions
from app.utils.host_information import get_os_name, get_package_manager
from app.utils.protocols import LoggerProtocol
from app.utils.timeout import timeout_wrapper

from .base import BaseInstall
from .deps import get_deps_from_config, get_installed_deps, install_dep
from .services import build_service_env_vars
from .messages import (
    clone_failed,
    created_env_file,
    env_file_creation_failed,
    env_file_permissions_failed,
    installation_failed,
    installing_nixopus,
    operation_timed_out,
    services_start_failed,
    ssh_setup_failed,
)
from .ssh import SSHConfig, generate_ssh_key_with_config


class DevelopmentInstall(BaseInstall):
    """Development installation flow - installs to current directory with auto-start"""

    def __init__(
        self,
        logger: LoggerProtocol = None,
        verbose: bool = False,
        timeout: int = 300,
        force: bool = False,
        dry_run: bool = False,
        config_file: str = None,
        repo: str = None,
        branch: str = None,
        install_path: str = None,
        api_port: int = None,
        view_port: int = None,
        db_port: int = None,
        redis_port: int = None,
        caddy_admin_port: int = None,
        caddy_http_port: int = None,
        caddy_https_port: int = None,
        supertokens_port: int = None,
        external_db_url: str = None,
    ):
        super().__init__(
            logger=logger,
            verbose=verbose,
            timeout=timeout,
            force=force,
            dry_run=dry_run,
            config_file=config_file,
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
            external_db_url=external_db_url,
        )

        # safe fallback incase cwd is not accessible
        if install_path:
            self.install_path = os.path.abspath(os.path.expanduser(install_path))
        else:
            try:
                self.install_path = os.getcwd()
            except (FileNotFoundError, OSError) as e:
                # cwd is not accessible
                # Fall back to user's home directory
                self.install_path = os.path.expanduser("~/nixopus-dev")
                os.makedirs(self.install_path, exist_ok=True)
                if logger:
                    logger.warning(f"Current directory is not accessible: {e}")
                    logger.info(f"Using default installation path: {self.install_path}")

        # Check platform and WSL requirement for Windows
        self._check_platform_support()

        # Load config from config.dev.yaml
        self._config = get_active_config(default_env="DEVELOPMENT")
        self._defaults = self._load_dev_defaults()

        if self.logger:
            self.logger.info(f"Development mode - installing to: {self.install_path}")

    def _check_platform_support(self):
        """Check if platform is supported for development"""
        if platform.system() != "Windows":
            return

        # Check if running in WSL
        is_wsl = False
        try:
            if os.path.exists("/proc/version"):
                with open("/proc/version", "r") as f:
                    is_wsl = "microsoft" in f.read().lower() or "wsl" in f.read().lower()
        except Exception:
            is_wsl = False

        if is_wsl:
            if self.verbose:
                self.logger.info("Running in WSL2 - full support available")
            return

        # Native Windows - show detailed guidance
        self.logger.warning("=" * 70)
        self.logger.warning("Running on Native Windows")
        self.logger.warning("=" * 70)
        self.logger.warning("")
        self.logger.warning("Nixopus development requires WSL2 for full functionality.")
        self.logger.warning("")
        self.logger.warning("What works on native Windows:")
        self.logger.warning("  + Running API/View containers with hot reload")
        self.logger.warning("  + Accessing at http://app.localhost")
        self.logger.warning("  + Database, Redis, SuperTokens, Caddy")
        self.logger.warning("")
        self.logger.warning("What requires WSL2:")
        self.logger.warning("  - Deploying applications (SSH/SFTP access needed)")
        self.logger.warning("  - Container-to-host filesystem access")
        self.logger.warning("  - Building Docker images from host directories")
        self.logger.warning("")
        self.logger.warning("Why WSL2?")
        self.logger.warning("  * Native SSH server for container-to-host communication")
        self.logger.warning("  * Unix-compatible filesystem paths")
        self.logger.warning("  * Full Docker Desktop integration")
        self.logger.warning("")
        self.logger.info("Install WSL2 (5 minutes):")
        self.logger.info("  1. Open PowerShell as Administrator")
        self.logger.info("  2. Run: wsl --install")
        self.logger.info("  3. Restart your computer")
        self.logger.info("  4. Run this command in WSL2 terminal")
        self.logger.info("")
        self.logger.info("Documentation:")
        self.logger.info("  https://docs.microsoft.com/en-us/windows/wsl/install")
        self.logger.info("")
        self.logger.error("Development installation requires macOS, Linux, or WSL2")
        self.logger.error("Native Windows is not supported due to SSH/filesystem requirements")
        raise typer.Exit(1)

    def _load_dev_defaults(self):
        """Load defaults from config.dev.yaml"""
        config_dir = get_yaml_value(self._config, NIXOPUS_CONFIG_DIR)
        source_path = get_yaml_value(self._config, DEFAULT_PATH)

        return {
            "ssh_key_type": get_yaml_value(self._config, SSH_KEY_TYPE),
            "ssh_key_size": get_yaml_value(self._config, SSH_KEY_SIZE),
            "ssh_passphrase": None,
            "service_name": "all",
            "service_detach": True,
            "required_ports": [int(port) for port in get_yaml_value(self._config, PORTS)],
            "repo_url": get_yaml_value(self._config, DEFAULT_REPO),
            "branch_name": get_yaml_value(self._config, DEFAULT_BRANCH),
            "source_path": source_path,
            "config_dir": config_dir,
            "api_env_file_path": get_yaml_value(self._config, API_ENV_FILE),
            "view_env_file_path": get_yaml_value(self._config, VIEW_ENV_FILE),
            "compose_file": get_yaml_value(self._config, DEFAULT_COMPOSE_FILE),
            "full_source_path": self.install_path,
            "ssh_key_path": os.path.expanduser("~/.ssh/id_rsa_nixopus"),
            "compose_file_path": os.path.join(self.install_path, "docker-compose-dev.yml"),
            "view_port": get_yaml_value(self._config, VIEW_PORT),
            "api_port": get_yaml_value(self._config, API_PORT),
            "proxy_port": get_yaml_value(self._config, PROXY_PORT),
            "nixopus_config_dir": os.path.join(self.install_path, "nixopus-dev"),
        }

    def _get_config(self, key: str, user_config=None, defaults=None):
        """Get config value with development-specific overrides"""
        # Development-specific path overrides
        if key == "compose_file_path":
            return os.path.join(self.install_path, "docker-compose-dev.yml")
        if key == "full_source_path":
            return self.install_path
        if key == "nixopus_config_dir":
            return os.path.join(self.install_path, "nixopus-dev")
        if key == "api_env_file_path":
            return os.path.join(self.install_path, "api", ".env")
        if key == "view_env_file_path":
            return os.path.join(self.install_path, "view", ".env")
        if key == "ssh_key_path":
            return os.path.expanduser("~/.ssh/id_rsa_nixopus")

        # Port overrides from CLI options
        if key == "api_port" and self.api_port is not None:
            return str(self.api_port)
        if key == "view_port" and self.view_port is not None:
            return str(self.view_port)
        if key == "db_port":
            if self.db_port is not None:
                return str(self.db_port)
            return str(get_config_value(self._config, "services.db.env.DB_PORT") or "5432")
        if key == "redis_port":
            if self.redis_port is not None:
                return str(self.redis_port)
            return str(get_config_value(self._config, "services.redis.env.REDIS_PORT") or "6379")
        if key == "proxy_port" and self.caddy_admin_port is not None:
            return str(self.caddy_admin_port)
        if key == "supertokens_api_port":
            if self.supertokens_port is not None:
                return str(self.supertokens_port)
            return str(get_config_value(self._config, "services.api.env.SUPERTOKENS_API_PORT") or "3567")
        if key == "services.caddy.env.CADDY_HTTP_PORT" and self.caddy_http_port is not None:
            return str(self.caddy_http_port)
        if key == "services.caddy.env.CADDY_HTTPS_PORT" and self.caddy_https_port is not None:
            return str(self.caddy_https_port)

        active_defaults = defaults or self._defaults
        if active_defaults and key in active_defaults:
            return active_defaults[key]

        try:
            return get_config_value(self._config, key)
        except KeyError:
            return super()._get_config(key)

    def run(self):
        """Execute development installation workflow"""
        steps = [
            ("Preflight checks", self._run_preflight_checks),
            ("Checking dependencies", self._check_and_install_dependencies),
            ("Cloning repository", self._setup_clone_and_config),
            ("Setting up proxy config", self._setup_proxy_config),
            ("Creating environment files", self._create_env_files),
            ("Generating SSH keys", self._setup_ssh),
            ("Starting all services", self._start_all_services),
            ("Loading proxy configuration", self._load_proxy),
            ("Validating services", self._validate_services),
        ]

        if self.force:

            def cleanup():
                compose_file = self._get_config("compose_file_path")
                if os.path.exists(compose_file):
                    try:
                        cleanup_docker_resources(
                            compose_file=compose_file,
                            logger=self.logger,
                            remove_images="all",
                            remove_volumes=True,
                            remove_orphans=True,
                        )
                    except Exception as e:
                        self.logger.warning(f"Docker cleanup failed: {e}")

            clone_index = next(i for i, (name, _) in enumerate(steps) if name == "Cloning repository")
            steps.insert(clone_index, ("Cleaning up Docker resources", cleanup))

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
            raise typer.Exit(1)

    def _handle_installation_error(self, error):
        """Handle installation errors with clean output"""
        if self.verbose:
            self.logger.error(f"{installation_failed}: {str(error)}")
        else:
            self.logger.error(str(error))

    def _run_preflight_checks(self):
        """Check ports and system requirements"""
        check_ports_from_config(logger=self.logger)

    def _check_and_install_dependencies(self):
        """Check dependencies and install only if missing"""
        deps = get_deps_from_config()
        os_name = get_os_name()
        package_manager = get_package_manager()

        if not package_manager:
            raise Exception("No supported package manager found")

        # Check which deps are installed
        installed = get_installed_deps(deps, os_name, package_manager, verbose=self.verbose)
        to_install = [dep for dep in deps if not installed.get(dep["name"])]

        if not to_install:
            if self.verbose:
                self.logger.info("All dependencies already installed")
            return

        # Install missing dependencies
        if not self.verbose:
            self.logger.info(f"Installing {len(to_install)} missing dependencies...")

        for dep in to_install:
            if self.verbose:
                self.logger.info(f"Installing {dep['name']}...")
            success = install_dep(dep, package_manager, self.logger, dry_run=self.dry_run)
            if not success and not self.dry_run:
                self.logger.warning(f"Failed to install {dep['name']}, continuing...")

        if not self.verbose:
            self.logger.info("Dependencies ready")

    def _setup_clone_and_config(self):
        """Clone repository to installation directory"""
        if self.dry_run:
            self.logger.info(f"[DRY RUN] Would clone {self._get_config('repo_url')} to {self._get_config('full_source_path')}")
            return

        try:
            with timeout_wrapper(self.timeout):
                success, error = clone_repository(
                    repo=self._get_config("repo_url"),
                    path=self._get_config("full_source_path"),
                    branch=self._get_config("branch_name"),
                    force=self.force,
                    logger=self.logger,
                )
        except TimeoutError:
            raise Exception(f"{clone_failed}: {operation_timed_out}")
        if not success:
            raise Exception(f"{clone_failed}: {error}")

    def _create_env_files(self):
        """Create environment files for backend and frontend"""
        api_env_file = self._get_config("api_env_file_path")
        view_env_file = self._get_config("view_env_file_path")

        if self.dry_run:
            self.logger.info(f"[DRY RUN] Would create environment files:")
            self.logger.info(f"  - API:  {api_env_file}")
            self.logger.info(f"  - View: {view_env_file}")
            return

        create_directory(get_directory_path(api_env_file), logger=self.logger)
        create_directory(get_directory_path(view_env_file), logger=self.logger)

        # Get combined env file path
        full_source_path = self._get_config("full_source_path")
        combined_env_file = os.path.join(full_source_path, ".env")
        create_directory(get_directory_path(combined_env_file), logger=self.logger)

        services = [
            ("api", "services.api.env", api_env_file),
            ("view", "services.view.env", view_env_file),
        ]

        # Create individual service env files
        for service_name, service_key, env_file in services:
            env_values = get_service_env_values(self._config, service_key)
            updated_env_values = self._update_environment_variables(env_values)
            success, error = write_env_file(env_file, updated_env_values, self.logger)
            if not success:
                raise Exception(f"{env_file_creation_failed} {service_name}: {error}")
            file_perm_success, file_perm_error = set_permissions(env_file, 0o644)
            if not file_perm_success:
                raise Exception(f"{env_file_permissions_failed} {service_name}: {file_perm_error}")
            self.logger.debug(created_env_file.format(service_name=service_name, env_file=env_file))

        # Create combined env file with both API and view variables (for docker-compose)
        api_env_values = get_service_env_values(self._config, "services.api.env")
        view_env_values = get_service_env_values(self._config, "services.view.env")

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

    def _update_environment_variables(self, env_values: dict) -> dict:
        """Update environment variables with development-specific values"""
        updated_env = env_values.copy()

        # Get values from config
        api_port = self._get_config("api_port") or "8080"
        view_port = self._get_config("view_port") or "3000"
        current_user = os.getenv("USER", "user")

        # Development-specific overrides
        key_map = {
            "SSH_HOST": "host.docker.internal",
            "SSH_USER": current_user,
            "SSH_PRIVATE_KEY": "/root/.ssh/id_rsa_nixopus",
            "SUPERTOKENS_API_DOMAIN": f"http://localhost:{api_port}",
            "SUPERTOKENS_WEBSITE_DOMAIN": f"http://localhost:{view_port}",
            "WEBSOCKET_URL": f"ws://localhost:{api_port}/ws",
            "API_URL": f"http://localhost:{api_port}/api",
            "NEXT_PUBLIC_API_URL": f"http://localhost:{api_port}/api",
            "NEXT_PUBLIC_WEBSITE_DOMAIN": f"http://localhost:{view_port}",
            "WEBHOOK_URL": f"http://localhost:{api_port}/api/v1/webhook",
        }

        for key, value in key_map.items():
            if key in updated_env:
                updated_env[key] = value

        return updated_env

    def _setup_ssh(self):
        """Generate SSH key and add to authorized_keys for localhost access"""
        config = SSHConfig(
            path=self._get_config("ssh_key_path"),
            key_type=self._get_config("ssh_key_type"),
            key_size=self._get_config("ssh_key_size"),
            passphrase=self._get_config("ssh_passphrase"),
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

        if not self.dry_run and self.verbose:
            self.logger.info("SSH key configured for local development with host access")

    def _setup_proxy_config(self):
        """Setup Caddy proxy configuration for development with localhost domains"""
        import json

        caddy_json_template = os.path.join(self.install_path, "helpers", "caddy.json")

        if self.dry_run:
            self.logger.info(f"[DRY RUN] Would setup proxy config at {caddy_json_template}")
            return

        if not os.path.exists(caddy_json_template):
            raise Exception(f"Caddy config template not found: {caddy_json_template}")

        with open(caddy_json_template, "r") as f:
            config_str = f.read()

        # Get ports from config
        view_port = self._get_config("view_port") or "3000"
        api_port = self._get_config("api_port") or "8080"

        # Use localhost domains for development
        view_domain = "app.localhost"
        api_domain = "api.localhost"

        # Use host.docker.internal for reverse proxy since containers talk to host
        app_reverse_proxy_url = f"host.docker.internal:{view_port}"
        api_reverse_proxy_url = f"host.docker.internal:{api_port}"

        # Replace placeholders
        config_str = config_str.replace("{env.APP_DOMAIN}", view_domain)
        config_str = config_str.replace("{env.API_DOMAIN}", api_domain)
        config_str = config_str.replace("{env.APP_REVERSE_PROXY_URL}", app_reverse_proxy_url)
        config_str = config_str.replace("{env.API_REVERSE_PROXY_URL}", api_reverse_proxy_url)

        # Parse and write back
        caddy_config = json.loads(config_str)

        # Ensure nixopus server has listen directive for both HTTP and HTTPS
        if "apps" in caddy_config and "http" in caddy_config["apps"]:
            if "servers" in caddy_config["apps"]["http"]:
                if "nixopus" in caddy_config["apps"]["http"]["servers"]:
                    server = caddy_config["apps"]["http"]["servers"]["nixopus"]
                    if "listen" not in server or not server["listen"]:
                        server["listen"] = [":80", ":443"]

        with open(caddy_json_template, "w") as f:
            json.dump(caddy_config, f, indent=2)

        if self.verbose:
            self.logger.info(f"Proxy config created for development:")
            self.logger.info(f"  - View:  http://{view_domain} → {app_reverse_proxy_url}")
            self.logger.info(f"  - API:   http://{api_domain} → {api_reverse_proxy_url}")

    def _load_proxy(self):
        """Load Caddy proxy configuration via Admin API"""
        proxy_port = int(self._get_config("proxy_port") or 2019)
        caddy_json_config = os.path.join(self.install_path, "helpers", "caddy.json")

        if self.dry_run:
            self.logger.info(f"[DRY RUN] Would load proxy config from {caddy_json_config}")
            return

        try:
            with timeout_wrapper(self.timeout):
                success, error = load_config(caddy_json_config, proxy_port, self.logger)
        except TimeoutError:
            raise Exception(f"Proxy load failed: {operation_timed_out}")

        if success:
            if not self.dry_run and self.verbose:
                self.logger.info("Caddy proxy configuration loaded successfully")
        else:
            raise Exception(f"Proxy load failed: {error}")

    def _start_all_services(self):
        """Start all services (API, View, DB, Redis, Caddy) using Docker Compose"""
        compose_file = self._get_config("compose_file_path")
        
        if self.dry_run:
            self.logger.info(f"[DRY RUN] Would start services using {compose_file}")
            return
        
        env_vars = build_service_env_vars(
            self.api_port,
            self.view_port,
            self.db_port,
            self.redis_port,
            self.caddy_admin_port,
            self.caddy_http_port,
            self.caddy_https_port,
            self.supertokens_port,
        )

        original_env = os.environ.copy()
        os.environ.update(env_vars)

        try:
            try:
                with timeout_wrapper(self.timeout):
                    success, error = start_services(
                        name=self._get_config("service_name"),
                        detach=self._get_config("service_detach"),
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

    def _validate_services(self):
        """Validate all Docker services are running and accessible through Caddy proxy"""
        if self.dry_run:
            self.logger.info("[DRY RUN] Would validate services")
            return

        if not self.verbose:
            self.logger.info("Validating services...")

        # Check API container health
        api_url = "http://api.localhost/api/v1/health"
        api_ready = False

        if self.verbose:
            self.logger.info("Checking API service...")

        for i in range(10):
            try:
                response = urllib.request.urlopen(api_url, timeout=5)
                if response.status == 200:
                    api_ready = True
                    break
            except Exception:
                if i < 9:
                    time.sleep(3)

        if self.verbose:
            if api_ready:
                self.logger.info("✓ API service ready at http://api.localhost")
            else:
                self.logger.warning("⚠ API service not responding yet (may need more time)")

        # Check View container health
        view_url = "http://app.localhost"
        view_ready = False

        if self.verbose:
            self.logger.info("Checking View service...")

        for i in range(10):
            try:
                response = urllib.request.urlopen(view_url, timeout=5)
                view_ready = True
                break
            except Exception:
                if i < 9:
                    time.sleep(3)

        if self.verbose:
            if view_ready:
                self.logger.info("✓ View service ready at http://app.localhost")
            else:
                self.logger.warning("⚠ View service not responding yet (may need more time)")

        # Check Docker containers
        try:
            result = subprocess.run(
                ["docker", "ps", "--filter", "name=nixopus", "--format", "{{.Names}}: {{.Status}}"],
                capture_output=True,
                text=True,
                timeout=5,
            )
            if result.returncode == 0 and self.verbose:
                self.logger.info("\nDocker Containers:")
                for line in result.stdout.strip().split("\n"):
                    if line:
                        self.logger.info(f"  • {line}")
        except Exception:
            pass

        if not self.verbose and (api_ready or view_ready):
            self.logger.info("Services validated")

    def _show_success_message(self):
        """Show success message with service URLs and commands"""
        self.logger.success("Installation Complete!")

        # Get ports from config
        api_port = self._get_config("api_port") or "8080"
        view_port = self._get_config("view_port") or "3000"
        db_port = self._get_config("db_port") or "5432"
        redis_port = self._get_config("redis_port") or "6379"
        supertokens_port = self._get_config("supertokens_api_port") or "3567"
        caddy_admin_port = self._get_config("proxy_port") or "2019"

        if not self.verbose:
            # Minimal output
            self.logger.info("")
            self.logger.info(" Development Environment Ready!")
            self.logger.info("")
            self.logger.info("Access via Caddy Proxy:")
            self.logger.info("  • Frontend:  http://app.localhost")
            self.logger.info("  • Backend:   http://api.localhost")
            self.logger.info("")
            self.logger.info("Direct Container Access:")
            self.logger.info(f"  • Frontend:  http://localhost:{view_port}")
            self.logger.info(f"  • Backend:   http://localhost:{api_port}")
            self.logger.info("")
            self.logger.info("View Logs:")
            self.logger.info("  • Frontend:  docker logs -f nixopus-view-dev")
            self.logger.info("  • Backend:   docker logs -f nixopus-api-dev")
            self.logger.info("")
            self.logger.info("Stop Services:")
            self.logger.info(f"  • All:       cd {self.install_path} && docker-compose -f docker-compose-dev.yml down")
        else:
            # Verbose output
            self.logger.info("")
            self.logger.info("=" * 70)
            self.logger.info(" Development Environment Successfully Installed!")
            self.logger.info("=" * 70)
            self.logger.info(f"Installation Path: {self.install_path}")
            self.logger.info("")
            self.logger.info(" Access via Caddy Reverse Proxy (Recommended):")
            self.logger.info("  • Frontend:    http://app.localhost")
            self.logger.info("  • Backend API: http://api.localhost")
            self.logger.info("  • Register:    http://app.localhost/register")
            self.logger.info("  • API Docs:    http://api.localhost/api/docs")
            self.logger.info("")
            self.logger.info(" Direct Container Access:")
            self.logger.info(f"  • Frontend:    http://localhost:{view_port}")
            self.logger.info(f"  • Backend:     http://localhost:{api_port}")
            self.logger.info(f"  • Database:    localhost:{db_port} (postgres/changeme)")
            self.logger.info(f"  • Redis:       localhost:{redis_port}")
            self.logger.info(f"  • SuperTokens: http://localhost:{supertokens_port}")
            self.logger.info(f"  • Caddy Admin: http://localhost:{caddy_admin_port}")
            self.logger.info("")
            self.logger.info("  View Logs:")
            self.logger.info("  • Frontend:  docker logs -f nixopus-view-dev")
            self.logger.info("  • Backend:   docker logs -f nixopus-api-dev")
            self.logger.info("  • Database:  docker logs -f nixopus-db")
            self.logger.info("  • Caddy:     docker logs -f nixopus-caddy")
            self.logger.info(f"  • All:       cd {self.install_path} && docker-compose -f docker-compose-dev.yml logs -f")
            self.logger.info("")
            self.logger.info("  Development Commands:")
            self.logger.info(f"  • Database:  docker exec -it nixopus-db psql -U postgres")
            self.logger.info("  • Redis:     docker exec -it nixopus-redis redis-cli")
            self.logger.info(f"  • Restart:   cd {self.install_path} && docker-compose -f docker-compose-dev.yml restart")
            self.logger.info(
                f"  • Rebuild:   cd {self.install_path} && docker-compose -f docker-compose-dev.yml up -d --build"
            )
            self.logger.info("")
            self.logger.info(" Hot Reload Enabled:")
            self.logger.info("  • Backend (Go):     Changes rebuild automatically with Air")
            self.logger.info("  • Frontend (Next):  Changes reload instantly with Turbopack")
            self.logger.info("")
            self.logger.info(" Configuration Files:")
            self.logger.info(f"  • Config Dir:  {os.path.join(self.install_path, 'nixopus-dev')}")
            self.logger.info(f"  • Backend:     {os.path.join(self.install_path, 'api', '.env')}")
            self.logger.info(f"  • Frontend:    {os.path.join(self.install_path, 'view', '.env')}")
            self.logger.info(f"  • Caddy:       {os.path.join(self.install_path, 'helpers', 'caddy.json')}")
            self.logger.info("  • SSH Key:     ~/.ssh/id_rsa_nixopus")
            self.logger.info("")
            self.logger.info(" Stop Services:")
            self.logger.info(f"  cd {self.install_path} && docker-compose -f docker-compose-dev.yml down")
            self.logger.info("")
            self.logger.info("=" * 70)
