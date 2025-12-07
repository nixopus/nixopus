import json
import os
from typing import Optional

import requests

from app.utils.config import CADDY_BASE_URL, LOAD_ENDPOINT, PROXY_PORT, get_active_config, get_yaml_value
from app.utils.protocols import LoggerProtocol
from app.utils.retry import retry_with_backoff, wait_for_condition

from .messages import (
    caddy_connection_failed,
    config_file_not_found,
    debug_config_loaded_success,
    debug_config_parsed,
    debug_loading_config_file,
    debug_posting_config,
    http_error,
    invalid_json_error,
    request_failed_error,
    unexpected_error,
)

_config = get_active_config()
default_proxy_port = get_yaml_value(_config, PROXY_PORT)
caddy_load_endpoint = get_yaml_value(_config, LOAD_ENDPOINT)
caddy_base_url = get_yaml_value(_config, CADDY_BASE_URL)


def _get_caddy_url(port: int, endpoint: str) -> str:
    """Build Caddy API URL."""
    return f"{caddy_base_url.format(port=port)}{endpoint}"


def _check_caddy_ready(proxy_port: int) -> bool:
    """Check if Caddy is ready to accept connections."""
    url = _get_caddy_url(proxy_port, "/config/")
    try:
        response = requests.get(url, timeout=5)
        # 200 = config exists, 404 = no config yet but Caddy is responding
        return response.status_code in (200, 404)
    except (requests.exceptions.ConnectionError, requests.exceptions.RequestException):
        return False


def _wait_for_caddy(
    proxy_port: int,
    logger: Optional[LoggerProtocol] = None,
) -> tuple[bool, Optional[str]]:
    """Wait for Caddy to be ready to accept connections with exponential backoff."""

    def on_retry(attempt: int, delay: float) -> None:
        if logger:
            logger.debug(f"Waiting for Caddy to be ready (attempt {attempt}, retrying in {delay:.1f}s)")

    success, error = wait_for_condition(
        check_func=lambda: _check_caddy_ready(proxy_port),
        on_retry=on_retry,
        timeout_message="Caddy not ready",
    )

    if success and logger:
        logger.debug("Caddy is ready")

    return success, error


def load_config(
    config_file: str,
    proxy_port: int = default_proxy_port,
    logger: Optional[LoggerProtocol] = None,
) -> tuple[bool, Optional[str]]:
    """Load Caddy proxy configuration from a JSON file with retry logic."""
    if not config_file:
        return False, "Configuration file is required"

    if not os.path.exists(config_file):
        error_msg = config_file_not_found.format(file=config_file)
        if logger:
            logger.debug(error_msg)
        return False, error_msg

    try:
        if logger:
            logger.debug(debug_loading_config_file.format(file=config_file))

        with open(config_file, "r") as f:
            config_data = json.load(f)

        if logger:
            logger.debug(debug_config_parsed)

    except FileNotFoundError:
        error_msg = config_file_not_found.format(file=config_file)
        if logger:
            logger.debug(error_msg)
        return False, error_msg
    except json.JSONDecodeError as e:
        error_msg = invalid_json_error.format(error=str(e))
        if logger:
            logger.debug(error_msg)
        return False, error_msg

    # Wait for Caddy to be ready before attempting to load config
    if logger:
        logger.debug("Waiting for Caddy to be ready...")
    ready, ready_error = _wait_for_caddy(proxy_port, logger)
    if not ready:
        return False, ready_error

    url = _get_caddy_url(proxy_port, caddy_load_endpoint)

    def post_config() -> tuple[bool, Optional[str]]:
        """Attempt to post config to Caddy."""
        try:
            if logger:
                logger.debug(debug_posting_config.format(url=url))

            response = requests.post(url, json=config_data, headers={"Content-Type": "application/json"}, timeout=10)

            if response.status_code == 200:
                return True, None
            else:
                error_msg = response.text.strip() if response.text else http_error.format(code=response.status_code)
                return False, error_msg

        except requests.exceptions.ConnectionError as e:
            return False, caddy_connection_failed.format(error=str(e))
        except requests.exceptions.RequestException as e:
            return False, request_failed_error.format(error=str(e))
        except Exception as e:
            return False, unexpected_error.format(error=str(e))

    def on_retry(attempt: int, delay: float, last_error: Optional[str]) -> None:
        if logger:
            logger.debug(f"Failed to load config (attempt {attempt}): {last_error}")
            logger.debug(f"Retrying in {delay:.1f}s...")

    success, error = retry_with_backoff(func=post_config, on_retry=on_retry)

    if success and logger:
        logger.debug(debug_config_loaded_success)

    return success, error
