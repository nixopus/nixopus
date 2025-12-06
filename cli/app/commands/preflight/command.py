import typer

from app.utils.lib import HostInformation
from app.utils.logger import create_logger, log_debug, log_error, log_info, log_success
from app.utils.timeout import TimeoutWrapper

from .deps import Deps, DepsConfig
from .messages import (
    debug_creating_deps_config,
    debug_creating_port_config,
    debug_deps_check_completed,
    debug_formatting_output,
    debug_initializing_deps_service,
    debug_initializing_port_service,
    debug_ports_check_completed,
    debug_preflight_check_completed,
    debug_starting_deps_check,
    debug_starting_ports_check,
    debug_starting_preflight_check,
    debug_timeout_wrapper_end,
    debug_timeout_wrapper_start,
    error_checking_deps,
    error_checking_ports,
    error_timeout_occurred,
    error_validation_failed,
    running_preflight_checks,
)
from .port import PortConfig, PortService
from .run import PreflightRunner

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
        logger = create_logger(verbose=verbose)
        log_debug(debug_starting_preflight_check, verbose=verbose)
        log_info(running_preflight_checks, verbose=verbose)

        log_debug(debug_timeout_wrapper_start.format(timeout=timeout), verbose=verbose)
        with TimeoutWrapper(timeout):
            preflight_runner = PreflightRunner(logger=logger, verbose=verbose)
            preflight_runner.check_ports_from_config()
            log_debug(debug_timeout_wrapper_end, verbose=verbose)
            log_debug(debug_preflight_check_completed, verbose=verbose)

        log_success("All preflight checks completed successfully", verbose=verbose)
    except TimeoutError as e:
        log_error(error_timeout_occurred.format(timeout=timeout), verbose=verbose)
        raise typer.Exit(1)
    except Exception as e:
        if not isinstance(e, typer.Exit):
            log_error(f"Unexpected error during preflight check: {e}", verbose=verbose)
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
        logger = create_logger(verbose=verbose)
        log_debug(debug_starting_ports_check, verbose=verbose)

        log_debug(debug_creating_port_config, verbose=verbose)
        config = PortConfig(ports=ports, host=host, verbose=verbose)

        log_debug(debug_initializing_port_service, verbose=verbose)
        port_service = PortService(config, logger=logger)

        log_debug(debug_timeout_wrapper_start.format(timeout=timeout), verbose=verbose)
        with TimeoutWrapper(timeout):
            results = port_service.check_ports()
        log_debug(debug_timeout_wrapper_end, verbose=verbose)

        log_debug(debug_formatting_output.format(format=output), verbose=verbose)
        formatted_output = port_service.formatter.format_output(results, output)

        log_success(formatted_output, verbose=verbose)
        log_debug(debug_ports_check_completed, verbose=verbose)

    except ValueError as e:
        log_error(error_validation_failed.format(error=e), verbose=verbose)
        raise typer.Exit(1)
    except TimeoutError as e:
        log_error(error_timeout_occurred.format(timeout=timeout), verbose=verbose)
        raise typer.Exit(1)
    except Exception as e:
        if not isinstance(e, typer.Exit):
            log_error(error_checking_ports.format(error=e), verbose=verbose)
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
        logger = create_logger(verbose=verbose)
        log_debug(debug_starting_deps_check, verbose=verbose)

        log_debug(debug_creating_deps_config, verbose=verbose)
        config = DepsConfig(
            deps=deps,
            verbose=verbose,
            output=output,
            os=HostInformation.get_os_name(),
            package_manager=HostInformation.get_package_manager(),
        )

        log_debug(debug_initializing_deps_service, verbose=verbose)
        deps_checker = Deps(logger=logger)

        log_debug(debug_timeout_wrapper_start.format(timeout=timeout), verbose=verbose)
        with TimeoutWrapper(timeout):
            results = deps_checker.check(config)
        log_debug(debug_timeout_wrapper_end, verbose=verbose)

        log_debug(debug_formatting_output.format(format=output), verbose=verbose)
        formatted_output = deps_checker.format_output(results, output)

        log_success(formatted_output, verbose=verbose)
        log_debug(debug_deps_check_completed, verbose=verbose)

    except ValueError as e:
        log_error(error_validation_failed.format(error=e), verbose=verbose)
        raise typer.Exit(1)
    except TimeoutError as e:
        log_error(error_timeout_occurred.format(timeout=timeout), verbose=verbose)
        raise typer.Exit(1)
    except Exception as e:
        if not isinstance(e, typer.Exit):
            log_error(error_checking_deps.format(error=e), verbose=verbose)
        raise typer.Exit(1)
