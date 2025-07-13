import subprocess
import os
from typing import Protocol, Optional
from pydantic import BaseModel, Field, field_validator

from app.utils.logger import Logger
from app.utils.protocols import LoggerProtocol
from app.utils.lib import DirectoryManager
from app.utils.output_formatter import OutputFormatter
from .messages import (
    path_already_exists_use_force,
    executing_command,
    successfully_cloned,
    git_clone_failed,
    unexpected_error_during_clone,
    dry_run_mode,
    dry_run_command_would_be_executed,
    dry_run_command,
    dry_run_repository,
    dry_run_branch,
    dry_run_target_path,
    dry_run_force_mode,
    path_exists_will_overwrite,
    path_exists_would_fail,
    target_path_not_exists,
    end_dry_run,
    cloning_repo_into_path,
    invalid_repository_url,
    invalid_path,
    invalid_repo,
    prerequisites_validation_failed,
    failed_to_prepare_target_directory,
    unknown_error,
    default_branch
)

class GitCloneProtocol(Protocol):
    def clone_repository(self, repo: str, path: str, branch: str = None) -> tuple[bool, str]:
        ...

class GitCommandBuilder:
    @staticmethod
    def build_clone_command(repo: str, path: str, branch: str = None) -> list[str]:
        cmd = ["git", "clone"]
        if branch:
            cmd.extend(["-b", branch])
        cmd.extend([repo, path])
        return cmd

class CloneFormatter:
    def __init__(self):
        self.output_formatter = OutputFormatter()
    
    def format_output(self, result: "CloneResult", output: str) -> str:
        if result.success:
            message = successfully_cloned.format(repo=result.repo, path=result.path)
            output_message = self.output_formatter.create_success_message(message, result.model_dump())
        else:
            error = result.error or unknown_error
            output_message = self.output_formatter.create_error_message(error, result.model_dump())
        
        return self.output_formatter.format_output(output_message, output)
    
    def format_dry_run(self, config: "CloneConfig") -> str:
        cmd = GitCommandBuilder.build_clone_command(config.repo, config.path, config.branch)
        
        output = []
        output.append(dry_run_mode)
        output.append(dry_run_command_would_be_executed)
        output.append(dry_run_command.format(command=' '.join(cmd)))
        output.append(dry_run_repository.format(repo=config.repo))
        output.append(dry_run_branch.format(branch=config.branch or default_branch))
        output.append(dry_run_target_path.format(path=config.path))
        output.append(dry_run_force_mode.format(force=config.force))
        
        self._add_path_status_message(output, config.path, config.force)
        
        output.append(end_dry_run)
        return "\n".join(output)
    
    def _add_path_status_message(self, output: list[str], path: str, force: bool) -> None:
        if os.path.exists(path):
            if force:
                output.append(path_exists_will_overwrite.format(path=path))
            else:
                output.append(path_exists_would_fail.format(path=path))
        else:
            output.append(target_path_not_exists.format(path=path))

class GitClone:
    def __init__(self, logger: LoggerProtocol):
        self.logger = logger
    
    def clone_repository(self, repo: str, path: str, branch: str = None) -> tuple[bool, str]:
        cmd = GitCommandBuilder.build_clone_command(repo, path, branch)
        
        try:
            self.logger.info(executing_command.format(command=' '.join(cmd)))
            result = subprocess.run(cmd, capture_output=True, text=True, check=True)
            self.logger.success(successfully_cloned.format(repo=repo, path=path))
            return True, None
        except subprocess.CalledProcessError as e:
            self.logger.error(git_clone_failed.format(error=e.stderr))
            return False, e.stderr
        except Exception as e:
            self.logger.error(unexpected_error_during_clone.format(error=e))
            return False, str(e)

class CloneResult(BaseModel):
    repo: str
    path: str
    branch: Optional[str]
    force: bool
    verbose: bool
    output: str
    success: bool = False
    error: Optional[str] = None

class CloneConfig(BaseModel):
    repo: str = Field(..., min_length=1, description="Repository URL to clone")
    branch: Optional[str] = Field("master", description="Branch to clone")
    path: str = Field(..., min_length=1, description="Target path for cloning")
    force: bool = Field(False, description="Force overwrite if path exists")
    verbose: bool = Field(False, description="Verbose output")
    output: str = Field("text", description="Output format: text, json")
    dry_run: bool = Field(False, description="Dry run mode")
    
    @field_validator("repo")
    @classmethod
    def validate_repo(cls, repo: str) -> str:
        stripped_repo = repo.strip()
        if not stripped_repo:
            raise ValueError(invalid_repo)
        
        if not cls._is_valid_repo_format(stripped_repo):
            raise ValueError(invalid_repository_url)
        return stripped_repo
    
    @staticmethod
    def _is_valid_repo_format(repo: str) -> bool:
        return (
            repo.startswith(('http://', 'https://', 'git://', 'ssh://')) or
            (repo.endswith('.git') and not repo.startswith('github.com:')) or
            ('@' in repo and ':' in repo and repo.count('@') == 1)
        )
    
    @field_validator("path")
    @classmethod
    def validate_path(cls, path: str) -> str:
        stripped_path = path.strip()
        if not stripped_path:
            raise ValueError(invalid_path)
        return stripped_path
    
    @field_validator("branch")
    @classmethod
    def validate_branch(cls, branch: str) -> Optional[str]:
        if not branch:
            return None
        stripped_branch = branch.strip()
        if not stripped_branch:
            return None
        return stripped_branch

class CloneService:
    def __init__(self, config: CloneConfig, logger: LoggerProtocol = None, cloner: GitCloneProtocol = None):
        self.config = config
        self.logger = logger or Logger(verbose=config.verbose)
        self.cloner = cloner or GitClone(self.logger)
        self.formatter = CloneFormatter()
        self.dir_manager = DirectoryManager()
    
    def _prepare_target_directory(self) -> bool:
        if self.config.force and os.path.exists(self.config.path):
            return self.dir_manager.remove_directory(
                self.config.path, 
                self.logger
            )
        return True
    
    def _validate_prerequisites(self) -> bool:
        if self.dir_manager.path_exists_and_not_force(self.config.path, self.config.force):
            self.logger.error(path_already_exists_use_force.format(path=self.config.path))
            return False
        return True
    
    def _create_result(self, success: bool, error: str = None) -> CloneResult:
        return CloneResult(
            repo=self.config.repo,
            path=self.config.path,
            branch=self.config.branch,
            force=self.config.force,
            verbose=self.config.verbose,
            output=self.config.output,
            success=success,
            error=error
        )
    
    def clone(self) -> CloneResult:
        self.logger.debug(cloning_repo_into_path.format(repo=self.config.repo, path=self.config.path))
        
        if not self._validate_prerequisites():
            return self._create_result(False, prerequisites_validation_failed)
        
        if not self._prepare_target_directory():
            return self._create_result(False, failed_to_prepare_target_directory)
        
        success, error = self.cloner.clone_repository(
            self.config.repo,
            self.config.path,   
            self.config.branch
        )
        
        return self._create_result(success, error)
    
    def clone_and_format(self) -> str:
        if self.config.dry_run:
            return self.formatter.format_dry_run(self.config)
        
        result = self.clone()
        return self.formatter.format_output(result, self.config.output)

class Clone:
    def __init__(self, logger: LoggerProtocol = None):
        self.logger = logger
        self.formatter = CloneFormatter()
    
    def clone(self, config: CloneConfig) -> CloneResult:
        service = CloneService(config, logger=self.logger)
        return service.clone()
    
    def format_output(self, result: CloneResult, output: str) -> str:
        return self.formatter.format_output(result, output)
