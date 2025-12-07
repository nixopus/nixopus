import os
 
from app.utils.config import Config
from app.utils.protocols import LoggerProtocol

from .config_utils import get_proxy_port
from .messages import configuration_key_has_no_default_value
from .validate import validate_domains, validate_repo


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
    
    def _get_port_override(self, port_param: int, config_key: str, default: str = None) -> str:
        if port_param is not None:
            return str(port_param)
        if default:
            return default
        try:
            return str(self._config.get(config_key))
        except (KeyError, ValueError):
            return None
    
    def _get_config(self, path: str):
        port_mappings = {
            "api_port": (self.api_port, "services.api.env.API_PORT"),
            "view_port": (self.view_port, "services.view.env.VIEW_PORT"),
            "db_port": (self.db_port, "services.db.env.DB_PORT"),
            "redis_port": (self.redis_port, "services.redis.env.REDIS_PORT"),
            "caddy_admin_port": (self.caddy_admin_port, "services.caddy.env.CADDY_ADMIN_PORT"),
            "caddy_http_port": (self.caddy_http_port, "services.caddy.env.CADDY_HTTP_PORT"),
            "caddy_https_port": (self.caddy_https_port, "services.caddy.env.CADDY_HTTPS_PORT"),
            "supertokens_api_port": (self.supertokens_port, "services.api.env.SUPERTOKENS_API_PORT"),
        }
        
        if path in port_mappings:
            param_value, config_key = port_mappings[path]
            return self._get_port_override(param_value, config_key)
        
        if path == "proxy_port":
            return str(get_proxy_port(self._config, self.caddy_admin_port))
        
        return self._config.get(path)
    
    def _validate_domains(self, api_domain: str = None, view_domain: str = None):
        validate_domains(api_domain, view_domain)
    
    def _validate_repo(self):
        validate_repo(self.repo)
    
    def _is_custom_repo_or_branch(self, default_repo: str, default_branch: str):
        repo_differs = self.repo is not None and self.repo != default_repo
        branch_differs = self.branch is not None and self.branch != default_branch
        return repo_differs or branch_differs

