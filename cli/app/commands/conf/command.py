import typer

from app.utils.logger import Logger
from app.utils.timeout import TimeoutWrapper

from .delete import Delete, DeleteConfig
from .list import List, ListConfig
from .messages import (
    argument_must_be_in_form,
    debug_action_created,
    debug_conf_command_invoked,
    debug_conf_operation_completed,
    debug_conf_operation_failed,
    debug_conf_operation_result,
    debug_config_created,
    debug_dry_run_completed,
    debug_dry_run_param,
    debug_env_file_param,
    debug_exception_caught,
    debug_exception_details,
    debug_executing_dry_run,
    debug_executing_with_timeout,
    debug_key_param,
    debug_key_value_parse_failed,
    debug_key_value_parsed,
    debug_output_param,
    debug_parsing_key_value,
    debug_service_param,
    debug_timeout_completed,
    debug_timeout_error,
    debug_timeout_param,
    debug_timeout_wrapper_created,
    debug_value_param,
    debug_verbose_param,
)
from .set import Set, SetConfig

conf_app = typer.Typer(help="Manage configuration")


@conf_app.command()
def list(
    service: str = typer.Option(
        "api", "--service", "-s", help="The name of the service to list configuration for, e.g api,view"
    ),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run"),
    env_file: str = typer.Option(None, "--env-file", "-e", help="Path to the environment file"),
    timeout: int = typer.Option(10, "--timeout", "-t", help="Timeout in seconds"),
):
    """List all configuration"""
    try:
        logger = Logger(verbose=verbose)

        logger.debug(debug_conf_command_invoked)
        logger.debug(debug_service_param.format(service=service))
        logger.debug(debug_verbose_param.format(verbose=verbose))
        logger.debug(debug_output_param.format(output=output))
        logger.debug(debug_dry_run_param.format(dry_run=dry_run))
        logger.debug(debug_env_file_param.format(env_file=env_file))
        logger.debug(debug_timeout_param.format(timeout=timeout))

        config = ListConfig(service=service, verbose=verbose, output=output, dry_run=dry_run, env_file=env_file)
        logger.debug(debug_config_created.format(config_type="ListConfig"))

        list_action = List(logger=logger)
        logger.debug(debug_action_created.format(action_type="List"))

        logger.debug(debug_timeout_wrapper_created.format(timeout=timeout))
        logger.debug(debug_executing_with_timeout.format(timeout=timeout))

        with TimeoutWrapper(timeout):
            if config.dry_run:
                logger.debug(debug_executing_dry_run)
                formatted_output = list_action.list_and_format(config)
                logger.info(formatted_output)
                logger.debug(debug_dry_run_completed)
            else:
                result = list_action.list(config)
                logger.debug(debug_conf_operation_result.format(success=result.success))

                if result.success:
                    formatted_output = list_action.format_output(result, output)
                    logger.success(formatted_output)
                    logger.debug(debug_conf_operation_completed)
                else:
                    logger.error(result.error)
                    logger.debug(debug_conf_operation_failed)
                    raise typer.Exit(1)

        logger.debug(debug_timeout_completed)

    except TimeoutError as e:
        logger.debug(debug_timeout_error.format(error=str(e)))
        logger.error(str(e))
        raise typer.Exit(1)
    except Exception as e:
        logger.debug(debug_exception_caught.format(error_type=type(e).__name__, error=str(e)))
        logger.debug(debug_exception_details.format(error=e))
        if not isinstance(e, typer.Exit):
            logger.error(str(e))
        raise typer.Exit(1)


