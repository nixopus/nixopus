import typer

from app.utils.config import Config, PROXY_PORT
from app.utils.logger import Logger

from .load import Load, LoadConfig
from .status import Status, StatusConfig
from .stop import Stop, StopConfig

proxy_app = typer.Typer(
    name="proxy",
    help="Manage Nixopus proxy (Caddy) configuration",
)

config = Config()
proxy_port = config.get_yaml_value(PROXY_PORT)

@proxy_app.command()
def load(
    proxy_port: int = typer.Option(proxy_port, "--proxy-port", "-p", help="Caddy admin port"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format: text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", help="Dry run"),
    config_file: str = typer.Option(None, "--config-file", "-c", help="Path to Caddy config file"),
):
    """Load Caddy proxy configuration"""
    logger = Logger(verbose=verbose)

    try:
        config = LoadConfig(proxy_port=proxy_port, verbose=verbose, output=output, dry_run=dry_run, config_file=config_file)

        load_service = Load(logger=logger)
        result = load_service.load(config)

        if result.success:
            logger.success(load_service.format_output(result, output))
        else:
            logger.error(result.error)
            raise typer.Exit(1)

    except Exception as e:
        logger.error(str(e))
        raise typer.Exit(1)


@proxy_app.command()
def status(
    proxy_port: int = typer.Option(proxy_port, "--proxy-port", "-p", help="Caddy admin port"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format: text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", help="Dry run"),
):
    """Check Caddy proxy status"""
    logger = Logger(verbose=verbose)

    try:
        config = StatusConfig(proxy_port=proxy_port, verbose=verbose, output=output, dry_run=dry_run)

        status_service = Status(logger=logger)
        result = status_service.status(config)

        if result.success:
            logger.success(status_service.format_output(result, output))
        else:
            logger.error(result.error)
            raise typer.Exit(1)

    except Exception as e:
        logger.error(str(e))
        raise typer.Exit(1)


@proxy_app.command()
def stop(
    proxy_port: int = typer.Option(proxy_port, "--proxy-port", "-p", help="Caddy admin port"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format: text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", help="Dry run"),
):
    """Stop Caddy proxy"""
    logger = Logger(verbose=verbose)

    try:
        config = StopConfig(proxy_port=proxy_port, verbose=verbose, output=output, dry_run=dry_run)

        stop_service = Stop(logger=logger)
        result = stop_service.stop(config)

        if result.success:
            logger.success(stop_service.format_output(result, output))
        else:
            logger.error(result.error)
            raise typer.Exit(1)

    except Exception as e:
        logger.error(str(e))
        raise typer.Exit(1)
