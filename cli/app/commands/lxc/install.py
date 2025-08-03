import typer
import os
import subprocess
import time
from typing import Optional
from rich.progress import Progress, SpinnerColumn, TextColumn, BarColumn, TaskProgressColumn

from app.utils.protocols import LoggerProtocol
from app.utils.config import Config
from app.utils.timeout import TimeoutWrapper
from app.utils.logger import Logger
from .config import (
    DEFAULT_CONTAINER_NAME, DEFAULT_LXC_IMAGE, DEFAULT_TIMEOUT,
    LXCInstallConfig, get_proxy_devices, get_command_template, get_packages_string,
    SECURITY_SETTINGS, REQUIRED_PACKAGES, SERVICES_TO_START, SERVICES_TO_ENABLE,
    CONTAINER_READY_WAIT_TIME, CONTAINER_RESTART_WAIT_TIME, DEFAULT_INSTALL_SCRIPT_URL,
    IP_DETECTION_CONFIG
)
from .messages import (
    lxc_install_completed, installation_failed, installation_timed_out,
    lxc_install_step_1, lxc_install_step_2, lxc_install_step_3,
    container_created_successfully, nixopus_installed_in_container,
    lxc_not_installed, container_exists_use_force,
    creating_container, removing_existing_container, waiting_for_container_ready,
    configuring_security_settings, restarting_container_for_security,
    failed_to_create_container, failed_to_setup_networking,
    installing_dependencies_in_container, updating_package_lists, installing_packages,
    starting_docker_service, enabling_docker_service, starting_ssh_service,
    enabling_ssh_service, failed_to_install_dependencies, installing_nixopus_in_container,
    downloading_install_script, failed_to_install_nixopus, getting_container_ip,
    container_ip_found, container_ip_not_found, debug_container_check,
    debug_security_config, debug_proxy_device,
    installing_nixopus_in_lxc, lxc_installation_completed, creating_lxc_container,
    setting_up_container_networking, installing_nixopus_in_container_step,
    failed_to_create_lxc_container, failed_to_setup_container_networking,
    failed_to_install_dependencies_in_container, failed_to_install_nixopus_in_container,
    container_name_info, container_ip_info, access_nixopus_through_urls,
    view_interface_access, api_access, direct_access_url,
    container_access_command, container_stop_command, container_start_command,
    error_context_during, installation_failed_with_context, installation_failed_simple,
    domain_validation_error, service_docker, service_ssh
)


