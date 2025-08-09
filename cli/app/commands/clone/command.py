import typer

from app.utils.logger import Logger
from app.utils.config import Config, DEFAULT_REPO, DEFAULT_BRANCH, DEFAULT_PATH, NIXOPUS_CONFIG_DIR
from app.utils.timeout import TimeoutWrapper

from .clone import Clone, CloneConfig
from .messages import (
    debug_clone_command_invoked,
    debug_repo_param,
    debug_branch_param,
    debug_path_param,
    debug_force_param,
    debug_verbose_param,
    debug_output_param,
    debug_dry_run_param,
    debug_timeout_param,
    debug_config_created,
    debug_action_created,
    debug_timeout_wrapper_created,
    debug_executing_with_timeout,
    debug_timeout_completed,
    debug_timeout_error,
    debug_executing_dry_run,
    debug_dry_run_completed,
    debug_clone_operation_result,
    debug_clone_operation_failed,
    debug_clone_operation_completed,
    debug_exception_caught,
    debug_exception_details,
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
        logger = Logger(verbose=verbose)
        logger.debug(debug_clone_command_invoked)
        logger.debug(debug_repo_param.format(repo=repo))
        logger.debug(debug_branch_param.format(branch=branch))
        logger.debug(debug_path_param.format(path=path))
        logger.debug(debug_force_param.format(force=force))
        logger.debug(debug_verbose_param.format(verbose=verbose))
        logger.debug(debug_output_param.format(output=output))
        logger.debug(debug_dry_run_param.format(dry_run=dry_run))
        logger.debug(debug_timeout_param.format(timeout=timeout))
        
        config = CloneConfig(repo=repo, branch=branch, path=path, force=force, verbose=verbose, output=output, dry_run=dry_run)
        logger.debug(debug_config_created.format(config_type="CloneConfig"))
        
        clone_operation = Clone(logger=logger)
        logger.debug(debug_action_created.format(action_type="Clone"))
        
        logger.debug(debug_timeout_wrapper_created.format(timeout=timeout))
        logger.debug(debug_executing_with_timeout.format(timeout=timeout))
        
        with TimeoutWrapper(timeout):
            if config.dry_run:
                logger.debug(debug_executing_dry_run)
                formatted_output = clone_operation.clone_and_format(config)
                logger.success(formatted_output)
                logger.debug(debug_dry_run_completed)
            else:
                result = clone_operation.clone(config)
                logger.debug(debug_clone_operation_result.format(success=result.success))
                
                if not result.success:
                    logger.error(result.output)
                    logger.debug(debug_clone_operation_failed)
                    raise typer.Exit(1)
                
                logger.debug(debug_clone_operation_completed)
                logger.success(result.output)
                
        logger.debug(debug_timeout_completed)
                
    except TimeoutError as e:
        logger.debug(debug_timeout_error.format(error=str(e)))
        if not isinstance(e, typer.Exit):
            logger.error(str(e))
        raise typer.Exit(1)
    except Exception as e:
        logger.debug(debug_exception_caught.format(error_type=type(e).__name__, error=str(e)))
        logger.debug(debug_exception_details.format(error=e))
        if not isinstance(e, typer.Exit):
            logger.error(str(e))
        raise typer.Exit(1)