@conf_app.command()
def delete(
    service: str = typer.Option(
        "api", "--service", "-s", help="The name of the service to delete configuration for, e.g api,view"
    ),
    key: str = typer.Argument(..., help="The key of the configuration to delete"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run"),
    env_file: str = typer.Option(None, "--env-file", "-e", help="Path to the environment file"),
    timeout: int = typer.Option(10, "--timeout", "-t", help="Timeout in seconds"),
):
    """Delete a configuration"""
    try:
        logger = Logger(verbose=verbose)

        logger.debug(debug_conf_command_invoked)
        logger.debug(debug_service_param.format(service=service))
        logger.debug(debug_key_param.format(key=key))
        logger.debug(debug_verbose_param.format(verbose=verbose))
        logger.debug(debug_output_param.format(output=output))
        logger.debug(debug_dry_run_param.format(dry_run=dry_run))
        logger.debug(debug_env_file_param.format(env_file=env_file))
        logger.debug(debug_timeout_param.format(timeout=timeout))

        config = DeleteConfig(service=service, key=key, verbose=verbose, output=output, dry_run=dry_run, env_file=env_file)
        logger.debug(debug_config_created.format(config_type="DeleteConfig"))

        delete_action = Delete(logger=logger)
        logger.debug(debug_action_created.format(action_type="Delete"))

        logger.debug(debug_timeout_wrapper_created.format(timeout=timeout))
        logger.debug(debug_executing_with_timeout.format(timeout=timeout))

        with TimeoutWrapper(timeout):
            if config.dry_run:
                logger.debug(debug_executing_dry_run)
                formatted_output = delete_action.delete_and_format(config)
                logger.info(formatted_output)
                logger.debug(debug_dry_run_completed)
            else:
                result = delete_action.delete(config)
                logger.debug(debug_conf_operation_result.format(success=result.success))

                if result.success:
                    formatted_output = delete_action.format_output(result, output)
                    logger.success(formatted_output)
                    logger.debug(debug_conf_operation_completed)
                else:
                    logger.error(result.error)
                    logger.debug(debug_conf_operation_failed)
                    raise typer.Exit(1)

        logger.debug(debug_timeout_completed)

    except TimeoutError as e:
        logger.debug(debug_timeout_error.format(error=str(e)))
        logger.error(str(e))
        raise typer.Exit(1)
    except Exception as e:
        logger.debug(debug_exception_caught.format(error_type=type(e).__name__, error=str(e)))
        logger.debug(debug_exception_details.format(error=e))
        if not isinstance(e, typer.Exit):
            logger.error(str(e))
        raise typer.Exit(1)


@conf_app.command()
def set(
    service: str = typer.Option(
        "api", "--service", "-s", help="The name of the service to set configuration for, e.g api,view"
    ),
    key_value: str = typer.Argument(..., help="Configuration in the form KEY=VALUE"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Verbose output"),
    output: str = typer.Option("text", "--output", "-o", help="Output format, text, json"),
    dry_run: bool = typer.Option(False, "--dry-run", "-d", help="Dry run"),
    env_file: str = typer.Option(None, "--env-file", "-e", help="Path to the environment file"),
    timeout: int = typer.Option(10, "--timeout", "-t", help="Timeout in seconds"),
):
    """Set a configuration"""
    try:
        logger = Logger(verbose=verbose)

        logger.debug(debug_conf_command_invoked)
        logger.debug(debug_service_param.format(service=service))
        logger.debug(debug_verbose_param.format(verbose=verbose))
        logger.debug(debug_output_param.format(output=output))
        logger.debug(debug_dry_run_param.format(dry_run=dry_run))
        logger.debug(debug_env_file_param.format(env_file=env_file))
        logger.debug(debug_timeout_param.format(timeout=timeout))
        logger.debug(debug_parsing_key_value.format(key_value=key_value))

        if "=" not in key_value:
            logger.debug(debug_key_value_parse_failed.format(key_value=key_value))
            logger.error(argument_must_be_in_form)
            raise typer.Exit(1)

        key, value = key_value.split("=", 1)
        logger.debug(debug_key_value_parsed.format(key=key, value=value))

        config = SetConfig(
            service=service, key=key, value=value, verbose=verbose, output=output, dry_run=dry_run, env_file=env_file
        )
        logger.debug(debug_config_created.format(config_type="SetConfig"))

        set_action = Set(logger=logger)
        logger.debug(debug_action_created.format(action_type="Set"))

        logger.debug(debug_timeout_wrapper_created.format(timeout=timeout))
        logger.debug(debug_executing_with_timeout.format(timeout=timeout))

        with TimeoutWrapper(timeout):
            if config.dry_run:
                logger.debug(debug_executing_dry_run)
                formatted_output = set_action.set_and_format(config)
                logger.info(formatted_output)
                logger.debug(debug_dry_run_completed)
            else:
                result = set_action.set(config)
                logger.debug(debug_conf_operation_result.format(success=result.success))

                if result.success:
                    formatted_output = set_action.format_output(result, output)
                    logger.success(formatted_output)
                    logger.debug(debug_conf_operation_completed)
                else:
                    logger.error(result.error)
                    logger.debug(debug_conf_operation_failed)
                    raise typer.Exit(1)

        logger.debug(debug_timeout_completed)

    except TimeoutError as e:
        logger.debug(debug_timeout_error.format(error=str(e)))
        logger.error(str(e))
        raise typer.Exit(1)
    except Exception as e:
        logger.debug(debug_exception_caught.format(error_type=type(e).__name__, error=str(e)))
        logger.debug(debug_exception_details.format(error=e))
        if not isinstance(e, typer.Exit):
            logger.error(str(e))
        raise typer.Exit(1)
