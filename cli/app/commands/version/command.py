import typer
from .version import VersionCommand
from app.utils.message import application_version_help

version_app = typer.Typer(
    help=application_version_help,
    invoke_without_command=True
)

@version_app.callback()
def version_callback(ctx: typer.Context):
    """Show version information (default)"""
    if ctx.invoked_subcommand is None:
        version_command = VersionCommand()
        version_command.run()

def main_version_callback(value: bool):
    if value:
        version_command = VersionCommand()
        version_command.run()
        raise typer.Exit()
