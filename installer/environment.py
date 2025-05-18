import os
import sys
import json
import secrets
import string
from pathlib import Path
import subprocess
from docker_setup import DockerSetup

class EnvironmentSetup:
    def __init__(self, domains, env="staging"):
        self.domains = domains
        self.project_root = Path(__file__).parent.parent
        self.env_file = self.project_root / ".env"
        self.env_sample = self.project_root / ".env.sample"
        self.config_dir = Path("/etc/nixopus-staging") if env == "staging" else Path("/etc/nixopus")
        self.ssh_dir = self.config_dir / "ssh"
        self.env = env
        self.context_name = f"nixopus-{self.env}"
        self.source_dir = self.config_dir / "source"
        self.docker_setup = DockerSetup(env)

    def generate_random_string(self, length=12):
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
            except Exception as e:
                raise Exception(f"Error generating SSH key: {str(e)}")
        
        return private_key_path, public_key_path

    def get_version(self):
        version_file = self.project_root / "version.txt"
        if version_file.exists():
            with open(version_file, 'r') as f:
                return f.read().strip()
        return "unknown"

    def setup_environment(self):
        db_name = f"nixopus_{self.generate_random_string(8)}"
        username = f"nixopus_{self.generate_random_string(8)}"
        password = self.generate_random_string(16)
        
        api_port = 8443 if self.env == "production" else 8444
        next_public_port = 7443 if self.env == "production" else 7444
        db_port = 5432 if self.env == "production" else 5433

        private_key_path, public_key_path = self.generate_ssh_key()
        self.setup_authorized_keys()
        local_ip = self.docker_setup.get_local_ip()

        docker_context = self.docker_setup.setup()

        base_env_vars = {
            "DB_NAME": db_name,
            "USERNAME": username,
            "PASSWORD": password,
            "HOST_NAME": "nixopus-db" if self.env == "production" else f"nixopus-{self.env}-db",
            "DB_PORT": str(db_port),
            "SSL_MODE": "disable",
            "API_PORT": str(api_port),
            "API_URL": f"https://{self.domains['api_domain']}/api",
            "WEBSOCKET_URL": f"wss://{self.domains['api_domain']}/ws",
            "WEBHOOK_URL": f"https://{self.domains['api_domain']}/api/v1/webhook",
            "NEXT_PUBLIC_PORT": str(next_public_port),
            "MOUNT_PATH": "/etc/nixopus/configs" if self.env == "production" else "/etc/nixopus-staging/configs",
            "PORT": str(api_port),
            "SSH_HOST": local_ip,
            "SSH_PORT": "22",
            "SSH_USER": "root",
            "SSH_PRIVATE_KEY": str(private_key_path),
            "DOCKER_HOST": f"tcp://{local_ip}:2376" if self.env == "production" else f"tcp://{local_ip}:2377",
            "DOCKER_TLS_VERIFY": "1",
            "DOCKER_CERT_PATH": str(self.docker_setup.docker_certs_dir),
            "DOCKER_CONTEXT": docker_context,
            "CADDY_ENDPOINT": "http://nixopus-caddy:2019",
            "CADDY_DATA_VOLUME": str(self.config_dir / "caddy" / "data"),
            "CADDY_CONFIG_VOLUME": str(self.config_dir / "caddy" / "config"),
            "DB_VOLUME": str(self.config_dir / "db"),
            "ALLOWED_ORIGIN": f"https://{self.domains['app_domain']}",
            "APP_VERSION": self.get_version(),
            "REDIS_URL": "redis://nixopus-redis:6379" if self.env == "production" else "redis://nixopus-staging-redis:6380"
        }

        with open(self.env_file, 'w') as f:
            for key, value in base_env_vars.items():
                f.write(f"{key}={value}\n")

        api_env_vars = base_env_vars.copy()
        api_env_vars["PORT"] = str(api_port)
        api_env_file = self.source_dir / "api" / ".env"
        api_env_file.parent.mkdir(parents=True, exist_ok=True)
        with open(api_env_file, 'w') as f:
            for key, value in api_env_vars.items():
                f.write(f"{key}={value}\n")
        
        view_env_vars = base_env_vars.copy()
        view_env_vars["PORT"] = str(next_public_port)
        view_env_file = self.source_dir / "view" / ".env"
        view_env_file.parent.mkdir(parents=True, exist_ok=True)
        with open(view_env_file, 'w') as f:
            for key, value in view_env_vars.items():
                f.write(f"{key}={value}\n")
        
        self.env_file.chmod(0o600)
        private_key_path.chmod(0o600)
        public_key_path.chmod(0o644)
        return base_env_vars