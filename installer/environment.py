import os
import sys
import json
import secrets
import string
from pathlib import Path
import subprocess
from docker_setup import DockerSetup
from dataclasses import dataclass
from typing import Dict, Optional

@dataclass
class EnvironmentConfig:
    env: str
    config_dir: Path
    api_port: int
    next_public_port: int
    db_port: int
    host_name: str
    redis_url: str
    mount_path: str
    docker_host: str
    docker_port: int

    @classmethod
    def create(cls, env: str) -> 'EnvironmentConfig':
        if env not in ["production", "staging"]:
            raise ValueError(f"Invalid environment: {env}. Must be either 'production' or 'staging'")
            
        is_production = env == "production"
        config_dir = Path("/etc/nixopus") if is_production else Path("/etc/nixopus-staging")
        return cls(
            env=env,
            config_dir=config_dir,
            api_port=8443 if is_production else 8444,
            next_public_port=7443 if is_production else 7444,
            db_port=5432 if is_production else 5433,
            host_name="nixopus-db" if is_production else f"nixopus-{env}-db",
            redis_url="redis://nixopus-redis:6379" if is_production else "redis://nixopus-staging-redis:6380",
            mount_path="/etc/nixopus/configs" if is_production else "/etc/nixopus-staging/configs",
            docker_host="tcp://{ip}:2376" if is_production else "tcp://{ip}:2377",
            docker_port=2376 if is_production else 2377
        )

