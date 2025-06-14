#!/usr/bin/env python3

import time
from pathlib import Path
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
        self.service_manager = None  # Will be initialized with environment later

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
    
    env = installer.input_parser.get_env_from_args(args)
    domains = installer.input_parser.get_domains_from_args(args)
    
    # enable this when domain is a required field, we can use ip address for now
    # if not domains:
    #     domains = installer.input_parser.ask_for_domain()
    
    email, password = installer.input_parser.get_admin_credentials_from_args(args)
    # if email is None and password is None:
    #     email, password = installer.input_parser.ask_admin_credentials()
    
    installer.service_manager = ServiceManager(installer.project_root, env)
    installer.service_manager.check_system_requirements()
    env_setup = EnvironmentSetup(domains,env) # TODO: make this dynamic 
    env_vars = env_setup.setup_environment()
    print("Environment setup completed!")
    installer.service_manager.start_services(env)
    installer.service_manager.verify_installation(env)
    if domains is not None:
        installer.service_manager.setup_caddy(domains,env)
    time.sleep(10)
    max_retries = 3
    retry_count = 0
    
    # setup admin if email and password are provided through args
    if email is not None and password is not None:
        while retry_count < max_retries:
            if installer.service_manager.check_api_up_status(env_vars["API_PORT"]):
                installer.service_manager.setup_admin(email, password, env_vars["API_PORT"])
                break
            retry_count += 1
            if retry_count < max_retries:
                time.sleep(2)
                print(f"Retrying API status check (attempt {retry_count + 1}/{max_retries})...")
    
    docker_setup = installer.service_manager.docker_setup
    if domains and isinstance(domains, dict) and domains.get("app_domain"):
        nixopus_accessible_at = domains["app_domain"]
    else:
        nixopus_accessible_at = f"{docker_setup.get_public_ip()}:{env_vars.get("NEXT_PUBLIC_PORT", "3000")}"
    print("\n\033[1mInstallation Complete!\033[0m")
    print(f"• Nixopus is accessible at: {nixopus_accessible_at}")
    
    # print("\n\033[1mAdmin Credentials:\033[0m")
    # print(f"• Email: {email}")
    # print(f"• Password: {password}")
    # print("\n\033[1mImportant:\033[0m Please save these credentials securely. You will need them to log in.")
    print("\n\033[1mThank you for installing Nixopus!\033[0m")
    print("\n\033[1mPlease visit the documentation at https://docs.nixopus.com for more information.\033[0m")
    print("\n\033[1mIf you have any questions, please visit the community forum at https://discord.gg/skdcq39Wpv\033[0m")
    print("\n\033[1mSee you in the community!\033[0m")

if __name__ == "__main__":
    main() 