import typer

from app.utils.logger import Logger

from .run import Install
from .ssh import SSH, SSHConfig

install_app = typer.Typer(help="Install Nixopus", invoke_without_command=True)


@install_app.callback()
def install_callback(ctx: typer.Context):
    """Install Nixopus"""
    if ctx.invoked_subcommand is None:
        install = Install()
        install.run()


def main_install_callback(value: bool):
    if value:
        install = Install()
        install.run()
        raise typer.Exit()

def ssh(
    path: str = typer.Option("~/.ssh/nixopus_ed25519", "--path", "-p", help="The SSH key path to generate"),
    key_type: str = typer.Option("ed25519", "--key-type", "-t", help="The SSH key type (rsa, ed25519, ecdsa)"),
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
        result = ssh_operation.generate(config)
        logger.success(result.output)
    except Exception as e:
        logger.error(e)
        raise typer.Exit(1)
