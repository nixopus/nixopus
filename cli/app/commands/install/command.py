import typer

from app.utils.config import Config
from app.utils.logger import Logger
from app.utils.timeout import TimeoutWrapper

from .deps import install_all_deps
from .run import Install
from .ssh import SSH, SSHConfig

install_app = typer.Typer(help="Install Nixopus", invoke_without_command=True)


@install_app.callback()
def install_callback(
    ctx: typer.Context,
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Show more details while installing"),
    timeout: int = typer.Option(300, "--timeout", "-t", help="How long to wait for each step (in seconds)"),
    force: bool = typer.Option(False, "--force", "-f", help="Replace files if they already exist"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="See what would happen, but don't make changes"),
    config_file: str = typer.Option(
        None, "--config-file", "-c", help="Path to custom config file (defaults to built-in config)"
    ),
    api_domain: str = typer.Option(
        None,
        "--api-domain",
        "-ad",
        help="The domain where the nixopus api will be accessible (e.g. api.nixopus.com), if not provided you can use the ip address of the server and the port (e.g. 192.168.1.100:8443)",
    ),
    view_domain: str = typer.Option(
        None,
        "--view-domain",
        "-vd",
        help="The domain where the nixopus view will be accessible (e.g. nixopus.com), if not provided you can use the ip address of the server and the port (e.g. 192.168.1.100:80)",
    ),
    repo: str = typer.Option(
        None, "--repo", "-r", help="GitHub repository URL to clone (defaults to config value)"
    ),
    branch: str = typer.Option(
        None, "--branch", "-b", help="Git branch to clone (defaults to config value)"
    ),
):
    """Install Nixopus"""
    if ctx.invoked_subcommand is None:
        logger = Logger(verbose=verbose)
        install = Install(
            logger=logger,
            verbose=verbose,
            timeout=timeout,
            force=force,
            dry_run=dry_run,
            config_file=config_file,
            api_domain=api_domain,
            view_domain=view_domain,
            repo=repo,
            branch=branch,
        )
        install.run()


def main_install_callback(value: bool):
    if value:
        logger = Logger(verbose=False)
        install = Install(logger=logger, verbose=False, timeout=300, force=False, dry_run=False, config_file=None)
        install.run()
        raise typer.Exit()


@install_app.command(name="ssh")
def ssh(
    path: str = typer.Option("~/.ssh/nixopus_rsa", "--path", "-p", help="The SSH key path to generate"),
    key_type: str = typer.Option("rsa", "--key-type", "-t", help="The SSH key type (rsa, ed25519, ecdsa)"),
    key_size: int = typer.Option(4096, "--key-size", "-s", help="The SSH key size"),
    passphrase: str = typer.Option(None, "--passphrase", "-P", help="The passphrase to use for the SSH key"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run"),
    force: bool = typer.Option(False, "--force", "-f", help="Force overwrite existing SSH key"),
    set_permissions: bool = typer.Option(True, "--set-permissions", "-S", help="Set proper file permissions"),
    add_to_authorized_keys: bool = typer.Option(
        False, "--add-to-authorized-keys", "-a", help="Add public key to authorized_keys"
    ),
    create_ssh_directory: bool = typer.Option(
        True, "--create-ssh-directory", "-c", help="Create .ssh directory if it doesn't exist"
    ),
    timeout: int = typer.Option(10, "--timeout", "-T", help="Timeout in seconds"),
):
    """Generate an SSH key pair with proper permissions and optional authorized_keys integration"""
    try:
        logger = Logger(verbose=verbose)
        config = SSHConfig(
            path=path,
            key_type=key_type,
            key_size=key_size,
            passphrase=passphrase,
            verbose=verbose,
            output=output,
            dry_run=dry_run,
            force=force,
            set_permissions=set_permissions,
            add_to_authorized_keys=add_to_authorized_keys,
            create_ssh_directory=create_ssh_directory,
        )
        ssh_operation = SSH(logger=logger)

        with TimeoutWrapper(timeout):
            result = ssh_operation.generate(config)

        logger.success(result.output)
    except TimeoutError as e:
        logger.error(e)
        raise typer.Exit(1)
    except Exception as e:
        logger.error(e)
        raise typer.Exit(1)


@install_app.command(name="deps")
def deps(
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run"),
    timeout: int = typer.Option(10, "--timeout", "-t", help="Timeout in seconds"),
):
    """Install dependencies"""
    try:
        logger = Logger(verbose=verbose)

        with TimeoutWrapper(timeout):
            result = install_all_deps(verbose=verbose, output=output, dry_run=dry_run)

        if output == "json":
            print(result)
        else:
            logger.success("All dependencies installed successfully.")
    except TimeoutError as e:
        logger.error(e)
        raise typer.Exit(1)
    except Exception as e:
        logger.error(e)
        raise typer.Exit(1)
