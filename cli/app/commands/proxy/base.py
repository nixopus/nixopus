import json
from typing import Generic, Optional, Protocol, TypeVar

import requests
from pydantic import BaseModel, Field, field_validator

from app.utils.config import CADDY_BASE_URL, CONFIG_ENDPOINT, LOAD_ENDPOINT, PROXY_PORT, STOP_ENDPOINT, Config
from app.utils.logger import Logger
from app.utils.output_formatter import OutputFormatter
from app.utils.protocols import LoggerProtocol

from .messages import (
    caddy_connection_failed,
    caddy_is_running,
    caddy_not_running,
    cannot_connect_to_caddy,
    config_file_not_found,
    debug_caddy_config_accessible,
    debug_caddy_load_failed,
    debug_caddy_load_response,
    debug_caddy_non_200,
    debug_caddy_response,
    debug_caddy_stop_failed,
    debug_caddy_stop_response,
    debug_caddy_stopped_success,
    debug_checking_caddy_status,
    debug_config_loaded_success,
    debug_config_parsed,
    debug_connection_refused,
    debug_loading_config_file,
    debug_posting_config,
    debug_request_failed,
    debug_stopping_caddy,
    debug_unexpected_error,
    http_error,
    invalid_json_error,
    port_must_be_between_1_and_65535,
    request_failed_error,
    unexpected_error,
)

TConfig = TypeVar("TConfig", bound=BaseModel)
TResult = TypeVar("TResult", bound=BaseModel)

config = Config()
proxy_port = config.get_yaml_value(PROXY_PORT)
caddy_config_endpoint = config.get_yaml_value(CONFIG_ENDPOINT)
caddy_load_endpoint = config.get_yaml_value(LOAD_ENDPOINT)
caddy_stop_endpoint = config.get_yaml_value(STOP_ENDPOINT)
caddy_base_url = config.get_yaml_value(CADDY_BASE_URL)


class CaddyServiceProtocol(Protocol):
    def check_status(self, port: int = proxy_port) -> tuple[bool, str]: ...

    def load_config(self, config_file: str, port: int = proxy_port) -> tuple[bool, str]: ...

    def stop_proxy(self, port: int = proxy_port) -> tuple[bool, str]: ...


class BaseCaddyCommandBuilder:
    @staticmethod
    def build_status_command(port: int = proxy_port) -> list[str]:
        return ["curl", "-X", "GET", f"{caddy_base_url.format(port=port)}{caddy_config_endpoint}"]

    @staticmethod
    def build_load_command(config_file: str, port: int = proxy_port) -> list[str]:
        return [
            "curl",
            "-X",
            "POST",
            f"{caddy_base_url.format(port=port)}{caddy_load_endpoint}",
            "-H",
            "Content-Type: application/json",
            "-d",
            f"@{config_file}",
        ]

    @staticmethod
    def build_stop_command(port: int = proxy_port) -> list[str]:
        return ["curl", "-X", "POST", f"{caddy_base_url.format(port=port)}{caddy_stop_endpoint}"]


class BaseFormatter:
    def __init__(self):
        self.output_formatter = OutputFormatter()

    def format_output(self, result: TResult, output: str, success_message: str, error_message: str) -> str:
        if result.success:
            message = success_message.format(port=result.proxy_port)
            output_message = self.output_formatter.create_success_message(message, result.model_dump())
        else:
            error = result.error or "Unknown error occurred"
            output_message = self.output_formatter.create_error_message(error, result.model_dump())

        return self.output_formatter.format_output(output_message, output)

    def format_dry_run(self, config: TConfig, command_builder, dry_run_messages: dict) -> str:
        if hasattr(command_builder, "build_status_command"):
            cmd = command_builder.build_status_command(getattr(config, "proxy_port", proxy_port))
        elif hasattr(command_builder, "build_load_command"):
            cmd = command_builder.build_load_command(
                getattr(config, "config_file", ""), getattr(config, "proxy_port", proxy_port)
            )
        elif hasattr(command_builder, "build_stop_command"):
            cmd = command_builder.build_stop_command(getattr(config, "proxy_port", proxy_port))
        else:
            cmd = command_builder.build_command(config)

        output = []
        output.append(dry_run_messages["mode"])
        output.append(dry_run_messages["command_would_be_executed"])
        output.append(f"{dry_run_messages['command']} {' '.join(cmd)}")
        output.append(f"{dry_run_messages['port']} {getattr(config, 'proxy_port', proxy_port)}")

        if hasattr(config, "config_file") and getattr(config, "config_file", None):
            output.append(f"{dry_run_messages['config_file']} {getattr(config, 'config_file')}")

        output.append(dry_run_messages["end"])
        return "\n".join(output)


