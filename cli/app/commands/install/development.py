import os
import platform
import subprocess
import time
import urllib.request

import typer
from rich.progress import BarColumn, Progress, SpinnerColumn, TaskProgressColumn, TextColumn

from app.commands.clone.clone import Clone, CloneConfig
from app.commands.conf.base import BaseEnvironmentManager
from app.commands.preflight.run import PreflightRunner
from app.commands.service.base import BaseDockerService
from app.commands.service.up import Up, UpConfig
from app.utils.config import (
    API_ENV_FILE,
    API_PORT,
    Config,
    DEFAULT_BRANCH,
    DEFAULT_COMPOSE_FILE,
    DEFAULT_PATH,
    DEFAULT_REPO,
    NIXOPUS_CONFIG_DIR,
    PORTS,
    SSH_FILE_PATH,
    SSH_KEY_SIZE,
    SSH_KEY_TYPE,
    VIEW_ENV_FILE,
    VIEW_PORT,
)
from app.utils.lib import FileManager, HostInformation
from app.utils.protocols import LoggerProtocol
from app.utils.timeout import TimeoutWrapper

from .base import BaseInstall
from .deps import get_deps_from_config, get_installed_deps, install_dep
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
from .ssh import SSH, SSHConfig


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
        )

        self.install_path = os.path.abspath(os.path.expanduser(install_path)) if install_path else os.getcwd()
        
        # Check platform and WSL requirement for Windows
        self._check_platform_support()
        
        # Load config from config.dev.yaml
        self._config = Config(default_env="DEVELOPMENT")
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
        
        # Native Windows - show warning
        self.logger.warning("Running on native Windows")
        self.logger.warning("")
        self.logger.warning("Nixopus development environment requires WSL2 for full feature support.")
        self.logger.warning("")
        self.logger.warning("Developmet setup not available on native Windows:")
        self.logger.warning("  - SSH file manager (container to host access)")
        self.logger.warning("  - Docker container management")
        self.logger.warning("")
        self.logger.warning("To install WSL2:")
        self.logger.warning("  1. Open PowerShell as Administrator")
        self.logger.warning("  2. Run: wsl --install")
        self.logger.warning("  3. Restart your computer")
        self.logger.warning("  4. Run this command again in WSL2 terminal")
        self.logger.warning("")
        self.logger.error("Development installation is only supported on macOS, Linux, or WSL2")
        self.logger.info("Visit: https://docs.microsoft.com/en-us/windows/wsl/install")
        raise typer.Exit(1)
    
    def _load_dev_defaults(self):
        """Load defaults from config.dev.yaml"""
        config_dir = self._config.get_yaml_value(NIXOPUS_CONFIG_DIR)
        source_path = self._config.get_yaml_value(DEFAULT_PATH)
        
        return {
            "ssh_key_type": self._config.get_yaml_value(SSH_KEY_TYPE),
            "ssh_key_size": self._config.get_yaml_value(SSH_KEY_SIZE),
            "ssh_passphrase": None,
            "service_name": "all",
            "service_detach": True,
            "required_ports": [int(port) for port in self._config.get_yaml_value(PORTS)],
            "repo_url": self._config.get_yaml_value(DEFAULT_REPO),
            "branch_name": self._config.get_yaml_value(DEFAULT_BRANCH),
            "source_path": source_path,
            "config_dir": config_dir,
            "api_env_file_path": self._config.get_yaml_value(API_ENV_FILE),
            "view_env_file_path": self._config.get_yaml_value(VIEW_ENV_FILE),
            "compose_file": self._config.get_yaml_value(DEFAULT_COMPOSE_FILE),
            "full_source_path": self.install_path,
            "ssh_key_path": os.path.expanduser("~/.ssh/id_rsa_nixopus"),
            "compose_file_path": os.path.join(self.install_path, "docker-compose-dev.yml"),
            "view_port": self._config.get_yaml_value(VIEW_PORT),
            "api_port": self._config.get_yaml_value(API_PORT),
            "nixopus_config_dir": os.path.join(self.install_path, "nixopus-dev"),
        }

    def _get_config(self, key: str):
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
            return os.path.join(self.install_path, "view", ".env.local")
        if key == "ssh_key_path":
            return os.path.expanduser("~/.ssh/id_rsa_nixopus")

        # Use parent's _get_config with dev defaults
        return super()._get_config(key, self._user_config, self._defaults)

    def run(self):
        """Execute development installation workflow"""
        steps = [
            ("Preflight checks", self._run_preflight_checks),
            ("Checking dependencies", self._check_and_install_dependencies),
            ("Cloning repository", self._setup_clone_and_config),
            ("Creating environment files", self._create_env_files),
            ("Generating SSH keys", self._setup_ssh),
            ("Starting backend services", self._start_backend),
            ("Starting frontend", self._start_frontend),
            ("Validating services", self._validate_services),
        ]

        if self.force:

            def cleanup():
                compose_file = self._get_config("compose_file_path")
                if os.path.exists(compose_file):
                    try:
                        BaseDockerService.cleanup_docker_resources(
                            logger=self.logger,
                            compose_file=compose_file,
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
        preflight_runner = PreflightRunner(logger=self.logger, verbose=self.verbose)
        preflight_runner.check_ports_from_config(
            config_key="required_ports", user_config=self._user_config, defaults=self._defaults
        )

    def _check_and_install_dependencies(self):
        """Check dependencies and install only if missing"""
        deps = get_deps_from_config()
        os_name = HostInformation.get_os_name()
        package_manager = HostInformation.get_package_manager()

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
        clone_config = CloneConfig(
            repo=self._get_config("repo_url"),
            branch=self._get_config("branch_name"),
            path=self._get_config("full_source_path"),
            force=self.force,
            verbose=self.verbose,
            output="text",
            dry_run=self.dry_run,
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
        """Create environment files for backend and frontend"""
        api_env_file = self._get_config("api_env_file_path")
        view_env_file = self._get_config("view_env_file_path")

        FileManager.create_directory(FileManager.get_directory_path(api_env_file), logger=self.logger)
        FileManager.create_directory(FileManager.get_directory_path(view_env_file), logger=self.logger)

        services = [
            ("api", "services.api.env", api_env_file),
            ("view", "services.view.env", view_env_file),
        ]

        env_manager = BaseEnvironmentManager(self.logger)

        for service_name, service_key, env_file in services:
            env_values = self._config.get_service_env_values(service_key)
            updated_env_values = self._update_environment_variables(env_values)
            success, error = env_manager.write_env_file(env_file, updated_env_values)
            if not success:
                raise Exception(f"{env_file_creation_failed} {service_name}: {error}")
            file_perm_success, file_perm_error = FileManager.set_permissions(env_file, 0o644)
            if not file_perm_success:
                raise Exception(f"{env_file_permissions_failed} {service_name}: {file_perm_error}")
            self.logger.debug(created_env_file.format(service_name=service_name, env_file=env_file))

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

        ssh_operation = SSH(logger=self.logger)
        try:
            with TimeoutWrapper(self.timeout):
                result = ssh_operation.generate(config)
        except TimeoutError:
            raise Exception(f"{ssh_setup_failed}: {operation_timed_out}")
        if not result.success:
            raise Exception(ssh_setup_failed)

        if not self.dry_run and self.verbose:
            self.logger.info("SSH key configured for local development with host access")

    def _start_backend(self):
        """Start backend services using Docker Compose"""
        config = UpConfig(
            name=self._get_config("service_name"),
            detach=self._get_config("service_detach"),
            env_file=None,
            verbose=self.verbose,
            output="text",
            dry_run=self.dry_run,
            compose_file=self._get_config("compose_file_path"),
        )

        up_service = Up(logger=self.logger)
        try:
            with TimeoutWrapper(self.timeout):
                result = up_service.up(config)
        except TimeoutError:
            raise Exception(f"{services_start_failed}: {operation_timed_out}")
        if not result.success:
            raise Exception(services_start_failed)

    def _start_frontend(self):
        """Install frontend dependencies and start dev server"""
        view_dir = os.path.join(self.install_path, "view")

        if not os.path.exists(view_dir):
            raise Exception(f"Frontend directory not found: {view_dir}")

        if self.dry_run:
            self.logger.info("[DRY RUN] Would install frontend dependencies and start server")
            return

        # Install dependencies
        if not self.verbose:
            self.logger.info("Installing frontend dependencies...")

        try:
            yarn_install = subprocess.run(
                ["yarn", "install", "--non-interactive"],
                cwd=view_dir,
                capture_output=not self.verbose,
                text=True,
                timeout=self.timeout,  # Use configurable timeout
            )

            if yarn_install.returncode != 0:
                raise Exception("Frontend dependency installation failed")
        except subprocess.TimeoutExpired:
            raise Exception("Frontend dependency installation timed out")
        except FileNotFoundError:
            raise Exception("yarn not found. Please install yarn first.")

        # Create logs directory
        log_dir = os.path.join(self.install_path, "logs")
        FileManager.create_directory(log_dir, logger=self.logger)

        # Start dev server in background
        log_file_path = os.path.join(log_dir, "frontend.log")
        pid_file_path = os.path.join(log_dir, "frontend.pid")

        if not self.verbose:
            self.logger.info("Starting frontend server...")

        try:
            log_file = open(log_file_path, "w")
            process = subprocess.Popen(
                ["yarn", "dev"], cwd=view_dir, stdout=log_file, stderr=subprocess.STDOUT, start_new_session=True
            )

            # Save PID
            with open(pid_file_path, "w") as f:
                f.write(str(process.pid))

            # Wait for startup
            time.sleep(5)

            # Check if process is still running
            if process.poll() is not None:
                raise Exception("Frontend server failed to start. Check logs/frontend.log")

            if not self.verbose:
                self.logger.info("Frontend server started")

        except FileNotFoundError:
            raise Exception("yarn not found. Please install yarn first.")
        except Exception as e:
            raise Exception(f"Failed to start frontend: {str(e)}")

    def _validate_services(self):
        """Validate backend and frontend are accessible"""
        if self.dry_run:
            self.logger.info("[DRY RUN] Would validate services")
            return

        # Get ports from config
        api_port = self._get_config("api_port") or "8080"
        view_port = self._get_config("view_port") or "3000"

        if not self.verbose:
            self.logger.info("Validating services...")

        # Check backend with retries
        backend_url = f"http://localhost:{api_port}/api/v1/health"
        backend_ready = False

        for i in range(5):
            try:
                response = urllib.request.urlopen(backend_url, timeout=5)
                if response.status == 200:
                    backend_ready = True
                    break
            except Exception:
                if i < 4:
                    time.sleep(2)

        if self.verbose:
            if backend_ready:
                self.logger.info(f"Backend ready at http://localhost:{api_port}")
            else:
                self.logger.warning("Backend not responding yet (may need more time)")

        # Check frontend
        frontend_url = f"http://localhost:{view_port}"
        frontend_ready = False

        for i in range(3):
            try:
                response = urllib.request.urlopen(frontend_url, timeout=5)
                frontend_ready = True
                break
            except Exception:
                if i < 2:
                    time.sleep(2)

        if self.verbose:
            if frontend_ready:
                self.logger.info(f"Frontend ready at http://localhost:{view_port}")
            else:
                self.logger.warning("Frontend not responding yet (may need more time)")

        if not self.verbose and (backend_ready or frontend_ready):
            self.logger.info("Services validated")

    def _show_success_message(self):
        """Show success message with service URLs and commands"""
        self.logger.success("Installation Complete!")

        # Get ports from config
        api_port = self._get_config("api_port") or "8080"
        view_port = self._get_config("view_port") or "3000"

        if not self.verbose:
            # Minimal output
            self.logger.info("")
            self.logger.info("Services Running:")
            self.logger.info(f"  • Frontend:  http://localhost:{view_port}")
            self.logger.info(f"  • Backend:   http://localhost:{api_port}")
            self.logger.info(f"  • Register:  http://localhost:{view_port}/register")
            self.logger.info("")
            self.logger.info("View Logs:")
            self.logger.info(f"  • Frontend:  tail -f {os.path.join(self.install_path, 'logs', 'frontend.log')}")
            self.logger.info("  • Backend:   docker logs -f nixopus-api-dev")
            self.logger.info("")
            self.logger.info("Stop Services:")
            self.logger.info(f"  • Frontend:  kill $(cat {os.path.join(self.install_path, 'logs', 'frontend.pid')})")
            self.logger.info(f"  • Backend:   cd {self.install_path} && docker-compose -f docker-compose-dev.yml down")
        else:
            # Verbose output
            self.logger.info("")
            self.logger.info(f"Development environment installed in: {self.install_path}")
            self.logger.info("")
            self.logger.info("Services Running:")
            self.logger.info(f"  • Frontend:  http://localhost:{view_port}")
            self.logger.info(f"  • Backend:   http://localhost:{api_port}")
            self.logger.info(f"  • Database:  localhost:5432 (postgres/changeme)")
            self.logger.info("  • Redis:     localhost:6379")
            self.logger.info(f"  • SuperTokens: http://localhost:3567")
            self.logger.info("")
            self.logger.info("Access Application:")
            self.logger.info(f"  • Main:      http://localhost:{view_port}")
            self.logger.info(f"  • Register:  http://localhost:{view_port}/register")
            self.logger.info(f"  • API Docs:  http://localhost:{api_port}/api/docs")
            self.logger.info("")
            self.logger.info("View Logs:")
            self.logger.info(f"  • Frontend:  tail -f {os.path.join(self.install_path, 'logs', 'frontend.log')}")
            self.logger.info("  • Backend:   docker logs -f nixopus-api-dev")
            self.logger.info("  • Database:  docker logs -f nixopus-db")
            self.logger.info(f"  • All:       cd {self.install_path} && docker-compose -f docker-compose-dev.yml logs -f")
            self.logger.info("")
            self.logger.info("Development Commands:")
            self.logger.info(f"  • Database:  docker exec -it nixopus-db psql -U postgres")
            self.logger.info("  • Redis:     docker exec -it nixopus-redis redis-cli")
            self.logger.info(f"  • Restart:   cd {self.install_path} && docker-compose -f docker-compose-dev.yml restart api")
            self.logger.info("")
            self.logger.info("Configuration:")
            self.logger.info(f"  • Config:    {os.path.join(self.install_path, 'nixopus-dev')}")
            self.logger.info(f"  • Backend:   {os.path.join(self.install_path, 'api', '.env')}")
            self.logger.info(f"  • Frontend:  {os.path.join(self.install_path, 'view', '.env.local')}")
            self.logger.info(f"  • Logs:      {os.path.join(self.install_path, 'logs')}")
            self.logger.info("  • SSH Key:   ~/.ssh/id_rsa_nixopus")
