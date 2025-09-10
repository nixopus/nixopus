import os
from typing import Dict, Optional, Protocol

from pydantic import BaseModel, Field

from app.utils.logger import Logger
from app.utils.output_formatter import OutputFormatter
from app.utils.protocols import LoggerProtocol

from .base import BaseAction, BaseConfig, BaseEnvironmentManager, BaseResult, BaseService
from .messages import (
    configuration_list_failed,
    configuration_list_title,
    configuration_listed,
    debug_config_file_exists,
    debug_config_file_not_exists,
    debug_config_file_read_failed,
    debug_config_file_read_success,
    debug_config_listed,
    debug_dry_run_simulation,
    debug_dry_run_simulation_complete,
    debug_listing_config,
    debug_no_config_to_list,
    debug_service_env_file_resolved,
    dry_run_list_config,
    dry_run_mode,
    end_dry_run,
    no_configuration_found,
)


class EnvironmentServiceProtocol(Protocol):
    def list_config(self, service: str, env_file: str = None) -> tuple[bool, Dict[str, str], str]: ...


class EnvironmentManager(BaseEnvironmentManager):
    def list_config(self, service: str, env_file: Optional[str] = None) -> tuple[bool, Dict[str, str], Optional[str]]:
        file_path = self.get_service_env_file(service, env_file)
        self.logger.debug(debug_service_env_file_resolved.format(file_path=file_path))

        if self.logger.verbose:
            if os.path.exists(file_path):
                self.logger.debug(debug_config_file_exists.format(file_path=file_path))
            else:
                self.logger.debug(debug_config_file_not_exists.format(file_path=file_path))

        success, config_dict, error = self.read_env_file(file_path)

        if success:
            self.logger.debug(debug_config_file_read_success.format(count=len(config_dict)))
        else:
            self.logger.debug(debug_config_file_read_failed.format(error=error))

        return success, config_dict, error


class ListResult(BaseResult):
    pass


class ListConfig(BaseConfig):
    pass


class ListService(BaseService[ListConfig, ListResult]):
    def __init__(
        self, config: ListConfig, logger: LoggerProtocol = None, environment_service: EnvironmentServiceProtocol = None
    ):
        super().__init__(config, logger, environment_service)
        self.environment_service = environment_service or EnvironmentManager(self.logger)
        self.formatter = OutputFormatter()

    def _create_result(self, success: bool, error: str = None, config_dict: Dict[str, str] = None) -> ListResult:
        return ListResult(
            service=self.config.service,
            verbose=self.config.verbose,
            output=self.config.output,
            success=success,
            error=error,
            config=config_dict or {},
        )

    def list(self) -> ListResult:
        return self.execute()

    def execute(self) -> ListResult:
        self.logger.debug(debug_listing_config.format(service=self.config.service))

        if self.config.dry_run:
            self.logger.debug(debug_dry_run_simulation)
            result = self._create_result(True)
            self.logger.debug(debug_dry_run_simulation_complete)
            return result

        success, config_dict, error = self.environment_service.list_config(self.config.service, self.config.env_file)

        if success:
            if config_dict:
                self.logger.debug(debug_config_listed.format(count=len(config_dict)))
            else:
                self.logger.debug(debug_no_config_to_list)
            self.logger.info(configuration_listed.format(service=self.config.service))
            return self._create_result(True, config_dict=config_dict)
        else:
            self.logger.error(configuration_list_failed.format(service=self.config.service, error=error))
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
            formatted = self._format_json(result)
        else:
            formatted = self._format_text(result, output_format)

        return formatted

    def _format_json(self, result: ListResult) -> str:
        import json

        output = {"service": result.service, "success": result.success, "error": result.error, "config": result.config}
        return json.dumps(output, indent=2)

    def _format_text(self, result: ListResult, output_format: str) -> str:
        if not result.success:
            return configuration_list_failed.format(service=result.service, error=result.error)

        if result.config:
            success_message = configuration_listed.format(service=result.service)
            title = configuration_list_title.format(service=result.service)
            headers = ("Key", "Value")

            return self.formatter.format_table_output(
                data=result.config, output_format=output_format, success_message=success_message, title=title, headers=headers
            )
        else:
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

    def list_and_format(self, config: ListConfig) -> str:
        service = ListService(config, logger=self.logger)
        return service.execute_and_format()
