from app.utils.protocols import DockerServiceProtocol, LoggerProtocol

from .base import BaseAction, BaseConfig, BaseDockerCommandBuilder, BaseDockerService, BaseFormatter, BaseResult, BaseService
from .messages import (
    dry_run_command,
    dry_run_command_would_be_executed,
    dry_run_env_file,
    dry_run_mode,
    dry_run_service,
    end_dry_run,
    service_restart_failed,
    services_restarted_successfully,
)


class DockerCommandBuilder(BaseDockerCommandBuilder):
    @staticmethod
    def build_restart_command(name: str = "all", env_file: str = None, compose_file: str = None) -> list[str]:
        return BaseDockerCommandBuilder.build_command("restart", name, env_file, compose_file)


class RestartFormatter(BaseFormatter):
    def format_output(self, result: "RestartResult", output: str) -> str:
        return super().format_output(result, output, services_restarted_successfully, service_restart_failed)

    def format_dry_run(self, config: "RestartConfig") -> str:
        dry_run_messages = {
            "mode": dry_run_mode,
            "command_would_be_executed": dry_run_command_would_be_executed,
            "command": dry_run_command,
            "service": dry_run_service,
            "env_file": dry_run_env_file,
            "end": end_dry_run,
        }
        return super().format_dry_run(config, DockerCommandBuilder(), dry_run_messages)


class DockerService(BaseDockerService):
    def __init__(self, logger: LoggerProtocol):
        super().__init__(logger, "restart")

    def restart_services(self, name: str = "all", env_file: str = None, compose_file: str = None) -> tuple[bool, str]:
        return self.execute_services(name, env_file, compose_file)


class RestartResult(BaseResult):
    pass


class RestartConfig(BaseConfig):
    pass


class RestartService(BaseService[RestartConfig, RestartResult]):
    def __init__(self, config: RestartConfig, logger: LoggerProtocol = None, docker_service: DockerServiceProtocol = None):
        super().__init__(config, logger, docker_service)
        self.docker_service = docker_service or DockerService(self.logger)
        self.formatter = RestartFormatter()

    def _create_result(self, success: bool, error: str = None, docker_output: str = None) -> RestartResult:
        return RestartResult(
            name=self.config.name,
            env_file=self.config.env_file,
            verbose=self.config.verbose,
            output=self.config.output,
            success=success,
            error=error,
            docker_output=docker_output,
        )

    def restart(self) -> RestartResult:
        return self.execute()

    def execute(self) -> RestartResult:
        self.logger.debug(f"Restarting services: {self.config.name}")

        success, docker_output = self.docker_service.restart_services(
            self.config.name, self.config.env_file, self.config.compose_file
        )

        error = None if success else docker_output
        return self._create_result(success, error, docker_output)

    def restart_and_format(self) -> str:
        return self.execute_and_format()

    def execute_and_format(self) -> str:
        if self.config.dry_run:
            return self.formatter.format_dry_run(self.config)

        result = self.execute()
        return self.formatter.format_output(result, self.config.output)


class Restart(BaseAction[RestartConfig, RestartResult]):
    def __init__(self, logger: LoggerProtocol = None):
        super().__init__(logger)
        self.formatter = RestartFormatter()

    def restart(self, config: RestartConfig) -> RestartResult:
        return self.execute(config)

    def execute(self, config: RestartConfig) -> RestartResult:
        service = RestartService(config, logger=self.logger)
        return service.execute()

    def format_output(self, result: RestartResult, output: str) -> str:
        return self.formatter.format_output(result, output)

    def format_dry_run(self, config: RestartConfig) -> str:
        return self.formatter.format_dry_run(config)
