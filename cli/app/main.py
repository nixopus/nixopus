import os
import time

import typer
from rich.console import Console
from rich.panel import Panel
from rich.text import Text

from app.commands.install.command import install_app
from app.commands.uninstall.uninstall import uninstall_app
from app.commands.update.update import update_app
from app.commands.version.command import get_version, main_version_callback, version_app
from app.utils.message import application_add_completion, application_description, application_name, application_version_help

app = typer.Typer(
    name=application_name,
    help=application_description,
    add_completion=application_add_completion,
)


@app.callback(invoke_without_command=True)
def main(
    ctx: typer.Context,
    version: bool = typer.Option(
        None,
        "--version",
        "-v",
        callback=main_version_callback,
        help=application_version_help,
    ),
):
    if ctx.invoked_subcommand is None:
        console = Console()

        ascii_art = r"""                                    
                              @%%@                                
                             @%--+%                              
                          @@%#=---=%%@                           
                        %%=-----------=%@                        
                      %=----------------=*%                      
                    @#--------------------=%                     
                    #----+#%#=-----=###=---=%                    
                   @=--=-.....+=-==.....==--#                    
                   %=-=....=-..=+=..=-...==-*@                   
                   @=-*...+%#:..=..-%*=...=-%@                   
                    %-+....*+.+=-+=.**....==%                    
           @%==#%   @#=+....*-------+....#=%   @%*=+%            
             @%--#@  %==*....%-+*+=#....*=+@  @#==#@             
             @%--+@  %=--==....+*=....+=--#@  %===#@             
             @=---+##=-------------=---====*##====#@             
              %--------------------===============%              
               @=-----=+----------=======#======*@               
                  @@@@*----+------========%@@@@                  
                 %%#%=---=*#=--=#==-=#+=====%%%%                 
                 @=----=*%@+---+@====#@%+=====#@                 
                    @@@  @#=--=@ %====%   @@@                    
                        @*==*%@   @%*==%%     
        """

        text = Text(ascii_art, style="bold cyan")
        panel = Panel(text, title="[bold white]Welcome to[/bold white]", border_style="cyan", padding=(1, 2))

        console.print(panel)

        cli_version = get_version()
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


app.add_typer(install_app, name="install")
app.add_typer(uninstall_app, name="uninstall")
app.add_typer(update_app, name="update")
app.add_typer(version_app, name="version")

if __name__ == "__main__":
    app()
