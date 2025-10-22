import platform
import subprocess

from app.commands.service.base import BaseDockerCommandBuilder, BaseDockerService
from app.utils.logger import Logger
from app.utils.config import Config, DEFAULT_COMPOSE_FILE
from app.utils.config import NIXOPUS_CONFIG_DIR
from .messages import (
    updating_nixopus,
    pulling_latest_images,
    images_pulled_successfully,
    failed_to_pull_images,
    starting_services,
    failed_to_start_services,
    nixopus_updated_successfully,
)

# TODO: Add support for staging or forked repository update is not availlable yet

PACKAGE_JSON_URL = "https://raw.githubusercontent.com/raghavyuva/nixopus/master/package.json"
RELEASES_BASE_URL = "https://github.com/raghavyuva/nixopus/releases/download"

class Update:
    def __init__(self, logger: Logger):
        self.logger = logger
        self.config = Config()

    def run(self):
        compose_file = self.config.get_yaml_value(DEFAULT_COMPOSE_FILE)
        compose_file_path = self.config.get_yaml_value(NIXOPUS_CONFIG_DIR) + "/" + compose_file
        env_file_path = self.config.get_yaml_value(NIXOPUS_CONFIG_DIR) + "/.env"
        self.logger.info(updating_nixopus)
        
        docker_service = BaseDockerService(self.logger, "pull")
        
        self.logger.debug(pulling_latest_images)
        success, output = docker_service.execute_services(compose_file=compose_file_path, env_file=env_file_path)
        
        if not success:
            self.logger.error(failed_to_pull_images.format(error=output))
            return
        
        self.logger.debug(images_pulled_successfully)
        
        docker_service_up = BaseDockerService(self.logger, "up")
        self.logger.debug(starting_services)
        success, output = docker_service_up.execute_services(compose_file=compose_file_path, env_file=env_file_path, detach=True)
        
        if not success:
            self.logger.error(failed_to_start_services.format(error=output))
            return
        
        self.logger.info(nixopus_updated_successfully)
    
    def update_cli(self):
        self.logger.info("Updating CLI tool...")
        try:
            import platform
            import urllib.request
            import json
            import tempfile
            import os
            
            arch = self._detect_arch()
            pkg_type = self._detect_os()
            
            with urllib.request.urlopen(PACKAGE_JSON_URL) as response:
                package_json = json.loads(response.read().decode())
            
            cli_version = package_json.get("cli-version")
            if not cli_version:
                self.logger.error("Could not find cli-version in package.json")
                return
            
            cli_packages = package_json.get("cli-packages", [])
            package_name = self._build_package_name(cli_version, arch, pkg_type)
            
            if package_name not in cli_packages:
                self.logger.error(f"Package {package_name} not found in available packages")
                return
            
            download_url = f"{RELEASES_BASE_URL}/nixopus-{cli_version}/{package_name}"
            
            with tempfile.NamedTemporaryFile(delete=False) as temp_file:
                urllib.request.urlretrieve(download_url, temp_file.name)
                
                if pkg_type == "tar":
                    self._install_tar_package(temp_file.name)
                else:
                    self._install_system_package(temp_file.name, pkg_type)
                
                os.unlink(temp_file.name)
            
            self.logger.info("CLI tool updated successfully")
            
        except Exception as e:
            self.logger.error(f"Failed to update CLI tool: {str(e)}")
    
    def _detect_arch(self):
        arch = platform.machine().lower()
        if arch in ["x86_64", "amd64"]:
            return "amd64"
        elif arch in ["aarch64", "arm64"]:
            return "arm64"
        else:
            raise Exception(f"Unsupported architecture: {arch}")
    
    def _detect_os(self):
        system = platform.system().lower()
        if system == "darwin":
            return "tar"
        elif system == "linux":
            try:
                subprocess.run(["apt", "--version"], capture_output=True, check=True)
                return "deb"
            except:
                try:
                    subprocess.run(["yum", "--version"], capture_output=True, check=True)
                    return "rpm"
                except:
                    try:
                        subprocess.run(["apk", "--version"], capture_output=True, check=True)
                        return "apk"
                    except:
                        return "tar"
        else:
            return "tar"
    
    def _build_package_name(self, version, arch, pkg_type):
        if pkg_type == "deb":
            return f"nixopus_{version}_{arch}.deb"
        elif pkg_type == "rpm":
            arch_name = "x86_64" if arch == "amd64" else "aarch64"
            return f"nixopus-{version}-1.{arch_name}.rpm"
        elif pkg_type == "apk":
            return f"nixopus_{version}_{arch}.apk"
        elif pkg_type == "tar":
            return f"nixopus-{version}.tar"
        else:
            raise Exception(f"Unknown package type: {pkg_type}")
    
    def _install_tar_package(self, package_path):
        subprocess.run(["tar", "-xf", package_path, "-C", "/tmp"], check=True)
        
        try:
            subprocess.run(["cp", "/tmp/usr/local/bin/nixopus", "/usr/local/bin/"], check=True)
            subprocess.run(["chmod", "+x", "/usr/local/bin/nixopus"], check=True)
        except subprocess.CalledProcessError:
            subprocess.run(["sudo", "mkdir", "-p", "/usr/local/bin"], check=True)
            subprocess.run(["sudo", "cp", "/tmp/usr/local/bin/nixopus", "/usr/local/bin/"], check=True)
            subprocess.run(["sudo", "chmod", "+x", "/usr/local/bin/nixopus"], check=True)
    
    def _install_system_package(self, package_path, pkg_type):
        if pkg_type == "deb":
            subprocess.run(["sudo", "dpkg", "-i", package_path], check=True)
            subprocess.run(["sudo", "apt-get", "install", "-f", "-y"], check=True)
        elif pkg_type == "rpm":
            try:
                subprocess.run(["sudo", "dnf", "install", "-y", package_path], check=True)
            except subprocess.CalledProcessError:
                subprocess.run(["sudo", "yum", "install", "-y", package_path], check=True)
        elif pkg_type == "apk":
            subprocess.run(["sudo", "apk", "add", "--allow-untrusted", package_path], check=True)
