import json
import os
import subprocess
from typing import Generic, Optional, Protocol, TypeVar

import requests
from pydantic import BaseModel, Field, field_validator

from app.utils.config import Config, PROXY_PORT, CONFIG_ENDPOINT, LOAD_ENDPOINT, STOP_ENDPOINT, CADDY_BASE_URL
from app.utils.logger import Logger
from app.utils.output_formatter import OutputFormatter
from app.utils.protocols import LoggerProtocol

from .messages import (
    caddy_connection_failed,
    caddy_load_failed,
    caddy_status_code_error,
    config_file_not_found,
    info_caddy_running,
    info_caddy_stopped,
    info_config_loaded,
    invalid_json_config,
    port_must_be_between_1_and_65535,
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
            cmd = command_builder.build_load_command(getattr(config, "config_file", ""), getattr(config, "proxy_port", proxy_port))
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
            response = requests.get(url, timeout=5)
            if response.status_code == 200:
                return True, info_caddy_running
            else:
                return False, caddy_status_code_error.format(code=response.status_code)
        except requests.exceptions.RequestException as e:
            return False, caddy_connection_failed.format(error=str(e))
        except Exception as e:
            return False, f"Unexpected error: {str(e)}"

    def load_config(self, config_file: str, port: int = proxy_port) -> tuple[bool, str]:
        try:
            with open(config_file, "r") as f:
                config_data = json.load(f)

            url = self._get_caddy_url(port, caddy_load_endpoint)
            response = requests.post(url, json=config_data, headers={"Content-Type": "application/json"}, timeout=10)

            if response.status_code == 200:
                return True, info_config_loaded
            else:
                return False, caddy_load_failed.format(code=response.status_code, response=response.text)
        except FileNotFoundError:
            return False, config_file_not_found.format(file=config_file)
        except json.JSONDecodeError as e:
            return False, invalid_json_config.format(error=str(e))
        except requests.exceptions.RequestException as e:
            return False, caddy_connection_failed.format(error=str(e))
        except Exception as e:
            return False, f"Unexpected error: {str(e)}"

    def stop_proxy(self, port: int = proxy_port) -> tuple[bool, str]:
        try:
            url = self._get_caddy_url(port, caddy_stop_endpoint)
            response = requests.post(url, timeout=5)
            if response.status_code == 200:
                return True, info_caddy_stopped
            else:
                return False, f"Failed to stop Caddy: {response.status_code}"
        except requests.exceptions.RequestException as e:
            return False, caddy_connection_failed.format(error=str(e))
        except Exception as e:
            return False, f"Unexpected error: {str(e)}"


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