class LXCManager:
    def __init__(self, logger: Logger):
        self.logger = logger

    def container_exists(self, name: str) -> bool:
        try:
            self.logger.debug(debug_container_check.format(name=name))
            command = get_command_template("lxc_info", name=name)
            result = subprocess.run(
                command,
                capture_output=True,
                text=True,
                check=False
            )
            return result.returncode == 0
        except FileNotFoundError:
            raise Exception(lxc_not_installed)

    def create_container(self, name: str, image: str = DEFAULT_LXC_IMAGE, force: bool = False) -> bool:
        try:
            if self.container_exists(name) and not force:
                self.logger.info(container_exists_use_force.format(name=name))
                return True
            
            if self.container_exists(name) and force:
                self.logger.info(removing_existing_container.format(name=name))
                command = get_command_template("lxc_delete", name=name)
                subprocess.run(command, check=True)
            
            self.logger.info(creating_container.format(name=name, image=image))
            command = get_command_template("lxc_launch", image=image, name=name)
            subprocess.run(command, check=True)
            
            self.logger.info(waiting_for_container_ready)
            time.sleep(CONTAINER_READY_WAIT_TIME)
            
            self.logger.info(configuring_security_settings)
            for key, value in SECURITY_SETTINGS.items():
                command = get_command_template("lxc_config_set", name=name, key=key, value=value)
                subprocess.run(command, check=True)
                self.logger.debug(debug_security_config.format(key=key, value=value, name=name))
            
            self.logger.info(restarting_container_for_security)
            command = get_command_template("lxc_restart", name=name)
            subprocess.run(command, check=True)
            time.sleep(CONTAINER_RESTART_WAIT_TIME)
            
            return True
            
        except subprocess.CalledProcessError as e:
            self.logger.error(failed_to_create_container.format(error=e))
            return False

    def setup_container_networking(self, name: str) -> bool:
        try:
            proxy_devices = get_proxy_devices()
            
            for device in proxy_devices:
                command = get_command_template(
                    "lxc_proxy_add",
                    name=name,
                    device_name=device.device_name,
                    host_port=device.host_port,
                    container_port=device.container_port
                )
                subprocess.run(command, check=True)
                self.logger.debug(debug_proxy_device.format(
                    device=device.device_name,
                    listen=f"tcp:0.0.0.0:{device.host_port}",
                    connect=f"tcp:127.0.0.1:{device.container_port}"
                ))
            
            return True
            
        except subprocess.CalledProcessError as e:
            self.logger.error(failed_to_setup_networking.format(error=e))
            return False

    def install_dependencies_in_container(self, name: str) -> bool:
        try:
            self.logger.info(installing_dependencies_in_container)
            
            self.logger.info(updating_package_lists)
            command = get_command_template("lxc_exec", name=name) + get_command_template("apt_update")
            subprocess.run(command, check=True)
            
            packages_str = get_packages_string()
            self.logger.info(installing_packages.format(packages=packages_str))
            command = get_command_template("lxc_exec", name=name) + get_command_template("apt_install") + REQUIRED_PACKAGES
            subprocess.run(command, check=True)
            
            for service_name, description in SERVICES_TO_START:
                if service_name == service_docker:
                    self.logger.info(starting_docker_service)
                elif service_name == service_ssh:
                    self.logger.info(starting_ssh_service)
                
                command = get_command_template("lxc_exec", name=name) + get_command_template("systemctl_start", service=service_name)
                subprocess.run(command, check=True)
            
            for service_name, description in SERVICES_TO_ENABLE:
                if service_name == service_docker:
                    self.logger.info(enabling_docker_service)
                elif service_name == service_ssh:
                    self.logger.info(enabling_ssh_service)
                    
                command = get_command_template("lxc_exec", name=name) + get_command_template("systemctl_enable", service=service_name)
                subprocess.run(command, check=True)
            
            return True
            
        except subprocess.CalledProcessError as e:
            self.logger.error(failed_to_install_dependencies.format(error=e))
            return False

    def install_nixopus_in_container(self, name: str, install_script_url: str = None) -> bool:
        try:
            self.logger.info(installing_nixopus_in_container)

            if not install_script_url:
                install_script_url = DEFAULT_INSTALL_SCRIPT_URL
            
            self.logger.info(downloading_install_script.format(url=install_script_url))
            
            command = get_command_template("lxc_exec", name=name) + get_command_template("curl_install", url=install_script_url)
            subprocess.run(command, check=True)
            
            return True
            
        except subprocess.CalledProcessError as e:
            self.logger.error(failed_to_install_nixopus.format(error=e))
            return False

    def get_container_ip(self, name: str) -> Optional[str]:
        try:
            self.logger.info(getting_container_ip)
            command = get_command_template("lxc_list_csv", name=name)
            result = subprocess.run(command, capture_output=True, text=True, check=True)
            
            output = result.stdout.strip()
            if output:
                parts = output.split()
                for part in parts:
                    if "." in part and part.count(".") == IP_DETECTION_CONFIG["ip_validation_parts"] - 1:
                        ip = part.strip("()")
                        try:
                            ip_parts = ip.split(".")
                            if (len(ip_parts) == IP_DETECTION_CONFIG["ip_validation_parts"] and
                                all(IP_DETECTION_CONFIG["ip_part_min"] <= int(p) <= IP_DETECTION_CONFIG["ip_part_max"] for p in ip_parts)):
                                self.logger.info(container_ip_found.format(ip=ip))
                                return ip
                        except ValueError:
                            continue
                            
            self.logger.warning(container_ip_not_found)
            return None
            
        except subprocess.CalledProcessError:
            self.logger.warning(container_ip_not_found)
            return None


