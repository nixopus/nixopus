import os
import ipaddress
import json
import shutil
import socket
from typing import Dict, Optional
from urllib.parse import urlparse, parse_qs, unquote

from app.utils.config import (
    CADDY_ADMIN_PORT,
    CADDY_CONFIG_VOLUME,
    DEFAULT_BRANCH,
    DEFAULT_COMPOSE_FILE,
    DEFAULT_PATH,
    DEFAULT_REPO,
    NIXOPUS_CONFIG_DIR,
    SSH_FILE_PATH,
    get_active_config,
    get_config_value,
)
from app.utils.directory_manager import create_directory
from app.utils.file_manager import set_permissions
from app.utils.host_information import get_public_ip
from app.utils.protocols import LoggerProtocol

from .config_schema import ENV_VAR_KEYS


def resolve_hostname_to_ipv4(hostname: str) -> str:
    try:
        ip = ipaddress.ip_address(hostname)
        if isinstance(ip, ipaddress.IPv4Address):
            return hostname
        elif isinstance(ip, ipaddress.IPv6Address):
            try:
                addr_info = socket.getaddrinfo(hostname, None, socket.AF_INET, socket.SOCK_STREAM)
                if addr_info:
                    return addr_info[0][4][0]
            except (socket.gaierror, OSError):
                pass
            return hostname
    except ValueError:
        pass
    
    try:
        addr_info = socket.getaddrinfo(hostname, None, socket.AF_INET, socket.SOCK_STREAM)
        if addr_info:
            ipv4_address = addr_info[0][4][0]
            return ipv4_address
    except (socket.gaierror, OSError):
        pass
    
    return hostname


def parse_db_url(db_url: str) -> Dict[str, str]:
    parsed = urlparse(db_url)
    
    username = unquote(parsed.username or "")
    password = unquote(parsed.password or "")
    hostname = parsed.hostname or ""
    port = parsed.port or 5432
    database = unquote(parsed.path.lstrip("/") or "")
    
    query_params = parse_qs(parsed.query)
    ssl_mode = query_params.get("sslmode", ["disable"])[0] if query_params.get("sslmode") else "disable"
    
    resolved_host = resolve_hostname_to_ipv4(hostname)
    
    return {
        "HOST_NAME": resolved_host,
        "DB_PORT": str(port),
        "USERNAME": username,
        "PASSWORD": password,
        "DB_NAME": database,
        "SSL_MODE": ssl_mode,
        "POSTGRESQL_CONNECTION_URI": db_url,
    }


def is_custom_repo_or_branch(repo: Optional[str], branch: Optional[str]) -> bool:
    temp_config = get_active_config()
    default_repo = get_config_value(temp_config, DEFAULT_REPO)
    default_branch = get_config_value(temp_config, DEFAULT_BRANCH)

    repo_differs = repo is not None and repo != default_repo
    branch_differs = branch is not None and branch != default_branch

    return repo_differs or branch_differs


def get_host_ip_or_default(host_ip: Optional[str]) -> str:
    if host_ip:
        return host_ip
    return get_public_ip()


def get_full_source_path(config: dict) -> str:
    return os.path.join(get_config_value(config, NIXOPUS_CONFIG_DIR), get_config_value(config, DEFAULT_PATH))


def get_ssh_key_path(config: dict) -> str:
    return os.path.join(get_config_value(config, NIXOPUS_CONFIG_DIR), get_config_value(config, SSH_FILE_PATH))


def get_compose_file_path(config: dict, use_staging: bool) -> str:
    compose_path = os.path.join(get_config_value(config, NIXOPUS_CONFIG_DIR), get_config_value(config, DEFAULT_COMPOSE_FILE))
    if use_staging:
        return compose_path.replace("docker-compose.yml", "docker-compose-staging.yml")
    return compose_path


def get_proxy_port(config: dict, caddy_admin_port: Optional[int]) -> int:
    if caddy_admin_port is not None:
        return caddy_admin_port
    try:
        return int(get_config_value(config, CADDY_ADMIN_PORT))
    except (KeyError, ValueError):
        return 2019


def get_supertokens_connection_uri(protocol: str, api_host: str, supertokens_api_port: int, host_ip: str) -> str:
    protocol = protocol.replace("https", "http")
    
    host_without_port = api_host
    
    if api_host.startswith(("http://", "https://")):
        parsed = urlparse(api_host)
        host_without_port = parsed.hostname or api_host
    elif ":" in api_host:
        parts = api_host.rsplit(":", 1)
        if len(parts) == 2 and parts[1].isdigit():
            host_without_port = parts[0]
    
    try:
        ipaddress.ip_address(host_without_port)
        return f"{protocol}://{host_ip}:{supertokens_api_port}"
    except ValueError:
        return f"{protocol}://{host_without_port}:{supertokens_api_port}"


