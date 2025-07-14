from typing import Protocol
from pydantic import BaseModel

from app.utils.logger import Logger
from app.utils.protocols import LoggerProtocol
from app.utils.output_formatter import OutputFormatter
from .base import (
    BaseCaddyCommandBuilder,
    BaseFormatter,
    BaseCaddyService,
    BaseConfig,
    BaseResult,
    BaseService,
    BaseAction
)
from .messages import (
    dry_run_mode,
    dry_run_command_would_be_executed,
    dry_run_command,
    dry_run_port,
    end_dry_run,
    proxy_stopped_successfully,
    proxy_stop_failed,
    debug_stop_proxy,
)

class CaddyServiceProtocol(Protocol):
    def stop_proxy(self, port: int = 2019) -> tuple[bool, str]:
        ...

class CaddyCommandBuilder(BaseCaddyCommandBuilder):
    @staticmethod
    def build_stop_command(port: int = 2019) -> list[str]:
        return BaseCaddyCommandBuilder.build_stop_command(port)

class StopFormatter(BaseFormatter):
    def format_output(self, result: "StopResult", output: str) -> str:
        return super().format_output(result, output, proxy_stopped_successfully, proxy_stop_failed)
    
    def format_dry_run(self, config: "StopConfig") -> str:
        dry_run_messages = {
            "mode": dry_run_mode,
            "command_would_be_executed": dry_run_command_would_be_executed,
            "command": dry_run_command,
            "port": dry_run_port,
            "end": end_dry_run
        }
        return super().format_dry_run(config, CaddyCommandBuilder(), dry_run_messages)

class CaddyService(BaseCaddyService):
    def __init__(self, logger: LoggerProtocol):
        super().__init__(logger)
    
    def stop_caddy(self, port: int = 2019) -> tuple[bool, str]:
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
            error=error
        )
    
    def stop(self) -> StopResult:
        return self.execute()
    
    def execute(self) -> StopResult:
        self.logger.debug(debug_stop_proxy.format(port=self.config.proxy_port))
        
        success, error = self.caddy_service.stop_caddy(self.config.proxy_port)
        
        return self._create_result(success, error)
    
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
