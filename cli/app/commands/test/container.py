import subprocess
from typing import Optional, Tuple

from .messages import (
    cleaning_up,
    configuring_proxy,
    container_created,
    container_started,
    creating_container,
    dry_run_would_execute,
    installing_dependencies,
    dependencies_installed,
    proxy_configured,
    setting_up_networking,
)
from .types import TestParams

def create_container(params: TestParams) -> Tuple[bool, Optional[str]]:
    if params.logger:
        params.logger.info(creating_container.format(name=params.container_name))

    if params.dry_run:
        if params.logger:
            params.logger.info(
                dry_run_would_execute.format(
                    action=f"lxc launch {params.image} {params.container_name}"
                )
            )
        return True, None

    try:
        image = params.image or params.distro
        if params.logger:
            params.logger.debug(f"Using image: {image}")
            params.logger.debug(f"Container name: {params.container_name}")
        
        cmd = ["lxc", "launch", image, params.container_name]
        if params.logger:
            params.logger.debug(f"Executing: {' '.join(cmd)}")
        
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        
        if params.logger:
            params.logger.success(container_created.format(name=params.container_name))
            params.logger.debug("Container launch completed successfully")
        
        return True, None
    except subprocess.CalledProcessError as e:
        error_msg = e.stderr if e.stderr else str(e)
        if params.logger:
            params.logger.debug(f"Container creation failed: {error_msg}")
        return False, f"Failed to create container: {error_msg}"

def _setup_dns_for_distro(container_name: str, pkg_manager: str) -> str:
    """Generate distribution-specific DNS setup script."""
    dns_servers = "nameserver 8.8.8.8\\nnameserver 8.8.4.4"
    
    stop_network_manager = pkg_manager in ("dnf", "yum") or pkg_manager == "unknown"
    stop_systemd_resolved = pkg_manager != "apk"
    double_check_symlink = pkg_manager in ("dnf", "yum")
    
    parts = [
        "[ -L /etc/resolv.conf ] && rm -f /etc/resolv.conf || true; "
    ]
    
    if stop_network_manager:
        parts.append("systemctl stop NetworkManager 2>/dev/null || true; ")
    
    if stop_systemd_resolved:
        parts.append("systemctl stop systemd-resolved 2>/dev/null || true; ")
    
    parts.append(f"printf '{dns_servers}\\n' > /etc/resolv.conf; ")
    
    if double_check_symlink:
        parts.append(
            "[ -L /etc/resolv.conf ] && rm -f /etc/resolv.conf && "
            f"printf '{dns_servers}\\n' > /etc/resolv.conf || true; "
        )
    
    parts.append(
        "for i in 1 2 3; do "
        "if getent hosts google.com >/dev/null 2>&1 || nslookup google.com >/dev/null 2>&1; "
        "then exit 0; fi; sleep 1; done; exit 1"
    )
    
    return "".join(parts)


def setup_networking(params: TestParams) -> Tuple[bool, Optional[str]]:
    if params.logger:
        params.logger.info(setting_up_networking.format(name=params.container_name))

    if params.dry_run:
        if params.logger:
            params.logger.info(
                dry_run_would_execute.format(
                    action=f"Configure proxy ports for {params.container_name}"
                )
            )
        return True, None

    try:
        if params.logger:
            params.logger.debug("Configuring container security settings...")
        
        subprocess.run(
            ["lxc", "config", "set", params.container_name, "security.privileged", "true"],
            check=True,
            capture_output=True,
        )
        subprocess.run(
            ["lxc", "config", "set", params.container_name, "security.nesting", "true"],
            check=True,
            capture_output=True,
        )
        
        if params.logger:
            params.logger.debug("Detecting distribution for DNS configuration...")
        
        success, pkg_manager, error = _detect_package_manager(params.container_name)
        if not success:
            if params.logger:
                params.logger.debug(f"Could not detect package manager: {error}, using fallback DNS setup")
            pkg_manager = "unknown"
        else:
            if params.logger:
                params.logger.debug(f"Detected package manager: {pkg_manager}")
        
        if params.logger:
            params.logger.debug("Configuring DNS settings...")
        
        dns_setup_script = _setup_dns_for_distro(params.container_name, pkg_manager)
        
        subprocess.run(
            [
                "lxc",
                "exec",
                params.container_name,
                "--",
                "sh",
                "-c",
                dns_setup_script
            ],
            check=True,
            capture_output=True,
        )

        if params.logger:
            params.logger.success(container_started.format(name=params.container_name))
            params.logger.debug("Container networking setup completed")
        
        return True, None
    except subprocess.CalledProcessError as e:
        error_msg = e.stderr if e.stderr else str(e)
        if params.logger:
            params.logger.debug(f"Networking setup failed: {error_msg}")
        return False, f"Failed to setup networking: {error_msg}"


