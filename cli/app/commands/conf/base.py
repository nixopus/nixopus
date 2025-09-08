import os
import shutil
import tempfile
from typing import Dict, Generic, Optional, Protocol, TypeVar

from pydantic import BaseModel, Field, field_validator

from app.utils.config import API_ENV_FILE, VIEW_ENV_FILE, Config
from app.utils.logger import Logger
from app.utils.protocols import LoggerProtocol

from .messages import (
    backup_created,
    backup_creation_failed,
    backup_file_not_found,
    backup_remove_failed,
    backup_removed,
    backup_restore_attempt,
    backup_restore_failed,
    backup_restore_success,
    file_not_exists,
    file_not_found,
    file_read_failed,
    file_write_failed,
    invalid_line_warning,
    invalid_service,
    read_error,
    read_success,
    reading_env_file,
)

TConfig = TypeVar("TConfig", bound=BaseModel)
TResult = TypeVar("TResult", bound=BaseModel)


class EnvironmentServiceProtocol(Protocol):
    def list_config(self, service: str, env_file: str = None) -> tuple[bool, Dict[str, str], str]: ...

    def set_config(self, service: str, key: str, value: str, env_file: str = None) -> tuple[bool, str]: ...

    def delete_config(self, service: str, key: str, env_file: str = None) -> tuple[bool, str]: ...


class BaseEnvironmentManager:
    def __init__(self, logger: LoggerProtocol):
        self.logger = logger

    def read_env_file(self, file_path: str) -> tuple[bool, Dict[str, str], Optional[str]]:
        self.logger.debug(reading_env_file.format(file_path=file_path))
        try:
            if not os.path.exists(file_path):
                self.logger.debug(file_not_exists.format(file_path=file_path))
                return False, {}, file_not_found.format(path=file_path)

            config = {}
            with open(file_path, "r") as f:
                for line_num, line in enumerate(f, 1):
                    line = line.strip()
                    if not line or line.startswith("#"):
                        continue

                    if "=" not in line:
                        self.logger.warning(invalid_line_warning.format(line_num=line_num, file_path=file_path, line=line))
                        continue

                    key, value = line.split("=", 1)
                    config[key.strip()] = value.strip()

            self.logger.debug(read_success.format(count=len(config), file_path=file_path))
            return True, config, None
        except Exception as e:
            self.logger.debug(read_error.format(file_path=file_path, error=e))
            return False, {}, file_read_failed.format(error=e)

    def _create_backup(self, file_path: str) -> tuple[bool, Optional[str], Optional[str]]:
        if not os.path.exists(file_path):
            return True, None, None

        try:
            backup_path = f"{file_path}.backup"
            shutil.copy2(file_path, backup_path)
            return True, backup_path, None
        except Exception as e:
            return False, None, backup_creation_failed.format(error=e)

    def _restore_backup(self, backup_path: str, file_path: str) -> tuple[bool, Optional[str]]:
        try:
            if os.path.exists(backup_path):
                shutil.copy2(backup_path, file_path)
                os.remove(backup_path)
                return True, None
            return False, backup_file_not_found.format(path=backup_path)
        except Exception as e:
            return False, backup_restore_failed.format(error=e)

    def _atomic_write(self, file_path: str, config: Dict[str, str]) -> tuple[bool, Optional[str]]:
        temp_path = None
        try:
            os.makedirs(os.path.dirname(file_path), exist_ok=True)

            with tempfile.NamedTemporaryFile(mode="w", delete=False, dir=os.path.dirname(file_path)) as temp_file:
                for key, value in sorted(config.items()):
                    temp_file.write(f"{key}={value}\n")
                temp_file.flush()
                try:
                    os.fsync(temp_file.fileno())
                except (OSError, AttributeError):
                    pass
                temp_path = temp_file.name

            os.replace(temp_path, file_path)
            return True, None
        except Exception as e:
            if temp_path and os.path.exists(temp_path):
                try:
                    os.unlink(temp_path)
                except:
                    pass
            return False, file_write_failed.format(error=e)

    def write_env_file(self, file_path: str, config: Dict[str, str]) -> tuple[bool, Optional[str]]:
        backup_created_flag = False
        backup_path = None

        try:
            success, backup_path, error = self._create_backup(file_path)
            if not success:
                return False, error

            backup_created_flag = True
            self.logger.debug(backup_created.format(backup_path=backup_path))

            success, error = self._atomic_write(file_path, config)
            if not success:
                if backup_created_flag and backup_path:
                    self.logger.warning(backup_restore_attempt)
                    restore_success, restore_error = self._restore_backup(backup_path, file_path)
                    if restore_success:
                        self.logger.debug(backup_restore_success)
                    else:
                        self.logger.error(backup_restore_failed.format(error=restore_error))
                return False, error

            if backup_created_flag and backup_path and os.path.exists(backup_path):
                try:
                    os.remove(backup_path)
                    self.logger.debug(backup_removed)
                except Exception as e:
                    self.logger.warning(backup_remove_failed.format(error=e))

            return True, None

        except Exception as e:
            return False, file_write_failed.format(error=e)

    def get_service_env_file(self, service: str, env_file: Optional[str] = None) -> str:
        if env_file:
            return env_file

        config = Config()
        if service == "api":
            default_path = config.get_yaml_value(API_ENV_FILE)
            return default_path
        elif service == "view":
            default_path = config.get_yaml_value(VIEW_ENV_FILE)
            return default_path
        else:
            raise ValueError(invalid_service.format(service=service))


class BaseConfig(BaseModel):
    service: str = Field("api", description="The name of the service to manage configuration for")
    key: Optional[str] = Field(None, description="The configuration key")
    value: Optional[str] = Field(None, description="The configuration value")
    verbose: bool = Field(False, description="Verbose output")
    output: str = Field("text", description="Output format: text, json")
    dry_run: bool = Field(False, description="Dry run mode")
    env_file: Optional[str] = Field(None, description="Path to the environment file")

    @field_validator("env_file")
    @classmethod
    def validate_env_file(cls, env_file: str) -> Optional[str]:
        if not env_file:
            return None
        stripped_env_file = env_file.strip()
        if not stripped_env_file:
            return None
        if not os.path.exists(stripped_env_file):
            raise ValueError(file_not_found.format(path=stripped_env_file))
        return stripped_env_file


class BaseResult(BaseModel):
    service: str
    key: Optional[str] = None
    value: Optional[str] = None
    config: Dict[str, str] = Field(default_factory=dict)
    verbose: bool
    output: str
    success: bool = False
    error: Optional[str] = None


class BaseService(Generic[TConfig, TResult]):
    def __init__(self, config: TConfig, logger: LoggerProtocol = None, environment_service: EnvironmentServiceProtocol = None):
        self.config = config
        self.logger = logger or Logger(verbose=config.verbose)
        self.environment_service = environment_service
        self.formatter = None

    def _create_result(self, success: bool, error: str = None, config_dict: Dict[str, str] = None) -> TResult:
        raise NotImplementedError

    def execute(self) -> TResult:
        raise NotImplementedError

    def execute_and_format(self) -> str:
        raise NotImplementedError


class BaseAction(Generic[TConfig, TResult]):
    def __init__(self, logger: LoggerProtocol = None):
        self.logger = logger
        self.formatter = None

    def execute(self, config: TConfig) -> TResult:
        raise NotImplementedError

    def format_output(self, result: TResult, output: str) -> str:
        raise NotImplementedError
