from rich.console import Console
from rich.panel import Panel
from rich.text import Text
from importlib.metadata import version

class VersionCommand:
    def __init__(self):
        self.console = Console()

    def run(self):
        """Display the version of the CLI"""
        cli_version = version('nixopus')
        
        version_text = Text()
        version_text.append("Nixopus CLI", style="bold blue")
        version_text.append(f" v{cli_version}", style="green")
        
        panel = Panel(
            version_text,
            title="[bold white]Version Info[/bold white]",
            border_style="blue",
            padding=(0, 1)
        )
        
        self.console.print(panel)
