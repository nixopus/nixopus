import typer

from app.utils.config import Config, PROXY_PORT
from app.utils.logger import Logger
from app.utils.timeout import TimeoutWrapper

from .load import Load, LoadConfig
from .status import Status, StatusConfig
from .stop import Stop, StopConfig
from .messages import operation_timed_out, unexpected_error

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
    timeout: int = typer.Option(10, "--timeout", "-t", help="Timeout in seconds"),
):
    """Load Caddy proxy configuration"""
    logger = Logger(verbose=verbose)
    
    try:
        config = LoadConfig(proxy_port=proxy_port, verbose=verbose, output=output, dry_run=dry_run, config_file=config_file)
        load_service = Load(logger=logger)
        
        with TimeoutWrapper(timeout):
            result = load_service.load(config)

        output_text = load_service.format_output(result, output)
        if result.success:
            logger.success(output_text)
        else:
            logger.error(output_text)
            raise typer.Exit(1)

    except TimeoutError:
        logger.error(operation_timed_out.format(timeout=timeout))
        raise typer.Exit(1)
    except ValueError as e:
        logger.error(str(e))
        raise typer.Exit(1)
    except Exception as e:
        if not isinstance(e, typer.Exit):
            logger.error(unexpected_error.format(error=str(e)))
        raise typer.Exit(1)


@proxy_app.command()
def status(
    proxy_port: int = typer.Option(proxy_port, "--proxy-port", "-p", help="Caddy admin port"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format: text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", help="Dry run"),
    timeout: int = typer.Option(10, "--timeout", "-t", help="Timeout in seconds"),
):
    """Check Caddy proxy status"""
    logger = Logger(verbose=verbose)

    try:
        config = StatusConfig(proxy_port=proxy_port, verbose=verbose, output=output, dry_run=dry_run)
        status_service = Status(logger=logger)
        
        with TimeoutWrapper(timeout):
            result = status_service.status(config)

        output_text = status_service.format_output(result, output)
        if result.success:
            logger.success(output_text)
        else:
            logger.error(output_text)
            raise typer.Exit(1)

    except TimeoutError:
        logger.error(operation_timed_out.format(timeout=timeout))
        raise typer.Exit(1)
    except ValueError as e:
        logger.error(str(e))
        raise typer.Exit(1)
    except Exception as e:
        if not isinstance(e, typer.Exit):
            logger.error(unexpected_error.format(error=str(e)))
        raise typer.Exit(1)


@proxy_app.command()
def stop(
    proxy_port: int = typer.Option(proxy_port, "--proxy-port", "-p", help="Caddy admin port"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format: text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", help="Dry run"),
    timeout: int = typer.Option(10, "--timeout", "-t", help="Timeout in seconds"),
):
    """Stop Caddy proxy"""
    logger = Logger(verbose=verbose)

    try:
        config = StopConfig(proxy_port=proxy_port, verbose=verbose, output=output, dry_run=dry_run)
        stop_service = Stop(logger=logger)
        
        with TimeoutWrapper(timeout):
            result = stop_service.stop(config)

        output_text = stop_service.format_output(result, output)
        if result.success:
            logger.success(output_text)
        else:
            logger.error(output_text)
            raise typer.Exit(1)

    except TimeoutError:
        logger.error(operation_timed_out.format(timeout=timeout))
        raise typer.Exit(1)
    except ValueError as e:
        logger.error(str(e))
        raise typer.Exit(1)
    except Exception as e:
        if not isinstance(e, typer.Exit):
            logger.error(unexpected_error.format(error=str(e)))
        raise typer.Exit(1)
