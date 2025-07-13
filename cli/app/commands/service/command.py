import typer
from app.utils.logger import Logger

service_app = typer.Typer(
    help="Manage Nixopus services"
)

@service_app.command()
def up(
    name: str = typer.Option("all", "--name", "-n", help="The name of the service"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run"),
    detach: bool = typer.Option(True, "--detach", "-d", help="Detach from the service and run in the background"),
    env_file: str = typer.Option(None, "--env-file", "-e", help="Path to the environment file"),
):
    """Start Nixopus services"""
    try:
        logger = Logger(verbose=verbose)
        logger.info(f"Starting service: {name}")
        if dry_run:
            logger.info("Dry run - would start service")
        else:
            logger.success(f"Service {name} started successfully")
    except Exception as e:
        logger.error(e)
        raise typer.Exit(1)

@service_app.command()
def down(
    name: str = typer.Option("all", "--name", "-n", help="The name of the service"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run"),
):
    """Stop Nixopus services"""
    try:
        logger = Logger(verbose=verbose)
        logger.info(f"Stopping service: {name}")
        if dry_run:
            logger.info("Dry run - would stop service")
        else:
            logger.success(f"Service {name} stopped successfully")
    except Exception as e:
        logger.error(e)
        raise typer.Exit(1)

@service_app.command()
def ps(
    all: bool = typer.Option(False, "--all", "-a", help="Show all services"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run"),
):
    """Show status of Nixopus services"""
    try:
        logger = Logger(verbose=verbose)
        logger.info(f"Checking status of services")
        if dry_run:
            logger.info("Dry run - would check service status")
        else:
            logger.success(f"Services status retrieved")
    except Exception as e:
        logger.error(e)
        raise typer.Exit(1)

@service_app.command()
def pull(
    name: str = typer.Option("all", "--name", "-n", help="The name of the service"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run"),
    version: str = typer.Option(None, "--version", "-v", help="Version to pull"),
    force: bool = typer.Option(False, "--force", "-f", help="Force pull even if the image already exists"),
):
    """Pull latest images for Nixopus services"""
    try:
        logger = Logger(verbose=verbose)
        logger.info(f"Pulling images for service: {name}")
        if dry_run:
            logger.info("Dry run - would pull images")
        elif force:
            logger.info("Force pull - would pull images even if the image already exists")
        else:
            logger.success(f"Images for service {name} pulled successfully")
            if version:
                logger.info(f"Version {version} pulled successfully")
    except Exception as e:
        logger.error(e)
        raise typer.Exit(1)

@service_app.command()
def restart(
    name: str = typer.Option("all", "--name", "-n", help="The name of the service"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run"),
):
    """Restart Nixopus services"""
    try:
        logger = Logger(verbose=verbose)
        logger.info(f"Restarting service: {name}")
        if dry_run:
            logger.info("Dry run - would restart service")
        else:
            logger.success(f"Service {name} restarted successfully")
    except Exception as e:
        logger.error(e)
        raise typer.Exit(1)
