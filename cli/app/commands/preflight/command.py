import typer

from app.utils.config import Config, DEPS, NIXOPUS_CONFIG_DIR, PORTS
from app.utils.lib import FileManager, HostInformation
from app.utils.logger import Logger
from .deps import Deps, DepsConfig
from .messages import error_checking_deps, error_checking_ports
from .port import PortConfig, PortService
from .run import preflight_run

preflight_app = typer.Typer(no_args_is_help=False)
config = Config()
ports_config = config.get_yaml_value(PORTS)
ports_list = [int(port) for port in ports_config]
nixopus_config_dir = config.get_yaml_value(NIXOPUS_CONFIG_DIR)
deps_list = list(config.get_yaml_value(DEPS).keys())

@preflight_app.callback(invoke_without_command=True)
def preflight_callback(ctx: typer.Context, output: str = typer.Option("text", "--output", "-o", help="Output format, text, json")):
    """Preflight checks for system compatibility"""
    if ctx.invoked_subcommand is None:
        preflight_run(logger=Logger(), ports_list=ports_list, deps_list=deps_list, nixopus_config_dir=nixopus_config_dir, output=output, host="localhost", timeout=1, verbose=False)

@preflight_app.command()
def ports(
    ports: list[int] = typer.Option(ports_list, "--ports", "-p", help="The list of ports to check (defaults to ports from config)"),
    host: str = typer.Option("localhost", "--host", "-h", help="The host to check"),
    timeout: int = typer.Option(1, "--timeout", "-t", help="The timeout in seconds for each port check"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
) -> None:
    """Check if list of ports are available on a host"""
    try:
        logger = Logger(verbose=verbose)
        config = PortConfig(ports=ports, host=host, timeout=timeout, verbose=verbose)
        port_service = PortService(config, logger=logger)
        results = port_service.check_ports()
        logger.success(port_service.formatter.format_output(results, output))
    except Exception as e:
        logger.error(error_checking_ports.format(error=e))
        raise typer.Exit(1)

@preflight_app.command()
def deps(
    deps: list[str] = typer.Option(deps_list, "--deps", "-d", help="The list of dependencies to check (defaults to dependencies from config)"),
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
            package_manager=HostInformation.get_package_manager(),
        )
        deps_checker = Deps(logger=logger)
        results = deps_checker.check(config)
        logger.highlight(deps_checker.format_output(results, output))
    except Exception as e:
        logger.error(error_checking_deps.format(error=e))
        raise typer.Exit(1)
