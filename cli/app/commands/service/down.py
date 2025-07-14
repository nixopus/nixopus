import subprocess
import os
from typing import Protocol, Optional
from pydantic import BaseModel, Field, field_validator

from app.utils.logger import Logger
from app.utils.protocols import LoggerProtocol
from .base import (
    BaseDockerCommandBuilder,
    BaseFormatter,
    BaseDockerService,
    BaseConfig,
    BaseResult,
    BaseService,
    BaseAction
)
from .messages import (
    dry_run_mode,
    dry_run_command_would_be_executed,
    dry_run_command,
    dry_run_service,
    dry_run_env_file,
    end_dry_run,
    services_stopped_successfully,
    service_stop_failed,
)

class DockerServiceProtocol(Protocol):
    def stop_services(self, name: str = "all", env_file: str = None, compose_file: str = None) -> tuple[bool, str]:
        ...

class DockerCommandBuilder(BaseDockerCommandBuilder):
    @staticmethod
    def build_down_command(name: str = "all", env_file: str = None, compose_file: str = None) -> list[str]:
        return BaseDockerCommandBuilder.build_command("down", name, env_file, compose_file)

class DownFormatter(BaseFormatter):
    def format_output(self, result: "DownResult", output: str) -> str:
        return super().format_output(result, output, services_stopped_successfully, service_stop_failed)
    
    def format_dry_run(self, config: "DownConfig") -> str:
        dry_run_messages = {
            "mode": dry_run_mode,
            "command_would_be_executed": dry_run_command_would_be_executed,
            "command": dry_run_command,
            "service": dry_run_service,
            "env_file": dry_run_env_file,
            "end": end_dry_run
        }
        return super().format_dry_run(config, DockerCommandBuilder(), dry_run_messages)

class DockerService(BaseDockerService):
    def __init__(self, logger: LoggerProtocol):
        super().__init__(logger, "down")
    
    def stop_services(self, name: str = "all", env_file: str = None, compose_file: str = None) -> tuple[bool, str]:
        return self.execute_services(name, env_file, compose_file)

class DownResult(BaseResult):
    pass

class DownConfig(BaseConfig):
    pass

class DownService(BaseService[DownConfig, DownResult]):
    def __init__(self, config: DownConfig, logger: LoggerProtocol = None, docker_service: DockerServiceProtocol = None):
        super().__init__(config, logger, docker_service)
        self.docker_service = docker_service or DockerService(self.logger)
        self.formatter = DownFormatter()
    
    def _create_result(self, success: bool, error: str = None) -> DownResult:
        return DownResult(
            name=self.config.name,
            env_file=self.config.env_file,
            verbose=self.config.verbose,
            output=self.config.output,
            success=success,
            error=error
        )
    
    def down(self) -> DownResult:
        return self.execute()
    
    def execute(self) -> DownResult:
        self.logger.debug(f"Stopping services: {self.config.name}")
        
        success, error = self.docker_service.stop_services(
            self.config.name,
            self.config.env_file,
            self.config.compose_file
        )
        
        return self._create_result(success, error)
    
    def down_and_format(self) -> str:
        return self.execute_and_format()
    
    def execute_and_format(self) -> str:
        if self.config.dry_run:
            return self.formatter.format_dry_run(self.config)
        
        result = self.execute()
        return self.formatter.format_output(result, self.config.output)

class Down(BaseAction[DownConfig, DownResult]):
    def __init__(self, logger: LoggerProtocol = None):
        super().__init__(logger)
        self.formatter = DownFormatter()
    
    def down(self, config: DownConfig) -> DownResult:
        return self.execute(config)
    
    def execute(self, config: DownConfig) -> DownResult:
        service = DownService(config, logger=self.logger)
        return service.execute()
    
    def format_output(self, result: DownResult, output: str) -> str:
        return self.formatter.format_output(result, output) 
