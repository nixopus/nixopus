import typer
from .test import TestCommand
from .messages import test_app_help

test_app = typer.Typer(
    help=test_app_help,
    invoke_without_command=True
)

@test_app.callback()
def test_callback(ctx: typer.Context, target: str = typer.Argument(None, help="Test target (e.g., version)")):
    """Run tests (only in DEVELOPMENT environment)"""
    if ctx.invoked_subcommand is None:
        test_command = TestCommand()
        test_command.run(target) 
