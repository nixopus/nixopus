import typer

from app.utils.config import DEFAULT_BRANCH, DEFAULT_PATH, DEFAULT_REPO, NIXOPUS_CONFIG_DIR, Config
from app.utils.logger import create_logger, log_debug, log_error, log_success
from app.utils.timeout import TimeoutWrapper

from .clone import Clone, CloneConfig
from .messages import (
    debug_action_created,
    debug_branch_param,
    debug_clone_command_invoked,
    debug_clone_operation_completed,
    debug_clone_operation_failed,
    debug_clone_operation_result,
    debug_config_created,
    debug_dry_run_completed,
    debug_dry_run_param,
    debug_exception_caught,
    debug_exception_details,
    debug_executing_dry_run,
    debug_executing_with_timeout,
    debug_force_param,
    debug_output_param,
    debug_path_param,
    debug_repo_param,
    debug_timeout_completed,
    debug_timeout_error,
    debug_timeout_param,
    debug_timeout_wrapper_created,
    debug_verbose_param,
)

config = Config()
nixopus_config_dir = config.get_yaml_value(NIXOPUS_CONFIG_DIR)
repo = config.get_yaml_value(DEFAULT_REPO)
branch = config.get_yaml_value(DEFAULT_BRANCH)
path = nixopus_config_dir + "/" + config.get_yaml_value(DEFAULT_PATH)

clone_app = typer.Typer(help="Clone a repository", invoke_without_command=True)


@clone_app.callback()
def clone_callback(
    repo: str = typer.Option(repo, "--repo", "-r", help="The repository to clone"),
    branch: str = typer.Option(branch, "--branch", "-b", help="The branch to clone"),
    path: str = typer.Option(path, "--path", "-p", help="The path to clone the repository to"),
    force: bool = typer.Option(False, "--force", "-f", help="Force the clone"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run"),
    timeout: int = typer.Option(10, "--timeout", "-t", help="Timeout in seconds"),
):
    """Clone a repository"""
    try:
        logger = create_logger(verbose=verbose)
        log_debug(debug_clone_command_invoked, verbose=verbose)
        log_debug(debug_repo_param.format(repo=repo), verbose=verbose)
        log_debug(debug_branch_param.format(branch=branch), verbose=verbose)
        log_debug(debug_path_param.format(path=path), verbose=verbose)
        log_debug(debug_force_param.format(force=force), verbose=verbose)
        log_debug(debug_verbose_param.format(verbose=verbose), verbose=verbose)
        log_debug(debug_output_param.format(output=output), verbose=verbose)
        log_debug(debug_dry_run_param.format(dry_run=dry_run), verbose=verbose)
        log_debug(debug_timeout_param.format(timeout=timeout), verbose=verbose)

        config = CloneConfig(repo=repo, branch=branch, path=path, force=force, verbose=verbose, output=output, dry_run=dry_run)
        log_debug(debug_config_created.format(config_type="CloneConfig"), verbose=verbose)

        clone_operation = Clone(logger=logger)
        log_debug(debug_action_created.format(action_type="Clone"), verbose=verbose)

        log_debug(debug_timeout_wrapper_created.format(timeout=timeout), verbose=verbose)
        log_debug(debug_executing_with_timeout.format(timeout=timeout), verbose=verbose)

        with TimeoutWrapper(timeout):
            if config.dry_run:
                log_debug(debug_executing_dry_run, verbose=verbose)
                formatted_output = clone_operation.clone_and_format(config)
                log_success(formatted_output, verbose=verbose)
                log_debug(debug_dry_run_completed, verbose=verbose)
            else:
                result = clone_operation.clone(config)
                log_debug(debug_clone_operation_result.format(success=result.success), verbose=verbose)

                if not result.success:
                    log_error(result.output, verbose=verbose)
                    log_debug(debug_clone_operation_failed, verbose=verbose)
                    raise typer.Exit(1)

                log_debug(debug_clone_operation_completed, verbose=verbose)
                log_success(result.output, verbose=verbose)

        log_debug(debug_timeout_completed, verbose=verbose)

    except TimeoutError as e:
        log_debug(debug_timeout_error.format(error=str(e)), verbose=verbose)
        if not isinstance(e, typer.Exit):
            log_error(str(e), verbose=verbose)
        raise typer.Exit(1)
    except Exception as e:
        log_debug(debug_exception_caught.format(error_type=type(e).__name__, error=str(e)), verbose=verbose)
        log_debug(debug_exception_details.format(error=e), verbose=verbose)
        if not isinstance(e, typer.Exit):
            log_error(str(e), verbose=verbose)
        raise typer.Exit(1)
