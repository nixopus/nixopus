from typing import Any, Dict, Optional, Tuple

import requests

from app.utils.protocols import LoggerProtocol


def make_post_request(
    url: str,
    json: Optional[Dict[str, Any]] = None,
    data: Optional[Any] = None,
    timeout: int = 10,
    headers: Optional[Dict[str, str]] = None,
    verify: bool = True,
    logger: Optional[LoggerProtocol] = None,
) -> Tuple[bool, Optional[requests.Response], Optional[str]]:
    try:
        response = requests.post(
            url,
            json=json,
            data=data,
            timeout=timeout,
            headers=headers,
            verify=verify,
        )
        return True, response, None
    except requests.RequestException as e:
        error_msg = str(e)
        if logger:
            logger.warning(f"POST request failed to {url}: {error_msg}")
        return False, None, error_msg
    except Exception as e:
        error_msg = str(e)
        if logger:
            logger.warning(f"Unexpected error during POST request to {url}: {error_msg}")
        return False, None, error_msg


def make_post_request_with_validation(
    url: str,
    json: Optional[Dict[str, Any]] = None,
    data: Optional[Any] = None,
    timeout: int = 10,
    headers: Optional[Dict[str, str]] = None,
    verify: bool = True,
    raise_on_error: bool = False,
    logger: Optional[LoggerProtocol] = None,
) -> Tuple[bool, Optional[requests.Response], Optional[str]]:
    success, response, error = make_post_request(
        url=url,
        json=json,
        data=data,
        timeout=timeout,
        headers=headers,
        verify=verify,
        logger=logger,
    )
    
    if not success:
        return False, None, error
    
    if raise_on_error:
        try:
            response.raise_for_status()
        except requests.HTTPError as e:
            error_msg = str(e)
            if logger:
                logger.warning(f"POST request returned error status to {url}: {error_msg}")
            return False, response, error_msg
    
    return True, response, None
