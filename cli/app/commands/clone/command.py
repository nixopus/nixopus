import typer

from app.utils.logger import Logger
from app.utils.config import Config, DEFAULT_REPO, DEFAULT_BRANCH, DEFAULT_PATH, NIXOPUS_CONFIG_DIR

from .clone import Clone, CloneConfig

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
):
    """Clone a repository"""
    try:
        logger = Logger(verbose=verbose)
        config = CloneConfig(repo=repo, branch=branch, path=path, force=force, verbose=verbose, output=output, dry_run=dry_run)
        clone_operation = Clone(logger=logger)
        result = clone_operation.clone(config)
        logger.success(result.output)
    except Exception as e:
        logger.error(e)
        raise typer.Exit(1)
