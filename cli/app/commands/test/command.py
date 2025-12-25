from typing import Optional
import typer
from app.utils.logger import create_logger
from app.utils.timeout import timeout_wrapper
from .run import run_test
from .types import TestParams

test_app = typer.Typer(help="Test Nixopus installation in LXD containers", invoke_without_command=True)

@test_app.callback()
def test_callback(
    ctx: typer.Context,
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Show more details"),
    timeout: int = typer.Option(600, "--timeout", "-t", help="Timeout in seconds"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Preview changes without executing"),
    image: str = typer.Option(None, "--image", "-i", help="LXD image to use (e.g., debian/11, ubuntu/22.04)"),
    container_name: str = typer.Option(None, "--container-name", "-n", help="Name for the test container"),
    app_port: int = typer.Option(None, "--app-port", help="External port for app proxy (default: auto)"),
    api_port: int = typer.Option(None, "--api-port", help="External port for API proxy (default: auto)"),
    proxy_url: str = typer.Option("127.0.0.1", "--proxy-url", help="Proxy URL for container networking"),
    internal_app_port: int = typer.Option(7443, "--internal-app-port", help="Internal app port in container"),
    internal_api_port: int = typer.Option(8443, "--internal-api-port", help="Internal API port in container"),
    distro: str = typer.Option("images:debian/13", "--distro", help="Distribution to use if image not specified"),
    repo: str = typer.Option(None, "--repo", "-r", help="GitHub repository URL for installation"),
    branch: str = typer.Option(None, "--branch", "-b", help="Git branch for installation"),
    health_check_timeout: int = typer.Option(300, "--health-check-timeout", help="Health check timeout in seconds (default: 300 for LXD)"),
):
    """Test Nixopus installation in an LXD container"""
    if ctx.invoked_subcommand is None:
        logger = create_logger(verbose=verbose)
        params = TestParams(
            logger=logger,
            verbose=verbose,
            timeout=timeout,
            dry_run=dry_run,
            image=image,
            container_name=container_name,
            app_port=app_port,
            api_port=api_port,
            proxy_url=proxy_url,
            internal_app_port=internal_app_port,
            internal_api_port=internal_api_port,
            distro=distro,
            repo=repo,
            branch=branch,
            health_check_timeout=health_check_timeout,
        )
        run_test(params)

@test_app.command(name="list-images")
def list_images(
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Show more details"),
):
    """List available LXD images"""
    logger = create_logger(verbose=verbose)
    params = TestParams(logger=logger, verbose=verbose)
    from .utils import list_lxd_images

    success, error, images = list_lxd_images(params)
    if not success:
        typer.echo(f"Error: {error}", err=True)
        raise typer.Exit(1)

    if images:
        typer.echo("Available LXD images:")
        for image in images:
            typer.echo(f"  - {image}")
    else:
        typer.echo("No images found")

@test_app.command(name="cleanup")
def cleanup(
    container_name: str = typer.Argument(..., help="Container name to cleanup"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Show more details"),
):
    """Clean up a test container"""
    logger = create_logger(verbose=verbose)
    params = TestParams(logger=logger, verbose=verbose, container_name=container_name)
    from .container import cleanup_container

    success, error = cleanup_container(params)
    if not success:
        typer.echo(f"Error: {error}", err=True)
        raise typer.Exit(1)

    typer.echo(f"Container {container_name} cleaned up successfully")
