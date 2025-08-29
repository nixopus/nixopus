import typer
from .conflict import ConflictConfig, ConflictService
from .messages import (
    conflict_check_help,
    error_checking_conflicts,
    conflicts_found_warning,
    no_conflicts_info,
    checking_conflicts_info,
)
from app.utils.logger import Logger
from app.utils.timeout import TimeoutWrapper

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
        logger = Logger(verbose=verbose)
        
        try:
            logger.info(checking_conflicts_info)

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
                logger.error(result)
                logger.warning(conflicts_found_warning.format(count=len(conflicts)))
                raise typer.Exit(1)
            else:
                logger.success(result)
                logger.info(no_conflicts_info)

        except TimeoutError as e:
            logger.error(str(e))
            raise typer.Exit(1)
        except Exception as e:
            logger.error(error_checking_conflicts.format(error=str(e)))
            raise typer.Exit(1)
