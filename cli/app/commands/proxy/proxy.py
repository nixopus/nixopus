import json
import os
from typing import Optional

import requests

from app.utils.config import CADDY_BASE_URL, LOAD_ENDPOINT, PROXY_PORT, get_active_config, get_yaml_value
from app.utils.protocols import LoggerProtocol

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


def load_config(
    config_file: str, proxy_port: int = default_proxy_port, logger: Optional[LoggerProtocol] = None
) -> tuple[bool, Optional[str]]:
    """Load Caddy proxy configuration from a JSON file."""
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

        url = _get_caddy_url(proxy_port, caddy_load_endpoint)
        if logger:
            logger.debug(debug_posting_config.format(url=url))

        response = requests.post(url, json=config_data, headers={"Content-Type": "application/json"}, timeout=10)
        
        if response.status_code == 200:
            if logger:
                logger.debug(debug_config_loaded_success)
            return True, None
        else:
            error_msg = response.text.strip() if response.text else http_error.format(code=response.status_code)
            if logger:
                logger.debug(f"Failed to load config: {error_msg}")
            return False, error_msg
            
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
    except requests.exceptions.ConnectionError as e:
        error_msg = caddy_connection_failed.format(error=str(e))
        if logger:
            logger.debug(error_msg)
        return False, error_msg
    except requests.exceptions.RequestException as e:
        error_msg = request_failed_error.format(error=str(e))
        if logger:
            logger.debug(error_msg)
        return False, error_msg
    except Exception as e:
        error_msg = unexpected_error.format(error=str(e))
        if logger:
            logger.debug(error_msg)
        return False, error_msg

