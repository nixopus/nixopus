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
    configuration_listed,
    configuration_list_failed,
    no_configuration_found,
    dry_run_mode,
    dry_run_list_config,
    end_dry_run
)

class EnvironmentServiceProtocol(Protocol):
    def list_config(self, service: str, env_file: str = None) -> tuple[bool, Dict[str, str], str]:
        ...

class EnvironmentManager(BaseEnvironmentManager):
    def list_config(self, service: str, env_file: Optional[str] = None) -> tuple[bool, Dict[str, str], Optional[str]]:
        file_path = self.get_service_env_file(service, env_file)
        return self.read_env_file(file_path)

class ListResult(BaseResult):
    pass

class ListConfig(BaseConfig):
    pass

class ListService(BaseService[ListConfig, ListResult]):
    def __init__(self, config: ListConfig, logger: LoggerProtocol = None, environment_service: EnvironmentServiceProtocol = None):
        super().__init__(config, logger, environment_service)
        self.environment_service = environment_service or EnvironmentManager(self.logger)
    
    def _create_result(self, success: bool, error: str = None, config_dict: Dict[str, str] = None) -> ListResult:
        return ListResult(
            service=self.config.service,
            verbose=self.config.verbose,
            output=self.config.output,
            success=success,
            error=error,
            config=config_dict or {}
        )
    
    def list(self) -> ListResult:
        return self.execute()
    
    def execute(self) -> ListResult:
        if self.config.dry_run:
            return self._create_result(True)
        
        success, config_dict, error = self.environment_service.list_config(
            self.config.service, self.config.env_file
        )
        
        if success:
            self.logger.info(configuration_listed.format(service=self.config.service))
            return self._create_result(True, config_dict=config_dict)
        else:
            self.logger.error(configuration_list_failed.format(
                service=self.config.service, error=error
            ))
            return self._create_result(False, error=error)
    
    def list_and_format(self) -> str:
        return self.execute_and_format()
    
    def execute_and_format(self) -> str:
        if self.config.dry_run:
            return self._format_dry_run()
        
        result = self.execute()
        return self._format_output(result, self.config.output)
    
    def _format_dry_run(self) -> str:
        lines = [dry_run_mode]
        lines.append(dry_run_list_config.format(service=self.config.service))
        lines.append(end_dry_run)
        return "\n".join(lines)
    
    def _format_output(self, result: ListResult, output_format: str) -> str:
        if output_format == "json":
            return self._format_json(result)
        else:
            return self._format_text(result)
    
    def _format_json(self, result: ListResult) -> str:
        import json
        output = {
            "service": result.service,
            "success": result.success,
            "error": result.error,
            "config": result.config
        }
        return json.dumps(output, indent=2)
    
    def _format_text(self, result: ListResult) -> str:
        if not result.success:
            return configuration_list_failed.format(service=result.service, error=result.error)
        
        if result.config:
            lines = [configuration_listed.format(service=result.service)]
            for key, value in sorted(result.config.items()):
                lines.append(f"  {key}={value}")
            return "\n".join(lines)
        
        return no_configuration_found.format(service=result.service)

class List(BaseAction[ListConfig, ListResult]):
    def __init__(self, logger: LoggerProtocol = None):
        super().__init__(logger)
    
    def list(self, config: ListConfig) -> ListResult:
        return self.execute(config)
    
    def execute(self, config: ListConfig) -> ListResult:
        service = ListService(config, logger=self.logger)
        return service.execute()
    
    def format_output(self, result: ListResult, output: str) -> str:
        service = ListService(result, logger=self.logger)
        return service._format_output(result, output) 
