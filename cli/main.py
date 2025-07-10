import typer
from commands.version.command import version_app, main_version_callback
from commands.preflight.command import preflight_app
from commands.test.command import test_app
from utils.message import application_name, application_description, application_add_completion, application_version_help

app = typer.Typer(
    name=application_name,
    help=application_description,
    add_completion=application_add_completion,
)

@app.callback()
def main(
    version: bool = typer.Option(
        None,
        "--version",
        "-v",
        callback=main_version_callback,
        help=application_version_help,
    )   
):
    pass

app.add_typer(test_app, name="test")
app.add_typer(preflight_app, name="preflight")
app.add_typer(version_app, name="version")

if __name__ == "__main__":
    app()
