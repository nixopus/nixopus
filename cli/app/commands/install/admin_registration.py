import json
import re
import secrets
import string
import sys
import time
from typing import Optional, Tuple

import requests
import typer

from app.utils.config import API_PORT
from app.utils.protocols import LoggerProtocol

from .config_utils import get_host_ip_or_default
from .environment import ConfigResolver
from .types import InstallParams


SPECIAL_CHARS = "!@#$%^&*()_+-=[]{};':\"\\|,.<>/?"
MAX_RETRY_ATTEMPTS = 3
INITIAL_RETRY_WAIT = 5
MAX_RETRY_WAIT = 60
REQUEST_TIMEOUT = 30
MIN_PASSWORD_LENGTH = 8
HEALTH_CHECK_RETRIES = 3
HEALTH_CHECK_WAIT = 2
MAX_EMAIL_PROMPT_ATTEMPTS = 5

def generate_secure_password(length: int = 16) -> str:
    if length < 8:
        length = 16
    uppercase = string.ascii_uppercase
    lowercase = string.ascii_lowercase
    digits = string.digits
    password_chars = uppercase + lowercase + digits + SPECIAL_CHARS
    password = [
        secrets.choice(uppercase),
        secrets.choice(lowercase),
        secrets.choice(digits),
        secrets.choice(SPECIAL_CHARS),
    ]
    remaining_length = length - len(password)
    password.extend(secrets.choice(password_chars) for _ in range(remaining_length))
    secrets.SystemRandom().shuffle(password)
    return "".join(password)


def validate_email_format(email: str) -> bool:
    pattern = r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"
    return bool(re.match(pattern, email))


def sanitize_username(username: str) -> str:
    sanitized = re.sub(r"[^a-zA-Z0-9_-]", "", username)
    return sanitized if sanitized else "admin"


def extract_username_from_email(email: str) -> str:
    if "@" not in email:
        return "admin"
    username_part = email.split("@")[0]
    return sanitize_username(username_part)


def get_api_base_url(api_domain: Optional[str], host_ip: str, api_port: str) -> str:
    secure = api_domain is not None
    api_host = api_domain if secure else f"{host_ip}:{api_port}"
    protocol = "https" if secure else "http"
    return f"{protocol}://{api_host}"


def convert_port_to_string(port) -> str:
    return str(port) if port else "8443"


def is_retryable_error(status_code: Optional[int], error: Optional[str]) -> bool:
    if status_code is None:
        return True
    if status_code == 429:
        return True
    if status_code >= 500:
        return True
    if status_code == 400 or status_code == 200:
        return False
    return False


def parse_error_response(response: requests.Response) -> str:
    content_type = response.headers.get("content-type", "").lower()
    if "application/json" in content_type:
        try:
            data = response.json()
            if isinstance(data, dict):
                return data.get("message") or data.get("error") or response.text
        except (json.JSONDecodeError, ValueError):
            pass
    return response.text or f"HTTP {response.status_code}"


def make_registration_request(api_base_url: str, email: str, password: str, username: str, timeout: int = REQUEST_TIMEOUT, verify_ssl: bool = True) -> Tuple[bool, Optional[str], Optional[int]]:
    url = f"{api_base_url}/auth/signup"
    payload = {
        "formFields": [
            {"id": "email", "value": email},
            {"id": "password", "value": password},
        ]
    }
    try:
        response = requests.post(url, json=payload, timeout=timeout, verify=verify_ssl)
        if response.status_code == 200:
            try:
                data = response.json()
                if data.get("status") == "OK":
                    return True, None, response.status_code
                if data.get("status") == "FIELD_ERROR":
                    errors = data.get("formFields", [])
                    error_msgs = [f.get("error", "") for f in errors if f.get("error")]
                    return False, "; ".join(error_msgs) if error_msgs else "Field validation failed", response.status_code
                if data.get("status") == "GENERAL_ERROR":
                    return False, data.get("message", "Registration failed"), response.status_code
            except (json.JSONDecodeError, ValueError):
                pass
            return True, None, response.status_code
        error_msg = parse_error_response(response)
        return False, error_msg, response.status_code
    except requests.exceptions.RequestException as e:
        return False, str(e), None


