import logging
import os
from typing import Any, Dict, Optional

from posthog import Posthog

# Set up a logger for PostHog tracking failures
_logger = logging.getLogger(__name__)


def get_posthog_client() -> Optional[Any]:
    api_key = os.environ.get("POSTHOG_API_KEY")
    if not api_key:
        _logger.debug("PostHog API key not configured (POSTHOG_API_KEY not set)")
        return None
    
    host = os.environ.get("POSTHOG_HOST", "https://us.i.posthog.com")
    
    try:
        return Posthog(api_key, host=host)
    except Exception as e:
        _logger.warning(f"Failed to initialize PostHog client: {e}")
        return None


def track_event(
    event_name: str,
    properties: Optional[Dict[str, Any]] = None,
    distinct_id: Optional[str] = None,
) -> None:
    if properties is None:
        properties = {}
    
    if distinct_id is None:
        distinct_id = "cli-installation"
    
    try:
        client = get_posthog_client()
        if client:
            client.capture(distinct_id=distinct_id, event=event_name, properties=properties)
            client.shutdown()
            _logger.debug(f"Successfully tracked event: {event_name}")
        else:
            _logger.debug(f"PostHog client not available, skipping event: {event_name}")
    except Exception as e:
        _logger.warning(f"Failed to track PostHog event '{event_name}': {e}", exc_info=False)


def track_installation_success(staging: bool = False, **kwargs: Any) -> None:
    properties = {
        "staging": staging,
        **kwargs,
    }
    track_event("installation_success", properties=properties)


def track_installation_failure(
    failed_step: Optional[str] = None,
    staging: bool = False,
    error_message: Optional[str] = None,
    **kwargs: Any,
) -> None:
    properties = {
        "staging": staging,
        **kwargs,
    }
    
    if failed_step:
        properties["failed_step"] = failed_step
    
    if error_message:
        properties["error"] = error_message[:500]  # Limit error message length
    
    track_event("installation_failure", properties=properties)


def track_staging_installation() -> None:
    track_event("staging_installation", properties={"staging": True})

