from typing import Protocol

from pydantic import BaseModel

from app.utils.config import Config, PROXY_PORT
from app.utils.logger import Logger
from app.utils.output_formatter import OutputFormatter
from app.utils.protocols import LoggerProtocol

from .base import BaseAction, BaseCaddyCommandBuilder, BaseCaddyService, BaseConfig, BaseFormatter, BaseResult, BaseService
from .messages import (
    debug_stop_proxy,
    dry_run_command,
    dry_run_command_would_be_executed,
    dry_run_mode,
    dry_run_port,
    end_dry_run,
    proxy_stop_failed,
    proxy_stopped_successfully,
)

config = Config()
proxy_port = config.get_yaml_value(PROXY_PORT)


class CaddyServiceProtocol(Protocol):
    def stop_proxy(self, port: int = proxy_port) -> tuple[bool, str]: ...


class CaddyCommandBuilder(BaseCaddyCommandBuilder):
    @staticmethod
    def build_stop_command(port: int = proxy_port) -> list[str]:
        return BaseCaddyCommandBuilder.build_stop_command(port)


class StopFormatter(BaseFormatter):
    def format_output(self, result: "StopResult", output: str) -> str:
        if output == "json":
            success_msg = "Caddy stopped successfully" if result.success else "Failed to stop Caddy"
            return super().format_output(result, output, success_msg, result.error or "Unknown error")
        
        if result.success:
            return "Caddy stopped successfully"
        else:
            return result.error or "Failed to stop Caddy"

    def format_dry_run(self, config: "StopConfig") -> str:
        dry_run_messages = {
            "mode": dry_run_mode,
            "command_would_be_executed": dry_run_command_would_be_executed,
            "command": dry_run_command,
            "port": dry_run_port,
            "end": end_dry_run,
        }
        return super().format_dry_run(config, CaddyCommandBuilder(), dry_run_messages)


class CaddyService(BaseCaddyService):
    def __init__(self, logger: LoggerProtocol):
        super().__init__(logger)

    def stop_caddy(self, port: int = proxy_port) -> tuple[bool, str]:
        return self.stop_proxy(port)


class StopResult(BaseResult):
    pass


class StopConfig(BaseConfig):
    pass


class StopService(BaseService[StopConfig, StopResult]):
    def __init__(self, config: StopConfig, logger: LoggerProtocol = None, caddy_service: CaddyServiceProtocol = None):
        super().__init__(config, logger, caddy_service)
        self.caddy_service = caddy_service or CaddyService(self.logger)
        self.formatter = StopFormatter()

    def _create_result(self, success: bool, error: str = None) -> StopResult:
        return StopResult(
            proxy_port=self.config.proxy_port,
            verbose=self.config.verbose,
            output=self.config.output,
            success=success,
            error=error,
        )

    def stop(self) -> StopResult:
        return self.execute()

    def execute(self) -> StopResult:
        success, message = self.caddy_service.stop_caddy(self.config.proxy_port)
        return self._create_result(success, None if success else message)

    def stop_and_format(self) -> str:
        return self.execute_and_format()

    def execute_and_format(self) -> str:
        if self.config.dry_run:
            return self.formatter.format_dry_run(self.config)

        result = self.execute()
        return self.formatter.format_output(result, self.config.output)


class Stop(BaseAction[StopConfig, StopResult]):
    def __init__(self, logger: LoggerProtocol = None):
        super().__init__(logger)
        self.formatter = StopFormatter()

    def stop(self, config: StopConfig) -> StopResult:
        return self.execute(config)

    def execute(self, config: StopConfig) -> StopResult:
        service = StopService(config, logger=self.logger)
        return service.execute()

    def format_output(self, result: StopResult, output: str) -> str:
        return self.formatter.format_output(result, output)
