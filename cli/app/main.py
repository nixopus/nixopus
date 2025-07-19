import typer
from importlib.metadata import version as get_version
from rich.console import Console
from rich.panel import Panel
from rich.text import Text

from app.commands.clone.command import clone_app
from app.commands.conf.command import conf_app
from app.commands.install.command import install_app
from app.commands.preflight.command import preflight_app
from app.commands.proxy.command import proxy_app
from app.commands.service.command import service_app
from app.commands.test.command import test_app
from app.commands.version.command import main_version_callback, version_app
from app.commands.version.version import VersionCommand
from app.utils.message import application_add_completion, application_description, application_name, application_version_help

app = typer.Typer(
    name=application_name,
    help=application_description,
    add_completion=application_add_completion,
)


@app.callback(invoke_without_command=True)
def main(
    version: bool = typer.Option(
        None,
        "--version",
        "-v",
        callback=main_version_callback,
        help=application_version_help,
    )
):
    console = Console()
    
    ascii_art = """
  _   _ _ _                           
 | \\ | (_)                          
 |  \\| |___  _____  _ __  _   _ ___ 
 | . ` | \\ \\/ / _ \\| '_ \\| | | / __|
 | |\\  | |>  < (_) | |_) | |_| \\__ \\
 |_| \\_|_/_/\\_\\___/| .__/ \\__,_|___/
                   | |              
                   |_|              
    """
    
    text = Text(ascii_art, style="bold cyan")
    panel = Panel(text, title="[bold white]Welcome to[/bold white]", border_style="cyan", padding=(1, 2))
    
    console.print(panel)
    
    cli_version = get_version("nixopus")
    version_text = Text()
    version_text.append("Version: ", style="bold white")
    version_text.append(f"v{cli_version}", style="green")
    
    description_text = Text()
    description_text.append(application_description, style="dim")
    
    console.print(version_text)
    console.print(description_text)
    console.print()
    
    help_text = Text()
    help_text.append("Run ", style="dim")
    help_text.append("nixopus --help", style="bold green")
    help_text.append(" to explore all available commands", style="dim")
    console.print(help_text)
    console.print()

app.add_typer(preflight_app, name="preflight")
app.add_typer(clone_app, name="clone")
app.add_typer(conf_app, name="conf")
app.add_typer(service_app, name="service")
app.add_typer(proxy_app, name="proxy")
app.add_typer(install_app, name="install")
app.add_typer(version_app, name="version")
app.add_typer(test_app, name="test")

if __name__ == "__main__":
    app()
