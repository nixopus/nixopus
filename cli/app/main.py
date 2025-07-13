import typer
from app.commands.version.command import version_app, main_version_callback
from app.commands.preflight.command import preflight_app
from app.commands.test.command import test_app
from app.commands.install.command import install_app
from app.utils.message import application_name, application_description, application_add_completion, application_version_help

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
app.add_typer(install_app, name="install")

if __name__ == "__main__":
    app()
