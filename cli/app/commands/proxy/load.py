import os
from typing import Optional, Protocol

from pydantic import Field, field_validator

from app.utils.config import Config, PROXY_PORT
from app.utils.protocols import LoggerProtocol

from .base import BaseAction, BaseCaddyCommandBuilder, BaseCaddyService, BaseConfig, BaseFormatter, BaseResult, BaseService
from .messages import (
    dry_run_command,
    dry_run_command_would_be_executed,
    dry_run_config_file,
    dry_run_mode,
    dry_run_port,
    end_dry_run,
)

config = Config()
proxy_port = config.get_yaml_value(PROXY_PORT)


class CaddyServiceProtocol(Protocol):
    def load_config(self, config_file: str, port: int = proxy_port) -> tuple[bool, str]: ...


class CaddyCommandBuilder(BaseCaddyCommandBuilder):
    @staticmethod
    def build_load_command(config_file: str, port: int = proxy_port) -> list[str]:
        return BaseCaddyCommandBuilder.build_load_command(config_file, port)


class LoadFormatter(BaseFormatter):
    def format_output(self, result: "LoadResult", output: str) -> str:
        if output == "json":
            success_msg = "Configuration loaded successfully" if result.success else "Failed to load configuration"
            return super().format_output(result, output, success_msg, result.error or "Unknown error")
        
        if result.success:
            return "Configuration loaded successfully"
        else:
            return result.error or "Failed to load configuration"

    def format_dry_run(self, config: "LoadConfig") -> str:
        dry_run_messages = {
            "mode": dry_run_mode,
            "command_would_be_executed": dry_run_command_would_be_executed,
            "command": dry_run_command,
            "port": dry_run_port,
            "config_file": dry_run_config_file,
            "end": end_dry_run,
        }
        return super().format_dry_run(config, CaddyCommandBuilder(), dry_run_messages)


class CaddyService(BaseCaddyService):
    def __init__(self, logger: LoggerProtocol):
        super().__init__(logger)

    def load_config_file(self, config_file: str, port: int = proxy_port) -> tuple[bool, str]:
        return self.load_config(config_file, port)


class LoadResult(BaseResult):
    config_file: Optional[str]


class LoadConfig(BaseConfig):
    config_file: Optional[str] = Field(None, description="Path to Caddy config file")

    @field_validator("config_file")
    @classmethod
    def validate_config_file(cls, config_file: str) -> Optional[str]:
        if not config_file:
            return None
        stripped_config_file = config_file.strip()
        if not stripped_config_file:
            return None
        if not os.path.exists(stripped_config_file):
            raise ValueError(f"Configuration file not found: {stripped_config_file}")
        return stripped_config_file


class LoadService(BaseService[LoadConfig, LoadResult]):
    def __init__(self, config: LoadConfig, logger: LoggerProtocol = None, caddy_service: CaddyServiceProtocol = None):
        super().__init__(config, logger, caddy_service)
        self.caddy_service = caddy_service or CaddyService(self.logger)
        self.formatter = LoadFormatter()

    def _create_result(self, success: bool, error: str = None) -> LoadResult:
        return LoadResult(
            proxy_port=self.config.proxy_port,
            config_file=self.config.config_file,
            verbose=self.config.verbose,
            output=self.config.output,
            success=success,
            error=error,
        )

    def load(self) -> LoadResult:
        return self.execute()

    def execute(self) -> LoadResult:
        if not self.config.config_file:
            return self._create_result(False, "Configuration file is required")

        success, message = self.caddy_service.load_config_file(self.config.config_file, self.config.proxy_port)
        return self._create_result(success, None if success else message)

    def load_and_format(self) -> str:
        return self.execute_and_format()

    def execute_and_format(self) -> str:
        if self.config.dry_run:
            return self.formatter.format_dry_run(self.config)

        result = self.execute()
        return self.formatter.format_output(result, self.config.output)


class Load(BaseAction[LoadConfig, LoadResult]):
    def __init__(self, logger: LoggerProtocol = None):
        super().__init__(logger)
        self.formatter = LoadFormatter()

    def load(self, config: LoadConfig) -> LoadResult:
        return self.execute(config)

    def execute(self, config: LoadConfig) -> LoadResult:
        service = LoadService(config, logger=self.logger)
        return service.execute()

    def format_output(self, result: LoadResult, output: str) -> str:
        return self.formatter.format_output(result, output)
