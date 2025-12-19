from typing import Optional

import typer

from app.utils.logger import create_logger, log_error, log_success
from app.utils.timeout import timeout_wrapper

from .deps import install_all_deps
from .run import run_installation
from .ssh import SSHConfig, format_ssh_output, generate_ssh_key_with_config
from .types import InstallParams

install_app = typer.Typer(help="Install Nixopus", invoke_without_command=True)


@install_app.callback()
def install_callback(
    ctx: typer.Context,
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Show more details while installing"),
    timeout: int = typer.Option(300, "--timeout", "-t", help="How long to wait for each step (in seconds)"),
    force: bool = typer.Option(False, "--force", "-f", help="Overwrite existing installation directory (DELETES all existing data)"),
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
    host_ip: str = typer.Option(
        None,
        "--host-ip",
        "-ip",
        help="The IP address of the server to use when no domains are provided (e.g. 10.0.0.154 or 192.168.1.100). If not provided, the public IP will be automatically detected.",
    ),
    api_port: int = typer.Option(None, "--api-port", help="Port for the API service (default: 8443)"),
    view_port: int = typer.Option(None, "--view-port", help="Port for the View/Frontend service (default: 7443)"),
    db_port: int = typer.Option(None, "--db-port", help="Port for the PostgreSQL database (default: 5432)"),
    redis_port: int = typer.Option(None, "--redis-port", help="Port for the Redis service (default: 6379)"),
    caddy_admin_port: int = typer.Option(None, "--caddy-admin-port", help="Port for Caddy admin API (default: 2019)"),
    caddy_http_port: int = typer.Option(None, "--caddy-http-port", help="Port for Caddy HTTP traffic (default: 80)"),
    caddy_https_port: int = typer.Option(None, "--caddy-https-port", help="Port for Caddy HTTPS traffic (default: 443)"),
    supertokens_port: int = typer.Option(None, "--supertokens-port", help="Port for SuperTokens service (default: 3567)"),
    repo: str = typer.Option(None, "--repo", "-r", help="GitHub repository URL to clone (defaults to config value)"),
    branch: str = typer.Option(None, "--branch", "-b", help="Git branch to clone (defaults to config value)"),
    external_db_url: str = typer.Option(None, "--external-db-url", help="External PostgreSQL database connection URL (e.g. postgresql://user:password@host:port/dbname?sslmode=require). If provided, local DB service will be excluded"),
    staging: bool = typer.Option(False, "--staging", "-s", help="Use staging docker-compose file (docker-compose-staging.yml)"),
    no_rollback: bool = typer.Option(False, "--no-rollback", help="Disable automatic rollback on installation failure"),
    verify_health: bool = typer.Option(True, "--verify-health", help="Verify that all services are healthy after starting (recommended)"),
    health_check_timeout: int = typer.Option(120, "--health-check-timeout", help="Maximum time to wait for services to become healthy (in seconds)"),
    admin_email: Optional[str] = typer.Option(None, "--admin-email", help="Email for admin user registration"),
    admin_password: Optional[str] = typer.Option(None, "--admin-password", help="Password for admin user registration"),
):
    """Install Nixopus for production"""
    if ctx.invoked_subcommand is None:
        logger = create_logger(verbose=verbose)
        params = InstallParams(
            logger=logger,
            verbose=verbose,
            timeout=timeout,
            force=force,
            dry_run=dry_run,
            config_file=config_file,
            api_domain=api_domain,
            view_domain=view_domain,
            host_ip=host_ip,
            repo=repo,
            branch=branch,
            api_port=api_port,
            view_port=view_port,
            db_port=db_port,
            redis_port=redis_port,
            caddy_admin_port=caddy_admin_port,
            caddy_http_port=caddy_http_port,
            caddy_https_port=caddy_https_port,
            supertokens_port=supertokens_port,
            external_db_url=external_db_url,
            staging=staging,
            no_rollback=no_rollback,
            verify_health=verify_health,
            health_check_timeout=health_check_timeout,
            admin_email=admin_email,
            admin_password=admin_password,
        )
        run_installation(params)


def main_install_callback(value: bool):
    if value:
        logger = create_logger(verbose=False)
        params = InstallParams(logger=logger)
        run_installation(params)
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
    logger = create_logger(verbose=verbose)
    try:
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
        with timeout_wrapper(timeout):
            result = generate_ssh_key_with_config(config, logger=logger)

        output = format_ssh_output(result, result.output)
        log_success(output, verbose=verbose)
    except TimeoutError as e:
        log_error(str(e), verbose=verbose)
        raise typer.Exit(1)
    except Exception as e:
        log_error(str(e), verbose=verbose)
        raise typer.Exit(1)


@install_app.command(name="deps")
def deps(
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run"),
    timeout: int = typer.Option(10, "--timeout", "-t", help="Timeout in seconds"),
):
    """Install dependencies"""
    logger = create_logger(verbose=verbose)
    try:

        with timeout_wrapper(timeout):
            result = install_all_deps(verbose=verbose, output=output, dry_run=dry_run)

        if output == "json":
            print(result)
        else:
            log_success("All dependencies installed successfully.", verbose=verbose)
    except TimeoutError as e:
        log_error(str(e), verbose=verbose)
        raise typer.Exit(1)
    except Exception as e:
        log_error(str(e), verbose=verbose)
        raise typer.Exit(1)
