from dataclasses import dataclass
from typing import List
from app.utils.config import Config

_config = Config()

DEFAULT_CONTAINER_NAME = "nixopus-lxc"
DEFAULT_LXC_IMAGE = "debian/11"
DEFAULT_TIMEOUT = 600
DEFAULT_INSTALL_SCRIPT_URL = "https://raw.githubusercontent.com/nixopus/nixopus/main/install.sh"

CONTAINER_READY_WAIT_TIME = 5
CONTAINER_RESTART_WAIT_TIME = 10

SECURITY_SETTINGS = {
    "security.privileged": "true",
    "security.nesting": "true",
}

PROXY_DEVICE_CONFIGS = [
    ("nixopus-view", "80", "3000", "Nixopus View Interface"),
    ("nixopus-api", "8443", "8443", "Nixopus API"),
    ("nixopus-proxy", "443", "2443", "Nixopus Proxy (Caddy)"),
    ("nixopus-docker", "2376", "2376", "Docker API"),
]

def _get_config_value_safe(path: str, default: str) -> str:
    try:
        return str(_config.get_yaml_value(path))
    except Exception:
        return default

VIEW_PORT = _get_config_value_safe("services.view.env.NEXT_PUBLIC_PORT", "3000")
API_PORT = _get_config_value_safe("services.api.env.PORT", "8443")
PROXY_PORT = _get_config_value_safe("services.caddy.env.PROXY_PORT", "2443")
DOCKER_PORT = _get_config_value_safe("services.api.env.DOCKER_PORT", "2376")

PROXY_DEVICE_CONFIGS = [
    ("nixopus-view", "80", VIEW_PORT, "Nixopus View Interface"),
    ("nixopus-api", "8443", API_PORT, "Nixopus API"),
    ("nixopus-proxy", "443", PROXY_PORT, "Nixopus Proxy (Caddy)"),
    ("nixopus-docker", "2376", DOCKER_PORT, "Docker API"),
]

REQUIRED_PACKAGES = [
    "curl",
    "wget", 
    "git",
    "python3",
    "python3-pip",
    "docker.io",
    "docker-compose",
    "openssh-server",
    "sudo",
    "systemctl"
]

SERVICES_TO_START = [
    ("docker", "Docker container runtime"),
    ("ssh", "SSH server for remote access"),
]

SERVICES_TO_ENABLE = [
    ("docker", "Docker container runtime"),
    ("ssh", "SSH server for remote access"),
]

IP_DETECTION_CONFIG = {
    "lxc_list_format": "csv",
    "lxc_list_columns": "4",  # IP column
    "ip_validation_parts": 4,
    "ip_part_min": 0,
    "ip_part_max": 255,
}

OPERATION_TIMEOUTS = {
    "container_creation": 300,
    "dependency_installation": 600,
    "nixopus_installation": 900,
    "service_start": 60,
    "networking_setup": 120,
}

COMMAND_TEMPLATES = {
    "lxc_info": ["lxc", "info", "{name}"],
    "lxc_launch": ["lxc", "launch", "{image}", "{name}"],
    "lxc_delete": ["lxc", "delete", "{name}", "--force"],
    "lxc_config_set": ["lxc", "config", "set", "{name}", "{key}", "{value}"],
    "lxc_restart": ["lxc", "restart", "{name}"],
    "lxc_exec": ["lxc", "exec", "{name}", "--"],
    "lxc_list": ["lxc", "list"],
    "lxc_list_csv": ["lxc", "list", "{name}", "--format", "csv", "--columns", "4"],
    "lxc_proxy_add": ["lxc", "config", "device", "add", "{name}", "{device_name}", "proxy", "listen=tcp:0.0.0.0:{host_port}", "connect=tcp:127.0.0.1:{container_port}"],
    
    "apt_update": ["apt", "update"],
    "apt_install": ["apt", "install", "-y"],
    "systemctl_start": ["systemctl", "start", "{service}"],
    "systemctl_enable": ["systemctl", "enable", "{service}"],
    "curl_install": ["bash", "-c", "curl -fsSL {url} | bash"],
}

MAX_RETRIES = {
    "container_creation": 3,
    "service_start": 3,
    "dependency_install": 2,
}

RETRY_DELAYS = {
    "container_creation": 5,
    "service_start": 3,
    "dependency_install": 10,
}

@dataclass
class LXCInstallConfig:
    name: str = DEFAULT_CONTAINER_NAME
    image: str = DEFAULT_LXC_IMAGE
    force: bool = False
    verbose: bool = False
    timeout: int = DEFAULT_TIMEOUT
    api_domain: str = None
    view_domain: str = None
    install_script_url: str = DEFAULT_INSTALL_SCRIPT_URL
    setup_proxy: bool = True
    dry_run: bool = False

@dataclass
class ProxyDeviceConfig:
    device_name: str
    host_port: str
    container_port: str
    description: str = ""

def get_proxy_devices() -> List[ProxyDeviceConfig]:
    return [
        ProxyDeviceConfig(
            device_name=device_name,
            host_port=host_port,
            container_port=container_port,
            description=description
        )
        for device_name, host_port, container_port, description in PROXY_DEVICE_CONFIGS
    ]

def get_formatted_proxy_ports() -> str:
    return ", ".join([
        f"{host_port}->{container_port}"
        for _, host_port, container_port, _ in PROXY_DEVICE_CONFIGS
    ])

def get_command_template(template_name: str, **kwargs) -> List[str]:
    if template_name not in COMMAND_TEMPLATES:
        raise ValueError(f"Unknown command template: {template_name}")
    
    template = COMMAND_TEMPLATES[template_name]
    return [part.format(**kwargs) for part in template]

def get_packages_string() -> str:
    return ", ".join(REQUIRED_PACKAGES)

__all__ = [
    'DEFAULT_CONTAINER_NAME',
    'DEFAULT_LXC_IMAGE', 
    'DEFAULT_TIMEOUT',
    'DEFAULT_INSTALL_SCRIPT_URL',
    'CONTAINER_READY_WAIT_TIME',
    'CONTAINER_RESTART_WAIT_TIME',
    'SECURITY_SETTINGS',
    'PROXY_DEVICE_CONFIGS',
    'REQUIRED_PACKAGES',
    'SERVICES_TO_START',
    'SERVICES_TO_ENABLE',
    'IP_DETECTION_CONFIG',
    'OPERATION_TIMEOUTS',
    'COMMAND_TEMPLATES',
    'MAX_RETRIES',
    'RETRY_DELAYS',
    'LXCInstallConfig',
    'ProxyDeviceConfig',
    'get_proxy_devices',
    'get_formatted_proxy_ports',
    'get_command_template',
    'get_packages_string',
    'VIEW_PORT',
    'API_PORT',
    'PROXY_PORT',
    'DOCKER_PORT',
]
