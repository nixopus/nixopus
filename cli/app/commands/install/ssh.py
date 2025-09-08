import os
import stat
import subprocess
from typing import Optional, Protocol

from pydantic import BaseModel, Field, field_validator

from app.utils.lib import FileManager
from app.utils.logger import Logger
from app.utils.output_formatter import OutputFormatter
from app.utils.protocols import LoggerProtocol

from .messages import (
    adding_to_authorized_keys,
    authorized_keys_updated,
    debug_ssh_authorized_keys_append,
    debug_ssh_authorized_keys_append_failed,
    debug_ssh_authorized_keys_created,
    debug_ssh_authorized_keys_creation_failed,
    debug_ssh_authorized_keys_enabled,
    debug_ssh_authorized_keys_exception,
    debug_ssh_authorized_keys_failed_abort,
    debug_ssh_authorized_keys_missing,
    debug_ssh_authorized_keys_path,
    debug_ssh_authorized_keys_read,
    debug_ssh_config_validation,
    debug_ssh_directory_check,
    debug_ssh_directory_created,
    debug_ssh_directory_creation,
    debug_ssh_directory_creation_enabled,
    debug_ssh_directory_creation_failed,
    debug_ssh_directory_creation_failed_abort,
    debug_ssh_directory_exception,
    debug_ssh_directory_missing,
    debug_ssh_dry_run_enabled,
    debug_ssh_force_disabled,
    debug_ssh_force_enabled,
    debug_ssh_generation_failed_abort,
    debug_ssh_generation_process_start,
    debug_ssh_key_directory_info,
    debug_ssh_key_exists,
    debug_ssh_key_generation_start,
    debug_ssh_key_generation_success,
    debug_ssh_key_not_exists,
    debug_ssh_keygen_availability,
    debug_ssh_keygen_availability_failed,
    debug_ssh_keygen_availability_result,
    debug_ssh_keygen_command_build,
    debug_ssh_keygen_version_info,
    debug_ssh_path_expansion,
    debug_ssh_permission_setting,
    debug_ssh_permissions_enabled,
    debug_ssh_permissions_exception,
    debug_ssh_permissions_failed_abort,
    debug_ssh_permissions_success,
    debug_ssh_prerequisites_completed,
    debug_ssh_prerequisites_failed_abort,
    debug_ssh_private_key_permissions,
    debug_ssh_private_key_permissions_failed,
    debug_ssh_process_completed,
    debug_ssh_process_stderr,
    debug_ssh_process_stdout,
    debug_ssh_public_key_path_info,
    debug_ssh_public_key_permissions,
    debug_ssh_public_key_permissions_failed,
    debug_ssh_public_key_read_failed,
    dry_run_command,
    dry_run_command_would_be_executed,
    dry_run_force_mode,
    dry_run_mode,
    dry_run_passphrase,
    dry_run_ssh_key,
    end_dry_run,
    executing_ssh_keygen,
    failed_to_add_to_authorized_keys,
    failed_to_append_to_authorized_keys,
    failed_to_read_public_key,
    generating_ssh_key,
    invalid_key_size,
    invalid_key_type,
    invalid_ssh_key_path,
    prerequisites_validation_failed,
    ssh_key_already_exists,
    ssh_keygen_failed,
    successfully_generated_ssh_key,
    unexpected_error_during_ssh_keygen,
    unknown_error,
)


class SSHKeyProtocol(Protocol):
    def generate_ssh_key(
        self, path: str, key_type: str = "rsa", key_size: int = 4096, passphrase: str = None
    ) -> tuple[bool, str]: ...


class SSHCommandBuilder:
    @staticmethod
    def build_ssh_keygen_command(path: str, key_type: str = "rsa", key_size: int = 4096, passphrase: str = None) -> list[str]:
        cmd = ["ssh-keygen", "-t", key_type, "-f", path, "-N"]
        if passphrase is not None:
            cmd.append(passphrase)
        else:
            cmd.append("")
        if key_type in ["rsa", "dsa", "ecdsa"]:
            cmd.extend(["-b", str(key_size)])
        return cmd