def handle_dry_run(email: str, logger: Optional[LoggerProtocol]) -> Tuple[bool, Optional[str], Optional[bool]]:
    if logger:
        logger.info(f"[DRY RUN] Would register admin user with email: {email}")
    return True, None, None


def show_registration_summary(email: str, password: str, is_generated: bool, logger: Optional[LoggerProtocol]) -> None:
    if not logger:
        return
    logger.success("Admin user registration completed successfully!")
    logger.highlight(f"Email: {email}")
    if is_generated:
        logger.highlight(f"Password: {password}")
        logger.info("Please save these credentials securely.")
    else:
        logger.info("Password was provided during installation.")
    logger.info("You can now log in to Nixopus using these credentials.")


def handle_success(email: str, logger: Optional[LoggerProtocol]) -> Tuple[bool, Optional[str]]:
    if logger:
        logger.info(f"Admin user registered successfully with email: {email}")
    return True, None


def handle_admin_exists(logger: Optional[LoggerProtocol]) -> Tuple[bool, Optional[str], Optional[bool]]:
    if logger:
        logger.info("Admin user already exists, skipping registration")
    return True, None, None

def is_admin_already_registered(status_code: int, error: Optional[str]) -> bool:
    if not error:
        return False
    error_lower = error.lower()
    patterns = [
        "admin already registered", 
        "admin exists", 
        "user already exists", 
        "already registered",
        "sign up is disabled",
    ]
    return any(pattern in error_lower for pattern in patterns)

def calculate_next_wait_time(wait_time: int) -> int:
    next_wait = wait_time * 2
    return min(next_wait, MAX_RETRY_WAIT)


def handle_retry_attempt(attempt: int, max_attempts: int, wait_time: int, logger: Optional[LoggerProtocol]) -> int:
    if logger:
        logger.info(f"Retrying registration (attempt {attempt + 1}/{max_attempts})...")
    time.sleep(wait_time)
    return calculate_next_wait_time(wait_time)


def check_api_readiness(api_base_url: str, timeout: int = REQUEST_TIMEOUT, verify_ssl: bool = True) -> bool:
    health_url = f"{api_base_url}/api/v1/health"
    for attempt in range(1, HEALTH_CHECK_RETRIES + 1):
        try:
            response = requests.get(health_url, timeout=timeout, verify=verify_ssl)
            if response.status_code == 200:
                return True
        except requests.exceptions.RequestException:
            pass
        if attempt < HEALTH_CHECK_RETRIES:
            time.sleep(HEALTH_CHECK_WAIT)
    return False


def get_effective_timeout(timeout: Optional[int]) -> int:
    if timeout and timeout > 0:
        return timeout
    return REQUEST_TIMEOUT


def register_admin_user(
    api_base_url: str,
    email: str,
    password: str,
    logger: Optional[LoggerProtocol] = None,
    dry_run: bool = False,
    timeout: Optional[int] = None,
    verify_ssl: bool = True,
    is_generated: bool = False,
) -> Tuple[bool, Optional[str], Optional[bool]]:
    if dry_run:
        success, error, _ = handle_dry_run(email, logger)
        return success, error, None
    effective_timeout = get_effective_timeout(timeout)
    if not check_api_readiness(api_base_url, effective_timeout, verify_ssl):
        error_msg = "API is not ready after health checks"
        if logger:
            logger.warning(f"Registration skipped: {error_msg}")
        return False, error_msg, None
    username = extract_username_from_email(email)
    wait_time = INITIAL_RETRY_WAIT
    error = None
    for attempt in range(1, MAX_RETRY_ATTEMPTS + 1):
        success, error, status_code = make_registration_request(api_base_url, email, password, username, effective_timeout, verify_ssl)
        if success:
            handle_success(email, logger)
            return True, None, is_generated
        if not is_retryable_error(status_code, error):
            if status_code and is_admin_already_registered(status_code, error):
                handle_admin_exists(logger)
                return True, None, None
            if logger:
                logger.warning(f"Registration failed: {error}")
            return False, error, None
        if attempt < MAX_RETRY_ATTEMPTS:
            wait_time = handle_retry_attempt(attempt, MAX_RETRY_ATTEMPTS, wait_time, logger)
    if logger:
        logger.warning(f"Failed to register admin user after {MAX_RETRY_ATTEMPTS} attempts: {error}")
    return False, error, None


