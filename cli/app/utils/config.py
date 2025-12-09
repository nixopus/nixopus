import os
import re
import sys
from typing import Any, Dict, Optional

import yaml

from app.utils.message import MISSING_CONFIG_KEY_MESSAGE


def get_config_file_path(default_env: str = "PRODUCTION") -> str:
    """Get the path to the config file based on environment"""
    config_file = "config.dev.yaml" if default_env.upper() == "DEVELOPMENT" else "config.prod.yaml"
    
    if getattr(sys, "frozen", False) and hasattr(sys, "_MEIPASS"):
        return os.path.join(sys._MEIPASS, "helpers", config_file)
    else:
        return os.path.abspath(os.path.join(os.path.dirname(__file__), "../../../helpers", config_file))


def load_config_file(config_file_path: str) -> Dict[str, Any]:
    """Load YAML config file"""
    with open(config_file_path, "r") as f:
        return yaml.safe_load(f) or {}


def get_active_config(user_config_file: Optional[str] = None, default_env: str = "PRODUCTION") -> Dict[str, Any]:
    """Get the active config (user config if provided, else default)"""
    if user_config_file:
        if not os.path.exists(user_config_file):
            raise FileNotFoundError(f"Config file not found: {user_config_file}")
        return load_config_file(user_config_file)
    
    config_file_path = get_config_file_path(default_env)
    return load_config_file(config_file_path)


def get_env(default_env: str = "PRODUCTION") -> str:
    """Get environment from ENV variable or default"""
    return os.environ.get("ENV", default_env)


def is_development(default_env: str = "PRODUCTION") -> bool:
    """Check if current environment is development"""
    return get_env(default_env).upper() == "DEVELOPMENT"


def get_config_value(
    config: Dict[str, Any],
    path: str,
) -> Any:
    """Get config value using dot notation path"""
    keys = path.split(".")
    value = config
    for key in keys:
        if isinstance(value, dict) and key in value:
            value = value[key]
        else:
            raise KeyError(MISSING_CONFIG_KEY_MESSAGE.format(path=path, key=key))
    
    if isinstance(value, str):
        value = expand_env_placeholders(value)
    
    return value


def get_service_env_values(config: Dict[str, Any], service_env_path: str) -> Dict[str, Any]:
    """Get service environment values as a dictionary"""
    env_config = get_config_value(config, service_env_path)
    if not isinstance(env_config, dict):
        raise ValueError(f"Expected dictionary at path '{service_env_path}'")
    return {key: expand_env_placeholders(value) if isinstance(value, str) else value for key, value in env_config.items()}


def load_yaml_config(user_config_file: Optional[str] = None, default_env: str = "PRODUCTION") -> Dict[str, Any]:
    """Return the active config dict (for backward compatibility)"""
    return get_active_config(user_config_file, default_env)


def get_yaml_value(
    config: Dict[str, Any],
    path: str,
) -> Any:
    """Alias for get_config_value() for backward compatibility"""
    return get_config_value(config, path)


def unflatten_config(flattened_config: dict) -> dict:
    """Convert flattened config back to nested structure"""
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
    """Expand environment placeholders in the form ${ENV_VAR:-default}"""
    # Supports nested expansions like ${VAR1:-${VAR2:-default}}
    pattern = re.compile(r"\$\{([A-Za-z_][A-Za-z0-9_]*)(:-([^}]*))?}")
    max_iterations = 10  # Prevent infinite loops
    iteration = 0

    def replacer(match):
        var_name = match.group(1)
        default = match.group(3) if match.group(2) else ""
        return os.environ.get(var_name, default)

    # Keep expanding until no more placeholders are found or max iterations reached
    while pattern.search(value) and iteration < max_iterations:
        value = pattern.sub(replacer, value)
        iteration += 1

    return value


# Config path constants
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
CADDY_ADMIN_PORT = "services.caddy.env.CADDY_ADMIN_PORT"
CADDY_HTTP_PORT = "services.caddy.env.CADDY_HTTP_PORT"
CADDY_HTTPS_PORT = "services.caddy.env.CADDY_HTTPS_PORT"
DOCKER_PORT = "services.api.env.DOCKER_PORT"
SUPERTOKENS_API_PORT = "services.api.env.SUPERTOKENS_API_PORT"