class SSHFormatter:
    def __init__(self):
        self.output_formatter = OutputFormatter()

    def format_output(self, result: "SSHResult", output: str) -> str:
        if result.success:
            message = successfully_generated_ssh_key.format(key=result.path)
            output_message = self.output_formatter.create_success_message(message, result.model_dump())
        else:
            error = result.error or unknown_error
            output_message = self.output_formatter.create_error_message(error, result.model_dump())

        return self.output_formatter.format_output(output_message, output)

    def format_dry_run(self, config: "SSHConfig") -> str:
        cmd = SSHCommandBuilder.build_ssh_keygen_command(config.path, config.key_type, config.key_size, config.passphrase)

        output = []
        output.append(dry_run_mode)
        output.append(dry_run_command_would_be_executed)
        output.append(dry_run_command.format(command=" ".join(cmd)))
        output.append(dry_run_ssh_key.format(key=config.path))
        output.append(f"Key type: {config.key_type}")
        output.append(f"Key size: {config.key_size}")
        if config.passphrase:
            output.append(dry_run_passphrase.format(passphrase="***"))
        output.append(dry_run_force_mode.format(force=config.force))
        output.append(end_dry_run)
        return "\n".join(output)


class SSHKeyManager:
    def __init__(self, logger: LoggerProtocol):
        self.file_manager = FileManager()
        self.logger = logger

    def _check_ssh_keygen_availability(self) -> tuple[bool, str]:
        self.logger.debug(debug_ssh_keygen_availability)
        try:
            result = subprocess.run(["ssh-keygen", "-h"], capture_output=True, text=True, check=False)
            availability = result.returncode == 0
            self.logger.debug(debug_ssh_keygen_availability_result.format(availability=availability))
            return availability, None
        except Exception as e:
            self.logger.debug(debug_ssh_keygen_availability_failed.format(error=e))
            return False, f"ssh-keygen not found: {e}"

    def _check_ssh_keygen_version(self) -> tuple[bool, str]:
        try:
            result = subprocess.run(["ssh-keygen", "-V"], capture_output=True, text=True, check=False)
            if result.returncode == 0:
                self.logger.debug(debug_ssh_keygen_version_info.format(version=result.stdout.strip()))
            return True, None
        except Exception:
            return True, None

    def generate_ssh_key(
        self, path: str, key_type: str = "rsa", key_size: int = 4096, passphrase: str = None, force: bool = False
    ) -> tuple[bool, str]:
        self.logger.debug(debug_ssh_key_generation_start.format(path=path))

        if force:
            if os.path.exists(path):
                os.remove(path)
            pub_path = path + ".pub"
            if os.path.exists(pub_path):
                os.remove(pub_path)

        cmd = SSHCommandBuilder.build_ssh_keygen_command(path, key_type, key_size, passphrase)
        self.logger.debug(debug_ssh_keygen_command_build.format(command=" ".join(cmd)))

        try:
            self.logger.debug(executing_ssh_keygen.format(command=" ".join(cmd)))
            result = subprocess.run(cmd, capture_output=True, text=True, check=True, timeout=30)
            self.logger.debug(debug_ssh_key_generation_success.format(path=path))
            return True, None
        except subprocess.TimeoutExpired:
            self.logger.error("ssh-keygen timed out")
            return False, "ssh-keygen timed out"
        except subprocess.CalledProcessError as e:
            self.logger.error(f"ssh-keygen failed. Command: {' '.join(cmd)}")
            self.logger.debug(debug_ssh_process_stdout.format(stdout=e.stdout))
            self.logger.debug(debug_ssh_process_stderr.format(stderr=e.stderr))
            self.logger.error(ssh_keygen_failed.format(error=e.stderr.strip() if e.stderr else str(e)))
            return False, e.stderr.strip() if e.stderr else str(e)
        except Exception as e:
            self.logger.error(f"Unexpected error running ssh-keygen. Command: {' '.join(cmd)}")
            self.logger.error(unexpected_error_during_ssh_keygen.format(error=e))
            return False, str(e)

    def set_key_permissions(self, private_key_path: str, public_key_path: str) -> tuple[bool, str]:
        self.logger.debug(debug_ssh_permission_setting.format(private_key=private_key_path, public_key=public_key_path))
        try:
            self.logger.debug(debug_ssh_private_key_permissions.format(path=private_key_path))
            private_success, private_error = self.file_manager.set_permissions(
                private_key_path, stat.S_IRUSR | stat.S_IWUSR, self.logger
            )
            if not private_success:
                self.logger.debug(debug_ssh_private_key_permissions_failed.format(error=private_error))
                return False, private_error

            self.logger.debug(debug_ssh_public_key_permissions.format(path=public_key_path))
            public_success, public_error = self.file_manager.set_permissions(
                public_key_path, stat.S_IRUSR | stat.S_IWUSR | stat.S_IRGRP | stat.S_IROTH, self.logger
            )
            if not public_success:
                self.logger.debug(debug_ssh_public_key_permissions_failed.format(error=public_error))
                return False, public_error

            self.logger.debug(debug_ssh_permissions_success)
            return True, None
        except Exception as e:
            self.logger.debug(debug_ssh_permissions_exception.format(error=e))
            return False, f"Failed to set permissions: {e}"

    def create_ssh_directory(self, ssh_dir: str) -> tuple[bool, str]:
        permissions = stat.S_IRUSR | stat.S_IWUSR | stat.S_IXUSR
        self.logger.debug(debug_ssh_directory_creation.format(directory=ssh_dir, permissions=oct(permissions)))
        try:
            self.logger.debug(debug_ssh_directory_check.format(directory=ssh_dir))
            success, error = self.file_manager.create_directory(ssh_dir, permissions, self.logger)
            if success:
                self.logger.debug(debug_ssh_directory_created.format(directory=ssh_dir))
            else:
                self.logger.debug(debug_ssh_directory_creation_failed.format(error=error))
            return success, error
        except Exception as e:
            self.logger.debug(debug_ssh_directory_exception.format(error=e))
            return False, f"Failed to create SSH directory: {e}"

    def add_to_authorized_keys(self, public_key_path: str) -> tuple[bool, str]:
        try:
            self.logger.debug(adding_to_authorized_keys)
            self.logger.debug(debug_ssh_authorized_keys_read.format(path=public_key_path))

            success, content, error = self.file_manager.read_file_content(public_key_path, self.logger)
            if not success:
                self.logger.debug(debug_ssh_public_key_read_failed.format(error=error))
                return False, error or failed_to_read_public_key

            ssh_dir = self.file_manager.expand_user_path("~/.ssh")
            authorized_keys_path = os.path.join(ssh_dir, "authorized_keys")
            self.logger.debug(debug_ssh_authorized_keys_path.format(path=authorized_keys_path))

            if not os.path.exists(ssh_dir):
                self.logger.debug(debug_ssh_directory_missing.format(directory=ssh_dir))
                success, error = self.create_ssh_directory(ssh_dir)
                if not success:
                    return False, error

            if not os.path.exists(authorized_keys_path):
                self.logger.debug(debug_ssh_authorized_keys_missing.format(path=authorized_keys_path))
                try:
                    with open(authorized_keys_path, "w") as f:
                        pass
                    os.chmod(authorized_keys_path, stat.S_IRUSR | stat.S_IWUSR)
                    self.logger.debug(debug_ssh_authorized_keys_created.format(path=authorized_keys_path))
                except Exception as e:
                    self.logger.debug(debug_ssh_authorized_keys_creation_failed.format(error=e))
                    return False, f"Failed to create authorized_keys file: {e}"

            self.logger.debug(debug_ssh_authorized_keys_append.format(path=authorized_keys_path))
            success, error = self.file_manager.append_to_file(authorized_keys_path, content, self.logger)
            if not success:
                self.logger.debug(debug_ssh_authorized_keys_append_failed.format(error=error))
                return False, error or failed_to_append_to_authorized_keys

            self.logger.debug(authorized_keys_updated)
            return True, None
        except Exception as e:
            error_msg = failed_to_add_to_authorized_keys.format(error=e)
            self.logger.debug(debug_ssh_authorized_keys_exception.format(error=e))
            self.logger.error(error_msg)
            return False, error_msg