def validate_password_strength(password: str) -> Tuple[bool, Optional[str]]:
    if len(password) < MIN_PASSWORD_LENGTH:
        return False, f"Password must be at least {MIN_PASSWORD_LENGTH} characters"
    has_upper = any(c.isupper() for c in password)
    has_lower = any(c.islower() for c in password)
    has_digit = any(c.isdigit() for c in password)
    has_special = any(c in SPECIAL_CHARS for c in password)
    if not (has_upper and has_lower and has_digit and has_special):
        return False, "Password must contain uppercase, lowercase, digit, and special character"
    return True, None


def get_registration_password(params: InstallParams) -> Tuple[str, bool]:
    if params.admin_password:
        is_valid, error = validate_password_strength(params.admin_password)
        if not is_valid:
            raise ValueError(f"Invalid password: {error}")
        return params.admin_password, False
    return generate_secure_password(), True


def is_interactive() -> bool:
    return sys.stdin.isatty() and sys.stdout.isatty()


def prompt_for_email(logger: Optional[LoggerProtocol]) -> Optional[str]:
    if not is_interactive():
        return None
    if logger:
        logger.info("Password provided but email is missing. Please provide admin email.")
    for attempt in range(1, MAX_EMAIL_PROMPT_ATTEMPTS + 1):
        email = typer.prompt("Admin email address")
        email = email.strip() if email else None
        if not email:
            continue
        if validate_email_format(email):
            return email
        if logger:
            remaining = MAX_EMAIL_PROMPT_ATTEMPTS - attempt
            if remaining > 0:
                logger.warning(f"Invalid email format. {remaining} attempts remaining.")
            else:
                logger.error("Maximum email prompt attempts reached")
                return None
    return None


def can_get_email(params: InstallParams) -> bool:
    if params.admin_email:
        return True
    if params.admin_password and is_interactive():
        return True
    return False


def get_email_with_prompt(params: InstallParams) -> Optional[str]:
    if params.admin_email:
        if validate_email_format(params.admin_email):
            return params.admin_email
        if params.logger:
            params.logger.warning("Invalid email format in provided email")
        return None
    if params.admin_password and is_interactive():
        return prompt_for_email(params.logger)
    return None


def skip_registration_reason(params: InstallParams) -> Optional[str]:
    if not params.verify_health:
        return "health verification disabled"
    if not can_get_email(params):
        if params.admin_password and not is_interactive():
            return "email required (non-interactive mode)"
        return "no email provided"
    return None


def build_api_base_url(config_resolver: ConfigResolver, params: InstallParams) -> str:
    host_ip = get_host_ip_or_default(params.host_ip)
    api_port = config_resolver.get(API_PORT)
    api_port_str = convert_port_to_string(api_port)
    return get_api_base_url(params.api_domain, host_ip, api_port_str)


def log_generated_password(password: str, logger: Optional[LoggerProtocol]) -> None:
    if logger:
        logger.info("Generated secure password for admin user")
        logger.highlight(f"Admin password: {password}")
        logger.info("Please save this password securely. You can change it after first login.")


def should_verify_ssl(params: InstallParams) -> bool:
    if params.api_domain:
        return True
    return False


def register_admin_user_step(config_resolver: ConfigResolver, params: InstallParams) -> None:
    skip_reason = skip_registration_reason(params)
    if skip_reason:
        if params.logger:
            params.logger.info(f"Skipping admin user registration ({skip_reason})")
        return
    email = get_email_with_prompt(params)
    if not email:
        if params.logger:
            params.logger.warning("Cannot proceed without email address")
        return
    if not validate_email_format(email):
        if params.logger:
            params.logger.warning("Invalid email format")
        return
    try:
        password, is_generated = get_registration_password(params)
    except ValueError as e:
        if params.logger:
            params.logger.error(str(e))
        return
    api_base_url = build_api_base_url(config_resolver, params)
    timeout = get_effective_timeout(params.timeout)
    verify_ssl = should_verify_ssl(params)
    success, error, returned_is_generated = register_admin_user(
        api_base_url,
        email,
        password,
        params.logger,
        params.dry_run,
        timeout,
        verify_ssl,
        is_generated,
    )
    if success and returned_is_generated is not None:
        show_registration_summary(email, password, returned_is_generated, params.logger)
    if not success and error and params.logger:
        params.logger.warning(f"Admin registration failed: {error}")

