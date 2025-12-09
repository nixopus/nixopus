import time
from typing import Callable, Optional, Tuple, TypeVar

T = TypeVar("T")

DEFAULT_MAX_RETRIES = 10
DEFAULT_INITIAL_DELAY = 2.0
DEFAULT_MAX_DELAY = 30.0
DEFAULT_BACKOFF_FACTOR = 1.5


def retry_with_backoff(
    func: Callable[[], Tuple[bool, Optional[str]]],
    max_retries: int = DEFAULT_MAX_RETRIES,
    initial_delay: float = DEFAULT_INITIAL_DELAY,
    max_delay: float = DEFAULT_MAX_DELAY,
    backoff_factor: float = DEFAULT_BACKOFF_FACTOR,
    on_retry: Optional[Callable[[int, float, Optional[str]], None]] = None,
) -> Tuple[bool, Optional[str]]:
    """
    Execute a function with exponential backoff retry logic.

    Args:
        func: A callable that returns (success: bool, error: Optional[str])
        max_retries: Maximum number of retry attempts
        initial_delay: Initial delay between retries in seconds
        max_delay: Maximum delay between retries in seconds
        backoff_factor: Multiplier for delay after each retry
        on_retry: Optional callback called before each retry with (attempt, delay, last_error)

    Returns:
        Tuple of (success: bool, error: Optional[str])
    """
    delay = initial_delay
    last_error: Optional[str] = None

    for attempt in range(1, max_retries + 1):
        try:
            success, error = func()
            if success:
                return True, None
            last_error = error
        except Exception as e:
            last_error = str(e)

        if attempt < max_retries:
            if on_retry:
                on_retry(attempt, delay, last_error)
            time.sleep(delay)
            delay = min(delay * backoff_factor, max_delay)

    return False, last_error


def wait_for_condition(
    check_func: Callable[[], bool],
    max_retries: int = DEFAULT_MAX_RETRIES,
    initial_delay: float = DEFAULT_INITIAL_DELAY,
    max_delay: float = DEFAULT_MAX_DELAY,
    backoff_factor: float = DEFAULT_BACKOFF_FACTOR,
    on_retry: Optional[Callable[[int, float], None]] = None,
    timeout_message: str = "Condition not met after max retries",
) -> Tuple[bool, Optional[str]]:
    """
    Wait for a condition to become true with exponential backoff.

    Args:
        check_func: A callable that returns True when condition is met
        max_retries: Maximum number of retry attempts
        initial_delay: Initial delay between retries in seconds
        max_delay: Maximum delay between retries in seconds
        backoff_factor: Multiplier for delay after each retry
        on_retry: Optional callback called before each retry with (attempt, delay)
        timeout_message: Error message if condition is never met

    Returns:
        Tuple of (success: bool, error: Optional[str])
    """
    delay = initial_delay

    for attempt in range(1, max_retries + 1):
        try:
            if check_func():
                return True, None
        except Exception:
            pass

        if attempt < max_retries:
            if on_retry:
                on_retry(attempt, delay)
            time.sleep(delay)
            delay = min(delay * backoff_factor, max_delay)

    return False, f"{timeout_message} ({max_retries} attempts)"

