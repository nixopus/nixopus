from typing import Protocol

from pydantic import Field

from app.utils.protocols import LoggerProtocol

from .base import BaseAction, BaseConfig, BaseDockerCommandBuilder, BaseDockerService, BaseFormatter, BaseResult, BaseService
from .messages import (
    dry_run_command,
    dry_run_command_would_be_executed,
    dry_run_detach_mode,
    dry_run_env_file,
    dry_run_mode,
    dry_run_service,
    end_dry_run,
    service_start_failed,
    services_started_successfully,
)


class DockerServiceProtocol(Protocol):
    def start_services(
        self, name: str = "all", detach: bool = True, env_file: str = None, compose_file: str = None
    ) -> tuple[bool, str]: ...


class DockerCommandBuilder(BaseDockerCommandBuilder):
    @staticmethod
    def build_up_command(name: str = "all", detach: bool = True, env_file: str = None, compose_file: str = None) -> list[str]:
        return BaseDockerCommandBuilder.build_command("up", name, env_file, compose_file, detach=detach)


class UpFormatter(BaseFormatter):
    def format_output(self, result: "UpResult", output: str) -> str:
        return super().format_output(result, output, services_started_successfully, service_start_failed)

    def format_dry_run(self, config: "UpConfig") -> str:
        dry_run_messages = {
            "mode": dry_run_mode,
            "command_would_be_executed": dry_run_command_would_be_executed,
            "command": dry_run_command,
            "service": dry_run_service,
            "detach_mode": dry_run_detach_mode,
            "env_file": dry_run_env_file,
            "end": end_dry_run,
        }
        return super().format_dry_run(config, DockerCommandBuilder(), dry_run_messages)


class DockerService(BaseDockerService):
    def __init__(self, logger: LoggerProtocol):
        super().__init__(logger, "up")

    def start_services(
        self, name: str = "all", detach: bool = False, env_file: str = None, compose_file: str = None
    ) -> tuple[bool, str]:
        return self.execute_services(name, env_file, compose_file, detach=detach)


class UpResult(BaseResult):
    detach: bool


class UpConfig(BaseConfig):
    detach: bool = Field(False, description="Run services in detached mode")


class UpService(BaseService[UpConfig, UpResult]):
    def __init__(self, config: UpConfig, logger: LoggerProtocol = None, docker_service: DockerServiceProtocol = None):
        super().__init__(config, logger, docker_service)
        self.docker_service = docker_service or DockerService(self.logger)
        self.formatter = UpFormatter()

    def _create_result(self, success: bool, error: str = None, docker_output: str = None) -> UpResult:
        return UpResult(
            name=self.config.name,
            detach=self.config.detach,
            env_file=self.config.env_file,
            verbose=self.config.verbose,
            output=self.config.output,
            success=success,
            error=error,
            docker_output=docker_output,
        )

    def up(self) -> UpResult:
        return self.execute()

    def execute(self) -> UpResult:
        self.logger.debug(f"Starting services: {self.config.name}")

        # Handle dry-run mode
        if self.config.dry_run:
            self.logger.debug("[DRY RUN] Would start services")
            return self._create_result(
                success=True,
                error=None,
                docker_output="[DRY RUN] Services would be started"
            )

        success, docker_output = self.docker_service.start_services(
            self.config.name, self.config.detach, self.config.env_file, self.config.compose_file
        )

        error = None if success else docker_output
        return self._create_result(success, error, docker_output)

    def up_and_format(self) -> str:
        return self.execute_and_format()

    def execute_and_format(self) -> str:
        if self.config.dry_run:
            return self.formatter.format_dry_run(self.config)

        result = self.execute()
        return self.formatter.format_output(result, self.config.output)


class Up(BaseAction[UpConfig, UpResult]):
    def __init__(self, logger: LoggerProtocol = None):
        super().__init__(logger)
        self.formatter = UpFormatter()

    def up(self, config: UpConfig) -> UpResult:
        return self.execute(config)

    def execute(self, config: UpConfig) -> UpResult:
        service = UpService(config, logger=self.logger)
        return service.execute()

    def format_output(self, result: UpResult, output: str) -> str:
        return self.formatter.format_output(result, output)

    def format_dry_run(self, config: UpConfig) -> str:
        return self.formatter.format_dry_run(config)
