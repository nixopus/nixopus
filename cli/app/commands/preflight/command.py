import typer
from .messages import error_checking_deps, error_checking_ports
from .port import PortConfig, PortService
from .deps import Deps, DepsConfig
from app.utils.lib import HostInformation
from app.utils.logger import Logger

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
):
    """Run all preflight checks"""
    pass

@preflight_app.command()
def ports(
    ports: list[int] = typer.Argument(..., help="The list of ports to check"),
    host: str = typer.Option("localhost", "--host", "-h", help="The host to check"),
    timeout: int = typer.Option(1, "--timeout", "-t", help="The timeout in seconds for each port check"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
) -> None:
    """Check if list of ports are available on a host"""
    try:
        logger = Logger(verbose=verbose)
        logger.debug(f"Checking ports: {ports}")
        config = PortConfig(ports=ports, host=host, timeout=timeout, verbose=verbose)
        port_service = PortService(config, logger=logger)
        results = port_service.check_ports()
        logger.success(port_service.formatter.format_output(results, output))
    except Exception as e:
        logger.error(error_checking_ports.format(error=e))
        raise typer.Exit(1)

@preflight_app.command()    
def deps(
    deps: list[str] = typer.Argument(..., help="The list of dependencies to check"),
    timeout: int = typer.Option(1, "--timeout", "-t", help="The timeout in seconds for each dependency check"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
) -> None:
    """Check if list of dependencies are available on the system"""
    try:
        logger = Logger(verbose=verbose)      
        config = DepsConfig(
            deps=deps,
            timeout=timeout,
            verbose=verbose,
            output=output,
            os=HostInformation.get_os_name(),
            package_manager=HostInformation.get_package_manager()
        )
        deps_checker = Deps(logger=logger)
        results = deps_checker.check(config)
        logger.success(deps_checker.format_output(results, output))
    except Exception as e:
        logger.error(error_checking_deps.format(error=e))
        raise typer.Exit(1)
