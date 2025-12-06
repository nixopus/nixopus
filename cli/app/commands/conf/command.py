import typer

from app.utils.logger import create_logger, log_debug, log_error, log_info, log_success
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
        logger = create_logger(verbose=verbose)

        log_debug(debug_conf_command_invoked, verbose=verbose)
        log_debug(debug_service_param.format(service=service), verbose=verbose)
        log_debug(debug_verbose_param.format(verbose=verbose), verbose=verbose)
        log_debug(debug_output_param.format(output=output), verbose=verbose)
        log_debug(debug_dry_run_param.format(dry_run=dry_run), verbose=verbose)
        log_debug(debug_env_file_param.format(env_file=env_file), verbose=verbose)
        log_debug(debug_timeout_param.format(timeout=timeout), verbose=verbose)

        config = ListConfig(service=service, verbose=verbose, output=output, dry_run=dry_run, env_file=env_file)
        log_debug(debug_config_created.format(config_type="ListConfig"), verbose=verbose)

        list_action = List(logger=logger)
        log_debug(debug_action_created.format(action_type="List"), verbose=verbose)

        log_debug(debug_timeout_wrapper_created.format(timeout=timeout), verbose=verbose)
        log_debug(debug_executing_with_timeout.format(timeout=timeout), verbose=verbose)

        with TimeoutWrapper(timeout):
            if config.dry_run:
                log_debug(debug_executing_dry_run, verbose=verbose)
                formatted_output = list_action.list_and_format(config)
                log_info(formatted_output, verbose=verbose)
                log_debug(debug_dry_run_completed, verbose=verbose)
            else:
                result = list_action.list(config)
                log_debug(debug_conf_operation_result.format(success=result.success), verbose=verbose)

                if result.success:
                    formatted_output = list_action.format_output(result, output)
                    log_success(formatted_output, verbose=verbose)
                    log_debug(debug_conf_operation_completed, verbose=verbose)
                else:
                    log_error(result.error, verbose=verbose)
                    log_debug(debug_conf_operation_failed, verbose=verbose)
                    raise typer.Exit(1)

        log_debug(debug_timeout_completed, verbose=verbose)

    except TimeoutError as e:
        log_debug(debug_timeout_error.format(error=str(e)), verbose=verbose)
        log_error(str(e), verbose=verbose)
        raise typer.Exit(1)
    except Exception as e:
        log_debug(debug_exception_caught.format(error_type=type(e).__name__, error=str(e)), verbose=verbose)
        log_debug(debug_exception_details.format(error=e), verbose=verbose)
        if not isinstance(e, typer.Exit):
            log_error(str(e), verbose=verbose)
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
        logger = create_logger(verbose=verbose)

        log_debug(debug_conf_command_invoked, verbose=verbose)
        log_debug(debug_service_param.format(service=service), verbose=verbose)
        log_debug(debug_key_param.format(key=key), verbose=verbose)
        log_debug(debug_verbose_param.format(verbose=verbose), verbose=verbose)
        log_debug(debug_output_param.format(output=output), verbose=verbose)
        log_debug(debug_dry_run_param.format(dry_run=dry_run), verbose=verbose)
        log_debug(debug_env_file_param.format(env_file=env_file), verbose=verbose)
        log_debug(debug_timeout_param.format(timeout=timeout), verbose=verbose)

        config = DeleteConfig(service=service, key=key, verbose=verbose, output=output, dry_run=dry_run, env_file=env_file)
        log_debug(debug_config_created.format(config_type="DeleteConfig"), verbose=verbose)

        delete_action = Delete(logger=logger)
        log_debug(debug_action_created.format(action_type="Delete"), verbose=verbose)

        log_debug(debug_timeout_wrapper_created.format(timeout=timeout), verbose=verbose)
        log_debug(debug_executing_with_timeout.format(timeout=timeout), verbose=verbose)

        with TimeoutWrapper(timeout):
            if config.dry_run:
                log_debug(debug_executing_dry_run, verbose=verbose)
                formatted_output = delete_action.delete_and_format(config)
                log_info(formatted_output, verbose=verbose)
                log_debug(debug_dry_run_completed, verbose=verbose)
            else:
                result = delete_action.delete(config)
                log_debug(debug_conf_operation_result.format(success=result.success), verbose=verbose)

                if result.success:
                    formatted_output = delete_action.format_output(result, output)
                    log_success(formatted_output, verbose=verbose)
                    log_debug(debug_conf_operation_completed, verbose=verbose)
                else:
                    log_error(result.error, verbose=verbose)
                    log_debug(debug_conf_operation_failed, verbose=verbose)
                    raise typer.Exit(1)

        log_debug(debug_timeout_completed, verbose=verbose)

    except TimeoutError as e:
        log_debug(debug_timeout_error.format(error=str(e)), verbose=verbose)
        log_error(str(e), verbose=verbose)
        raise typer.Exit(1)
    except Exception as e:
        log_debug(debug_exception_caught.format(error_type=type(e).__name__, error=str(e)), verbose=verbose)
        log_debug(debug_exception_details.format(error=e), verbose=verbose)
        if not isinstance(e, typer.Exit):
            log_error(str(e), verbose=verbose)
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
        logger = create_logger(verbose=verbose)

        log_debug(debug_conf_command_invoked, verbose=verbose)
        log_debug(debug_service_param.format(service=service), verbose=verbose)
        log_debug(debug_verbose_param.format(verbose=verbose), verbose=verbose)
        log_debug(debug_output_param.format(output=output), verbose=verbose)
        log_debug(debug_dry_run_param.format(dry_run=dry_run), verbose=verbose)
        log_debug(debug_env_file_param.format(env_file=env_file), verbose=verbose)
        log_debug(debug_timeout_param.format(timeout=timeout), verbose=verbose)
        log_debug(debug_parsing_key_value.format(key_value=key_value), verbose=verbose)

        if "=" not in key_value:
            log_debug(debug_key_value_parse_failed.format(key_value=key_value), verbose=verbose)
            log_error(argument_must_be_in_form, verbose=verbose)
            raise typer.Exit(1)

        key, value = key_value.split("=", 1)
        log_debug(debug_key_value_parsed.format(key=key, value=value), verbose=verbose)

        config = SetConfig(
            service=service, key=key, value=value, verbose=verbose, output=output, dry_run=dry_run, env_file=env_file
        )
        log_debug(debug_config_created.format(config_type="SetConfig"), verbose=verbose)

        set_action = Set(logger=logger)
        log_debug(debug_action_created.format(action_type="Set"), verbose=verbose)

        log_debug(debug_timeout_wrapper_created.format(timeout=timeout), verbose=verbose)
        log_debug(debug_executing_with_timeout.format(timeout=timeout), verbose=verbose)

        with TimeoutWrapper(timeout):
            if config.dry_run:
                log_debug(debug_executing_dry_run, verbose=verbose)
                formatted_output = set_action.set_and_format(config)
                log_info(formatted_output, verbose=verbose)
                log_debug(debug_dry_run_completed, verbose=verbose)
            else:
                result = set_action.set(config)
                log_debug(debug_conf_operation_result.format(success=result.success), verbose=verbose)

                if result.success:
                    formatted_output = set_action.format_output(result, output)
                    log_success(formatted_output, verbose=verbose)
                    log_debug(debug_conf_operation_completed, verbose=verbose)
                else:
                    log_error(result.error, verbose=verbose)
                    log_debug(debug_conf_operation_failed, verbose=verbose)
                    raise typer.Exit(1)

        log_debug(debug_timeout_completed, verbose=verbose)

    except TimeoutError as e:
        log_debug(debug_timeout_error.format(error=str(e)), verbose=verbose)
        log_error(str(e), verbose=verbose)
        raise typer.Exit(1)
    except Exception as e:
        log_debug(debug_exception_caught.format(error_type=type(e).__name__, error=str(e)), verbose=verbose)
        log_debug(debug_exception_details.format(error=e), verbose=verbose)
        if not isinstance(e, typer.Exit):
            log_error(str(e), verbose=verbose)
        raise typer.Exit(1)