class SSHResult(BaseModel):
    path: str
    key_type: str
    key_size: int
    passphrase: Optional[str]
    force: bool
    verbose: bool
    output: str
    success: bool = False
    error: Optional[str] = None
    set_permissions: bool = True
    add_to_authorized_keys: bool = False
    create_ssh_directory: bool = True


class SSHConfig(BaseModel):
    path: str = Field(..., min_length=1, description="SSH key path to generate")
    key_type: str = Field("rsa", description="SSH key type (rsa, ed25519, ecdsa)")
    key_size: int = Field(4096, description="SSH key size")
    passphrase: Optional[str] = Field(None, description="Passphrase for the SSH key")
    force: bool = Field(False, description="Force overwrite existing SSH key")
    verbose: bool = Field(False, description="Verbose output")
    output: str = Field("text", description="Output format: text, json")
    dry_run: bool = Field(False, description="Dry run mode")
    set_permissions: bool = Field(True, description="Set proper file permissions")
    add_to_authorized_keys: bool = Field(False, description="Add public key to authorized_keys")
    create_ssh_directory: bool = Field(True, description="Create .ssh directory if it doesn't exist")

    @field_validator("path")
    @classmethod
    def validate_path(cls, path: str) -> str:
        stripped_path = path.strip()
        if not stripped_path:
            raise ValueError(invalid_ssh_key_path)

        if not cls._is_valid_key_path(stripped_path):
            raise ValueError(invalid_ssh_key_path)
        return stripped_path

    @staticmethod
    def _is_valid_key_path(key_path: str) -> bool:
        return (
            key_path.startswith(("~", "/", "./"))
            or os.path.isabs(key_path)
            or key_path.endswith((".pem", ".key", "_rsa", "_ed25519"))
        )

    @field_validator("key_type")
    @classmethod
    def validate_key_type(cls, key_type: str) -> str:
        valid_types = ["rsa", "ed25519", "ecdsa", "dsa"]
        if key_type.lower() not in valid_types:
            raise ValueError(invalid_key_type)
        return key_type.lower()

    @field_validator("key_size")
    @classmethod
    def validate_key_size(cls, key_size: int, info) -> int:
        key_type = info.data.get("key_type", "rsa")

        if key_type == "ed25519":
            return 256
        elif key_type == "ecdsa":
            if key_size not in [256, 384, 521]:
                raise ValueError(invalid_key_size)
        elif key_type == "dsa":
            if key_size != 1024:
                raise ValueError(invalid_key_size)
        else:
            if key_size < 1024 or key_size > 16384:
                raise ValueError(invalid_key_size)

        return key_size

    @field_validator("passphrase")
    @classmethod
    def validate_passphrase(cls, passphrase: str) -> Optional[str]:
        if not passphrase:
            return None
        stripped_passphrase = passphrase.strip()
        if not stripped_passphrase:
            return None
        return stripped_passphrase


