from typing import Any, Dict, List

from app.utils.config import Config
from app.utils.protocols import LoggerProtocol

from .messages import ports_unavailable
from .port import PortConfig, PortService


class PreflightRunner:
    """Centralized preflight check runner for port availability"""

    def __init__(self, logger: LoggerProtocol = None, verbose: bool = False):
        self.logger = logger
        self.verbose = verbose
        self.config = Config()

    def run_port_checks(self, ports: List[int], host: str = "localhost") -> List[Dict[str, Any]]:
        """Run port availability checks and return results"""
        port_config = PortConfig(ports=ports, host=host, verbose=self.verbose)
        port_service = PortService(port_config, logger=self.logger)
        return port_service.check_ports()

    def check_required_ports(self, ports: List[int], host: str = "localhost") -> None:
        """Check required ports and raise exception if any are unavailable"""
        port_results = self.run_port_checks(ports, host)
        unavailable_ports = [result for result in port_results if not result.get("is_available", True)]

        if unavailable_ports:
            error_msg = f"{ports_unavailable}: {[p['port'] for p in unavailable_ports]}"
            raise Exception(error_msg)

    def check_ports_from_config(
        self, config_key: str = "required_ports", user_config: dict = None, defaults: dict = None
    ) -> None:
        """Check ports using configuration values"""
        if user_config is not None and defaults is not None:
            ports = self.config.get_config_value(config_key, user_config, defaults)
        else:
            ports = self.config.get_yaml_value("ports")

        self.check_required_ports(ports)
