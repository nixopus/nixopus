import signal
from app.commands.install.messages import timeout_error


class TimeoutWrapper:
    """Context manager for timeout operations"""
    
    def __init__(self, timeout: int):
        self.timeout = timeout
        self.original_handler = None
    
    def __enter__(self):
        if self.timeout > 0:
            def timeout_handler(signum, frame):
                raise TimeoutError(timeout_error.format(timeout=self.timeout))
            
            self.original_handler = signal.signal(signal.SIGALRM, timeout_handler)
            signal.alarm(self.timeout)
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        if self.timeout > 0:
            signal.alarm(0)
            signal.signal(signal.SIGALRM, self.original_handler) 
