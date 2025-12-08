import typer

from app.utils.logger import create_logger, log_error, log_success, log_warning
from app.utils.timeout import timeout_wrapper

from .deps import install_all_deps
from .run import Install
from .development import DevelopmentInstall
from .ssh import SSHConfig, format_ssh_output, generate_ssh_key_with_config

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
    development: bool = typer.Option(
        False,
        "--development",
        "-D",
        help="Use development workflow (local setup, dev compose, dev env)",
    ),
    dev_path: str = typer.Option(
        None,
        "--dev-path",
        help="Installation directory for development workflow (defaults to current directory)",
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
    api_port: int = typer.Option(None, "--api-port", help="Port for the API service (default: 8443 for production, 8080 for development)"),
    view_port: int = typer.Option(None, "--view-port", help="Port for the View/Frontend service (default: 7443 for production, 3000 for development)"),
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
):
    """Install Nixopus for production"""
    if ctx.invoked_subcommand is None:
        logger = create_logger(verbose=verbose)
        if development:
            # Warn when incompatible production-only options are provided alongside --development
            if api_domain or view_domain:
                log_warning("Ignoring --api-domain/--view-domain in development mode", verbose=verbose)
            dev_install = DevelopmentInstall(
                logger=logger,
                verbose=verbose,
                timeout=timeout,
                force=force,
                dry_run=dry_run,
                config_file=config_file,
                repo=repo,
                branch=branch,
                install_path=dev_path,
                api_port=api_port,
                view_port=view_port,
                db_port=db_port,
                redis_port=redis_port,
                caddy_admin_port=caddy_admin_port,
                caddy_http_port=caddy_http_port,
                caddy_https_port=caddy_https_port,
                supertokens_port=supertokens_port,
                external_db_url=external_db_url,
            )
            dev_install.run()
        else:
            install = Install(
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
            )
            install.run()


def main_install_callback(value: bool):
    if value:
        logger = create_logger(verbose=False)
        install = Install(
            logger=logger,
            verbose=False,
            timeout=300,
            force=False,
            dry_run=False,
            config_file=None,
        )
        install.run()
        raise typer.Exit()


@install_app.command(name="development")
def development(
    path: str = typer.Option(None, "--path", "-p", help="Installation directory (defaults to current directory)"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Show more details while installing"),
    timeout: int = typer.Option(1800, "--timeout", "-t", help="How long to wait for each step (in seconds)"),
    force: bool = typer.Option(False, "--force", "-f", help="Replace files if they already exist"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="See what would happen, but don't make changes"),
    config_file: str = typer.Option(
        None, "--config-file", "-c", help="Path to custom config file (defaults to config.dev.yaml)"
    ),
    repo: str = typer.Option(None, "--repo", "-r", help="GitHub repository URL to clone (defaults to config value)"),
    branch: str = typer.Option(None, "--branch", "-b", help="Git branch to clone (defaults to config value)"),
    api_port: int = typer.Option(None, "--api-port", help="Port for the API service (default: 8080)"),
    view_port: int = typer.Option(None, "--view-port", help="Port for the View/Frontend service (default: 3000)"),
    db_port: int = typer.Option(None, "--db-port", help="Port for the PostgreSQL database (default: 5432)"),
    redis_port: int = typer.Option(None, "--redis-port", help="Port for the Redis service (default: 6379)"),
    caddy_admin_port: int = typer.Option(None, "--caddy-admin-port", help="Port for Caddy admin API (default: 2019)"),
    caddy_http_port: int = typer.Option(None, "--caddy-http-port", help="Port for Caddy HTTP traffic (default: 80)"),
    caddy_https_port: int = typer.Option(None, "--caddy-https-port", help="Port for Caddy HTTPS traffic (default: 443)"),
    supertokens_port: int = typer.Option(None, "--supertokens-port", help="Port for SuperTokens service (default: 3567)"),
    external_db_url: str = typer.Option(None, "--external-db-url", help="External PostgreSQL database connection URL (e.g. postgresql://user:password@host:port/dbname?sslmode=require). If provided, local DB service will be excluded"),
):
    """Install Nixopus for local development in specified or current directory"""
    logger = create_logger(verbose=verbose)
    install = DevelopmentInstall(
        logger=logger,
        verbose=verbose,
        timeout=timeout,
        force=force,
        dry_run=dry_run,
        config_file=config_file,
        repo=repo,
        branch=branch,
        install_path=path,
        api_port=api_port,
        view_port=view_port,
        db_port=db_port,
        redis_port=redis_port,
        caddy_admin_port=caddy_admin_port,
        caddy_http_port=caddy_http_port,
        caddy_https_port=caddy_https_port,
        supertokens_port=supertokens_port,
        external_db_url=external_db_url,
    )
    install.run()


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
