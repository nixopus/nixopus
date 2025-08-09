import typer

from .message import DEBUG_MESSAGE, ERROR_MESSAGE, HIGHLIGHT_MESSAGE, INFO_MESSAGE, SUCCESS_MESSAGE, WARNING_MESSAGE


class Logger:
    """Wrapper for typer.secho to log messages to the console"""

    def __init__(self, verbose: bool = False, quiet: bool = False):
        if verbose and quiet:
            raise ValueError("Cannot have both verbose and quiet options enabled")
        self.verbose = verbose
        self.quiet = quiet

    def _should_print(self, require_verbose: bool = False) -> bool:
        """Helper method to determine if message should be printed"""
        if self.quiet:
            return False
        if require_verbose and not self.verbose:
            return False
        return True

    def info(self, message: str) -> None:
        """Prints an info message"""
        if self._should_print():
            typer.secho(INFO_MESSAGE.format(message=message), fg=typer.colors.BLUE)

    def debug(self, message: str) -> None:
        """Prints a debug message if verbose is enabled"""
        if self._should_print(require_verbose=True):
            typer.secho(DEBUG_MESSAGE.format(message=message), fg=typer.colors.CYAN)

    def warning(self, message: str) -> None:
        """Prints a warning message"""
        if self._should_print():
            typer.secho(WARNING_MESSAGE.format(message=message), fg=typer.colors.YELLOW)

    def error(self, message: str) -> None:
        """Prints an error message"""
        if self._should_print():
            typer.secho(ERROR_MESSAGE.format(message=message), fg=typer.colors.RED)

    def success(self, message: str) -> None:
        """Prints a success message"""
        if self._should_print():
            typer.secho(SUCCESS_MESSAGE.format(message=message), fg=typer.colors.GREEN)

    def highlight(self, message: str) -> None:
        """Prints a highlighted message"""
        if self._should_print():
            typer.secho(HIGHLIGHT_MESSAGE.format(message=message), fg=typer.colors.MAGENTA)
