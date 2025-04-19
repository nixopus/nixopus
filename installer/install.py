#!/usr/bin/env python3

import os
import sys
import subprocess
import platform
from pathlib import Path
import re
from validation import Validation
from environment import EnvironmentSetup
import shutil
import json
import string
import secrets
import argparse
import requests

class Installer:
    def __init__(self):
        self.required_docker_version = "20.10.0"
        self.required_compose_version = "2.0.0"
        self.project_root = Path(__file__).parent.parent
        self.env_file = self.project_root / ".env"
        self.env_sample = self.project_root / ".env.sample"
        self.parser = self._setup_arg_parser()
    
    def _setup_arg_parser(self):
        parser = argparse.ArgumentParser(description='Nixopus Installation Wizard')
        parser.add_argument('--api-domain', help='The domain where the nixopus api will be accessible (e.g. nixopusapi.example.com)')
        parser.add_argument('--app-domain', help='The domain where the nixopus app will be accessible (e.g. nixopus.example.com)')
        parser.add_argument('--email', '-e', help='The email to create the admin account with')
        parser.add_argument('--password', '-p', help='The password to create the admin account with')
        return parser
    
    def get_domains_from_args(self, args):
        if args.api_domain or args.app_domain:
            validation = Validation()
            base_domain = None
            
            try:
                if args.api_domain and args.app_domain:
                    validation.validate_domain(args.api_domain)
                    validation.validate_domain(args.app_domain)
                    base_domain = args.api_domain.split('.', 1)[1] if '.' in args.api_domain else args.api_domain
                    return {
                        "api_domain": args.api_domain,
                        "app_domain": args.app_domain,
                        "base_domain": base_domain
                    }
                elif args.api_domain:
                    validation.validate_domain(args.api_domain)
                    base_domain = args.api_domain.split('.', 1)[1] if '.' in args.api_domain else args.api_domain
                    return {
                        "api_domain": args.api_domain,
                        "app_domain": f"nixopus.{base_domain}",
                        "base_domain": base_domain
                    }
                elif args.app_domain:
                    validation.validate_domain(args.app_domain)
                    base_domain = args.app_domain.split('.', 1)[1] if '.' in args.app_domain else args.app_domain
                    return {
                        "api_domain": f"nixopusapi.{base_domain}",
                        "app_domain": args.app_domain,
                        "base_domain": base_domain
                    }
            except SystemExit:
                print("Invalid domain provided. Please try again with valid domains.")
                return None
        return None
    
    def get_admin_credentials_from_args(self, args):
        if args.email or args.password:
            validation = Validation()
            
            if args.email and args.password:
                validation.validate_email(args.email)
                validation.validate_password(args.password)
                return args.email, args.password
            elif args.email:
                validation.validate_email(args.email)
                password = input("Please enter the password for the admin(generates a strong password if left blank): ")
                if not password:
                    password = self.generate_strong_password()
                validation.validate_password(password)
                return args.email, password
            elif args.password:
                validation.validate_password(args.password)
                while True:
                    email = input("Please enter the email for the admin: ")
                    try:
                        validation.validate_email(email)
                        return email, args.password
                    except SystemExit:
                        print("Please enter a valid email address")
                        continue
        return None, None
    
    # this script will only work with root privileges
    def ask_for_sudo(self):
        if os.geteuid() != 0:
            print("Please run the script with sudo privileges")
            sys.exit(1)
    
    def generate_strong_password(self):
        while True:
            password = ''.join(secrets.choice(
                string.ascii_letters + string.digits + string.punctuation
            ) for _ in range(16))
            if (any(c.isupper() for c in password) and
                any(c.islower() for c in password) and
                any(c.isdigit() for c in password) and
                any(c in string.punctuation for c in password)):
                return password

    def ask_admin_credentials(self):
        validation = Validation()
        while True:
            email = input("Please enter the email for the admin: ")
            try:
                validation.validate_email(email)
                break
            except SystemExit:
                print("Please enter a valid email address")
                continue
                
        password = input("Please enter the password for the admin(generates a strong password if left blank): ")
        if not password:
            password = self.generate_strong_password()
        validation.validate_password(password)
        return email, password
    
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

    def check_system_requirements(self):
        print("Checking system requirements...")
        
        # we will only support linux for now
        system = platform.system()
        if system not in ["Linux"]:
            print(f"Error: Unsupported operating system: {system}")
            sys.exit(1)

        self.check_docker_version()
        self.check_docker_compose_version()
        self.check_curl_installed()

        print("System requirements check passed!")

    def _version_check(self, version_string, required_version):
        version = re.search(r'\d+\.\d+\.\d+', version_string)
        if not version:
            return False
        return tuple(map(int, version.group().split('.'))) >= tuple(map(int, required_version.split('.')))

    def setup_environment(self, domain):
        print("\nSetting up environment...")
        env_setup = EnvironmentSetup(domain)
        env_vars = env_setup.setup_environment()
        print("Environment setup completed!")
        return env_vars

    def start_services(self):
        print("\nStarting services...")
        try:
            # Check if docker daemon is running
            try:
                subprocess.run(["docker", "info"], check=True, capture_output=True)
            except subprocess.CalledProcessError:
                print("Error: Docker daemon is not running. Please start the Docker service and try again.")
                sys.exit(1)

            os.environ["DOCKER_HOST"] = "tcp://localhost:2376"
            os.environ["DOCKER_TLS_VERIFY"] = "1"
            os.environ["DOCKER_CERT_PATH"] = "/etc/nixopus/docker-certs"
            
            compose_cmd = ["docker", "compose"] if shutil.which("docker") else ["docker-compose"]
            
            print("Pulling required images...")
            pull_result = subprocess.run(
                compose_cmd + ["pull"],
                capture_output=True,
                text=True,
                cwd=self.project_root
            )
            
            if pull_result.returncode != 0:
                print("Error pulling images:")
                print(pull_result.stderr)
                print("\nTroubleshooting tips:")
                print("1. Check your internet connection")
                print("2. Verify Docker is running: sudo systemctl status docker")
                print("3. Try pulling images manually: docker pull <image-name>")
                sys.exit(1)
            
            print("Starting services...")
            result = subprocess.run(
                compose_cmd + ["up", "-d"],
                capture_output=True,
                text=True,
                cwd=self.project_root
            )
            
            if result.returncode != 0:
                print("Error starting services:")
                print(result.stderr)
                print("\nTroubleshooting tips:")
                print("1. Check if ports are already in use")
                print("2. Verify Docker has enough resources")
                print("3. Check Docker logs: docker-compose logs")
                sys.exit(1)
        except Exception as e:
            print(f"Error starting services: {str(e)}")
            print("\nTroubleshooting tips:")
            print("1. Check if Docker is installed and running")
            print("2. Verify you have sufficient permissions")
            print("3. Check system resources (memory, disk space)")
            sys.exit(1)

    def verify_installation(self):
        print("\nVerifying installation...")
        try:
            result = subprocess.run(["docker", "ps", "--format", "{{.Names}} {{.Status}}"], capture_output=True, text=True)
            if result.returncode != 0:
                print("Error verifying installation:")
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
            print(f"Error verifying installation: {str(e)}")
            sys.exit(1)
    
    def setup_caddy(self, domains):
        print("\nSetting up Proxy...")
        try:
            with open('helpers/caddy.json', 'r') as f:
                config = json.load(f)
            
            os.environ['APP_DOMAIN'] = domains['app_domain']
            os.environ['API_DOMAIN'] = domains['api_domain']
            
            result = subprocess.run(
                ['curl', '-X', 'POST', 'http://localhost:2019/load',
                 '-H', 'Content-Type: application/json',
                 '-d', json.dumps(config)],
                capture_output=True,
                text=True
            )
            
            if result.returncode == 0:
                print("Caddy configuration loaded successfully")
            else:
                print("Failed to load Caddy configuration:")
                print(result.stderr)
        except Exception as e:
            print(f"Error setting up Caddy: {str(e)}")
    
    def setup_admin(self, email, password, api_domain):
        print("\nSetting up admin...")
        username = email.split('@')[0]
        
        try:
            response = requests.post(
                f"https://{api_domain}/api/v1/auth/register",
                json={
                    "email": email,
                    "password": password,
                    "type": "admin",
                    "username": username,
                    "organization": ""
                },
                headers={"Content-Type": "application/json"}
            )
            
            if response.status_code == 200:
                print("Admin setup completed successfully")
                return
                
            if response.status_code == 400 and "admin already registered" in response.text:
                print("Admin already registered")
                return
                
            error_msg = response.json().get("message", "Unknown error")
            print(f"API Error: {error_msg}")
            raise Exception(f"API Error: {error_msg}")
            
        except requests.exceptions.RequestException as e:
            print(f"Request failed: {str(e)}")
            raise Exception(f"Failed to connect to API: {str(e)}")
        except json.JSONDecodeError as e:
            print(f"Invalid JSON response: {response.text}")
            raise Exception(f"Invalid response from API: {str(e)}")

def main():
    installer = Installer()
    args = installer.parser.parse_args()
    
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
    
    domains = installer.get_domains_from_args(args)
    if not domains:
        validation = Validation()
        while True:
            domain = input("Please enter the base domain (e.g. nixopus.com): ")
            try:
                validation.validate_domain(domain)
                domains = {
                    "api_domain": f"nixopusapi.{domain}",
                    "app_domain": f"nixopus.{domain}",
                    "base_domain": domain
                }
                break
            except SystemExit:
                print("Please enter a valid domain name")
                continue
    
    email, password = installer.get_admin_credentials_from_args(args)
    if not email or not password:
        email, password = installer.ask_admin_credentials()
    
    installer.check_system_requirements()
    env_vars = installer.setup_environment(domains)
    installer.start_services()
    installer.verify_installation()
    installer.setup_caddy(domains)
    
    try:
        installer.setup_admin(email, password, domains["api_domain"])
    except Exception as e:
        print(f"{str(e)}")
        sys.exit(1)
    
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