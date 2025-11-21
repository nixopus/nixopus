import os
import re

from app.utils.config import (
    API_PORT,
    CADDY_ADMIN_PORT,
    CADDY_HTTP_PORT,
    CADDY_HTTPS_PORT,
    Config,
    PROXY_PORT,
    SUPERTOKENS_API_PORT,
    VIEW_PORT,
)
from app.utils.protocols import LoggerProtocol

from .messages import configuration_key_has_no_default_value


class BaseInstall:
    """Base class with shared logic for both production and development installations"""
    
    def __init__(
        self,
        logger: LoggerProtocol = None,
        verbose: bool = False,
        timeout: int = 300,
        force: bool = False,
        dry_run: bool = False,
        config_file: str = None,
        repo: str = None,
        branch: str = None,
        api_port: int = None,
        view_port: int = None,
        db_port: int = None,
        redis_port: int = None,
        caddy_admin_port: int = None,
        caddy_http_port: int = None,
        caddy_https_port: int = None,
        supertokens_port: int = None,
    ):
        self.logger = logger
        self.verbose = verbose
        self.timeout = timeout
        self.force = force
        self.dry_run = dry_run
        self.config_file = config_file
        self.repo = repo
        self.branch = branch
        self.api_port = api_port
        self.view_port = view_port
        self.db_port = db_port
        self.redis_port = redis_port
        self.caddy_admin_port = caddy_admin_port
        self.caddy_http_port = caddy_http_port
        self.caddy_https_port = caddy_https_port
        self.supertokens_port = supertokens_port
        self._config = Config()
        self._config.load_user_config(self.config_file)
        self._user_config = None
        self.progress = None
        self.main_task = None
    
    def _get_config(self, path: str):
        """Base config getter - override in subclasses for specific behavior"""
        # Override port values if provided via command line
        if path == API_PORT and self.api_port is not None:
            return str(self.api_port)
        if path == VIEW_PORT and self.view_port is not None:
            return str(self.view_port)
        if path == "services.db.env.DB_PORT" and self.db_port is not None:
            return str(self.db_port)
        if path == "services.redis.env.REDIS_PORT" and self.redis_port is not None:
            return str(self.redis_port)
        
        # Handle PROXY_PORT (which is an alias for CADDY_ADMIN_PORT)
        if path == PROXY_PORT:
            if self.caddy_admin_port is not None:
                return str(self.caddy_admin_port)
            # Fall back to CADDY_ADMIN_PORT from config, default to 2019
            try:
                return str(self._config.get(CADDY_ADMIN_PORT))
            except (KeyError, ValueError):
                return "2019"
        
        if path == CADDY_ADMIN_PORT and self.caddy_admin_port is not None:
            return str(self.caddy_admin_port)
        if path == CADDY_HTTP_PORT and self.caddy_http_port is not None:
            return str(self.caddy_http_port)
        if path == CADDY_HTTPS_PORT and self.caddy_https_port is not None:
            return str(self.caddy_https_port)
        if path == SUPERTOKENS_API_PORT and self.supertokens_port is not None:
            return str(self.supertokens_port)
        
        # Handle simple key lookups for port overrides
        if path == "db_port" and self.db_port is not None:
            return str(self.db_port)
        if path == "redis_port" and self.redis_port is not None:
            return str(self.redis_port)
        if path == "proxy_port" and self.caddy_admin_port is not None:
            return str(self.caddy_admin_port)
        if path == "api_port" and self.api_port is not None:
            return str(self.api_port)
        if path == "view_port" and self.view_port is not None:
            return str(self.view_port)
        if path == "caddy_admin_port" and self.caddy_admin_port is not None:
            return str(self.caddy_admin_port)
        if path == "caddy_http_port" and self.caddy_http_port is not None:
            return str(self.caddy_http_port)
        if path == "caddy_https_port" and self.caddy_https_port is not None:
            return str(self.caddy_https_port)
        if path == "supertokens_api_port" and self.supertokens_port is not None:
            return str(self.supertokens_port)
        
        return self._config.get(path)
    
    def _validate_domains(self, api_domain: str = None, view_domain: str = None):
        """Validate domain format"""
        if (api_domain is None) != (view_domain is None):
            raise ValueError("Both api_domain and view_domain must be provided together, or neither should be provided")
        
        if api_domain and view_domain:
            domain_pattern = re.compile(
                r"^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?))*$"
            )
            if not domain_pattern.match(api_domain) or not domain_pattern.match(view_domain):
                raise ValueError("Invalid domain format. Domains must be valid hostnames")
    
    def _validate_repo(self):
        """Validate repository URL format"""
        if self.repo:
            if not (
                self.repo.startswith(("http://", "https://", "git://", "ssh://"))
                or (self.repo.endswith(".git") and not self.repo.startswith("github.com:"))
                or ("@" in self.repo and ":" in self.repo and self.repo.count("@") == 1)
            ):
                raise ValueError("Invalid repository URL format")
    
    def _is_custom_repo_or_branch(self, default_repo: str, default_branch: str):
        """Check if custom repository or branch is provided"""
        repo_differs = self.repo is not None and self.repo != default_repo
        branch_differs = self.branch is not None and self.branch != default_branch
        return repo_differs or branch_differs

