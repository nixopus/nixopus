import typer
from app.utils.logger import Logger
from .list import List, ListConfig
from .set import Set, SetConfig
from .delete import Delete, DeleteConfig

conf_app = typer.Typer(help="Manage configuration")

@conf_app.command()
def list(
    service: str = typer.Option("api", "--service", "-s", help="The name of the service to list configuration for, e.g api,view"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run"),
    env_file: str = typer.Option(None, "--env-file", "-e", help="Path to the environment file"),
):
    """List all configuration"""
    logger = Logger(verbose=verbose)
    
    try:
        config = ListConfig(
            service=service,
            verbose=verbose,
            output=output,
            dry_run=dry_run,
            env_file=env_file
        )
        
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
    service: str = typer.Option("api", "--service", "-s", help="The name of the service to delete configuration for, e.g api,view"),
    key: str = typer.Option(None, "--key", "-k", help="The key of the configuration to delete"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run"),
    env_file: str = typer.Option(None, "--env-file", "-e", help="Path to the environment file"),
):
    """Delete a configuration"""
    logger = Logger(verbose=verbose)
    
    try:
        config = DeleteConfig(
            service=service,
            key=key,
            verbose=verbose,
            output=output,
            dry_run=dry_run,
            env_file=env_file
        )
        
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
    service: str = typer.Option("api", "--service", "-s", help="The name of the service to set configuration for, e.g api,view"),
    key: str = typer.Option(None, "--key", "-k", help="The key of the configuration to set"),
    value: str = typer.Option(None, "--value", "-v", help="The value of the configuration to set"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run"),
    env_file: str = typer.Option(None, "--env-file", "-e", help="Path to the environment file"),
):
    """Set a configuration"""
    logger = Logger(verbose=verbose)
    
    try:
        config = SetConfig(
            service=service,
            key=key,
            value=value,
            verbose=verbose,
            output=output,
            dry_run=dry_run,
            env_file=env_file
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
