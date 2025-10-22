from app.commands.update.run import Update
from app.utils.logger import Logger
import typer

update_app = typer.Typer(help="Update Nixopus", invoke_without_command=True)


@update_app.callback()
def update_callback(
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Show more details while updating"),
):
    """Update Nixopus"""
    logger = Logger(verbose=verbose)
    update = Update(logger=logger)
    update.run()


@update_app.command()
def cli(
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Show more details while updating"),
):
    """Update CLI tool"""
    logger = Logger(verbose=verbose)
    update = Update(logger=logger)
    update.update_cli()
