#!/usr/bin/env python3

import os
import sys
import subprocess
from pathlib import Path
import shutil
import json

class Updater:
    def __init__(self):
        self.project_root = Path(__file__).parent.parent
        self.env_file = self.project_root / ".env"
        self.repo = None
    
    def ask_for_sudo(self):
        if os.geteuid() != 0:
            print("Please run the script with sudo privileges")
            sys.exit(1)
    
    def check_docker_installed(self):
        if not shutil.which("docker"):
            print("Error: Docker is not installed")
            sys.exit(1)
    
    def check_docker_compose_installed(self):
        if not shutil.which("docker-compose") and not shutil.which("docker"):
            print("Error: Docker Compose is not installed")
            sys.exit(1)
    
    def check_system_requirements(self):
        print("Checking system requirements...")
        self.check_docker_installed()
        self.check_docker_compose_installed()
        print("System requirements check passed!")
    
    def update_services(self):
        print("\nUpdating services...")
        try:
            os.environ["DOCKER_HOST"] = "tcp://localhost:2376"
            os.environ["DOCKER_TLS_VERIFY"] = "1"
            os.environ["DOCKER_CERT_PATH"] = "/etc/nixopus/docker-certs"
            
            compose_cmd = ["docker", "compose"] if shutil.which("docker") else ["docker-compose"]
            
            pull_result = subprocess.run(
                compose_cmd + ["pull"],
                capture_output=True,
                text=True,
                cwd=self.project_root
            )
            
            if pull_result.returncode != 0:
                print("Error pulling images:")
                print(pull_result.stderr)
                sys.exit(1)
            
            result = subprocess.run(
                compose_cmd + ["up", "--build", "-d"],
                capture_output=True,
                text=True,
                cwd=self.project_root
            )
            
            if result.returncode != 0:
                print("Error updating services:")
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