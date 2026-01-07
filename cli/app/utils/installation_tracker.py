from typing import Any, Dict, Optional, Tuple

from app.commands.version.command import get_version
from app.utils.build_utils import detect_architecture
from app.utils.host_information import get_os_name
from app.utils.http import make_post_request_with_validation
from app.utils.installation_tracker_messages import (
    tracking_failed,
    tracking_invalid_event_type,
    tracking_system_info_failed,
)
from app.utils.protocols import LoggerProtocol

INSTALLATION_TRACKING_URL = "https://nixopus-api.nixopus.com/api/cli/installations"
DEFAULT_TRACKING_TIMEOUT = 10
VALID_EVENT_TYPES = ["installation_success", "installation_success_staging", "installation_failure"]
MAX_ERROR_MESSAGE_LENGTH = 500


def _get_system_info(logger: Optional[LoggerProtocol] = None) -> Tuple[bool, Optional[Dict[str, str]], Optional[str]]:
    try:
        system_info = {
            "os": get_os_name(),
            "arch": detect_architecture(),
            "version": get_version(),
        }
        return True, system_info, None
    except Exception as e:
        error_msg = tracking_system_info_failed.format(error=str(e))
        if logger:
            logger.warning(error_msg)
        return False, None, error_msg


def track_installation_event(
    event_type: str,
    logger: Optional[LoggerProtocol] = None,
    timeout: int = DEFAULT_TRACKING_TIMEOUT,
    **kwargs: Any,
) -> Tuple[bool, Optional[str]]:
    try:
        if event_type not in VALID_EVENT_TYPES:
            error_msg = tracking_invalid_event_type.format(event_type=event_type, valid_types=VALID_EVENT_TYPES)
            if logger:
                logger.warning(error_msg)
            return False, error_msg
        
        success, system_info, error = _get_system_info(logger)
        if not success:
            return False, error
        
        payload = {
            "event_type": event_type,
            "os": system_info["os"],
            "arch": system_info["arch"],
            "version": system_info["version"],
            **kwargs,
        }
        
        success, response, error = make_post_request_with_validation(
            url=INSTALLATION_TRACKING_URL,
            json=payload,
            headers={
                "Accept": "application/json, application/xml",
                "Content-Type": "application/json",
            },
            timeout=timeout,
            raise_on_error=True,
            logger=logger,
        )
        
        if success:
            if logger:
                logger.debug(f"Successfully tracked installation event: {event_type}")
            return True, None
        
        error_msg = f"{tracking_failed}: {error}" if error else tracking_failed
        if logger:
            logger.warning(error_msg)
        return False, error_msg
    except Exception as e:
        error_msg = f"{tracking_failed}: {str(e)}"
        if logger:
            logger.warning(error_msg)
        return False, error_msg


def track_installation_success(
    staging: bool = False,
    logger: Optional[LoggerProtocol] = None,
    timeout: int = DEFAULT_TRACKING_TIMEOUT,
    **kwargs: Any,
) -> Tuple[bool, Optional[str]]:
    event_type = "installation_success_staging" if staging else "installation_success"
    return track_installation_event(event_type, logger=logger, timeout=timeout, **kwargs)


def track_installation_failure(
    failed_step: Optional[str] = None,
    staging: bool = False,
    error_message: Optional[str] = None,
    logger: Optional[LoggerProtocol] = None,
    timeout: int = DEFAULT_TRACKING_TIMEOUT,
    **kwargs: Any,
) -> Tuple[bool, Optional[str]]:
    properties = {**kwargs}
    
    if failed_step:
        properties["failed_step"] = failed_step
    
    if error_message:
        properties["error"] = error_message[:MAX_ERROR_MESSAGE_LENGTH]
    
    if staging:
        properties["staging"] = True
    
    return track_installation_event("installation_failure", logger=logger, timeout=timeout, **properties)


def track_staging_installation(
    logger: Optional[LoggerProtocol] = None,
    timeout: int = DEFAULT_TRACKING_TIMEOUT,
) -> Tuple[bool, Optional[str]]:
    return track_installation_success(staging=True, logger=logger, timeout=timeout)

