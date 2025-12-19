from dataclasses import dataclass
from typing import Optional

from app.utils.protocols import LoggerProtocol


@dataclass
class InstallParams:
    logger: Optional[LoggerProtocol] = None
    verbose: bool = False
    timeout: int = 300
    force: bool = False
    dry_run: bool = False
    config_file: Optional[str] = None
    api_domain: Optional[str] = None
    view_domain: Optional[str] = None
    host_ip: Optional[str] = None
    repo: Optional[str] = None
    branch: Optional[str] = None
    api_port: Optional[int] = None
    view_port: Optional[int] = None
    db_port: Optional[int] = None
    redis_port: Optional[int] = None
    caddy_admin_port: Optional[int] = None
    caddy_http_port: Optional[int] = None
    caddy_https_port: Optional[int] = None
    supertokens_port: Optional[int] = None
    external_db_url: Optional[str] = None
    staging: bool = False
    no_rollback: bool = False
    verify_health: bool = True
    health_check_timeout: int = 120
    admin_email: Optional[str] = None
    admin_password: Optional[str] = None

