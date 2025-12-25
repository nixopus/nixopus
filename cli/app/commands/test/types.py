from dataclasses import dataclass
from typing import Optional

from app.utils.protocols import LoggerProtocol


@dataclass
class TestParams:
    logger: Optional[LoggerProtocol] = None
    verbose: bool = False
    timeout: int = 600
    dry_run: bool = False
    image: Optional[str] = None
    container_name: Optional[str] = None
    app_port: Optional[int] = None
    api_port: Optional[int] = None
    proxy_url: str = "127.0.0.1"
    internal_app_port: int = 7443
    internal_api_port: int = 8443
    distro: str = "images:debian/13"
    repo: Optional[str] = None
    branch: Optional[str] = None
    health_check_timeout: int = 300

