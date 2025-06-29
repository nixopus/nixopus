import os
import sys
import json
import secrets
import string
import logging
from pathlib import Path
import subprocess
from docker_setup import DockerSetup
from ssh_setup import SSHSetup, SSHConfig
from dataclasses import dataclass
from typing import Dict, Optional
from base_config import BaseConfig

@dataclass
class URLConfig:
    pattern: str
    protocols: Dict[str, str]

@dataclass
class DirectoryConfig:
    ssh: str
    source: str
    api: str
    view: str
    db: str
    caddy: Dict[str, str]

@dataclass
class FileConfig:
    env: str
    env_sample: str
    permissions: Dict[str, str]

@dataclass
class ErrorConfig:
    invalid_environment: str
    config_not_found: str
    invalid_json: str
    env_not_found: str
    missing_keys: str
    invalid_type: str
    invalid_url_type: str
    invalid_dir_type: str
    invalid_subdir_type: str
    ssh_keygen_failed: str
    ssh_keygen_not_found: str
    ssh_key_error: str
    auth_keys_error: str
    file_read_error: str
    file_write_error: str
    invalid_domain: str
    setup_error: str

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
    ssh_port: int
    ssh_user: str
    ssh_key_bits: int
    ssh_key_type: str
    caddy_endpoint: str
    caddy_data_volume: str
    caddy_config_volume: str
    db_name_prefix: str
    db_user_prefix: str
    db_name_length: int
    db_user_length: int
    db_password_length: int
    db_ssl_mode: str
    version_file_path: str
    urls: Dict[str, URLConfig]
    directories: DirectoryConfig
    files: FileConfig
    errors: ErrorConfig

    def get_url(self, url_type: str, host: str, secure: bool = True) -> str:
        if url_type not in self.urls:
            raise ValueError(self.errors.invalid_url_type.format(type=url_type))
            
        url_config = self.urls[url_type]
        protocol = url_config.protocols["secure" if secure else "insecure"]
        return url_config.pattern.format(protocol=protocol, host=host)

    def get_path(self, dir_type: str, sub_type: Optional[str] = None) -> Path:
        if dir_type not in self.directories.__dict__:
            raise ValueError(self.errors.invalid_dir_type.format(type=dir_type))
            
        base_path = getattr(self.directories, dir_type)
        if isinstance(base_path, dict) and sub_type:
            if sub_type not in base_path:
                raise ValueError(self.errors.invalid_subdir_type.format(type=sub_type))
            return self.config_dir / base_path[sub_type]
        return self.config_dir / base_path

    def get_permission(self, file_type: str) -> int:
        if file_type not in self.files.permissions:
            raise ValueError(f"Invalid file type for permissions: {file_type}")
        return int(self.files.permissions[file_type], 8)

    @classmethod
    def create(cls, env: str) -> 'EnvironmentConfig':
        VALID_ENVIRONMENTS = ["production", "staging"]
        REQUIRED_CONFIG_KEYS = [
            "config_dir", "api_port", "next_public_port", "db_port", "host_name",
            "redis_url", "mount_path", "docker_host", "docker_port", "ssh",
            "caddy", "database", "version", "urls", "directories", "files", "errors"
        ]

        config_path = Path(__file__).parent.parent / "helpers" / "config.json"
        base_config = BaseConfig[EnvironmentConfig](
            config_path=config_path,
            env=env,
            required_keys=REQUIRED_CONFIG_KEYS,
            valid_environments=VALID_ENVIRONMENTS
        )

        config = base_config.load_config()
        env_config = config[env]

        try:
            urls = {
                url_type: URLConfig(
                    pattern=url_config["pattern"],
                    protocols=url_config["protocols"]
                )
                for url_type, url_config in env_config["urls"].items()
            }
            
            directories = DirectoryConfig(
                ssh=env_config["directories"]["ssh"],
                source=env_config["directories"]["source"],
                api=env_config["directories"]["api"],
                view=env_config["directories"]["view"],
                db=env_config["directories"]["db"],
                caddy=env_config["directories"]["caddy"]
            )

            files = FileConfig(
                env=env_config["files"]["env"],
                env_sample=env_config["files"]["env_sample"],
                permissions=env_config["files"]["permissions"]
            )

            errors = ErrorConfig(
                invalid_environment=env_config["errors"]["invalid_environment"],
                config_not_found=env_config["errors"]["config_not_found"],
                invalid_json=env_config["errors"]["invalid_json"],
                env_not_found=env_config["errors"]["env_not_found"],
                missing_keys=env_config["errors"]["missing_keys"],
                invalid_type=env_config["errors"]["invalid_type"],
                invalid_url_type=env_config["errors"]["invalid_url_type"],
                invalid_dir_type=env_config["errors"]["invalid_dir_type"],
                invalid_subdir_type=env_config["errors"]["invalid_subdir_type"],
                ssh_keygen_failed=env_config["errors"]["ssh_keygen_failed"],
                ssh_keygen_not_found=env_config["errors"]["ssh_keygen_not_found"],
                ssh_key_error=env_config["errors"]["ssh_key_error"],
                auth_keys_error=env_config["errors"]["auth_keys_error"],
                file_read_error=env_config["errors"]["file_read_error"],
                file_write_error=env_config["errors"]["file_write_error"],
                invalid_domain=env_config["errors"]["invalid_domain"],
                setup_error=env_config["errors"]["setup_error"]
            )
            
            return cls(
                env=env,
                config_dir=Path(env_config["config_dir"]),
                api_port=int(env_config["api_port"]),
                next_public_port=int(env_config["next_public_port"]),
                db_port=int(env_config["db_port"]),
                host_name=str(env_config["host_name"]),
                redis_url=str(env_config["redis_url"]),
                mount_path=str(env_config["mount_path"]),
                docker_host=str(env_config["docker_host"]),
                docker_port=int(env_config["docker_port"]),
                ssh_port=int(env_config["ssh"]["port"]),
                ssh_user=str(env_config["ssh"]["user"]),
                ssh_key_bits=int(env_config["ssh"]["key_bits"]),
                ssh_key_type=str(env_config["ssh"]["key_type"]),
                caddy_endpoint=str(env_config["caddy"]["endpoint"]),
                caddy_data_volume=str(env_config["caddy"]["data_volume"]),
                caddy_config_volume=str(env_config["caddy"]["config_volume"]),
                db_name_prefix=str(env_config["database"]["name_prefix"]),
                db_user_prefix=str(env_config["database"]["user_prefix"]),
                db_name_length=int(env_config["database"]["name_length"]),
                db_user_length=int(env_config["database"]["user_length"]),
                db_password_length=int(env_config["database"]["password_length"]),
                db_ssl_mode=str(env_config["database"]["ssl_mode"]),
                version_file_path=str(env_config["version"]["file_path"]),
                urls=urls,
                directories=directories,
                files=files,
                errors=errors
            )
        except (ValueError, TypeError) as e:
            raise Exception(env_config["errors"]["invalid_type"].format(env=env, error=str(e))) from e

