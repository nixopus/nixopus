import json
import subprocess
import time
from dataclasses import dataclass
from enum import Enum
from typing import Dict, List, Optional, Tuple

from app.utils.protocols import LoggerProtocol


class HealthStatus(Enum):
    HEALTHY = "healthy"
    UNHEALTHY = "unhealthy"
    STARTING = "starting"
    NONE = "none"   
    UNKNOWN = "unknown"

@dataclass
class ContainerHealth:
    name: str
    status: HealthStatus
    running: bool
    exit_code: Optional[int] = None
    error: Optional[str] = None


def _run_docker_command(cmd: List[str], logger: Optional[LoggerProtocol]) -> Tuple[bool, str]:
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, check=False)
        
        if result.returncode != 0:
            error = result.stderr.strip() or "Command failed"
            if logger:
                logger.debug(f"Command failed: {' '.join(cmd)}")
                logger.debug(f"Error: {error}")
            return False, error
        
        return True, result.stdout.strip()
    
    except Exception as e:
        if logger:
            logger.debug(f"Exception running command: {str(e)}")
        return False, str(e)


def _parse_container_inspect_output(output: str) -> Tuple[bool, Optional[int], str]:
    parts = [p.strip() for p in output.split("|")]
    
    is_running = parts[0].lower() == "true" if parts else False
    
    exit_code = None
    if len(parts) > 1 and parts[1]:
        try:
            exit_code = int(parts[1])
        except ValueError:
            pass
    
    health_str = parts[2] if len(parts) > 2 else ""
    
    return is_running, exit_code, health_str


def _map_health_string_to_enum(health_str: str) -> HealthStatus:
    if not health_str or health_str == "<no value>":
        return HealthStatus.NONE
    
    try:
        return HealthStatus(health_str)
    except ValueError:
        return HealthStatus.UNKNOWN


def get_container_health_status(
    container_name: str,
    logger: Optional[LoggerProtocol] = None,
) -> ContainerHealth:
    if logger:
        logger.debug(f"Checking health of container: {container_name}")
    
    cmd = [
        "docker", "inspect",
        "--format", "{{.State.Running}}|{{.State.ExitCode}}|{{.State.Health.Status}}",
        container_name
    ]
    
    success, output = _run_docker_command(cmd, logger)
    
    # Early return: Container not found or inspect failed
    if not success:
        return ContainerHealth(
            name=container_name,
            status=HealthStatus.UNKNOWN,
            running=False,
            error=output
        )
    
    is_running, exit_code, health_str = _parse_container_inspect_output(output)
    health_status = _map_health_string_to_enum(health_str)
    
    if logger:
        logger.debug(
            f"Container {container_name}: running={is_running}, "
            f"exit_code={exit_code}, health={health_status.value}"
        )
    
    return ContainerHealth(
        name=container_name,
        status=health_status,
        running=is_running,
        exit_code=exit_code
    )


def get_compose_services(
    compose_file: str,
    profiles: Optional[List[str]] = None,
    logger: Optional[LoggerProtocol] = None,
) -> List[str]:
    cmd = ["docker", "compose", "-f", compose_file]
    
    if profiles:
        for profile in profiles:
            cmd.extend(["--profile", profile])
    
    cmd.extend(["config", "--services"])
    
    success, output = _run_docker_command(cmd, logger)
    
    # Early return: Failed to get services
    if not success:
        if logger:
            logger.debug(f"Failed to get services from compose file: {output}")
        return []
    
    services = [s.strip() for s in output.split("\n") if s.strip()]
    
    if logger:
        logger.debug(f"Found services: {', '.join(services)}")
    
    return services


def get_container_name_for_service(
    service_name: str,
    compose_file: str,
    logger: Optional[LoggerProtocol] = None,
) -> Optional[str]:
    cmd = [
        "docker", "compose", "-f", compose_file,
        "ps", "--format", "json", service_name
    ]
    
    success, output = _run_docker_command(cmd, logger)
    
    # Early return: Service not found
    if not success or not output:
        if logger:
            logger.debug(f"Service {service_name} not found or not running")
        return None
    
    # Parse first line of JSON output (may have multiple containers per service)
    lines = output.split("\n")
    if not lines:
        return None
    
    try:
        data = json.loads(lines[0])
        container_name = data.get("Name")
        
        if logger:
            logger.debug(f"Service {service_name} -> container {container_name}")
        
        return container_name
    
    except json.JSONDecodeError:
        if logger:
            logger.debug(f"Failed to parse container info for {service_name}")
        return None


def _is_container_healthy(health: ContainerHealth, logger: Optional[LoggerProtocol]) -> bool:
    if not health.running:
        if logger:
            logger.debug(f"Container {health.name} is not running")
        return False
    
    if health.status == HealthStatus.HEALTHY:
        return True
    
    if health.status == HealthStatus.NONE:
        if logger:
            logger.debug(f"Container {health.name} has no healthcheck (assuming healthy)")
        return True
    
    return False


