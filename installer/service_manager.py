import subprocess
import shutil
import json
import requests
import os
import sys
import platform
from pathlib import Path
from docker_setup import DockerSetup
from base_config import BaseConfig
from service_config import ServiceConfig
import time

class ServiceManager:
    def __init__(self, project_root, env="staging", debug=False):
        self.project_root = project_root
        self.debug = debug
        self.docker_setup = DockerSetup(env, debug)
        self.config = self._load_config(env)

    def debug_print(self, message):
        if self.debug:
            print(f"[DEBUG] {message}")

    def _load_config(self, env: str) -> ServiceConfig:
        self.debug_print("Loading service configuration...")
        config_path = Path(__file__).parent.parent / "helpers" / "config.json"
        base_config = BaseConfig[ServiceConfig](
            config_path=config_path,
            env=env,
            required_keys=[
                "config_dir", "docker", "source", "compose", "containers", "caddy", "api", "system"
            ],
            valid_environments=["production", "staging"]
        )
        return base_config.create(ServiceConfig)

    def check_system_requirements(self):
        self.debug_print("Checking system requirements...")
        
        system = platform.system()
        if system not in self.config.system['supported_os']:
            print(f"Error: Unsupported operating system: {system}")
            sys.exit(1)

        self.check_required_tools()

        self.debug_print("System requirements check passed!")
            
    def check_required_tools(self):
        self.debug_print("Checking required tools...")
        for tool in self.config.system['required_tools']:
            if not shutil.which(tool):
                print(f"Error: {tool} is not installed")
                sys.exit(1)

    def start_services(self, env):
        self.debug_print("Starting services...")
        try:
            try:
                subprocess.run(["docker", "info"], check=True, capture_output=True)
            except subprocess.CalledProcessError:
                print("Error: Docker daemon is not running. Please start the Docker service and try again.")
                sys.exit(1)

            os.environ["DOCKER_HOST"] = f"tcp://localhost:{self.config.docker['port']}"
            os.environ["DOCKER_TLS_VERIFY"] = "1"
            os.environ["DOCKER_CERT_PATH"] = self.config.docker['cert_path']
            os.environ["DOCKER_CONTEXT"] = self.config.docker['context']
            
            compose_file = os.path.join(self.config.source, self.config.compose['file'])
            
            self.debug_print(f"Using Docker Compose file: {compose_file}")
            if not os.path.exists(compose_file):
                print(f"Error: Docker Compose file not found at {compose_file}")
                sys.exit(1)
                
            compose_cmd = ["docker", "compose", "-f", compose_file]
            
            if env == "staging":
                self.debug_print("Building and starting staging services...")
                result = subprocess.run(
                    compose_cmd + ["up", "--build", "-d"],
                    capture_output=True,
                    text=True,
                    cwd=self.project_root
                )
                if result.returncode != 0:
                    print("Error building and starting services:")
                    print(result.stderr)
                    raise Exception("Failed to build and start services")
            else:
                self.debug_print("Pulling production images...")
                pull_result = subprocess.run(
                    compose_cmd + ["pull"],
                    capture_output=True,
                    text=True,
                    cwd=self.project_root
                )
                if pull_result.returncode != 0:
                    print("Error pulling images:")
                    print(pull_result.stderr)
                    raise Exception("Failed to pull images")
                
                self.debug_print("Starting services...")
                result = subprocess.run(
                    compose_cmd + ["up", "-d"],
                    capture_output=True,
                    text=True,
                    cwd=self.project_root
                )
                if result.returncode != 0:
                    print("Error starting services:")
                    print(result.stderr)
                    raise Exception("Failed to start services")
        except Exception as e:
            print(f"Error starting services: {str(e)}")
            sys.exit(1)
            
    def verify_installation(self, env):
        self.debug_print("Verifying installation...")
        try:
            result = subprocess.run(["docker", "ps", "--format", "{{.Names}} {{.Status}}"], capture_output=True, text=True)
            if result.returncode != 0:
                print("Error verifying installation:")
                print(result.stderr)
                sys.exit(1)
                
            running_containers = result.stdout.splitlines()
            required_containers = self.config.containers
            
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

            self.debug_print("All services are running successfully!")
        except Exception as e:
            print(f"Error verifying installation: {str(e)}")
            sys.exit(1)
    
    def setup_caddy(self, domains, env):
        self.debug_print("Setting up Proxy...")
        try:
            with open(self.project_root / self.config.caddy['config_path'], 'r') as f:
                config_str = f.read()
                config_str = config_str.replace('{env.APP_DOMAIN}', domains['app_domain'])
                config_str = config_str.replace('{env.API_DOMAIN}', domains['api_domain'])
                app_reverse_proxy_url = self.config.caddy["reverse_proxy"]["app"]
                api_reverse_proxy_url = self.config.caddy["reverse_proxy"]["api"]
                config_str = config_str.replace('{env.APP_REVERSE_PROXY_URL}', app_reverse_proxy_url)
                config_str = config_str.replace('{env.API_REVERSE_PROXY_URL}', api_reverse_proxy_url)
                new_config = json.loads(config_str)
                self.debug_print("Loading Caddy configuration...")
                response = requests.post(
                    f'http://localhost:{self.config.caddy["admin_port"]}/load',
                    json=new_config,
                    headers={'Content-Type': 'application/json'}
                )
                if response.status_code != 200:
                    print("Failed to create server configuration:")
                    print(response.text)
                    raise Exception("Failed to create server configuration")
            self.debug_print("Caddy configuration loaded successfully")
        except requests.exceptions.RequestException as e:
            print(f"Error connecting to Caddy: {str(e)}")
        except Exception as e:
            print(f"Error setting up Caddy: {str(e)}")
    
    def check_api_up_status(self, port):
        self.debug_print(f"Checking API status on port {port}...")
        try:
            response = requests.get(f"http://localhost:{port}{self.config.api['health_endpoint']}", verify=False)
            if response.status_code == 200:
                self.debug_print("API is up and running")
                return True
            self.debug_print("API is not responding")
            return False
        except requests.exceptions.RequestException as e:
            self.debug_print(f"Error checking API status: {str(e)}")
            return False
    
    def setup_admin(self, email, password, port):
        self.debug_print("Setting up admin...")
        username = email.split('@')[0]
        
        try:
            self.debug_print(f"Creating admin account for {email}...")
            response = requests.post(
                f"http://localhost:{port}{self.config.api['register_endpoint']}",
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
                self.debug_print("Admin setup completed successfully")
                return
                
            if response.status_code == 400 and "admin already registered" in response.text:
                self.debug_print("Admin already registered")
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
