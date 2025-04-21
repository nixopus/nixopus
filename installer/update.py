#!/usr/bin/env python3

import os
import sys
import subprocess
from pathlib import Path
import shutil
import json
import platform
import re

class Updater:
    def __init__(self):
        self.project_root = Path(__file__).parent.parent
        self.env_file = self.project_root / ".env"
        self.required_docker_version = "20.10.0"
        self.required_compose_version = "2.0.0"
    
    def ask_for_sudo(self):
        if os.geteuid() != 0:
            print("Please run the script with sudo privileges")
            sys.exit(1)
    
    def check_docker_version(self):
        try:
            result = subprocess.run(["docker", "--version"], check=True, capture_output=True, text=True)
            version_string = result.stdout.strip()
            if not self._version_check(version_string, self.required_docker_version):
                print(f"Error: Docker version {self.required_docker_version} or higher is required")
                print(f"Current version: {version_string}")
                sys.exit(1)
        except subprocess.CalledProcessError as e:
            print(f"Error: Docker is not installed or not working properly")
            print(e.stderr.decode())
            sys.exit(1)
            
    def check_docker_compose_version(self):
        try:
            result = subprocess.run(["docker-compose", "--version"], check=True, capture_output=True, text=True)
            version_string = result.stdout.strip()
            if not self._version_check(version_string, self.required_compose_version):
                print(f"Error: Docker Compose version {self.required_compose_version} or higher is required")
                print(f"Current version: {version_string}")
                sys.exit(1)
        except subprocess.CalledProcessError as e:
            print(f"Error: Docker Compose is not installed or not working properly")
            print(e.stderr.decode())
            sys.exit(1)
            
    def check_curl_installed(self):
        if not shutil.which("curl"):
            print("Error: Curl is not installed")
            sys.exit(1)

    def _version_check(self, version_string, required_version):
        version = re.search(r'\d+\.\d+\.\d+', version_string)
        if not version:
            return False
        return tuple(map(int, version.group().split('.'))) >= tuple(map(int, required_version.split('.')))
    
    def check_system_requirements(self):
        print("Checking system requirements...")
        
        system = platform.system()
        if system not in ["Linux"]:
            print(f"Error: Unsupported operating system: {system}")
            sys.exit(1)

        self.check_docker_version()
        self.check_docker_compose_version()
        self.check_curl_installed()

        print("System requirements check passed!")
    
    def update_services(self):
        print("\nUpdating services...")
        try:
            try:
                subprocess.run(["docker", "info"], check=True, capture_output=True)
            except subprocess.CalledProcessError:
                print("Error: Docker daemon is not running. Please start the Docker service and try again.")
                sys.exit(1)

            os.environ["DOCKER_HOST"] = "tcp://localhost:2376"
            os.environ["DOCKER_TLS_VERIFY"] = "1"
            os.environ["DOCKER_CERT_PATH"] = "/etc/nixopus/docker-certs"
            
            services = {
                "nixopus-api-container": "nixopus-api:latest",
                "nixopus-db-container": "nixopus-db:latest",
                "nixopus-view-container": "nixopus-view:latest",
                "nixopus-caddy-container": "nixopus-caddy:latest"
            }
            
            for container_name, image_name in services.items():
                print(f"\nUpdating {container_name}...")
                
                inspect_result = subprocess.run(
                    ["docker", "inspect", container_name],
                    capture_output=True,
                    text=True
                )
                
                if inspect_result.returncode == 0:
                    container_config = json.loads(inspect_result.stdout)[0]
                    env_vars = container_config.get("Config", {}).get("Env", [])
                    volumes = container_config.get("HostConfig", {}).get("Binds", [])
                    ports = container_config.get("HostConfig", {}).get("PortBindings", {})
                    networks = container_config.get("NetworkSettings", {}).get("Networks", {})
                    
                    subprocess.run(["docker", "stop", container_name], capture_output=True)
                    subprocess.run(["docker", "rm", container_name], capture_output=True)
                    
                    pull_result = subprocess.run(
                        ["docker", "pull", image_name],
                        capture_output=True,
                        text=True
                    )
                    
                    if pull_result.returncode != 0:
                        print(f"Error pulling image {image_name}:")
                        print(pull_result.stderr)
                        sys.exit(1)
                    
                    run_cmd = ["docker", "run", "-d", "--name", container_name]
                    
                    for env in env_vars:
                        run_cmd.extend(["-e", env])
                    
                    for volume in volumes:
                        run_cmd.extend(["-v", volume])
                    
                    for container_port, host_ports in ports.items():
                        for host_port in host_ports:
                            run_cmd.extend(["-p", f"{host_port['HostPort']}:{container_port.split('/')[0]}"])
                    
                    for network_name in networks.keys():
                        run_cmd.extend(["--network", network_name])
                    
                    run_cmd.append(image_name)
                    
                    result = subprocess.run(
                        run_cmd,
                        capture_output=True,
                        text=True
                    )
                    
                    if result.returncode != 0:
                        print(f"Error starting container {container_name}:")
                        print(result.stderr)
                        sys.exit(1)
                else:
                    pull_result = subprocess.run(
                        ["docker", "pull", image_name],
                        capture_output=True,
                        text=True
                    )
                    
                    if pull_result.returncode != 0:
                        print(f"Error pulling image {image_name}:")
                        print(pull_result.stderr)
                        sys.exit(1)
                    
                    result = subprocess.run(
                        ["docker", "run", "-d", "--name", container_name, image_name],
                        capture_output=True,
                        text=True
                    )
                    
                    if result.returncode != 0:
                        print(f"Error starting container {container_name}:")
                        print(result.stderr)
                        sys.exit(1)
                
            print("Services updated successfully!")
        except Exception as e:
            print(f"Error updating services: {str(e)}")
            sys.exit(1)
    
    def verify_update(self):
        print("\nVerifying update...")
        try:
            result = subprocess.run(["docker", "ps", "--format", "{{.Names}} {{.Status}}"], capture_output=True, text=True)
            if result.returncode != 0:
                print("Error verifying update:")
                print(result.stderr)
                sys.exit(1)
                
            running_containers = result.stdout.splitlines()
            required_containers = {
                "nixopus-api-container": "API service",
                "nixopus-db-container": "Database service",
                "nixopus-view-container": "View service",
                "nixopus-caddy-container": "Caddy service"
            }
            
            missing_containers = []
            for container, service_name in required_containers.items():
                container_running = any(
                    line.startswith(container) and "Up" in line
                    for line in running_containers
                )
                if not container_running:
                    missing_containers.append(service_name)

            if missing_containers:
                print("Error: The following services are not running:")
                for service in missing_containers:
                    print(f"  - {service}")
                sys.exit(1)

            print("✓ All services are running successfully!")
        except Exception as e:
            print(f"Error verifying update: {str(e)}")
            sys.exit(1)
    
    def setup_caddy(self):
        print("\nSetting up Proxy...")
        try:
            with open('api/helpers/caddy.json', 'r') as f:
                config = json.dumps(json.load(f))
            
            result = subprocess.run(
                ['curl', '-X', 'POST', 'http://localhost:2019/load',
                 '-H', 'Content-Type: application/json',
                 '-d', config],
                capture_output=True,
                text=True
            )
            
            if result.returncode == 0:
                print("✓ Caddy configuration loaded successfully")
            else:
                print("✗ Failed to load Caddy configuration:")
                print(result.stderr)
        except Exception as e:
            print(f"✗ Error setting up Caddy: {str(e)}")

