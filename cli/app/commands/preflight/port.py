import re
import socket
import json
from typing import Any, List, Optional, Protocol, TypedDict, Union

from pydantic import BaseModel, Field, field_validator

from app.utils.lib import ParallelProcessor
from app.utils.logger import Logger
from app.utils.output_formatter import OutputFormatter
from app.utils.protocols import LoggerProtocol

from .messages import available, error_checking_port, host_must_be_localhost_or_valid_ip_or_domain, not_available, invalid_output_format


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
    output: str = Field("text", description="Output format, text, json")

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

    def format_output(self, results: List[PortCheckResult], output: str) -> str:
        if output == "text":
            return self._format_text_output(results)
        elif output == "json":
            return self._format_json_output(results)
        else:
            raise ValueError(invalid_output_format.format(output=output))

    def _format_text_output(self, results: List[PortCheckResult]) -> str:   
        unavailable_ports = [r for r in results if not r["is_available"]]
        
        if not unavailable_ports:
            return "All ports are available"
        
        output_lines = ["Unavailable Ports:"]
        for result in sorted(unavailable_ports, key=lambda x: x["port"]):
            error_msg = f" ({result['error']})" if result.get("error") else ""
            output_lines.append(f"  • Port {result['port']}: {result['status']}{error_msg}")
        
        return "\n".join(output_lines)

    def _format_json_output(self, results: List[PortCheckResult]) -> str:
        available_ports = [r for r in results if r["is_available"]]
        unavailable_ports = [r for r in results if not r["is_available"]]
        
        summary = {
            "total_ports": len(results),
            "available_count": len(available_ports),
            "unavailable_count": len(unavailable_ports),
            "available_ports": [r["port"] for r in available_ports],
            "unavailable_ports": [{"port": r["port"], "error": r.get("error")} for r in unavailable_ports],
            "results": results
        }
        
        return json.dumps(summary, indent=2)


class PortChecker:
    def __init__(self, logger: LoggerProtocol):
        self.logger = logger

    def is_port_available(self, host: str, port: int) -> bool:
        try:
            with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
                sock.settimeout(1)
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
            "is_available": status == available,
        }


class PortService:
    def __init__(self, config: PortConfig, logger: LoggerProtocol = None, checker: PortCheckerProtocol = None):
        self.config = config
        self.logger = logger or Logger(verbose=config.verbose)
        self.checker = checker or PortChecker(self.logger)
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
            error_handler=error_handler,
        )
        return sorted(results, key=lambda x: x["port"])

    def check_and_format(self) -> str:
        results = self.check_ports()
        return self.formatter.format_output(results, self.config.output)