def build_env_variable_map(
    host_ip: str,
    api_domain: Optional[str],
    view_domain: Optional[str],
    api_port: str,
    view_port: str,
    supertokens_api_port: str,
    ssh_key_path: str,
) -> Dict[str, str]:
    secure = api_domain is not None and view_domain is not None
    api_host = api_domain if secure else f"{host_ip}:{api_port}"
    view_host = view_domain if secure else f"{host_ip}:{view_port}"
    protocol = "https" if secure else "http"
    ws_protocol = "wss" if secure else "ws"
    
    env_map = {
        "ALLOWED_ORIGIN": f"{protocol}://{view_host}",
        "SSH_HOST": host_ip,
        "SSH_PRIVATE_KEY": ssh_key_path,
        "WEBSOCKET_URL": f"{ws_protocol}://{api_host}/ws",
        "API_URL": f"{protocol}://{api_host}/api",
        "WEBHOOK_URL": f"{protocol}://{api_host}/api/v1/webhook",
        "VIEW_DOMAIN": f"{protocol}://{view_host}",
        "SUPERTOKENS_API_KEY": ENV_VAR_KEYS["SUPERTOKENS_API_KEY"].default,
        "SUPERTOKENS_API_DOMAIN": f"{protocol}://{api_host}/api",
        "SUPERTOKENS_WEBSITE_DOMAIN": f"{protocol}://{view_host}",
        "SUPERTOKENS_CONNECTION_URI": get_supertokens_connection_uri(
            protocol, api_host, supertokens_api_port, host_ip
        ),
    }
    
    return {k: v for k, v in env_map.items() if k in ENV_VAR_KEYS}


def update_environment_variables(
    env_values: dict,
    host_ip: str,
    api_domain: Optional[str],
    view_domain: Optional[str],
    api_port: str,
    view_port: str,
    supertokens_api_port: str,
    ssh_key_path: str,
    external_db_url: Optional[str] = None,
) -> dict:
    updated_env = env_values.copy()
    
    if external_db_url:
        db_config = parse_db_url(external_db_url)
        updated_env.update(db_config)
    
    env_map = build_env_variable_map(
        host_ip=host_ip,
        api_domain=api_domain,
        view_domain=view_domain,
        api_port=api_port,
        view_port=view_port,
        supertokens_api_port=supertokens_api_port,
        ssh_key_path=ssh_key_path,
    )

    for key, value in env_map.items():
        if key in updated_env:
            updated_env[key] = value

    return updated_env


def setup_proxy_config(
    full_source_path: str,
    host_ip: str,
    view_domain: Optional[str],
    api_domain: Optional[str],
    view_port: str,
    api_port: str,
) -> str:
    caddy_json_template = os.path.join(full_source_path, "helpers", "caddy.json")
    
    with open(caddy_json_template, "r") as f:
        config_str = f.read()

    view_domain_or_ip = view_domain if view_domain is not None else host_ip
    api_domain_or_ip = api_domain if api_domain is not None else host_ip

    config_str = config_str.replace("{env.APP_DOMAIN}", view_domain_or_ip)
    config_str = config_str.replace("{env.API_DOMAIN}", api_domain_or_ip)

    app_reverse_proxy_url = f"{host_ip}:{view_port}"
    api_reverse_proxy_url = f"{host_ip}:{api_port}"
    config_str = config_str.replace("{env.APP_REVERSE_PROXY_URL}", app_reverse_proxy_url)
    config_str = config_str.replace("{env.API_REVERSE_PROXY_URL}", api_reverse_proxy_url)

    caddy_config = json.loads(config_str)
    with open(caddy_json_template, "w") as f:
        json.dump(caddy_config, f, indent=2)
    
    return caddy_json_template


def copy_caddyfile_to_target(full_source_path: str, config: dict, logger: Optional[LoggerProtocol] = None):
    try:
        source_caddyfile = os.path.join(full_source_path, "helpers", "Caddyfile")
        target_dir = get_config_value(config, CADDY_CONFIG_VOLUME)
        target_caddyfile = os.path.join(target_dir, "Caddyfile")
        create_directory(target_dir, logger=logger)
        if os.path.exists(source_caddyfile):
            shutil.copy2(source_caddyfile, target_caddyfile)
            set_permissions(target_caddyfile, 0o644, logger=logger)
            if logger:
                logger.debug(f"Copied Caddyfile from {source_caddyfile} to {target_caddyfile}")
        else:
            if logger:
                logger.warning(f"Source Caddyfile not found at {source_caddyfile}")
    except Exception as e:
        if logger:
            logger.error(f"Failed to copy Caddyfile: {str(e)}")


def get_access_url(view_domain: Optional[str], api_domain: Optional[str], host_ip: str, view_port: str) -> str:
    if view_domain:
        return f"https://{view_domain}"
    elif api_domain:
        return f"https://{api_domain}"
    else:
        return f"http://{host_ip}:{view_port}"
