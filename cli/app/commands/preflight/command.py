import typer

from app.utils.lib import HostInformation
from app.utils.logger import Logger
from app.utils.timeout import TimeoutWrapper

from .deps import Deps, DepsConfig
from .run import PreflightRunner
from .messages import (
    debug_starting_preflight_check,
    debug_preflight_check_completed,
    debug_starting_ports_check,
    debug_ports_check_completed,
    debug_starting_deps_check,
    debug_deps_check_completed,
    debug_creating_port_config,
    debug_creating_deps_config,
    debug_initializing_port_service,
    debug_initializing_deps_service,
    debug_timeout_wrapper_start,
    debug_timeout_wrapper_end,
    debug_formatting_output,
    error_checking_deps,
    error_checking_ports,
    error_timeout_occurred,
    error_validation_failed,
    running_preflight_checks,
)
from .port import PortConfig, PortService

preflight_app = typer.Typer(no_args_is_help=False)


@preflight_app.callback(invoke_without_command=True)
def preflight_callback(ctx: typer.Context):
    """Preflight checks for system compatibility"""
    if ctx.invoked_subcommand is None:
        ctx.invoke(check)


@preflight_app.command()
def check(
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text,json"),
    timeout: int = typer.Option(10, "--timeout", "-t", help="Timeout in seconds"),
):
    """Run all preflight checks"""
    try:
        logger = Logger(verbose=verbose)
        logger.debug(debug_starting_preflight_check)
        logger.info(running_preflight_checks)
        
        logger.debug(debug_timeout_wrapper_start.format(timeout=timeout))
        with TimeoutWrapper(timeout):
            preflight_runner = PreflightRunner(logger=logger, verbose=verbose)
            preflight_runner.check_ports_from_config()
            logger.debug(debug_timeout_wrapper_end)
            logger.debug(debug_preflight_check_completed)
        
        logger.success("All preflight checks completed successfully")
    except TimeoutError as e:
        logger.error(error_timeout_occurred.format(timeout=timeout))
        raise typer.Exit(1)
    except Exception as e:
        if not isinstance(e, typer.Exit):
            logger.error(f"Unexpected error during preflight check: {e}")
        raise typer.Exit(1)

@preflight_app.command()
def ports(
    ports: list[int] = typer.Argument(..., help="The list of ports to check"),
    host: str = typer.Option("localhost", "--host", "-h", help="The host to check"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    timeout: int = typer.Option(10, "--timeout", "-t", help="Timeout in seconds"),
) -> None:
    """Check if list of ports are available on a host"""
    try:
        logger = Logger(verbose=verbose)
        logger.debug(debug_starting_ports_check)
        
        logger.debug(debug_creating_port_config)
        config = PortConfig(ports=ports, host=host, verbose=verbose)
        
        logger.debug(debug_initializing_port_service)
        port_service = PortService(config, logger=logger)
        
        logger.debug(debug_timeout_wrapper_start.format(timeout=timeout))
        with TimeoutWrapper(timeout):
            results = port_service.check_ports()
        logger.debug(debug_timeout_wrapper_end)
        
        logger.debug(debug_formatting_output.format(format=output))
        formatted_output = port_service.formatter.format_output(results, output)
        
        logger.success(formatted_output)
        logger.debug(debug_ports_check_completed)
        
    except ValueError as e:
        logger.error(error_validation_failed.format(error=e))
        raise typer.Exit(1)
    except TimeoutError as e:
        logger.error(error_timeout_occurred.format(timeout=timeout))
        raise typer.Exit(1)
    except Exception as e:
        if not isinstance(e, typer.Exit):
            logger.error(error_checking_ports.format(error=e))
        raise typer.Exit(1)


@preflight_app.command()
def deps(
    deps: list[str] = typer.Argument(..., help="The list of dependencies to check"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    timeout: int = typer.Option(10, "--timeout", "-t", help="Timeout in seconds"),
) -> None:
    """Check if list of dependencies are available on the system"""
    try:
        logger = Logger(verbose=verbose)
        logger.debug(debug_starting_deps_check)
        
        logger.debug(debug_creating_deps_config)
        config = DepsConfig(
            deps=deps,
            verbose=verbose,
            output=output,
            os=HostInformation.get_os_name(),
            package_manager=HostInformation.get_package_manager(),
        )
        
        logger.debug(debug_initializing_deps_service)
        deps_checker = Deps(logger=logger)
        
        logger.debug(debug_timeout_wrapper_start.format(timeout=timeout))
        with TimeoutWrapper(timeout):
            results = deps_checker.check(config)
        logger.debug(debug_timeout_wrapper_end)
        
        logger.debug(debug_formatting_output.format(format=output))
        formatted_output = deps_checker.format_output(results, output)
        
        logger.success(formatted_output)
        logger.debug(debug_deps_check_completed)
        
    except ValueError as e:
        logger.error(error_validation_failed.format(error=e))
        raise typer.Exit(1)
    except TimeoutError as e:
        logger.error(error_timeout_occurred.format(timeout=timeout))
        raise typer.Exit(1)
    except Exception as e:
        if not isinstance(e, typer.Exit):
            logger.error(error_checking_deps.format(error=e))
        raise typer.Exit(1)
