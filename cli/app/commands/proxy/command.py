import typer

from app.utils.config import PROXY_PORT, Config
from app.utils.logger import create_logger, log_error, log_success
from app.utils.timeout import TimeoutWrapper

from .load import Load, LoadConfig
from .messages import operation_timed_out, unexpected_error
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
    timeout: int = typer.Option(10, "--timeout", "-t", help="Timeout in seconds"),
):
    """Load Caddy proxy configuration"""
    logger = create_logger(verbose=verbose)

    try:
        config = LoadConfig(proxy_port=proxy_port, verbose=verbose, output=output, dry_run=dry_run, config_file=config_file)
        load_service = Load(logger=logger)

        with TimeoutWrapper(timeout):
            result = load_service.load(config)

        output_text = load_service.format_output(result, output)
        if result.success:
            log_success(output_text, verbose=verbose)
        else:
            log_error(output_text, verbose=verbose)
            raise typer.Exit(1)

    except TimeoutError:
        log_error(operation_timed_out.format(timeout=timeout), verbose=verbose)
        raise typer.Exit(1)
    except ValueError as e:
        log_error(str(e), verbose=verbose)
        raise typer.Exit(1)
    except Exception as e:
        if not isinstance(e, typer.Exit):
            log_error(unexpected_error.format(error=str(e)), verbose=verbose)
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
    logger = create_logger(verbose=verbose)

    try:
        config = StatusConfig(proxy_port=proxy_port, verbose=verbose, output=output, dry_run=dry_run)
        status_service = Status(logger=logger)

        with TimeoutWrapper(timeout):
            result = status_service.status(config)

        output_text = status_service.format_output(result, output)
        if result.success:
            log_success(output_text, verbose=verbose)
        else:
            log_error(output_text, verbose=verbose)
            raise typer.Exit(1)

    except TimeoutError:
        log_error(operation_timed_out.format(timeout=timeout), verbose=verbose)
        raise typer.Exit(1)
    except ValueError as e:
        log_error(str(e), verbose=verbose)
        raise typer.Exit(1)
    except Exception as e:
        if not isinstance(e, typer.Exit):
            log_error(unexpected_error.format(error=str(e)), verbose=verbose)
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
    logger = create_logger(verbose=verbose)

    try:
        config = StopConfig(proxy_port=proxy_port, verbose=verbose, output=output, dry_run=dry_run)
        stop_service = Stop(logger=logger)

        with TimeoutWrapper(timeout):
            result = stop_service.stop(config)

        output_text = stop_service.format_output(result, output)
        if result.success:
            log_success(output_text, verbose=verbose)
        else:
            log_error(output_text, verbose=verbose)
            raise typer.Exit(1)

    except TimeoutError:
        log_error(operation_timed_out.format(timeout=timeout), verbose=verbose)
        raise typer.Exit(1)
    except ValueError as e:
        log_error(str(e), verbose=verbose)
        raise typer.Exit(1)
    except Exception as e:
        if not isinstance(e, typer.Exit):
            log_error(unexpected_error.format(error=str(e)), verbose=verbose)
        raise typer.Exit(1)
