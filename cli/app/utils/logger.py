import typer
from types import SimpleNamespace

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


def create_logger(verbose: bool = False, quiet: bool = False):
    """Create a LoggerProtocol-compatible object using functional functions"""
    validate_logger_flags(verbose, quiet)

    # Use closure to capture verbose/quiet and return object with methods

    logger_obj = SimpleNamespace()
    logger_obj.info = lambda msg: log_info(msg, verbose, quiet)
    logger_obj.debug = lambda msg: log_debug(msg, verbose, quiet)
    logger_obj.warning = lambda msg: log_warning(msg, verbose, quiet)
    logger_obj.error = lambda msg: log_error(msg, verbose, quiet)
    logger_obj.success = lambda msg: log_success(msg, verbose, quiet)
    logger_obj.highlight = lambda msg: log_highlight(msg, verbose, quiet)

    return logger_obj