def _detect_package_manager(container_name: str) -> Tuple[bool, Optional[str], Optional[str]]:
    detection_script = (
        "if command -v apt-get >/dev/null 2>&1; then echo apt-get; "
        "elif command -v dnf >/dev/null 2>&1; then echo dnf; "
        "elif command -v yum >/dev/null 2>&1; then echo yum; "
        "elif command -v apk >/dev/null 2>&1; then echo apk; "
        "elif command -v pacman >/dev/null 2>&1; then echo pacman; "
        "else echo unknown; fi"
    )
    
    try:
        result = subprocess.run(
            ["lxc", "exec", container_name, "--", "sh", "-c", detection_script],
            capture_output=True,
            text=True,
            check=True,
            timeout=10,
        )
        pkg_manager = result.stdout.strip()
        
        if pkg_manager == "unknown":
            return False, None, "Could not detect package manager"
        
        return True, pkg_manager, None
    except subprocess.CalledProcessError as e:
        error_msg = e.stderr if e.stderr else str(e)
        return False, None, f"Failed to detect package manager: {error_msg}"
    except subprocess.TimeoutExpired:
        return False, None, "Timeout detecting package manager"
    except FileNotFoundError:
        return False, None, "lxc command not found. Is LXD installed?"


def _build_install_command(pkg_manager: str) -> Tuple[bool, Optional[str], Optional[str]]:
    pkg_configs = {
        "apt-get": {
            "update": "DEBIAN_FRONTEND=noninteractive apt-get update -qq",
            "install": "DEBIAN_FRONTEND=noninteractive apt-get install -y -qq",
            "packages": "git openssh-server openssl curl ca-certificates python3 python3-pip",
        },
        "dnf": {
            "update": None,
            "install": "dnf install -y -q",
            "packages": "git openssh-server openssl curl ca-certificates python3 python3-pip",
        },
        "yum": {
            "update": None,
            "install": "yum install -y -q",
            "packages": "git openssh-server openssh-clients openssl curl ca-certificates python3 python3-pip",
        },
        "apk": {
            "update": None,
            "install": "apk add --no-cache",
            "packages": "git openssh openssl curl ca-certificates python3 py3-pip",
        },
        "pacman": {
            "update": None,
            "install": "pacman -S --noconfirm",
            "packages": "git openssh openssl curl ca-certificates python python-pip",
        },
    }
    
    if pkg_manager not in pkg_configs:
        return False, None, f"Unsupported package manager: {pkg_manager}"
    
    config = pkg_configs[pkg_manager]
    parts = []
    
    if config["update"]:
        parts.append(config["update"])
    
    parts.append(f"{config['install']} {config['packages']} || exit 1")
    
    return True, " && ".join(parts), None


