from typing import Optional
from pydantic import Field

from app.utils.logger import Logger
from app.utils.protocols import LoggerProtocol, DockerServiceProtocol
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
    services_status_retrieved,
    service_status_failed,
    unknown_error
)

class DockerCommandBuilder(BaseDockerCommandBuilder):
    @staticmethod
    def build_ps_command(name: str = "all", env_file: str = None, compose_file: str = None) -> list[str]:
        return BaseDockerCommandBuilder.build_command("ps", name, env_file, compose_file)

class PsFormatter(BaseFormatter):
    def format_output(self, result: "PsResult", output: str) -> str:
        return super().format_output(result, output, services_status_retrieved, service_status_failed)
    
    def format_dry_run(self, config: "PsConfig") -> str:
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
        super().__init__(logger, "ps")
    
    def show_services_status(self, name: str = "all", env_file: str = None, compose_file: str = None) -> tuple[bool, str]:
        return self.execute_services(name, env_file, compose_file)

class PsResult(BaseResult):
    pass

class PsConfig(BaseConfig):
    pass

class PsService(BaseService[PsConfig, PsResult]):
    def __init__(self, config: PsConfig, logger: LoggerProtocol = None, docker_service: DockerServiceProtocol = None):
        super().__init__(config, logger, docker_service)
        self.docker_service = docker_service or DockerService(self.logger)
        self.formatter = PsFormatter()
    
    def _create_result(self, success: bool, error: str = None) -> PsResult:
        return PsResult(
            name=self.config.name,
            env_file=self.config.env_file,
            verbose=self.config.verbose,
            output=self.config.output,
            success=success,
            error=error
        )
    
    def ps(self) -> PsResult:
        return self.execute()
    
    def execute(self) -> PsResult:
        self.logger.debug(f"Checking status of services: {self.config.name}")
        
        success, error = self.docker_service.show_services_status(
            self.config.name,
            self.config.env_file,
            self.config.compose_file
        )
        
        return self._create_result(success, error)
    
    def ps_and_format(self) -> str:
        return self.execute_and_format()
    
    def execute_and_format(self) -> str:
        if self.config.dry_run:
            return self.formatter.format_dry_run(self.config)
        
        result = self.execute()
        return self.formatter.format_output(result, self.config.output)

class Ps(BaseAction[PsConfig, PsResult]):
    def __init__(self, logger: LoggerProtocol = None):
        super().__init__(logger)
        self.formatter = PsFormatter()
    
    def ps(self, config: PsConfig) -> PsResult:
        return self.execute(config)
    
    def execute(self, config: PsConfig) -> PsResult:
        service = PsService(config, logger=self.logger)
        return service.execute()
    
    def format_output(self, result: PsResult, output: str) -> str:
        return self.formatter.format_output(result, output) 
