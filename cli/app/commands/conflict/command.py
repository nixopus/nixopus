import typer

from app.utils.logger import create_logger, log_error, log_info, log_success, log_warning
from app.utils.timeout import TimeoutWrapper

from .conflict import ConflictConfig, ConflictService
from .messages import (
    checking_conflicts_info,
    conflict_check_help,
    conflicts_found_warning,
    error_checking_conflicts,
    no_conflicts_info,
)

conflict_app = typer.Typer(help=conflict_check_help, no_args_is_help=False)


@conflict_app.callback(invoke_without_command=True)
def conflict_callback(
    ctx: typer.Context,
    config_file: str = typer.Option(
        None,
        "--config-file",
        "-c",
        help="Path to configuration file (defaults to built-in config)",
    ),
    timeout: int = typer.Option(5, "--timeout", "-t", help="Timeout for tool checks in seconds"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format (text/json)"),
) -> None:
    """Check for tool version conflicts"""
    if ctx.invoked_subcommand is None:
        # Initialize logger once and reuse throughout
        logger = create_logger(verbose=verbose)

        try:
            log_info(checking_conflicts_info, verbose=verbose)

            config = ConflictConfig(
                config_file=config_file,
                verbose=verbose,
                output=output,
            )

            service = ConflictService(config, logger=logger)

            with TimeoutWrapper(timeout):
                result = service.check_and_format(output)
                # Check if there are any conflicts and exit with appropriate code
                results = service.check_conflicts()
                conflicts = [r for r in results if r.conflict]

            if conflicts:
                log_error(result, verbose=verbose)
                log_warning(conflicts_found_warning.format(count=len(conflicts)), verbose=verbose)
                raise typer.Exit(1)
            else:
                log_success(result, verbose=verbose)
                log_info(no_conflicts_info, verbose=verbose)

        except TimeoutError as e:
            log_error(str(e), verbose=verbose)
            raise typer.Exit(1)
        except Exception as e:
            log_error(error_checking_conflicts.format(error=str(e)), verbose=verbose)
            raise typer.Exit(1)
