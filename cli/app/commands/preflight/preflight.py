import shutil
import socket
from typing import Any, Dict, List, Optional

from app.utils.config import get_active_config, get_config_value
from app.utils.parallel_processor import process_parallel
from app.utils.protocols import LoggerProtocol

from .messages import (
    available,
    debug_dep_check_result,
    debug_deps_check_completed,
    debug_port_check_result,
    debug_processing_deps,
    debug_processing_ports,
    error_checking_dependency,
    error_checking_port,
    error_socket_connection_failed,
    not_available,
    ports_unavailable,
)


def _is_port_available(host: str, port: int, logger: Optional[LoggerProtocol] = None) -> bool:
    """Check if a port is available on the given host."""
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
            sock.settimeout(1)
            result = sock.connect_ex((host, port))
            return result != 0
    except Exception as e:
        if logger and logger.verbose:
            logger.error(error_socket_connection_failed.format(port=port, error=e))
        return False


def _check_port(port: int, host: str, logger: Optional[LoggerProtocol] = None) -> Dict[str, Any]:
    """Check a single port and return result."""
    try:
        is_avail = _is_port_available(host, port, logger)
        status = available if is_avail else not_available
        
        if logger:
            logger.debug(debug_port_check_result.format(port=port, status=status))
        
        return {
            "port": port,
            "status": status,
            "host": host if host != "localhost" else None,
            "error": None,
            "is_available": is_avail,
        }
    except Exception as e:
        if logger and logger.verbose:
            logger.error(error_checking_port.format(port=port, error=str(e)))
        return {
            "port": port,
            "status": not_available,
            "host": host if host != "localhost" else None,
            "error": str(e),
            "is_available": False,
        }


def check_ports(ports: List[int], host: str = "localhost", logger: Optional[LoggerProtocol] = None) -> List[Dict[str, Any]]:
    """Check multiple ports and return results."""
    if logger:
        logger.debug(debug_processing_ports.format(count=len(ports)))

    def process_port(port: int) -> Dict[str, Any]:
        return _check_port(port, host, logger)

    def error_handler(port: int, error: Exception) -> Dict[str, Any]:
        if logger and logger.verbose:
            logger.error(error_checking_port.format(port=port, error=str(error)))
        return {
            "port": port,
            "status": not_available,
            "host": host if host != "localhost" else None,
            "error": str(error),
            "is_available": False,
        }

    results = process_parallel(
        items=ports,
        processor_func=process_port,
        max_workers=min(len(ports), 50),
        error_handler=error_handler,
    )
    return sorted(results, key=lambda x: x["port"])


def check_required_ports(ports: List[int], host: str = "localhost", logger: Optional[LoggerProtocol] = None) -> None:
    """Check required ports and raise exception if any are unavailable."""
    port_results = check_ports(ports, host, logger)
    unavailable_ports = [result for result in port_results if not result.get("is_available", True)]

    if unavailable_ports:
        error_msg = f"{ports_unavailable}: {[p['port'] for p in unavailable_ports]}"
        raise Exception(error_msg)


def check_ports_from_config(logger: Optional[LoggerProtocol] = None) -> None:
    """Check ports using configuration values."""
    config = get_active_config()
    ports = get_config_value(config, "ports")
    ports = [int(port) for port in ports] if isinstance(ports, list) else [int(ports)]
    check_required_ports(ports, logger=logger)


def _check_dependency(dep: str, logger: Optional[LoggerProtocol] = None) -> Dict[str, Any]:
    """Check if a dependency (command) is available."""
    try:
        is_available = shutil.which(dep) is not None
        status = available if is_available else not_available
        
        if logger:
            logger.debug(debug_dep_check_result.format(dep=dep, status=status))
        
        return {
            "dependency": dep,
            "status": status,
            "is_available": is_available,
            "error": None,
        }
    except Exception as e:
        if logger and logger.verbose:
            logger.error(error_checking_dependency.format(dep=dep, error=str(e)))
        return {
            "dependency": dep,
            "status": not_available,
            "is_available": False,
            "error": str(e),
        }


def check_dependencies(deps: List[str], logger: Optional[LoggerProtocol] = None) -> List[Dict[str, Any]]:
    """Check multiple dependencies and return results."""
    if logger:
        logger.debug(debug_processing_deps.format(count=len(deps)))

    def process_dep(dep: str) -> Dict[str, Any]:
        return _check_dependency(dep, logger)

    def error_handler(dep: str, error: Exception) -> Dict[str, Any]:
        if logger and logger.verbose:
            logger.error(error_checking_dependency.format(dep=dep, error=str(error)))
        return {
            "dependency": dep,
            "status": not_available,
            "is_available": False,
            "error": str(error),
        }

    results = process_parallel(
        items=deps,
        processor_func=process_dep,
        max_workers=min(len(deps), 50),
        error_handler=error_handler,
    )
    
    if logger:
        logger.debug(debug_deps_check_completed)
    
    return results


def check_required_dependencies(deps: List[str], logger: Optional[LoggerProtocol] = None) -> None:
    """Check required dependencies and raise exception if any are unavailable."""
    dep_results = check_dependencies(deps, logger)
    unavailable_deps = [result for result in dep_results if not result.get("is_available", True)]

    if unavailable_deps:
        error_msg = f"Dependencies unavailable: {[d['dependency'] for d in unavailable_deps]}"
        raise Exception(error_msg)

