import os
import re

from app.utils.config import Config
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
    ):
        self.logger = logger
        self.verbose = verbose
        self.timeout = timeout
        self.force = force
        self.dry_run = dry_run
        self.config_file = config_file
        self.repo = repo
        self.branch = branch
        self._config = Config()
        self._config.load_user_config(self.config_file)
        self.progress = None
        self.main_task = None
    
    def _get_config(self, path: str):
        """Base config getter - override in subclasses for specific behavior"""
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

