import typer

from .message import DEBUG_MESSAGE, ERROR_MESSAGE, HIGHLIGHT_MESSAGE, INFO_MESSAGE, SUCCESS_MESSAGE, WARNING_MESSAGE


def validate_logger_flags(verbose: bool, quiet: bool) -> None:
    """Validate that verbose and quiet are not both enabled"""
    if verbose and quiet:
        raise ValueError("Cannot have both verbose and quiet options enabled")


def _should_print(verbose: bool, quiet: bool, require_verbose: bool = False) -> bool:
    """Helper function to determine if message should be printed"""
    if quiet:
        return False
    if require_verbose and not verbose:
        return False
    return True


def log_info(message: str, verbose: bool = False, quiet: bool = False) -> None:
    """Prints an info message"""
    if _should_print(verbose, quiet):
        typer.secho(INFO_MESSAGE.format(message=message), fg=typer.colors.BLUE)


def log_debug(message: str, verbose: bool = False, quiet: bool = False) -> None:
    """Prints a debug message if verbose is enabled"""
    if _should_print(verbose, quiet, require_verbose=True):
        typer.secho(DEBUG_MESSAGE.format(message=message), fg=typer.colors.CYAN)


def log_warning(message: str, verbose: bool = False, quiet: bool = False) -> None:
    """Prints a warning message"""
    if _should_print(verbose, quiet):
        typer.secho(WARNING_MESSAGE.format(message=message), fg=typer.colors.YELLOW)


def log_error(message: str, verbose: bool = False, quiet: bool = False) -> None:
    """Prints an error message"""
    if _should_print(verbose, quiet):
        typer.secho(ERROR_MESSAGE.format(message=message), fg=typer.colors.RED)


def log_success(message: str, verbose: bool = False, quiet: bool = False) -> None:
    """Prints a success message"""
    if _should_print(verbose, quiet):
        typer.secho(SUCCESS_MESSAGE.format(message=message), fg=typer.colors.GREEN)


def log_highlight(message: str, verbose: bool = False, quiet: bool = False) -> None:
    """Prints a highlighted message"""
    if _should_print(verbose, quiet):
        typer.secho(HIGHLIGHT_MESSAGE.format(message=message), fg=typer.colors.MAGENTA)


class Logger:
    """Compatibility wrapper class for LoggerProtocol - delegates to functional functions"""

    def __init__(self, verbose: bool = False, quiet: bool = False):
        validate_logger_flags(verbose, quiet)
        self.verbose = verbose
        self.quiet = quiet

    def info(self, message: str) -> None:
        """Prints an info message"""
        log_info(message, self.verbose, self.quiet)

    def debug(self, message: str) -> None:
        """Prints a debug message if verbose is enabled"""
        log_debug(message, self.verbose, self.quiet)

    def warning(self, message: str) -> None:
        """Prints a warning message"""
        log_warning(message, self.verbose, self.quiet)

    def error(self, message: str) -> None:
        """Prints an error message"""
        log_error(message, self.verbose, self.quiet)

    def success(self, message: str) -> None:
        """Prints a success message"""
        log_success(message, self.verbose, self.quiet)

    def highlight(self, message: str) -> None:
        """Prints a highlighted message"""
        log_highlight(message, self.verbose, self.quiet)