def main():
    updater = Updater()
    
    print("\033[36m  _   _ _ _                           \033[0m")
    print("\033[36m | \ | (_)                          \033[0m")
    print("\033[36m |  \| |___  _____  _ __  _   _ ___ \033[0m")
    print("\033[36m | . \` | \ \/ / _ \| '_ \| | | / __|\033[0m")
    print("\033[36m | |\  | |>  < (_) | |_) | |_| \__ \033[0m")
    print("\033[36m |_| \_|_/_/\_\___/| .__/ \__,_|___/\033[0m")
    print("\033[36m                   | |              \033[0m")
    print("\033[36m                   |_|              \033[0m")
    print("\n")
    print("\033[1mWelcome to Nixopus Update Wizard\033[0m")
    print("This wizard will guide you through the update process of Nixopus services.")
    print("Please follow the prompts carefully to complete the update.\n")
    
    updater.ask_for_sudo()
    updater.check_system_requirements()
    updater.update_services()
    updater.verify_update()
    updater.setup_caddy()
    
    print("\n\033[1mUpdate Complete!\033[0m")
    print("\n\033[1mYour Nixopus services have been successfully updated to the latest version.\033[0m")
    print("\n\033[1mThank you for using Nixopus!\033[0m")

if __name__ == "__main__":
    main() 