def preinstall_dependencies(params: TestParams) -> Tuple[bool, Optional[str]]:
    if params.logger:
        params.logger.info(installing_dependencies.format(name=params.container_name))

    if params.dry_run:
        if params.logger:
            params.logger.info(
                dry_run_would_execute.format(
                    action=f"Pre-install dependencies in {params.container_name}"
                )
            )
        return True, None

    success, pkg_manager, error = _detect_package_manager(params.container_name)
    if not success:
        return False, error or "Failed to detect package manager"

    if params.logger:
        params.logger.debug(f"Detected package manager: {pkg_manager}")

    success, install_cmd, error = _build_install_command(pkg_manager)
    if not success:
        return False, error
    
    docker_install_cmd = (
        "curl -fsSL https://get.docker.com -o /tmp/get-docker.sh && "
        "sh /tmp/get-docker.sh && "
        "rm -f /tmp/get-docker.sh || exit 1"
    )
    
    full_install_cmd = f"{install_cmd} && {docker_install_cmd}"

    DEPENDENCY_INSTALL_TIMEOUT = 600
    try:
        if params.logger:
            params.logger.debug(f"Installing dependencies: {install_cmd}")
            params.logger.debug("Installing Docker...")
        
        result = subprocess.run(
            ["lxc", "exec", params.container_name, "--", "sh", "-c", full_install_cmd],
            capture_output=True,
            text=True,
            check=False,
            timeout=DEPENDENCY_INSTALL_TIMEOUT,
        )
        
        if result.returncode != 0:
            error_output = result.stderr if result.stderr else result.stdout
            return False, f"Failed to install dependencies: {error_output}"
        
        if params.logger:
            params.logger.success(dependencies_installed.format(name=params.container_name))
        
        return True, None
    except subprocess.TimeoutExpired:
        return False, f"Timeout installing dependencies (exceeded {DEPENDENCY_INSTALL_TIMEOUT}s)"
    except Exception as e:
        return False, f"Error installing dependencies: {str(e)}"

def configure_proxy_ports(params: TestParams) -> Tuple[bool, Optional[str]]:
    if params.logger:
        params.logger.info(configuring_proxy.format(name=params.container_name))
        params.logger.debug(f"App proxy: {params.app_port} -> {params.internal_app_port}")
        params.logger.debug(f"API proxy: {params.api_port} -> {params.internal_api_port}")

    if params.dry_run:
        if params.logger:
            params.logger.info(
                dry_run_would_execute.format(
                    action=f"Configure proxy: {params.app_port}->{params.internal_app_port}, {params.api_port}->{params.internal_api_port}"
                )
            )
        return True, None

    try:
        if params.logger:
            params.logger.debug("Configuring LXD proxy devices...")
        if params.app_port:
            app_proxy_name = f"app-proxy-{params.container_name}"
            subprocess.run(
                [
                    "lxc",
                    "config",
                    "device",
                    "add",
                    params.container_name,
                    app_proxy_name,
                    "proxy",
                    f"listen=tcp:0.0.0.0:{params.app_port}",
                    f"connect=tcp:{params.proxy_url}:{params.internal_app_port}",
                ],
                check=True,
                capture_output=True,
            )
            if params.logger:
                params.logger.info(
                    proxy_configured.format(
                        external=params.app_port, internal=params.internal_app_port
                    )
                )

        if params.api_port:
            api_proxy_name = f"api-proxy-{params.container_name}"
            subprocess.run(
                [
                    "lxc",
                    "config",
                    "device",
                    "add",
                    params.container_name,
                    api_proxy_name,
                    "proxy",
                    f"listen=tcp:0.0.0.0:{params.api_port}",
                    f"connect=tcp:{params.proxy_url}:{params.internal_api_port}",
                ],
                check=True,
                capture_output=True,
            )
            if params.logger:
                params.logger.info(
                    proxy_configured.format(
                        external=params.api_port, internal=params.internal_api_port
                    )
                )

        return True, None
    except subprocess.CalledProcessError as e:
        error_msg = e.stderr if hasattr(e, 'stderr') and e.stderr else str(e)
        return False, f"Failed to configure proxy: {error_msg}"


def cleanup_container(params: TestParams) -> Tuple[bool, Optional[str]]:
    if not params.container_name:
        return True, None

    if params.logger:
        params.logger.info(cleaning_up.format(name=params.container_name))

    if params.dry_run:
        if params.logger:
            params.logger.info(
                dry_run_would_execute.format(
                    action=f"Delete container {params.container_name}"
                )
            )
        return True, None

    try:
        subprocess.run(
            ["lxc", "delete", "--force", params.container_name],
            check=True,
            capture_output=True,
        )
        return True, None
    except subprocess.CalledProcessError as e:
        error_msg = e.stderr if hasattr(e, 'stderr') and e.stderr else str(e)
        return False, f"Failed to cleanup container: {error_msg}"
