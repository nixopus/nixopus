import typer
from typing import Optional

from .install import lxc_install_command
from .config import (
    DEFAULT_CONTAINER_NAME, DEFAULT_LXC_IMAGE, DEFAULT_TIMEOUT
)

lxc_app = typer.Typer(help="LXC container management for Nixopus")

@lxc_app.command(name="install")
def lxc_install(
    name: str = typer.Argument(DEFAULT_CONTAINER_NAME, help="Container name"),
    image: str = typer.Option(DEFAULT_LXC_IMAGE, "--image", "-i", help="LXC image to use"),
    force: bool = typer.Option(False, "--force", "-f", help="Force recreate if container exists"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    timeout: int = typer.Option(DEFAULT_TIMEOUT, "--timeout", "-t", help="Timeout for installation (seconds)"),
    api_domain: Optional[str] = typer.Option(None, "--api-domain", help="API domain for nixopus"),
    view_domain: Optional[str] = typer.Option(None, "--view-domain", help="View domain for nixopus"),
    install_script_url: Optional[str] = typer.Option(None, "--install-url", help="Custom nixopus install script URL"),
    setup_proxy: bool = typer.Option(True, "--setup-proxy/--no-proxy", help="Setup proxy devices"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run mode"),
):
    
    lxc_install_command(
        name=name,
        image=image,
        force=force,
        verbose=verbose,
        timeout=timeout,
        api_domain=api_domain,
        view_domain=view_domain,
        install_script_url=install_script_url,
        setup_proxy=setup_proxy,
        dry_run=dry_run
    )