def _is_container_unhealthy(health: ContainerHealth, logger: Optional[LoggerProtocol]) -> bool:
    if health.error == "Service container not found":
        if logger:
            logger.debug(f"Service {health.name} container not found yet (still starting)")
        return False
    
    if not health.running:
        if logger:
            logger.debug(f"Container {health.name} is not running")
        return True
    
    if health.status == HealthStatus.UNHEALTHY:
        if logger:
            logger.debug(f"Container {health.name} is unhealthy")
        return True
    
    if health.status == HealthStatus.UNKNOWN:
        if logger:
            logger.debug(f"Container {health.name} has unknown status")
        return True
    
    return False


def _evaluate_all_container_health(
    health_statuses: Dict[str, ContainerHealth],
    logger: Optional[LoggerProtocol]
) -> Tuple[bool, bool]:
    all_healthy = True
    any_unhealthy = False
    
    for service, health in health_statuses.items():
        if _is_container_unhealthy(health, logger):
            any_unhealthy = True
            all_healthy = False
            continue
        
        if not _is_container_healthy(health, logger):
            all_healthy = False
    
    return all_healthy, any_unhealthy


def _discover_container_names(
    services: List[str],
    compose_file: str,
    logger: Optional[LoggerProtocol],
) -> Dict[str, Optional[str]]:
    """Discover container names for all services. Returns None for services without containers."""
    container_names: Dict[str, Optional[str]] = {}
    for service in services:
        container_name = get_container_name_for_service(service, compose_file, logger)
        container_names[service] = container_name
    return container_names


def _check_all_containers_health(
    container_names: Dict[str, Optional[str]],
    logger: Optional[LoggerProtocol],
) -> Dict[str, ContainerHealth]:
    health_statuses: Dict[str, ContainerHealth] = {}
    for service, container_name in container_names.items():
        if container_name is None:
            health_statuses[service] = ContainerHealth(
                name=service,
                status=HealthStatus.UNKNOWN,
                running=False,
                error="Service container not found"
            )
        else:
            health_statuses[service] = get_container_health_status(container_name, logger)
    return health_statuses


def _log_if_enabled(logger: Optional[LoggerProtocol], level: str, message: str) -> None:
    """Helper to reduce repetitive logger checks."""
    if logger:
        if level == "debug":
            logger.debug(message)
        elif level == "error":
            logger.error(message)


def wait_for_healthy_services(
    compose_file: str,
    timeout: int = 120,
    check_interval: int = 2,
    profiles: Optional[List[str]] = None,
    logger: Optional[LoggerProtocol] = None,
) -> Tuple[bool, Dict[str, ContainerHealth]]:
    _log_if_enabled(logger, "debug", f"Waiting for services to become healthy (timeout: {timeout}s)")
    
    # Discover services from compose file
    services = get_compose_services(compose_file, profiles, logger)
    if not services:
        _log_if_enabled(logger, "error", "No services found in compose file")
        return False, {}
    
    # Initialize tracking
    start_time = time.time()
    time.sleep(1)  # Give containers a moment to start
    
    # Poll until success, failure, or timeout
    while True:
        elapsed = time.time() - start_time
        
        # Re-discover container names on each iteration to handle race conditions
        container_names = _discover_container_names(services, compose_file, logger)
        
        # Check health of all services (including those without containers)
        health_statuses = _check_all_containers_health(container_names, logger)
        all_healthy, any_unhealthy = _evaluate_all_container_health(health_statuses, logger)
        
        # Return immediately if all healthy or any unhealthy
        if all_healthy:
            _log_if_enabled(logger, "debug", "All services are healthy")
            return True, health_statuses
        
        if any_unhealthy:
            _log_if_enabled(logger, "debug", "Some services are unhealthy")
            return False, health_statuses
        
        # Check timeout AFTER checking health to avoid exceeding timeout
        if elapsed > timeout:
            _log_if_enabled(logger, "debug", f"Health check timed out after {timeout}s")
            break
        
        # Still starting, wait and check again
        _log_if_enabled(logger, "debug", "Services still starting, waiting...")
        time.sleep(check_interval)
    
    # Timeout reached - return final status (already have it from last iteration)
    return False, health_statuses  

def format_health_report(health_statuses: Dict[str, ContainerHealth]) -> str:
    if not health_statuses:
        return "No services found"
    
    lines = ["Service Health Status:"]
    
    for service, health in health_statuses.items():
        is_healthy = health.running and health.status in [HealthStatus.HEALTHY, HealthStatus.NONE]
        status_icon = "✓" if is_healthy else "✗"
        status_text = health.status.value if health.status != HealthStatus.NONE else "running"
        
        line = f"  {status_icon} {service}: {status_text}"
        
        if not health.running:
            line += f" (exit code: {health.exit_code})"
        
        if health.error:
            line += f" - {health.error}"
        
        lines.append(line)
    
    return "\n".join(lines)
