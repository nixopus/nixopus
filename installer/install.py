#!/usr/bin/env python3

import os
import sys
from pathlib import Path
from validation import Validation
from environment import EnvironmentSetup
from input_parser import InputParser
from service_manager import ServiceManager

class Installer:
    def __init__(self):
        self.required_docker_version = "20.10.0"
        self.required_compose_version = "2.0.0"
        self.project_root = Path(__file__).parent.parent
        self.env_file = self.project_root / ".env"
        self.env_sample = self.project_root / ".env.sample"
        self.input_parser = InputParser()
        self.service_manager = ServiceManager(self.project_root)
    
    # this script will only work with root privileges
    def ask_for_sudo(self):
        if os.geteuid() != 0:
            print("Please run the script with sudo privileges")
            sys.exit(1)

    def setup_environment(self, domain):
        print("\nSetting up environment...")
        env_setup = EnvironmentSetup(domain)
        env_vars = env_setup.setup_environment()
        print("Environment setup completed!")
        return env_vars

def main():
    installer = Installer()
    args = installer.input_parser.parser.parse_args()
    
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
    
    installer.ask_for_sudo()
    
    domains = installer.input_parser.get_domains_from_args(args)
    if not domains:
        domains = installer.input_parser.ask_for_domain()
    
    email, password = installer.input_parser.get_admin_credentials_from_args(args)
    if email is None and password is None:
        email, password = installer.input_parser.ask_admin_credentials()
    
    installer.service_manager.check_system_requirements()
    env_vars = installer.setup_environment(domains)
    installer.service_manager.start_services()
    installer.service_manager.verify_installation()
    installer.service_manager.setup_caddy(domains)
    installer.service_manager.setup_admin(email, password, domains["api_domain"])
    
    print("\n\033[1mInstallation Complete!\033[0m")
    print("\n\033[1mAccess Information:\033[0m")
    print(f"• API is accessible at: {domains['api_domain']}")
    print(f"• App is accessible at: {domains['app_domain']}")
    
    print("\n\033[1mAdmin Credentials:\033[0m")
    print(f"• Email: {email}")
    print(f"• Password: {password}")
    print("\n\033[1mImportant:\033[0m Please save these credentials securely. You will need them to log in.")
    print("\n\033[1mThank you for installing Nixopus!\033[0m")
    print("\n\033[1mPlease visit the documentation at https://docs.nixopus.com for more information.\033[0m")
    print("\n\033[1mIf you have any questions, please visit the community forum at https://discord.gg/skdcq39Wpv\033[0m")
    print("\n\033[1mSee you in the community!\033[0m")

if __name__ == "__main__":
    main() 