class EnvironmentSetup:
    def __init__(self, domains: Optional[Dict[str, str]], env: str = "staging", debug: bool = False):
        self.domains = domains
        self.project_root = Path(__file__).parent.parent
        self.config = EnvironmentConfig.create(env)
        self.env_file = self.project_root / self.config.files.env
        self.env_sample = self.project_root / self.config.files.env_sample
        self.ssh_dir = self.config.get_path("ssh")
        self.context_name = f"nixopus-{self.config.env}"
        self.source_dir = self.config.get_path("source")
        self.logger = logging.getLogger("nixopus")
        if debug:
            self.logger.setLevel(logging.DEBUG)
        self.docker_setup = DockerSetup(env, debug)
        self.ssh_setup = SSHSetup(
            SSHConfig(
                port=self.config.ssh_port,
                user=self.config.ssh_user,
                key_bits=self.config.ssh_key_bits,
                key_type=self.config.ssh_key_type,
                errors=self.config.errors
            ),
            self.ssh_dir
        )

    def generate_random_string(self, length=12):
        if length <= 0:
            raise ValueError("Length must be positive")
            
        alphabet = string.ascii_letters + string.digits
        return ''.join(secrets.choice(alphabet) for _ in range(length))
    
    def get_version(self):
        self.logger.debug("Getting version from version file")
        version_file = self.project_root / self.config.version_file_path
        if version_file.exists():
            try:
                with open(version_file, 'r') as f:
                    version = f.read().strip()
                    self.logger.debug(f"Version: {version}")
                    return version
            except IOError as e:
                self.logger.debug(f"Error reading version file: {e}")
                raise Exception(self.config.errors.file_read_error.format(error=str(e)))
        self.logger.debug("Version file not found, returning 'unknown'")
        return "unknown"

    def setup_environment(self):
        try:
            self.logger.debug("Starting environment setup")
            db_name = f"{self.config.db_name_prefix}{self.generate_random_string(self.config.db_name_length)}"
            username = f"{self.config.db_user_prefix}{self.generate_random_string(self.config.db_user_length)}"
            password = self.generate_random_string(self.config.db_password_length)
            
            self.logger.debug("Generating SSH keys")
            private_key_path, public_key_path = self.ssh_setup.generate_key()
            self.ssh_setup.setup_authorized_keys(public_key_path, self.config.files.permissions)
            
            self.logger.debug("Getting public IP")
            local_ip = self.docker_setup.get_public_ip()
            self.logger.debug(f"Public IP: {local_ip}")

            self.logger.debug("Setting up Docker context")
            docker_context = self.docker_setup.setup()

            domain_not_provided = self.domains is None
            if not domain_not_provided:
                self.logger.debug("Validating domains")
                if not all(isinstance(domain, str) and '.' in domain for domain in self.domains.values()):
                    raise ValueError(self.config.errors.invalid_domain)

            api_host = self.domains['api_domain'] if not domain_not_provided else f"{local_ip}:{self.config.api_port}"
            app_host = self.domains['app_domain'] if not domain_not_provided else f"{local_ip}:{self.config.next_public_port}"
            
            self.logger.debug("Generating URLs")
            api_url = self.config.get_url("api", api_host, not domain_not_provided)
            websocket_url = self.config.get_url("websocket", api_host, not domain_not_provided)
            webhook_url = self.config.get_url("webhook", api_host, not domain_not_provided)
            allowed_origin = self.config.get_url("app", app_host, not domain_not_provided)
            
            self.logger.debug("Setting up environment variables")
            base_env_vars = {
                "DB_NAME": db_name,
                "USERNAME": username,
                "PASSWORD": password,
                "HOST_NAME": self.config.host_name,
                "DB_PORT": str(self.config.db_port),
                "SSL_MODE": self.config.db_ssl_mode,
                "API_PORT": str(self.config.api_port),
                "API_URL": api_url,
                "WEBSOCKET_URL": websocket_url,
                "WEBHOOK_URL": webhook_url,
                "NEXT_PUBLIC_PORT": str(self.config.next_public_port),
                "MOUNT_PATH": self.config.mount_path,
                "PORT": str(self.config.api_port),
                "SSH_HOST": local_ip,
                "SSH_PORT": str(self.config.ssh_port),
                "SSH_USER": self.config.ssh_user,
                "SSH_PRIVATE_KEY": str(private_key_path),
                "DOCKER_HOST": self.config.docker_host.format(ip=local_ip),
                "DOCKER_TLS_VERIFY": "1",
                "DOCKER_CERT_PATH": str(self.docker_setup.docker_certs_dir),
                "DOCKER_CONTEXT": docker_context,
                "CADDY_ENDPOINT": self.config.caddy_endpoint,
                "CADDY_DATA_VOLUME": self.config.get_path("caddy", "data"),
                "CADDY_CONFIG_VOLUME": self.config.get_path("caddy", "config"),
                "DB_VOLUME": self.config.get_path("db"),
                "ALLOWED_ORIGIN": allowed_origin,
                "APP_VERSION": self.get_version(),
                "REDIS_URL": self.config.redis_url
            }

            try:
                self.logger.debug("Writing environment files")
                with open(self.env_file, 'w') as f:
                    for key, value in base_env_vars.items():
                        f.write(f"{key}={value}\n")

                api_env_vars = base_env_vars.copy()
                api_env_vars["PORT"] = str(self.config.api_port)
                api_env_file = self.config.get_path("api") / self.config.files.env
                api_env_file.parent.mkdir(parents=True, exist_ok=True)
                with open(api_env_file, 'w') as f:
                    for key, value in api_env_vars.items():
                        f.write(f"{key}={value}\n")
                
                view_env_vars = base_env_vars.copy()
                view_env_vars["PORT"] = str(self.config.next_public_port)
                view_env_file = self.config.get_path("view") / self.config.files.env
                view_env_file.parent.mkdir(parents=True, exist_ok=True)
                with open(view_env_file, 'w') as f:
                    for key, value in view_env_vars.items():
                        f.write(f"{key}={value}\n")
                
                self.logger.debug("Setting file permissions")
                self.env_file.chmod(self.config.get_permission("env"))
                private_key_path.chmod(self.config.get_permission("private_key"))
                public_key_path.chmod(self.config.get_permission("public_key"))
                self.logger.debug("Environment setup completed successfully")
                return base_env_vars
            except IOError as e:
                self.logger.debug(f"Error writing environment files: {e}")
                raise Exception(self.config.errors.file_write_error.format(error=str(e)))
        except ValueError as e:
            self.logger.debug(f"Validation error: {e}")
            raise e
        except Exception as e:
            self.logger.debug(f"Setup error: {e}")
            raise Exception(self.config.errors.setup_error.format(error=str(e)))