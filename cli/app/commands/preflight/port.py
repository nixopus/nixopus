import re
import socket
from typing import Any, List, Optional, Protocol, TypedDict, Union

from pydantic import BaseModel, Field, field_validator

from app.utils.lib import ParallelProcessor
from app.utils.logger import Logger
from app.utils.output_formatter import OutputFormatter
from app.utils.protocols import LoggerProtocol

from .messages import (
    available,
    error_checking_port,
    host_must_be_localhost_or_valid_ip_or_domain,
    not_available,
    debug_processing_ports,
    debug_port_check_result,
    error_socket_connection_failed,
)


class PortCheckerProtocol(Protocol):
    def check_port(self, port: int, config: "PortConfig") -> "PortCheckResult": ...


class PortCheckResult(TypedDict):
    port: int
    status: str
    host: Optional[str]
    error: Optional[str]
    is_available: bool


class PortConfig(BaseModel):
    ports: List[int] = Field(..., min_length=1, max_length=65535, description="List of ports to check")
    host: str = Field("localhost", min_length=1, description="Host to check")
    verbose: bool = Field(False, description="Verbose output")

    @field_validator("host")
    @classmethod
    def validate_host(cls, v: str) -> str:
        if v.lower() == "localhost":
            return v
        ip_pattern = r"^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$"
        if re.match(ip_pattern, v):
            return v
        domain_pattern = r"^[a-zA-Z]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$"
        if re.match(domain_pattern, v):
            return v
        raise ValueError(host_must_be_localhost_or_valid_ip_or_domain)


class PortFormatter:
    def __init__(self):
        self.output_formatter = OutputFormatter()

    def format_output(self, data: Union[str, List[PortCheckResult], Any], output_type: str) -> str:
        if isinstance(data, list):
            messages = []
            for item in data:
                if item.get("is_available", False):
                    message = f"Port {item['port']}: {item['status']}"
                    data = {"port": item['port'], "status": item['status'], "is_available": item['is_available']}
                    messages.append(self.output_formatter.create_success_message(message, data))
                else:
                    error = f"Port {item['port']}: {item['status']}"
                    data = {"port": item['port'], "status": item['status'], "is_available": item['is_available']}
                    if item.get('error'):
                        data['error'] = item['error']
                    messages.append(self.output_formatter.create_error_message(error, data))
            return self.output_formatter.format_output(messages, output_type)
        else:
            return str(data)


class PortChecker:
    def __init__(self, logger: LoggerProtocol):
        self.logger = logger

    def is_port_available(self, host: str, port: int) -> bool:
        try:
            with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
                sock.settimeout(1)
                result = sock.connect_ex((host, port))
                return result != 0
        except Exception as e:
            if self.logger.verbose:
                self.logger.error(error_socket_connection_failed.format(port=port, error=e))
            return False

    def check_port(self, port: int, config: PortConfig) -> PortCheckResult:
        try:
            status = available if self.is_port_available(config.host, port) else not_available
            self.logger.debug(debug_port_check_result.format(port=port, status=status))
            return self._create_result(port, config, status)
        except Exception as e:
            if self.logger.verbose:
                self.logger.error(error_checking_port.format(port=port, error=str(e)))
            return self._create_result(port, config, not_available, str(e))

    def _create_result(self, port: int, config: PortConfig, status: str, error: Optional[str] = None) -> PortCheckResult:
        return {
            "port": port,
            "status": status,
            "host": config.host if config.verbose else None,
            "error": error,
            "is_available": status == available,
        }


class PortService:
    def __init__(self, config: PortConfig, logger: LoggerProtocol = None, checker: PortCheckerProtocol = None):
        self.config = config
        self.logger = logger or Logger(verbose=config.verbose)
        self.checker = checker or PortChecker(self.logger)
        self.formatter = PortFormatter()

    def check_ports(self) -> List[PortCheckResult]:
        self.logger.debug(debug_processing_ports.format(count=len(self.config.ports)))

        def process_port(port: int) -> PortCheckResult:
            return self.checker.check_port(port, self.config)

        def error_handler(port: int, error: Exception) -> PortCheckResult:
            if self.logger.verbose:
                self.logger.error(error_checking_port.format(port=port, error=str(error)))
            return self.checker._create_result(port, self.config, not_available, str(error))

        results = ParallelProcessor.process_items(
            items=self.config.ports,
            processor_func=process_port,
            max_workers=min(len(self.config.ports), 50),
            error_handler=error_handler,
        )
        return sorted(results, key=lambda x: x["port"])

    def check_and_format(self, output_type: str) -> str:
        results = self.check_ports()
        return self.formatter.format_output(results, output_type)
