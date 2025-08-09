import typer

from app.utils.logger import Logger
from app.utils.timeout import TimeoutWrapper
from .run import Uninstall

uninstall_app = typer.Typer(help="Uninstall Nixopus", invoke_without_command=True)


@uninstall_app.callback()
def uninstall_callback(
    ctx: typer.Context,
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Show more details while uninstalling"),
    timeout: int = typer.Option(300, "--timeout", "-t", help="How long to wait for each step (in seconds)"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="See what would happen, but don't make changes"),
    force: bool = typer.Option(False, "--force", "-f", help="Remove files without confirmation prompts"),
):
    """Uninstall Nixopus completely from the system"""
    if ctx.invoked_subcommand is None:
        logger = Logger(verbose=verbose)
        uninstall = Uninstall(
            logger=logger, 
            verbose=verbose, 
            timeout=timeout, 
            dry_run=dry_run,
            force=force
        )
        uninstall.run()