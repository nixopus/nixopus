from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Callable, Dict, List, Optional


class EnvVarType(Enum):
    PORT = "port"
    URL = "url"
    DOMAIN = "domain"
    PATH = "path"
    STRING = "string"
    BOOLEAN = "boolean"


@dataclass
class EnvVarDefinition:
    key: str
    var_type: EnvVarType
    description: str
    required: bool = False
    default: Optional[Any] = None
    transform: Optional[Callable[[Any], str]] = None
    depends_on: List[str] = field(default_factory=list)
    computed: bool = False


@dataclass
class PortMapping:
    env_key: str
    param_name: str
    description: str
    aliases: List[str] = field(default_factory=list)


@dataclass
class ConfigPath:
    key: str
    description: str
    resolver: Optional[Callable] = None
    depends_on: List[str] = field(default_factory=list)


PORT_MAPPINGS = [
    PortMapping("API_PORT", "api_port", "API service port"),
    PortMapping("VIEW_PORT", "view_port", "Frontend view port"),
    PortMapping("NEXT_PUBLIC_PORT", "view_port", "Next.js public port", aliases=["VIEW_PORT"]),
    PortMapping("DB_PORT", "db_port", "Database port"),
    PortMapping("REDIS_PORT", "redis_port", "Redis cache port"),
    PortMapping("CADDY_ADMIN_PORT", "caddy_admin_port", "Caddy admin API port"),
    PortMapping("CADDY_HTTP_PORT", "caddy_http_port", "Caddy HTTP port"),
    PortMapping("CADDY_HTTPS_PORT", "caddy_https_port", "Caddy HTTPS port"),
    PortMapping("SUPERTOKENS_PORT", "supertokens_port", "SuperTokens service port"),
]


ENV_VAR_KEYS = {
    "ALLOWED_ORIGIN": EnvVarDefinition(
        key="ALLOWED_ORIGIN",
        var_type=EnvVarType.URL,
        description="CORS allowed origin URL",
        computed=True,
        depends_on=["VIEW_DOMAIN", "VIEW_PORT"],
    ),
    "SSH_HOST": EnvVarDefinition(
        key="SSH_HOST",
        var_type=EnvVarType.STRING,
        description="SSH host IP address",
        computed=True,
        depends_on=["HOST_IP"],
    ),
    "SSH_PRIVATE_KEY": EnvVarDefinition(
        key="SSH_PRIVATE_KEY",
        var_type=EnvVarType.PATH,
        description="Path to SSH private key file",
        computed=True,
        depends_on=["SSH_KEY_PATH"],
    ),
    "WEBSOCKET_URL": EnvVarDefinition(
        key="WEBSOCKET_URL",
        var_type=EnvVarType.URL,
        description="WebSocket connection URL",
        computed=True,
        depends_on=["API_DOMAIN", "API_PORT"],
    ),
    "API_URL": EnvVarDefinition(
        key="API_URL",
        var_type=EnvVarType.URL,
        description="API base URL",
        computed=True,
        depends_on=["API_DOMAIN", "API_PORT"],
    ),
    "WEBHOOK_URL": EnvVarDefinition(
        key="WEBHOOK_URL",
        var_type=EnvVarType.URL,
        description="Webhook callback URL",
        computed=True,
        depends_on=["API_DOMAIN", "API_PORT"],
    ),
    "VIEW_DOMAIN": EnvVarDefinition(
        key="VIEW_DOMAIN",
        var_type=EnvVarType.URL,
        description="Full view/frontend domain URL",
        computed=True,
        depends_on=["VIEW_DOMAIN", "VIEW_PORT"],
    ),
    "SUPERTOKENS_API_KEY": EnvVarDefinition(
        key="SUPERTOKENS_API_KEY",
        var_type=EnvVarType.STRING,
        description="SuperTokens API authentication key",
        default="NixopusSuperTokensAPIKey",
        computed=False,
    ),
    "SUPERTOKENS_API_DOMAIN": EnvVarDefinition(
        key="SUPERTOKENS_API_DOMAIN",
        var_type=EnvVarType.URL,
        description="SuperTokens API domain URL",
        computed=True,
        depends_on=["API_DOMAIN", "API_PORT"],
    ),
    "SUPERTOKENS_WEBSITE_DOMAIN": EnvVarDefinition(
        key="SUPERTOKENS_WEBSITE_DOMAIN",
        var_type=EnvVarType.URL,
        description="SuperTokens website domain URL",
        computed=True,
        depends_on=["VIEW_DOMAIN", "VIEW_PORT"],
    ),
    "SUPERTOKENS_CONNECTION_URI": EnvVarDefinition(
        key="SUPERTOKENS_CONNECTION_URI",
        var_type=EnvVarType.URL,
        description="SuperTokens connection URI",
        computed=True,
        depends_on=["API_DOMAIN", "API_PORT", "SUPERTOKENS_PORT", "HOST_IP"],
    ),
}

for port_def in PORT_MAPPINGS:
    ENV_VAR_KEYS[port_def.env_key] = EnvVarDefinition(
        key=port_def.env_key,
        var_type=EnvVarType.PORT,
        description=port_def.description,
        required=False,
        transform=lambda x: str(x) if x is not None else None,
    )


CONFIG_PATHS = {
    "full_source_path": ConfigPath(
        "full_source_path",
        "Full path to source code directory",
        depends_on=["NIXOPUS_CONFIG_DIR", "DEFAULT_PATH"],
    ),
    "ssh_key_path": ConfigPath(
        "ssh_key_path",
        "Path to SSH private key",
        depends_on=["NIXOPUS_CONFIG_DIR", "SSH_FILE_PATH"],
    ),
    "compose_file_path": ConfigPath(
        "compose_file_path",
        "Path to docker-compose file",
        depends_on=["NIXOPUS_CONFIG_DIR", "DEFAULT_COMPOSE_FILE", "repo", "branch"],
    ),
}


def build_port_env_vars(**params) -> Dict[str, str]:
    env_vars = {}
    for mapping in PORT_MAPPINGS:
        value = params.get(mapping.param_name)
        if value is not None:
            env_vars[mapping.env_key] = str(value)
    return env_vars


def get_port_mapping(param_name: str) -> Optional[PortMapping]:
    for mapping in PORT_MAPPINGS:
        if mapping.param_name == param_name:
            return mapping
    return None


def get_all_port_mappings() -> List[PortMapping]:
    return PORT_MAPPINGS.copy()


def get_env_var_definition(key: str) -> Optional[EnvVarDefinition]:
    return ENV_VAR_KEYS.get(key)


def get_all_env_vars() -> Dict[str, EnvVarDefinition]:
    return ENV_VAR_KEYS.copy()


def get_env_vars_by_type(var_type: EnvVarType) -> Dict[str, EnvVarDefinition]:
    return {k: v for k, v in ENV_VAR_KEYS.items() if v.var_type == var_type}


def get_computed_env_vars() -> Dict[str, EnvVarDefinition]:
    return {k: v for k, v in ENV_VAR_KEYS.items() if v.computed}


def get_port_env_vars() -> Dict[str, EnvVarDefinition]:
    return get_env_vars_by_type(EnvVarType.PORT)


def get_config_path(key: str) -> Optional[ConfigPath]:
    return CONFIG_PATHS.get(key)


def get_all_config_paths() -> Dict[str, ConfigPath]:
    return CONFIG_PATHS.copy()

