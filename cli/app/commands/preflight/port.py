import re, socket
from typing import List, TypedDict, Union, Any, Optional, Protocol
from pydantic import BaseModel, Field, field_validator
from .messages import available, not_available, error_checking_port, host_must_be_localhost_or_valid_ip_or_domain
from app.utils.logger import Logger
from app.utils.protocols import LoggerProtocol
from app.utils.lib import ParallelProcessor
from app.utils.output_formatter import OutputFormatter

class PortCheckerProtocol(Protocol):
    def check_port(self, port: int, config: "PortConfig") -> "PortCheckResult":
        ...

class PortCheckResult(TypedDict):
    port: int
    status: str
    host: Optional[str]
    error: Optional[str]
    is_available: bool

class PortConfig(BaseModel):
    ports: List[int] = Field(..., min_length=1, max_length=65535, description="List of ports to check")
    host: str = Field("localhost", min_length=1, description="Host to check")
    timeout: int = Field(1, gt=0, le=60, description="Timeout in seconds")
    verbose: bool = Field(False, description="Verbose output")

    @field_validator('host')
    @classmethod
    def validate_host(cls, v: str) -> str:
        if v.lower() == "localhost":
            return v
        ip_pattern = r'^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$'
        if re.match(ip_pattern, v):
            return v
        domain_pattern = r'^[a-zA-Z]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$'
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
                if item.get('is_available', False):
                    message = f"Port {item['port']}: {item['status']}"
                    messages.append(self.output_formatter.create_success_message(message, item))
                else:
                    error = f"Port {item['port']}: {item['status']}"
                    messages.append(self.output_formatter.create_error_message(error, item))
            return self.output_formatter.format_output(messages, output_type)
        else:
            return str(data)

class PortChecker:
    def __init__(self, logger: LoggerProtocol, timeout: int):
        self.logger = logger
        self.timeout = timeout
    
    def is_port_available(self, host: str, port: int) -> bool:
        try:
            with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
                sock.settimeout(self.timeout)
                result = sock.connect_ex((host, port))
                return result != 0
        except Exception:
            return False
    
    def check_port(self, port: int, config: PortConfig) -> PortCheckResult:
        self.logger.debug(f"Checking port {port} on host {config.host}")
        try:
            status = available if self.is_port_available(config.host, port) else not_available
            return self._create_result(port, config, status)
        except Exception as e:
            self.logger.error(error_checking_port.format(port=port, error=str(e)))
            return self._create_result(port, config, not_available, str(e))
    
    def _create_result(self, port: int, config: PortConfig, status: str, error: Optional[str] = None) -> PortCheckResult:
        return {
            "port": port,
            "status": status,
            "host": config.host if config.verbose else None,
            "error": error,
            "is_available": status == available
        }

class PortService:
    def __init__(self, config: PortConfig, logger: LoggerProtocol = None, checker: PortCheckerProtocol = None):
        self.config = config
        self.logger = logger or Logger(verbose=config.verbose)
        self.checker = checker or PortChecker(self.logger, config.timeout)
        self.formatter = PortFormatter()
    
    def check_ports(self) -> List[PortCheckResult]:
        self.logger.debug(f"Checking ports: {self.config.ports}")
        def process_port(port: int) -> PortCheckResult:
            return self.checker.check_port(port, self.config)
        def error_handler(port: int, error: Exception) -> PortCheckResult:
            self.logger.error(error_checking_port.format(port=port, error=str(error)))
            return self.checker._create_result(port, self.config, not_available, str(error))
        results = ParallelProcessor.process_items(
            items=self.config.ports,
            processor_func=process_port,
            max_workers=min(len(self.config.ports), 50),
            error_handler=error_handler
        )
        return sorted(results, key=lambda x: x["port"])
    
    def check_and_format(self, output_type: str) -> str:
        results = self.check_ports()
        return self.formatter.format_output(results, output_type)
