from typing import Protocol, Optional, Dict
from pydantic import BaseModel, Field

from app.utils.logger import Logger
from app.utils.protocols import LoggerProtocol
from .base import (
    BaseEnvironmentManager,
    BaseConfig,
    BaseResult,
    BaseService,
    BaseAction
)
from .messages import (
    configuration_set,
    configuration_set_failed,
    key_required,
    value_required,
    dry_run_mode,
    dry_run_set_config,
    end_dry_run
)

class EnvironmentServiceProtocol(Protocol):
    def set_config(self, service: str, key: str, value: str, env_file: str = None) -> tuple[bool, str]:
        ...

class EnvironmentManager(BaseEnvironmentManager):
    def set_config(self, service: str, key: str, value: str, env_file: Optional[str] = None) -> tuple[bool, Optional[str]]:
        file_path = self.get_service_env_file(service, env_file)
        
        success, config, error = self.read_env_file(file_path)
        if not success:
            return False, error
        
        config[key] = value
        return self.write_env_file(file_path, config)

class SetResult(BaseResult):
    pass

class SetConfig(BaseConfig):
    key: str = Field(..., description="The key of the configuration to set")
    value: str = Field(..., description="The value of the configuration to set")

class SetService(BaseService[SetConfig, SetResult]):
    def __init__(self, config: SetConfig, logger: LoggerProtocol = None, environment_service: EnvironmentServiceProtocol = None):
        super().__init__(config, logger, environment_service)
        self.environment_service = environment_service or EnvironmentManager(self.logger)
    
    def _create_result(self, success: bool, error: str = None, config_dict: Dict[str, str] = None) -> SetResult:
        return SetResult(
            service=self.config.service,
            key=self.config.key,
            value=self.config.value,
            verbose=self.config.verbose,
            output=self.config.output,
            success=success,
            error=error,
            config=config_dict or {}
        )
    
    def set(self) -> SetResult:
        return self.execute()
    
    def execute(self) -> SetResult:
        if not self.config.key:
            return self._create_result(False, error=key_required)
        
        if not self.config.value:
            return self._create_result(False, error=value_required)
        
        if self.config.dry_run:
            return self._create_result(True)
        
        success, error = self.environment_service.set_config(
            self.config.service, self.config.key, self.config.value, self.config.env_file
        )
        
        if success:
            self.logger.info(configuration_set.format(
                service=self.config.service, key=self.config.key, value=self.config.value
            ))
            return self._create_result(True)
        else:
            self.logger.error(configuration_set_failed.format(
                service=self.config.service, error=error
            ))
            return self._create_result(False, error=error)
    
    def set_and_format(self) -> str:
        return self.execute_and_format()
    
    def execute_and_format(self) -> str:
        if self.config.dry_run:
            return self._format_dry_run()
        
        result = self.execute()
        return self._format_output(result, self.config.output)
    
    def _format_dry_run(self) -> str:
        lines = [dry_run_mode]
        lines.append(dry_run_set_config.format(
            service=self.config.service,
            key=self.config.key,
            value=self.config.value
        ))
        lines.append(end_dry_run)
        return "\n".join(lines)
    
    def _format_output(self, result: SetResult, output_format: str) -> str:
        if output_format == "json":
            return self._format_json(result)
        else:
            return self._format_text(result)
    
    def _format_json(self, result: SetResult) -> str:
        import json
        output = {
            "service": result.service,
            "key": result.key,
            "value": result.value,
            "success": result.success,
            "error": result.error
        }
        return json.dumps(output, indent=2)
    
    def _format_text(self, result: SetResult) -> str:
        if not result.success:
            return configuration_set_failed.format(service=result.service, error=result.error)
        
        return configuration_set.format(service=result.service, key=result.key, value=result.value)

class Set(BaseAction[SetConfig, SetResult]):
    def __init__(self, logger: LoggerProtocol = None):
        super().__init__(logger)
    
    def set(self, config: SetConfig) -> SetResult:
        return self.execute(config)
    
    def execute(self, config: SetConfig) -> SetResult:
        service = SetService(config, logger=self.logger)
        return service.execute()
    
    def format_output(self, result: SetResult, output: str) -> str:
        service = SetService(result, logger=self.logger)
        return service._format_output(result, output) 