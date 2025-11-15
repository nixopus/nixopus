import os
import re
import sys

import yaml

from app.utils.message import MISSING_CONFIG_KEY_MESSAGE


class Config:
    def __init__(self, default_env="PRODUCTION"):
        self.default_env = default_env
        self._yaml_config = None
        self._user_config_file = None
        self._cache = {}

        config_file = "config.dev.yaml" if default_env.upper() == "DEVELOPMENT" else "config.prod.yaml"

        if getattr(sys, "frozen", False) and hasattr(sys, "_MEIPASS"):
            self._yaml_path = os.path.join(sys._MEIPASS, "helpers", config_file)
        else:
            self._yaml_path = os.path.abspath(os.path.join(os.path.dirname(__file__), "../../../helpers", config_file))

    def get_env(self):
        return os.environ.get("ENV", self.default_env)

    def is_development(self):
        return self.get_env().upper() == "DEVELOPMENT"

    def load_user_config(self, config_file: str):
        """Set user config file to replace default config."""
        if config_file and not os.path.exists(config_file):
            raise FileNotFoundError(f"Config file not found: {config_file}")
        self._user_config_file = config_file
        self._yaml_config = None
        self._cache = {}

    def _get_active_config(self):
        """Get the active config (user config if provided, else default)."""
        if self._user_config_file:
            if self._yaml_config is None:
                with open(self._user_config_file, "r") as f:
                    self._yaml_config = yaml.safe_load(f)
            return self._yaml_config

        if self._yaml_config is None:
            with open(self._yaml_path, "r") as f:
                self._yaml_config = yaml.safe_load(f)
        return self._yaml_config

    def get(self, path: str):
        """Get config value using dot notation path."""
        if path in self._cache:
            return self._cache[path]

        config = self._get_active_config()
        keys = path.split(".")
        for key in keys:
            if isinstance(config, dict) and key in config:
                config = config[key]
            else:
                raise KeyError(MISSING_CONFIG_KEY_MESSAGE.format(path=path, key=key))

        if isinstance(config, str):
            config = expand_env_placeholders(config)

        self._cache[path] = config
        return config

    def get_service_env_values(self, service_env_path: str):
        """Get service environment values as a dictionary."""
        env_config = self.get(service_env_path)
        if not isinstance(env_config, dict):
            raise ValueError(f"Expected dictionary at path '{service_env_path}'")
        return {key: expand_env_placeholders(value) if isinstance(value, str) else value for key, value in env_config.items()}

    def load_yaml_config(self):
        """Return the active config dict (for backward compatibility)."""
        return self._get_active_config()

    def get_yaml_value(self, path: str):
        """Alias for get() for backward compatibility."""
        return self.get(path)

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