class EnvironmentSetup:
    def __init__(self, domains: Optional[Dict[str, str]], env: str = "staging"):
        self.domains = domains
        self.project_root = Path(__file__).parent.parent
        self.env_file = self.project_root / ".env"
        self.env_sample = self.project_root / ".env.sample"
        self.config = EnvironmentConfig.create(env)
        self.ssh_dir = self.config.config_dir / "ssh"
        self.context_name = f"nixopus-{self.config.env}"
        self.source_dir = self.config.config_dir / "source"
        self.docker_setup = DockerSetup(env)

    def generate_random_string(self, length=12):
        if length <= 0:
            raise ValueError("Length must be positive")
            
        alphabet = string.ascii_letters + string.digits
        return ''.join(secrets.choice(alphabet) for _ in range(length))
    
    def setup_authorized_keys(self):
        try:
            ssh_config_dir = Path.home() / ".ssh"
            ssh_config_dir.mkdir(mode=0o700, parents=True, exist_ok=True)
            authorized_keys_path = ssh_config_dir / "authorized_keys"
            
            _, public_key_path = self.generate_ssh_key()
            with open(public_key_path, 'r') as pk_file:
                public_key_content = pk_file.read().strip()
                
            if authorized_keys_path.exists():
                with open(authorized_keys_path, 'r') as auth_file:
                    existing_content = auth_file.read()
                    if public_key_content in existing_content:
                        return
                        
            with open(authorized_keys_path, 'a+') as auth_file:
                auth_file.write(f"\n{public_key_content}\n")
                
            authorized_keys_path.chmod(0o600)
            
        except Exception as e:
            raise Exception(f"Error setting up authorized_keys: {str(e)}")

    def generate_ssh_key(self):
        self.ssh_dir.mkdir(parents=True, exist_ok=True)
        private_key_path = self.ssh_dir / "id_rsa"
        public_key_path = self.ssh_dir / "id_rsa.pub"
        
        if not private_key_path.exists():
            try:
                result = subprocess.run(
                    ["ssh-keygen", "-t", "rsa", "-b", "4096", "-f", str(private_key_path), "-N", ""],
                    capture_output=True,
                    text=True
                )
                if result.returncode != 0:
                    raise Exception("Failed to generate SSH key")
            except FileNotFoundError as e:
                raise Exception(f"ssh-keygen not found: {str(e)}")
            except Exception as e:
                raise Exception(f"Error generating SSH key: {str(e)}")
        
        return private_key_path, public_key_path

    def get_version(self):
        version_file = self.project_root / "version.txt"
        if version_file.exists():
            try:
                with open(version_file, 'r') as f:
                    return f.read().strip()
            except IOError as e:
                raise Exception(f"File read error: {str(e)}")
        return "unknown"

    def setup_environment(self):
        try:
            db_name = f"nixopus_{self.generate_random_string(8)}"
            username = f"nixopus_{self.generate_random_string(8)}"
            password = self.generate_random_string(16)
            
            private_key_path, public_key_path = self.generate_ssh_key()
            self.setup_authorized_keys()
            local_ip = self.docker_setup.get_public_ip()

            docker_context = self.docker_setup.setup()

            domain_not_provided = self.domains is None
            if not domain_not_provided:
                if not all(isinstance(domain, str) and '.' in domain for domain in self.domains.values()):
                    raise ValueError("Invalid domain format. Domains must be valid hostnames")

            api_url = f"https://{self.domains['api_domain']}/api" if not domain_not_provided else f"http://{local_ip}:{self.config.api_port}/api"
            websocket_url = f"wss://{self.domains['api_domain']}/ws" if not domain_not_provided else f"ws://{local_ip}:{self.config.api_port}/ws"
            webhook_url = f"https://{self.domains['api_domain']}/api/v1/webhook" if not domain_not_provided else f"http://{local_ip}:{self.config.api_port}/api/v1/webhook"
            allowed_origin = f"https://{self.domains['app_domain']}" if not domain_not_provided else f"http://{local_ip}:{self.config.next_public_port}"
            
            base_env_vars = {
                "DB_NAME": db_name,
                "USERNAME": username,
                "PASSWORD": password,
                "HOST_NAME": self.config.host_name,
                "DB_PORT": str(self.config.db_port),
                "SSL_MODE": "disable",
                "API_PORT": str(self.config.api_port),
                "API_URL": api_url,
                "WEBSOCKET_URL": websocket_url,
                "WEBHOOK_URL": webhook_url,
                "NEXT_PUBLIC_PORT": str(self.config.next_public_port),
                "MOUNT_PATH": self.config.mount_path,
                "PORT": str(self.config.api_port),
                "SSH_HOST": local_ip,
                "SSH_PORT": "22",
                "SSH_USER": "root",
                "SSH_PRIVATE_KEY": str(private_key_path),
                "DOCKER_HOST": self.config.docker_host.format(ip=local_ip),
                "DOCKER_TLS_VERIFY": "1",
                "DOCKER_CERT_PATH": str(self.docker_setup.docker_certs_dir),
                "DOCKER_CONTEXT": docker_context,
                "CADDY_ENDPOINT": "http://nixopus-caddy:2019",
                "CADDY_DATA_VOLUME": str(self.config.config_dir / "caddy" / "data"),
                "CADDY_CONFIG_VOLUME": str(self.config.config_dir / "caddy" / "config"),
                "DB_VOLUME": str(self.config.config_dir / "db"),
                "ALLOWED_ORIGIN": allowed_origin,
                "APP_VERSION": self.get_version(),
                "REDIS_URL": self.config.redis_url
            }

            try:
                with open(self.env_file, 'w') as f:
                    for key, value in base_env_vars.items():
                        f.write(f"{key}={value}\n")

                api_env_vars = base_env_vars.copy()
                api_env_vars["PORT"] = str(self.config.api_port)
                api_env_file = self.source_dir / "api" / ".env"
                api_env_file.parent.mkdir(parents=True, exist_ok=True)
                with open(api_env_file, 'w') as f:
                    for key, value in api_env_vars.items():
                        f.write(f"{key}={value}\n")
                
                view_env_vars = base_env_vars.copy()
                view_env_vars["PORT"] = str(self.config.next_public_port)
                view_env_file = self.source_dir / "view" / ".env"
                view_env_file.parent.mkdir(parents=True, exist_ok=True)
                with open(view_env_file, 'w') as f:
                    for key, value in view_env_vars.items():
                        f.write(f"{key}={value}\n")
                
                self.env_file.chmod(0o600)
                private_key_path.chmod(0o600)
                public_key_path.chmod(0o644)
                return base_env_vars
            except IOError as e:
                raise Exception(f"File write error: {str(e)}")
        except ValueError as e:
            raise e
        except Exception as e:
            raise Exception(f"Error setting up environment: {str(e)}")