class LXCInstall:
    def __init__(self, logger: LoggerProtocol, config: LXCInstallConfig):
        self.logger = logger
        self.config = config
        self.lxc_manager = LXCManager(logger)
        self._config = Config()
        self.progress = None
        self.main_task = None

    def run(self):
        steps = [
            (creating_lxc_container, self._create_container),
            (setting_up_container_networking, self._setup_networking),
            (installing_nixopus_in_container_step, self._install_nixopus_in_container),
        ]
        
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
                self.main_task = progress.add_task(installing_nixopus_in_lxc, total=len(steps))
                
                for i, (step_name, step_func) in enumerate(steps):
                    progress.update(self.main_task, description=f"{installing_nixopus_in_lxc} - {step_name} ({i+1}/{len(steps)})")
                    try:
                        step_func()
                        progress.advance(self.main_task, 1)
                    except Exception as e:
                        progress.update(self.main_task, description=f"Failed at {step_name}")
                        raise
                
                progress.update(self.main_task, completed=True, description=lxc_installation_completed)
            
            self._show_success_message()
            
        except Exception as e:
            self._handle_installation_error(e)
            self.logger.error(f"{installation_failed}: {str(e)}")
            raise typer.Exit(1)

    def _create_container(self):
        self.logger.info(lxc_install_step_1)
        if not self.config.dry_run:
            success = self.lxc_manager.create_container(
                self.config.name, 
                self.config.image, 
                self.config.force
            )
            if not success:
                raise Exception(failed_to_create_lxc_container)
        self.logger.info(container_created_successfully.format(name=self.config.name))

    def _setup_networking(self):
        if self.config.setup_proxy:
            self.logger.info(lxc_install_step_2)
            if not self.config.dry_run:
                success = self.lxc_manager.setup_container_networking(self.config.name)
                if not success:
                    raise Exception(failed_to_setup_container_networking)

    def _install_nixopus_in_container(self):
        self.logger.info(lxc_install_step_3)
        if not self.config.dry_run:
            success = self.lxc_manager.install_dependencies_in_container(self.config.name)
            if not success:
                raise Exception(failed_to_install_dependencies_in_container)
            
            self._run_nixopus_install_in_container()
        
        self.logger.info(nixopus_installed_in_container.format(name=self.config.name))

    def _run_nixopus_install_in_container(self):
        if self.config.install_script_url:
            success = self.lxc_manager.install_nixopus_in_container(
                self.config.name, 
                self.config.install_script_url
            )
        else:
            success = self.lxc_manager.install_nixopus_in_container(self.config.name)
        
        if not success:
            raise Exception(failed_to_install_nixopus_in_container)

    def _show_success_message(self):
        container_ip = self.lxc_manager.get_container_ip(self.config.name)
        
        self.logger.success(lxc_install_completed)
        self.logger.info(container_name_info.format(name=self.config.name))
        
        if container_ip:
            from .config import VIEW_PORT, API_PORT
            self.logger.info(container_ip_info.format(ip=container_ip))
            
            if self.config.setup_proxy:
                self.logger.info(access_nixopus_through_urls)
                self.logger.info(view_interface_access.format(ip=container_ip, port=VIEW_PORT))
                self.logger.info(api_access.format(ip=container_ip, port=API_PORT))
            else:
                self.logger.info(direct_access_url.format(ip=container_ip, port=VIEW_PORT))
        
        self.logger.info(container_access_command.format(name=self.config.name))
        self.logger.info(container_stop_command.format(name=self.config.name))
        self.logger.info(container_start_command.format(name=self.config.name))

    def _handle_installation_error(self, error, context=""):
        context_msg = error_context_during.format(context=context) if context else ""
        if self.config.verbose:
            self.logger.error(installation_failed_with_context.format(message=installation_failed, context=context_msg, error=str(error)))
        else:
            self.logger.error(installation_failed_simple.format(message=installation_failed, context=context_msg))


def lxc_install_command(
    name: str = typer.Argument(DEFAULT_CONTAINER_NAME, help="Container name"),
    image: str = typer.Option(DEFAULT_LXC_IMAGE, "--image", "-i", help="LXC image to use"),
    force: bool = typer.Option(False, "--force", "-f", help="Force recreate if container exists"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    timeout: int = typer.Option(DEFAULT_TIMEOUT, "--timeout", "-t", help="Timeout for installation (seconds)"),
    api_domain: Optional[str] = typer.Option(None, "--api-domain", help="API domain for nixopus"),
    view_domain: Optional[str] = typer.Option(None, "--view-domain", help="View domain for nixopus"),
    install_script_url: Optional[str] = typer.Option(None, "--install-url", help="Custom nixopus install script URL"),
    setup_proxy: bool = typer.Option(True, "--setup-proxy/--no-proxy", help="Setup proxy devices"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run mode"),
):
    
    if (api_domain is None) != (view_domain is None):
        raise typer.BadParameter(domain_validation_error)
    
    logger = Logger(verbose=verbose)
    
    config = LXCInstallConfig(
        name=name,
        image=image,
        force=force,
        verbose=verbose,
        timeout=timeout,
        api_domain=api_domain,
        view_domain=view_domain,
        install_script_url=install_script_url,
        setup_proxy=setup_proxy,
        dry_run=dry_run
    )
    
    installer = LXCInstall(logger, config)
    
    try:
        with TimeoutWrapper(timeout):
            installer.run()
    except TimeoutError:
        logger.error(installation_timed_out.format(timeout=timeout))
        raise typer.Exit(1)
    except Exception as e:
        logger.error(installation_failed.format(error=str(e)))
        raise typer.Exit(1)