class SSHService:
    def __init__(self, config: SSHConfig, logger: LoggerProtocol = None, ssh_manager: SSHKeyProtocol = None):
        self.logger = logger or Logger(verbose=config.verbose)
        self.config = config
        self.ssh_manager = ssh_manager or SSHKeyManager(self.logger)
        self.formatter = SSHFormatter()
        self.file_manager = FileManager()

    def _validate_prerequisites(self) -> bool:
        self.logger.debug(
            debug_ssh_config_validation.format(
                path=self.config.path, key_type=self.config.key_type, key_size=self.config.key_size
            )
        )

        expanded_key_path = self.file_manager.expand_user_path(self.config.path)
        self.logger.debug(debug_ssh_path_expansion.format(original=self.config.path, expanded=expanded_key_path))

        if os.path.exists(expanded_key_path):
            self.logger.debug(debug_ssh_key_exists.format(path=expanded_key_path))
            if not self.config.force:
                self.logger.debug(debug_ssh_force_disabled)
                self.logger.error(ssh_key_already_exists.format(path=self.config.path))
                return False
            else:
                self.logger.debug(debug_ssh_force_enabled)
        else:
            self.logger.debug(debug_ssh_key_not_exists.format(path=expanded_key_path))

        self.logger.debug(debug_ssh_prerequisites_completed)
        return True

    def _create_result(self, success: bool, error: str = None) -> SSHResult:
        return SSHResult(
            path=self.config.path,
            key_type=self.config.key_type,
            key_size=self.config.key_size,
            passphrase=self.config.passphrase,
            force=self.config.force,
            verbose=self.config.verbose,
            output=self.config.output,
            success=success,
            error=error,
            set_permissions=self.config.set_permissions,
            add_to_authorized_keys=self.config.add_to_authorized_keys,
            create_ssh_directory=self.config.create_ssh_directory,
        )

    def generate_ssh_key(self) -> SSHResult:
        self.logger.debug(generating_ssh_key.format(key=self.config.path))

        if not self._validate_prerequisites():
            self.logger.debug(debug_ssh_prerequisites_failed_abort)
            return self._create_result(False, prerequisites_validation_failed)

        if self.config.dry_run:
            self.logger.debug(debug_ssh_dry_run_enabled)
            dry_run_output = self.formatter.format_dry_run(self.config)
            return self._create_result(True, dry_run_output)

        expanded_path = self.file_manager.expand_user_path(self.config.path)
        ssh_dir = self.file_manager.get_directory_path(expanded_path)
        self.logger.debug(debug_ssh_key_directory_info.format(directory=ssh_dir))

        if self.config.create_ssh_directory:
            self.logger.debug(debug_ssh_directory_creation_enabled.format(directory=ssh_dir))
            success, error = self.ssh_manager.create_ssh_directory(ssh_dir)
            if not success:
                self.logger.debug(debug_ssh_directory_creation_failed_abort.format(error=error))
                return self._create_result(False, error)

        self.logger.debug(debug_ssh_generation_process_start)
        success, error = self.ssh_manager.generate_ssh_key(
            self.config.path, self.config.key_type, self.config.key_size, self.config.passphrase, self.config.force
        )

        if not success:
            return self._create_result(False, error)

        if self.config.set_permissions:
            self.logger.debug(debug_ssh_permissions_enabled)
            public_key_path = self.file_manager.get_public_key_path(expanded_path)
            self.logger.debug(debug_ssh_public_key_path_info.format(path=public_key_path))
            success, error = self.ssh_manager.set_key_permissions(expanded_path, public_key_path)
            if not success:
                self.logger.debug(debug_ssh_permissions_failed_abort.format(error=error))
                return self._create_result(False, error)

        if self.config.add_to_authorized_keys:
            self.logger.debug(debug_ssh_authorized_keys_enabled)
            public_key_path = self.file_manager.get_public_key_path(expanded_path)
            success, error = self.ssh_manager.add_to_authorized_keys(public_key_path)
            if not success:
                self.logger.debug(debug_ssh_authorized_keys_failed_abort.format(error=error))
                return self._create_result(False, error)

        self.logger.debug(debug_ssh_process_completed)
        return self._create_result(True)

    def generate_and_format(self) -> str:
        result = self.generate_ssh_key()
        return self.formatter.format_output(result, self.config.output)


class SSH:
    def __init__(self, logger: LoggerProtocol = None):
        self.logger = logger or Logger()

    def generate(self, config: SSHConfig) -> SSHResult:
        service = SSHService(config, self.logger)
        return service.generate_ssh_key()

    def format_output(self, result: SSHResult, output: str) -> str:
        formatter = SSHFormatter()
        return formatter.format_output(result, output)
