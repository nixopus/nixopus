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
    configuration_deleted,
    configuration_delete_failed,
    key_required_delete,
    dry_run_mode,
    dry_run_delete_config,
    end_dry_run,
    config_key_not_found
)

class EnvironmentServiceProtocol(Protocol):
    def delete_config(self, service: str, key: str, env_file: str = None) -> tuple[bool, str]:
        ...

class EnvironmentManager(BaseEnvironmentManager):
    def delete_config(self, service: str, key: str, env_file: Optional[str] = None) -> tuple[bool, Optional[str]]:
        file_path = self.get_service_env_file(service, env_file)
        
        success, config, error = self.read_env_file(file_path)
        if not success:
            return False, error
        
        if key not in config:
            return False, config_key_not_found.format(key=key)
        
        del config[key]
        return self.write_env_file(file_path, config)

class DeleteResult(BaseResult):
    pass

class DeleteConfig(BaseConfig):
    key: str = Field(..., description="The key of the configuration to delete")

class DeleteService(BaseService[DeleteConfig, DeleteResult]):
    def __init__(self, config: DeleteConfig, logger: LoggerProtocol = None, environment_service: EnvironmentServiceProtocol = None):
        super().__init__(config, logger, environment_service)
        self.environment_service = environment_service or EnvironmentManager(self.logger)
    
    def _create_result(self, success: bool, error: str = None, config_dict: Dict[str, str] = None) -> DeleteResult:
        return DeleteResult(
            service=self.config.service,
            key=self.config.key,
            verbose=self.config.verbose,
            output=self.config.output,
            success=success,
            error=error,
            config=config_dict or {}
        )
    
    def delete(self) -> DeleteResult:
        return self.execute()
    
    def execute(self) -> DeleteResult:
        if not self.config.key:
            return self._create_result(False, error=key_required_delete)
        
        if self.config.dry_run:
            return self._create_result(True)
        
        success, error = self.environment_service.delete_config(
            self.config.service, self.config.key, self.config.env_file
        )
        
        if success:
            self.logger.info(configuration_deleted.format(
                service=self.config.service, key=self.config.key
            ))
            return self._create_result(True)
        else:
            self.logger.error(configuration_delete_failed.format(
                service=self.config.service, error=error
            ))
            return self._create_result(False, error=error)
    
    def delete_and_format(self) -> str:
        return self.execute_and_format()
    
    def execute_and_format(self) -> str:
        if self.config.dry_run:
            return self._format_dry_run()
        
        result = self.execute()
        return self._format_output(result, self.config.output)
    
    def _format_dry_run(self) -> str:
        lines = [dry_run_mode]
        lines.append(dry_run_delete_config.format(
            service=self.config.service,
            key=self.config.key
        ))
        lines.append(end_dry_run)
        return "\n".join(lines)
    
    def _format_output(self, result: DeleteResult, output_format: str) -> str:
        if output_format == "json":
            return self._format_json(result)
        else:
            return self._format_text(result)
    
    def _format_json(self, result: DeleteResult) -> str:
        import json
        output = {
            "service": result.service,
            "key": result.key,
            "success": result.success,
            "error": result.error
        }
        return json.dumps(output, indent=2)
    
    def _format_text(self, result: DeleteResult) -> str:
        if not result.success:
            return configuration_delete_failed.format(service=result.service, error=result.error)
        
        return configuration_deleted.format(service=result.service, key=result.key)

class Delete(BaseAction[DeleteConfig, DeleteResult]):
    def __init__(self, logger: LoggerProtocol = None):
        super().__init__(logger)
    
    def delete(self, config: DeleteConfig) -> DeleteResult:
        return self.execute(config)
    
    def execute(self, config: DeleteConfig) -> DeleteResult:
        service = DeleteService(config, logger=self.logger)
        return service.execute()
    
    def format_output(self, result: DeleteResult, output: str) -> str:
        service = DeleteService(result, logger=self.logger)
        return service._format_output(result, output) 
