import typer

from app.utils.logger import Logger
from app.utils.timeout import TimeoutWrapper

from .delete import Delete, DeleteConfig
from .list import List, ListConfig
from .messages import argument_must_be_in_form
from .set import Set, SetConfig

conf_app = typer.Typer(help="Manage configuration")


@conf_app.command()
def list(
    service: str = typer.Option(
        "api", "--service", "-s", help="The name of the service to list configuration for, e.g api,view"
    ),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run"),
    env_file: str = typer.Option(None, "--env-file", "-e", help="Path to the environment file"),
    timeout: int = typer.Option(10, "--timeout", "-t", help="Timeout in seconds"),
):
    """List all configuration"""
    logger = Logger(verbose=verbose)

    try:
        config = ListConfig(service=service, verbose=verbose, output=output, dry_run=dry_run, env_file=env_file)

        list_action = List(logger=logger)
        
        with TimeoutWrapper(timeout):
            result = list_action.list(config)

        if result.success:
            logger.success(list_action.format_output(result, output))
        else:
            logger.error(result.error)
            raise typer.Exit(1)

    except TimeoutError as e:
        logger.error(e)
        raise typer.Exit(1)
    except Exception as e:
        logger.error(str(e))
        raise typer.Exit(1)


@conf_app.command()
def delete(
    service: str = typer.Option(
        "api", "--service", "-s", help="The name of the service to delete configuration for, e.g api,view"
    ),
    key: str = typer.Argument(..., help="The key of the configuration to delete"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run"),
    env_file: str = typer.Option(None, "--env-file", "-e", help="Path to the environment file"),
    timeout: int = typer.Option(10, "--timeout", "-t", help="Timeout in seconds"),
):
    """Delete a configuration"""
    logger = Logger(verbose=verbose)

    try:
        config = DeleteConfig(service=service, key=key, verbose=verbose, output=output, dry_run=dry_run, env_file=env_file)

        delete_action = Delete(logger=logger)
        
        with TimeoutWrapper(timeout):
            result = delete_action.delete(config)

        if result.success:
            logger.success(delete_action.format_output(result, output))
        else:
            logger.error(result.error)
            raise typer.Exit(1)

    except TimeoutError as e:
        logger.error(e)
        raise typer.Exit(1)
    except Exception as e:
        logger.error(str(e))
        raise typer.Exit(1)


@conf_app.command()
def set(
    service: str = typer.Option(
        "api", "--service", "-s", help="The name of the service to set configuration for, e.g api,view"
    ),
    key_value: str = typer.Argument(..., help="Configuration in the form KEY=VALUE"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run"),
    env_file: str = typer.Option(None, "--env-file", "-e", help="Path to the environment file"),
    timeout: int = typer.Option(10, "--timeout", "-t", help="Timeout in seconds"),
):
    """Set a configuration"""
    logger = Logger(verbose=verbose)

    try:
        if "=" not in key_value:
            logger.error(argument_must_be_in_form)
            raise typer.Exit(1)
        key, value = key_value.split("=", 1)

        config = SetConfig(
            service=service, key=key, value=value, verbose=verbose, output=output, dry_run=dry_run, env_file=env_file
        )

        set_action = Set(logger=logger)
        
        with TimeoutWrapper(timeout):
            result = set_action.set(config)

        if result.success:
            logger.success(set_action.format_output(result, output))
        else:
            logger.error(result.error)
            raise typer.Exit(1)

    except TimeoutError as e:
        logger.error(e)
        raise typer.Exit(1)
    except Exception as e:
        logger.error(str(e))
        raise typer.Exit(1)
