import typer

from app.commands.clone.command import clone_app
from app.commands.conf.command import conf_app
from app.commands.install.command import install_app
from app.commands.preflight.command import preflight_app
from app.commands.proxy.command import proxy_app
from app.commands.service.command import service_app
from app.commands.test.command import test_app
from app.commands.version.command import main_version_callback, version_app
from app.utils.message import application_add_completion, application_description, application_name, application_version_help

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
