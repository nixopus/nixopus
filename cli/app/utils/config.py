import os
import re
import sys

import yaml

from app.utils.message import MISSING_CONFIG_KEY_MESSAGE


class Config:
    def __init__(self, default_env="PRODUCTION"):
        self.default_env = default_env
        self._yaml_config = None
        self._cache = {}

        # Determine config file based on environment
        config_file = "config.dev.yaml" if default_env.upper() == "DEVELOPMENT" else "config.prod.yaml"

        # Check if running as PyInstaller bundle
        if getattr(sys, "frozen", False) and hasattr(sys, "_MEIPASS"):
            # Running as PyInstaller bundle
            self._yaml_path = os.path.join(sys._MEIPASS, "helpers", config_file)
        else:
            # Running as normal Python script
            self._yaml_path = os.path.abspath(os.path.join(os.path.dirname(__file__), "../../../helpers", config_file))

    def get_env(self):
        return os.environ.get("ENV", self.default_env)

    def is_development(self):
        return self.get_env().upper() == "DEVELOPMENT"

    def load_yaml_config(self):
        if self._yaml_config is None:
            with open(self._yaml_path, "r") as f:
                self._yaml_config = yaml.safe_load(f)
        return self._yaml_config

    def get_yaml_value(self, path: str):
        config = self.load_yaml_config()
        keys = path.split(".")
        for key in keys:
            if isinstance(config, dict) and key in config:
                config = config[key]
            else:
                raise KeyError(MISSING_CONFIG_KEY_MESSAGE.format(path=path, key=key))
        if isinstance(config, str):
            config = expand_env_placeholders(config)
        return config

    def get_service_env_values(self, service_env_path: str):
        config = self.get_yaml_value(service_env_path)
        return {key: expand_env_placeholders(value) for key, value in config.items()}

    def load_user_config(self, config_file: str):
        """Load and parse user config file, returning flattened config dict."""
        if not config_file:
            return {}

        if not os.path.exists(config_file):
            raise FileNotFoundError(f"Config file not found: {config_file}")

        with open(config_file, "r") as f:
            user_config = yaml.safe_load(f)

        flattened = {}
        self.flatten_config(user_config, flattened)
        return flattened

    def flatten_config(self, config: dict, result: dict, prefix: str = ""):
        """Flatten nested config dict into dot notation keys."""
        for key, value in config.items():
            new_key = f"{prefix}.{key}" if prefix else key
            if isinstance(value, dict):
                self.flatten_config(value, result, new_key)
            else:
                result[new_key] = value

    def unflatten_config(self, flattened_config: dict) -> dict:
        """Convert flattened config back to nested structure."""
        nested = {}
        for key, value in flattened_config.items():
            keys = key.split(".")
            current = nested
            for k in keys[:-1]:
                if k not in current:
                    current[k] = {}
                current = current[k]
            current[keys[-1]] = value
        return nested

    def get_config_value(self, key: str, user_config: dict, defaults: dict):
        """Get config value from user config with fallback to defaults and caching."""
        if key in self._cache:
            return self._cache[key]

        # Key mappings for user config lookup
        key_mappings = {
            "proxy_port": "services.caddy.env.PROXY_PORT",
            "repo_url": "clone.repo",
            "branch_name": "clone.branch",
            "source_path": "clone.source-path",
            "config_dir": "nixopus-config-dir",
            "api_env_file_path": "services.api.env.API_ENV_FILE",
            "view_env_file_path": "services.view.env.VIEW_ENV_FILE",
            "compose_file": "compose-file-path",
            "required_ports": "ports",
        }

        config_path = key_mappings.get(key, key)
        user_value = user_config.get(config_path)
        value = user_value if user_value is not None else defaults.get(key)

        if value is None and key not in ["ssh_passphrase"]:
            raise ValueError(f"Configuration key '{key}' has no default value")

        self._cache[key] = value
        return value


def expand_env_placeholders(value: str) -> str:
    # Expand environment placeholders in the form ${ENV_VAR:-default}
    pattern = re.compile(r"\$\{([A-Za-z_][A-Za-z0-9_]*)(:-([^}]*))?}")

    def replacer(match):
        var_name = match.group(1)
        default = match.group(3) if match.group(2) else ""
        return os.environ.get(var_name, default)

    return pattern.sub(replacer, value)


VIEW_ENV_FILE = "services.view.env.VIEW_ENV_FILE"
API_ENV_FILE = "services.api.env.API_ENV_FILE"
DEFAULT_REPO = "clone.repo"
DEFAULT_BRANCH = "clone.branch"
DEFAULT_PATH = "clone.source-path"
DEFAULT_COMPOSE_FILE = "compose-file-path"
NIXOPUS_CONFIG_DIR = "nixopus-config-dir"
PROXY_PORT = "services.caddy.env.PROXY_PORT"
CADDY_BASE_URL = "services.caddy.env.BASE_URL"
CONFIG_ENDPOINT = "services.caddy.env.CONFIG_ENDPOINT"
LOAD_ENDPOINT = "services.caddy.env.LOAD_ENDPOINT"
STOP_ENDPOINT = "services.caddy.env.STOP_ENDPOINT"
DEPS = "deps"
PORTS = "ports"
API_SERVICE = "services.api"
VIEW_SERVICE = "services.view"
SSH_KEY_SIZE = "ssh_key_size"
SSH_KEY_TYPE = "ssh_key_type"
SSH_FILE_PATH = "ssh_file_path"
VIEW_PORT = "services.view.env.NEXT_PUBLIC_PORT"
API_PORT = "services.api.env.PORT"
CADDY_CONFIG_VOLUME = "services.caddy.env.CADDY_CONFIG_VOLUME"
DOCKER_PORT = "services.api.env.DOCKER_PORT"
SUPERTOKENS_API_PORT = "services.api.env.SUPERTOKENS_API_PORT"
