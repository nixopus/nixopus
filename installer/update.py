#!/usr/bin/env python3

import os
import sys
import subprocess
from pathlib import Path
import shutil

def install_git_package():
    try:
        import git
    except ImportError:
        print("Installing gitpython package...")
        try:
            subprocess.check_call([sys.executable, "-m", "pip", "install", "gitpython"])
            print("gitpython package installed successfully!")
        except subprocess.CalledProcessError:
            print("Error: Failed to install gitpython package")
            sys.exit(1)

install_git_package()
import git

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
    
    def update_from_github(self):
        print("\nUpdating from GitHub...")
        try:
            self.repo = git.Repo(self.project_root)
            
            self.repo.git.stash()
            
            self.repo.remotes.origin.pull()
            
            self.repo.git.stash('pop')
            
            print("GitHub update completed successfully!")
        except Exception as e:
            print(f"Error updating from GitHub: {str(e)}")
            sys.exit(1)
    
    def update_services(self):
        print("\nUpdating services...")
        try:
            self.update_from_github()
            os.environ["DOCKER_HOST"] = "tcp://localhost:2376"
            os.environ["DOCKER_TLS_VERIFY"] = "1"
            os.environ["DOCKER_CERT_PATH"] = "/etc/nixopus/docker-certs"
            
            compose_cmd = ["docker", "compose"] if shutil.which("docker") else ["docker-compose"]
            
            result = subprocess.run(compose_cmd + ["up", "--build", "-d"], capture_output=True, text=True, cwd=self.project_root)
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
            result = subprocess.run(["docker", "ps"], capture_output=True, text=True)
            if result.returncode != 0:
                print("Error verifying update:")
                print(result.stderr)
                sys.exit(1)
                
            containers = result.stdout
            required_containers = ["nixopus-api-container", "nixopus-db-container", "nixopus-view-container"]
            
            for container in required_containers:
                if container not in containers:
                    print(f"Error: {container} is not running")
                    sys.exit(1)

            print("Update verified successfully!")
        except Exception as e:
            print(f"Error verifying update: {str(e)}")
            sys.exit(1)

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
    
    print("\n\033[1mUpdate Complete!\033[0m")
    print("\n\033[1mYour Nixopus services have been successfully updated to the latest version.\033[0m")
    print("\n\033[1mThank you for using Nixopus!\033[0m")

if __name__ == "__main__":
    main() 