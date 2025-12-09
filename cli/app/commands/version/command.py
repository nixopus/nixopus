import sys
from importlib.metadata import PackageNotFoundError, version
from pathlib import Path

import typer
from rich.console import Console
from rich.panel import Panel
from rich.text import Text

from app.utils.message import application_version_help

version_app = typer.Typer(help=application_version_help, invoke_without_command=True)


def get_version() -> str:
    """Get version from package metadata or bundled version.txt"""
    try:
        return version("nixopus")
    except PackageNotFoundError:
        # Read from version.txt (works in both dev and PyInstaller bundle)
        if getattr(sys, 'frozen', False) and hasattr(sys, '_MEIPASS'):
            # Running from PyInstaller bundle
            version_file = Path(sys._MEIPASS) / "version.txt"
        else:
            # Running from source
            version_file = Path(__file__).parent.parent.parent.parent.parent / "version.txt"

        if version_file.exists():
            return version_file.read_text().strip().lstrip("v")

        return "unknown"


def run_version():
    """Display the version of the CLI"""
    console = Console()
    cli_version = get_version()

    version_text = Text()
    version_text.append("Nixopus CLI", style="bold blue")
    version_text.append(f" v{cli_version}", style="green")

    panel = Panel(version_text, title="[bold white]Version Info[/bold white]", border_style="blue", padding=(0, 1))

    console.print(panel)


@version_app.callback()
def version_callback(ctx: typer.Context):
    """Show version information (default)"""
    if ctx.invoked_subcommand is None:
        run_version()


def main_version_callback(value: bool):
    if value:
        run_version()
        raise typer.Exit()
