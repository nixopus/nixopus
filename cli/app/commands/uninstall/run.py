import os
import shutil
import subprocess
from pathlib import Path

import typer
from rich.progress import BarColumn, Progress, SpinnerColumn, TaskProgressColumn, TextColumn

from app.commands.service.down import Down, DownConfig
from app.utils.config import DEFAULT_COMPOSE_FILE, NIXOPUS_CONFIG_DIR, SSH_FILE_PATH, Config
from app.utils.protocols import LoggerProtocol
from app.utils.timeout import TimeoutWrapper

from .messages import (
    authorized_keys_not_found,
    compose_file_not_found_skip,
    config_dir_not_exist_skip,
    config_directory_removal_failed,
    failed_at_step,
    operation_timed_out,
    removed_config_dir,
    removed_private_key,
    removed_public_key,
    removed_ssh_key_from,
    services_stop_failed,
    skipped_removal_config_dir,
    ssh_key_not_found_in_authorized_keys,
    ssh_keys_removal_failed,
    ssh_public_key_not_found_skip,
    uninstall_completed,
    uninstall_completed_info,
    uninstall_dry_run_mode,
    uninstall_failed,
    uninstall_thank_you,
    uninstalling_nixopus,
)

_config = Config()
_config_dir = _config.get_yaml_value(NIXOPUS_CONFIG_DIR)
_compose_file = _config.get_yaml_value(DEFAULT_COMPOSE_FILE)
_ssh_key_path = _config_dir + "/" + _config.get_yaml_value(SSH_FILE_PATH)


class Uninstall:
    def __init__(
        self,
        logger: LoggerProtocol = None,
        verbose: bool = False,
        timeout: int = 300,
        dry_run: bool = False,
        force: bool = False,
    ):
        self.logger = logger
        self.verbose = verbose
        self.timeout = timeout
        self.dry_run = dry_run
        self.force = force
        self.progress = None
        self.main_task = None

    def run(self):
        steps = [
            ("Stopping services", self._stop_services),
            ("Removing SSH keys", self._remove_ssh_keys),
            ("Removing configuration directory", self._remove_config_directory),
        ]

        try:
            if self.dry_run:
                self.logger.info(uninstall_dry_run_mode)
                for step_name, _ in steps:
                    self.logger.info(f"Would execute: {step_name}")
                return

            with Progress(
                SpinnerColumn(),
                TextColumn("[progress.description]{task.description}"),
                BarColumn(),
                TaskProgressColumn(),
                transient=True,
                refresh_per_second=2,
            ) as progress:
                self.progress = progress
                self.main_task = progress.add_task(uninstalling_nixopus, total=len(steps))

                for i, (step_name, step_func) in enumerate(steps):
                    progress.update(self.main_task, description=f"{uninstalling_nixopus} - {step_name} ({i+1}/{len(steps)})")
                    try:
                        step_func()
                        progress.advance(self.main_task, 1)
                    except Exception as e:
                        progress.update(self.main_task, description=failed_at_step.format(step_name=step_name))
                        raise

                progress.update(self.main_task, completed=True, description=uninstall_completed)

            self._show_success_message()

        except Exception as e:
            self._handle_uninstall_error(e)
            self.logger.error(f"{uninstall_failed}: {str(e)}")
            raise typer.Exit(1)

    def _handle_uninstall_error(self, error, context=""):
        context_msg = f" during {context}" if context else ""
        if self.verbose:
            self.logger.error(f"{uninstall_failed}{context_msg}: {str(error)}")
        else:
            self.logger.error(f"{uninstall_failed}{context_msg}")

    def _stop_services(self):
        compose_file_path = os.path.join(_config_dir, _compose_file)

        if not os.path.exists(compose_file_path):
            self.logger.debug(compose_file_not_found_skip.format(compose_file_path=compose_file_path))
            return

        try:
            config = DownConfig(
                name="all", env_file=None, verbose=self.verbose, output="text", dry_run=False, compose_file=compose_file_path
            )

            down_service = Down(logger=self.logger)

            with TimeoutWrapper(self.timeout):
                result = down_service.down(config)

            if not result.success:
                raise Exception(f"{services_stop_failed}: {result.error}")

        except TimeoutError:
            raise Exception(f"{services_stop_failed}: {operation_timed_out}")

    def _remove_ssh_keys(self):
        ssh_key_path = Path(_ssh_key_path)
        public_key_path = ssh_key_path.with_suffix(".pub")

        if not public_key_path.exists():
            self.logger.debug(ssh_public_key_not_found_skip.format(public_key_path=public_key_path))
            return

        try:
            with open(public_key_path, "r") as f:
                public_key_content = f.read().strip()

            authorized_keys_path = Path.home() / ".ssh" / "authorized_keys"

            if not authorized_keys_path.exists():
                self.logger.debug(authorized_keys_not_found)
                return

            with open(authorized_keys_path, "r") as f:
                lines = f.readlines()

            original_count = len(lines)
            filtered_lines = [line for line in lines if public_key_content not in line]

            if len(filtered_lines) < original_count:
                with open(authorized_keys_path, "w") as f:
                    f.writelines(filtered_lines)
                self.logger.debug(removed_ssh_key_from.format(authorized_keys_path=authorized_keys_path))
            else:
                self.logger.debug(ssh_key_not_found_in_authorized_keys)

            if ssh_key_path.exists():
                ssh_key_path.unlink()
                self.logger.debug(removed_private_key.format(ssh_key_path=ssh_key_path))

            if public_key_path.exists():
                public_key_path.unlink()
                self.logger.debug(removed_public_key.format(public_key_path=public_key_path))

        except Exception as e:
            raise Exception(f"{ssh_keys_removal_failed}: {str(e)}")

    def _remove_config_directory(self):
        config_dir_path = Path(_config_dir)

        if not config_dir_path.exists():
            self.logger.debug(config_dir_not_exist_skip.format(config_dir_path=config_dir_path))
            return

        try:
            if self.force or self._confirm_removal(config_dir_path):
                shutil.rmtree(config_dir_path)
                self.logger.debug(removed_config_dir.format(config_dir_path=config_dir_path))
            else:
                self.logger.info(skipped_removal_config_dir.format(config_dir_path=config_dir_path))

        except Exception as e:
            raise Exception(f"{config_directory_removal_failed}: {str(e)}")

    def _confirm_removal(self, path: Path) -> bool:
        if self.force:
            return True

        response = typer.confirm(f"Remove configuration directory {path}? This action cannot be undone.")
        return response

    def _show_success_message(self):
        self.logger.success(uninstall_completed)
        self.logger.info(uninstall_completed_info)
        self.logger.info(uninstall_thank_you)
