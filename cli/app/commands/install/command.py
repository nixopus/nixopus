import typer
from .run import Install
from .clone import Clone, CloneConfig
from app.utils.logger import Logger

install_app = typer.Typer(
    help="Install Nixopus",
    invoke_without_command=True
)

@install_app.callback()
def install_callback(ctx: typer.Context):
    """Install Nixopus"""
    if ctx.invoked_subcommand is None:
        install = Install()
        install.run()

def main_install_callback(value: bool):
    if value:
        install = Install()
        install.run()
        raise typer.Exit()

@install_app.command()
def clone(
    repo: str = typer.Option(..., help="The repository to clone"),
    branch: str = typer.Option(..., help="The branch to clone"),
    path: str = typer.Option(..., help="The path to clone the repository to"),
    force: bool = typer.Option(False, help="Force the clone"),
    verbose: bool = typer.Option(False, help="Verbose output"),
    output: str = typer.Option("text", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, help="Dry run"),
):
    """Clone a repository"""
    config = CloneConfig(
        repo=repo,
        branch=branch,
        path=path,
        force=force,
        verbose=verbose,
        output=output,
        dry_run=dry_run
    )
    clone_operation = Clone(config)
    clone_operation.run()
