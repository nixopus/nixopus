#!/usr/bin/env python3

import time
import logging
from pathlib import Path
from environment import EnvironmentSetup
from input_parser import InputParser
from service_manager import ServiceManager
import sys

class Installer:
    def __init__(self):
        self.required_docker_version = "20.10.0"
        self.required_compose_version = "2.0.0"
        self.project_root = Path(__file__).parent.parent
        self.env_file = self.project_root / ".env"
        self.env_sample = self.project_root / ".env.sample"
        self.input_parser = InputParser()
        self.service_manager = None  # Will be initialized with environment later
        self.logger = logging.getLogger("nixopus")

def main():
    installer = Installer()
    args = installer.input_parser.parse_args()
    
    print("\033[36m  _   _ _ _                           \033[0m")
    print("\033[36m | \\ | (_)                          \033[0m")
    print("\033[36m |  \\| |___  _____  _ __  _   _ ___ \033[0m")
    print("\033[36m | . ` | \\ \\/ / _ \\| '_ \\| | | / __|\033[0m")
    print("\033[36m | |\\  | |>  < (_) | |_) | |_| \\__ \033[0m")
    print("\033[36m |_| \\_|_/_/\\_\\___/| .__/ \\__,_|___/\033[0m")
    print("\033[36m                   | |              \033[0m")
    print("\033[36m                   |_|              \033[0m")
    print("\n")
    print("\033[1mWelcome to Nixopus Installation Wizard\033[0m")
    print("This wizard will guide you through the installation process of Nixopus.")
    print("Please follow the prompts carefully to complete the setup.\n")
    
    installer.logger.debug("Starting installation process...")
    
    env = installer.input_parser.get_env_from_args(args)
    domains = installer.input_parser.get_domains_from_args(args)
    email, password = installer.input_parser.get_admin_credentials_from_args(args)
    
    installer.logger.debug("Initializing service manager...")
    installer.service_manager = ServiceManager(installer.project_root, env, args.debug)
    
    installer.logger.debug("Checking system requirements...")
    installer.service_manager.check_system_requirements()
    
    installer.logger.debug("Setting up environment...")
    env_setup = EnvironmentSetup(domains, env, args.debug)
    env_vars = env_setup.setup_environment()
    
    installer.logger.debug("Environment setup completed!")
    
    installer.logger.debug("Starting services...")
    installer.service_manager.start_services(env)
    
    installer.logger.debug("Verifying installation...")
    installer.service_manager.verify_installation(env)
    
    if domains is not None:
        installer.logger.debug("Setting up Caddy reverse proxy...")
        installer.service_manager.setup_caddy(domains, env)
    
    installer.logger.debug("Waiting for services to start...")
    time.sleep(10)
    
    max_retries = 3
    retry_count = 0
    
    if email is not None and password is not None:
        installer.logger.debug("Setting up admin account...")
        while retry_count < max_retries:
            if installer.service_manager.check_api_up_status(env_vars["API_PORT"]):
                installer.logger.debug("API is up, creating admin account...")
                installer.service_manager.setup_admin(email, password, env_vars["API_PORT"])
                break
            retry_count += 1
            if retry_count < max_retries:
                installer.logger.debug(f"Retrying API status check (attempt {retry_count + 1}/{max_retries})...")
                time.sleep(2)
    
    docker_setup = installer.service_manager.docker_setup
    if domains and isinstance(domains, dict) and domains.get("app_domain"):
        nixopus_accessible_at = domains["app_domain"]
        installer.logger.debug(f"Using domain for access: {nixopus_accessible_at}")
    else:
        app_port = str(env_vars.get("APP_PORT", ""))
        public_ip = docker_setup.get_public_ip()
        nixopus_accessible_at = (
            public_ip
            if app_port in {"80", "443"}
            else f"{public_ip}:{app_port}"
        )
        installer.logger.debug(f"Using IP and port for access: {nixopus_accessible_at}")
    
    print("\n\033[1mInstallation Complete!\033[0m")
    print(f"• Nixopus is accessible at: {nixopus_accessible_at}")
    print("\n\033[1mThank you for installing Nixopus!\033[0m")
    print("\n\033[1mPlease visit the documentation at https://docs.nixopus.com for more information.\033[0m")
    print("\n\033[1mIf you have any questions, please visit the community forum at https://discord.gg/skdcq39Wpv\033[0m")
    print("\n\033[1mSee you in the community!\033[0m")

if __name__ == "__main__":
    main() 