class BaseCaddyService:
    def __init__(self, logger: LoggerProtocol):
        self.logger = logger

    def _get_caddy_url(self, port: int, endpoint: str) -> str:
        return f"{caddy_base_url.format(port=port)}{endpoint}"

    def check_status(self, port: int = proxy_port) -> tuple[bool, str]:
        try:
            url = self._get_caddy_url(port, caddy_config_endpoint)
            self.logger.debug(debug_checking_caddy_status.format(url=url))

            response = requests.get(url, timeout=5)
            self.logger.debug(debug_caddy_response.format(code=response.status_code))

            if response.status_code == 200:
                self.logger.debug(debug_caddy_config_accessible)
                return True, caddy_is_running
            else:
                self.logger.debug(debug_caddy_non_200.format(code=response.status_code))
                return False, http_error.format(code=response.status_code)
        except requests.exceptions.ConnectionError:
            self.logger.debug(debug_connection_refused.format(port=port))
            return False, caddy_not_running
        except requests.exceptions.RequestException as e:
            self.logger.debug(debug_request_failed.format(error=str(e)))
            return False, request_failed_error.format(error=str(e))
        except Exception as e:
            self.logger.debug(debug_unexpected_error.format(error=str(e)))
            return False, unexpected_error.format(error=str(e))

    def load_config(self, config_file: str, port: int = proxy_port) -> tuple[bool, str]:
        try:
            self.logger.debug(debug_loading_config_file.format(file=config_file))
            with open(config_file, "r") as f:
                config_data = json.load(f)
            self.logger.debug(debug_config_parsed)

            url = self._get_caddy_url(port, caddy_load_endpoint)
            self.logger.debug(debug_posting_config.format(url=url))

            response = requests.post(url, json=config_data, headers={"Content-Type": "application/json"}, timeout=10)
            self.logger.debug(debug_caddy_load_response.format(code=response.status_code))

            if response.status_code == 200:
                self.logger.debug(debug_config_loaded_success)
                return True, "Configuration loaded"
            else:
                error_msg = response.text.strip() if response.text else http_error.format(code=response.status_code)
                self.logger.debug(debug_caddy_load_failed.format(error=error_msg))
                return False, error_msg
        except FileNotFoundError:
            error_msg = config_file_not_found.format(file=config_file)
            self.logger.debug(error_msg)
            return False, error_msg
        except json.JSONDecodeError as e:
            error_msg = invalid_json_error.format(error=str(e))
            self.logger.debug(error_msg)
            return False, error_msg
        except requests.exceptions.ConnectionError:
            error_msg = caddy_connection_failed.format(error=str(e))
            self.logger.debug(error_msg)
            return False, error_msg
        except requests.exceptions.RequestException as e:
            error_msg = request_failed_error.format(error=str(e))
            self.logger.debug(error_msg)
            return False, error_msg
        except Exception as e:
            error_msg = unexpected_error.format(error=str(e))
            self.logger.debug(error_msg)
            return False, error_msg

    def stop_proxy(self, port: int = proxy_port) -> tuple[bool, str]:
        try:
            url = self._get_caddy_url(port, caddy_stop_endpoint)
            self.logger.debug(debug_stopping_caddy.format(url=url))

            response = requests.post(url, timeout=5)
            self.logger.debug(debug_caddy_stop_response.format(code=response.status_code))

            if response.status_code == 200:
                self.logger.debug(debug_caddy_stopped_success)
                return True, "Caddy stopped"
            else:
                error_msg = http_error.format(code=response.status_code)
                self.logger.debug(debug_caddy_stop_failed.format(error=error_msg))
                return False, error_msg
        except requests.exceptions.ConnectionError:
            error_msg = cannot_connect_to_caddy.format(port=port)
            self.logger.debug(error_msg)
            return False, error_msg
        except requests.exceptions.RequestException as e:
            error_msg = request_failed_error.format(error=str(e))
            self.logger.debug(error_msg)
            return False, error_msg
        except Exception as e:
            error_msg = unexpected_error.format(error=str(e))
            self.logger.debug(error_msg)
            return False, error_msg


class BaseConfig(BaseModel):
    proxy_port: int = Field(proxy_port, description="Caddy admin port")
    verbose: bool = Field(False, description="Verbose output")
    output: str = Field("text", description="Output format: text, json")
    dry_run: bool = Field(False, description="Dry run mode")

    @field_validator("proxy_port")
    @classmethod
    def validate_proxy_port(cls, port: int) -> int:
        if port < 1 or port > 65535:
            raise ValueError(port_must_be_between_1_and_65535)
        return port


class BaseResult(BaseModel):
    proxy_port: int
    verbose: bool
    output: str
    success: bool = False
    error: Optional[str] = None


class BaseService(Generic[TConfig, TResult]):
    def __init__(self, config: TConfig, logger: LoggerProtocol = None, caddy_service: CaddyServiceProtocol = None):
        self.config = config
        self.logger = logger or Logger(verbose=config.verbose)
        self.caddy_service = caddy_service
        self.formatter = None

    def _create_result(self, success: bool, error: str = None) -> TResult:
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
