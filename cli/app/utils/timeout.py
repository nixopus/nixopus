import signal
from contextlib import contextmanager
from typing import Generator

from app.commands.install.messages import timeout_error


@contextmanager
def timeout_wrapper(timeout: int) -> Generator[None, None, None]:
    """Context manager for timeout operations using functional approach"""
    if timeout > 0:
        def timeout_handler(signum, frame):
            raise TimeoutError(timeout_error.format(timeout=timeout))

        original_handler = signal.signal(signal.SIGALRM, timeout_handler)
        signal.alarm(timeout)
        
        try:
            yield
        finally:
            signal.alarm(0)
            signal.signal(signal.SIGALRM, original_handler)
    else:
        yield
