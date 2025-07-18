import typer

from app.utils.config import API_SERVICE, Config, VIEW_SERVICE, API_ENV_FILE, VIEW_ENV_FILE
from app.utils.logger import Logger
from app.utils.lib import FileManager
from app.commands.conf.base import BaseEnvironmentManager

from .delete import Delete, DeleteConfig
from .list import List, ListConfig
from .messages import all_env_vars_set_successfully, argument_must_be_in_form, failed_to_set_env_vars
from .set import Set, SetConfig
from .run import conf_run

conf_app = typer.Typer(help="Manage configuration")
config = Config()
api_env_file = config.get_yaml_value(API_ENV_FILE)
view_env_file = config.get_yaml_value(VIEW_ENV_FILE)

@conf_app.callback(invoke_without_command=True)
def conf_callback(ctx: typer.Context, output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"), verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output")):
    if ctx.invoked_subcommand is None:
        conf_run(verbose=verbose, output=output)
        
@conf_app.command()
def list(
    service: str = typer.Option(
        "api", "--service", "-s", help="The name of the service to list configuration for, e.g api,view"
    ),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run"),
    env_file: str = typer.Option(None, "--env-file", "-e", help="Path to the environment file"),
):
    """List all configuration"""
    logger = Logger(verbose=verbose)

    try:
        config = ListConfig(service=service, verbose=verbose, output=output, dry_run=dry_run, env_file=env_file)

        list_action = List(logger=logger)
        result = list_action.list(config)

        if result.success:
            logger.success(list_action.format_output(result, output))
        else:
            logger.error(result.error)
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
):
    """Delete a configuration"""
    logger = Logger(verbose=verbose)

    try:
        config = DeleteConfig(service=service, key=key, verbose=verbose, output=output, dry_run=dry_run, env_file=env_file)

        delete_action = Delete(logger=logger)
        result = delete_action.delete(config)

        if result.success:
            logger.success(delete_action.format_output(result, output))
        else:
            logger.error(result.error)
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
        result = set_action.set(config)

        if result.success:
            logger.success(set_action.format_output(result, output))
        else:
            logger.error(result.error)
            raise typer.Exit(1)

    except Exception as e:
        logger.error(str(e))
        raise typer.Exit(1)
