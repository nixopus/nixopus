import os
import subprocess
from typing import Generic, Optional, Protocol, TypeVar

from pydantic import BaseModel, Field, field_validator

from app.utils.logger import Logger
from app.utils.output_formatter import OutputFormatter
from app.utils.protocols import LoggerProtocol
from .messages import (
    service_action_info, 
    service_action_success, 
    service_action_failed, 
    service_action_unexpected_error, 
    environment_file_not_found, 
    compose_file_not_found,
    docker_command_executing,
    docker_command_completed,
    docker_command_failed,
    docker_command_stdout,
    docker_command_stderr,
    docker_unexpected_error,
)

TConfig = TypeVar("TConfig", bound=BaseModel)
TResult = TypeVar("TResult", bound=BaseModel)


class DockerServiceProtocol(Protocol):
    def execute_services(
        self, name: str = "all", env_file: str = None, compose_file: str = None, **kwargs
    ) -> tuple[bool, str]: ...


class BaseDockerCommandBuilder:
    @staticmethod
    def build_command(action: str, name: str = "all", env_file: str = None, compose_file: str = None, **kwargs) -> list[str]:
        cmd = ["docker", "compose"]
        if compose_file:
            cmd.extend(["-f", compose_file])
        cmd.append(action)

        if action == "up" and kwargs.get("detach", False):
            cmd.append("-d")

        if env_file:
            cmd.extend(["--env-file", env_file])

        if name != "all":
            cmd.append(name)

        return cmd


class BaseFormatter:
    def __init__(self):
        self.output_formatter = OutputFormatter()

    def format_output(self, result: TResult, output: str, success_message: str, error_message: str) -> str:
        if result.success:
            if output == "json":
                # For JSON, return formatted data structure
                message = success_message.format(services=result.name)
                output_message = self.output_formatter.create_success_message(message, result.model_dump())
                return self.output_formatter.format_output(output_message, output)
            else:
                # For text format, return only docker output or empty (command.py handles success message)
                if result.verbose and result.docker_output and result.docker_output.strip():
                    return f"Docker Command Output:\n{result.docker_output.strip()}"
                return ""
        else:
            # Always format errors the same way
            error = result.error or "Unknown error occurred"
            output_message = self.output_formatter.create_error_message(error, result.model_dump())
            return self.output_formatter.format_output(output_message, output)

    def format_dry_run(self, config: TConfig, command_builder, dry_run_messages: dict) -> str:
        if hasattr(command_builder, "build_up_command"):
            cmd = command_builder.build_up_command(
                getattr(config, "name", "all"),
                getattr(config, "detach", True),
                getattr(config, "env_file", None),
                getattr(config, "compose_file", None),
            )
        elif hasattr(command_builder, "build_down_command"):
            cmd = command_builder.build_down_command(
                getattr(config, "name", "all"), getattr(config, "env_file", None), getattr(config, "compose_file", None)
            )
        elif hasattr(command_builder, "build_ps_command"):
            cmd = command_builder.build_ps_command(
                getattr(config, "name", "all"), getattr(config, "env_file", None), getattr(config, "compose_file", None)
            )
        elif hasattr(command_builder, "build_restart_command"):
            cmd = command_builder.build_restart_command(
                getattr(config, "name", "all"), getattr(config, "env_file", None), getattr(config, "compose_file", None)
            )
        else:
            cmd = command_builder.build_command(config)

        output = []
        output.append(dry_run_messages["mode"])
        output.append(dry_run_messages["command_would_be_executed"])
        output.append(f"{dry_run_messages['command']} {' '.join(cmd)}")
        output.append(f"{dry_run_messages['service']} {getattr(config, 'name', 'all')}")

        if hasattr(config, "detach"):
            output.append(f"{dry_run_messages.get('detach_mode', 'Detach mode:')} {getattr(config, 'detach', True)}")

        if getattr(config, "env_file", None):
            output.append(f"{dry_run_messages['env_file']} {getattr(config, 'env_file')}")

        output.append(dry_run_messages["end"])
        return "\n".join(output)


class BaseDockerService:
    def __init__(self, logger: LoggerProtocol, action: str):
        self.logger = logger
        self.action = action

    def execute_services(
        self, name: str = "all", env_file: str = None, compose_file: str = None, **kwargs
    ) -> tuple[bool, str]:
        cmd = BaseDockerCommandBuilder.build_command(self.action, name, env_file, compose_file, **kwargs)
        
        self.logger.debug(docker_command_executing.format(command=' '.join(cmd)))
        
        try:
            self.logger.info(service_action_info.format(action=self.action, name=name))
            
            if self.action == "up" and not kwargs.get("detach", False):
                process = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, bufsize=1, universal_newlines=True)
                
                output_lines = []
                self.logger.info("Docker container logs:")
                self.logger.info("-" * 50)
                
                for line in process.stdout:
                    self.logger.info(line.rstrip())  # Stream logs through logger
                    output_lines.append(line.rstrip())
                
                return_code = process.wait()
                
                full_output = '\n'.join(output_lines)
                
                if return_code == 0:
                    self.logger.debug(docker_command_completed.format(action=self.action))
                    if full_output.strip():
                        self.logger.debug(docker_command_stdout.format(output=full_output.strip()))
                    return True, full_output
                else:
                    self.logger.debug(docker_command_failed.format(return_code=return_code))
                    if full_output.strip():
                        self.logger.debug(docker_command_stderr.format(output=full_output.strip()))
                    self.logger.error(service_action_failed.format(action=self.action, error=full_output or f"Process exited with code {return_code}"))
                    return False, full_output or f"Process exited with code {return_code}"
            else:
                result = subprocess.run(cmd, capture_output=True, text=True, check=True)
                
                self.logger.debug(docker_command_completed.format(action=self.action))
                
                if result.stdout.strip():
                    self.logger.debug(docker_command_stdout.format(output=result.stdout.strip()))
                
                if result.stderr.strip():
                    self.logger.debug(docker_command_stderr.format(output=result.stderr.strip()))
                
                return True, result.stdout or result.stderr
                
        except subprocess.CalledProcessError as e:
            self.logger.debug(docker_command_failed.format(return_code=e.returncode))
            
            if e.stdout and e.stdout.strip():
                self.logger.debug(docker_command_stdout.format(output=e.stdout.strip()))
            
            if e.stderr and e.stderr.strip():
                self.logger.debug(docker_command_stderr.format(output=e.stderr.strip()))
            
            self.logger.error(service_action_failed.format(action=self.action, error=e.stderr or str(e)))
            return False, e.stderr or e.stdout or str(e)
        except Exception as e:
            self.logger.debug(docker_unexpected_error.format(action=self.action, error=str(e)))
            self.logger.error(service_action_unexpected_error.format(action=self.action, error=e))
            return False, str(e)


class BaseConfig(BaseModel):
    name: str = Field("all", description="Name of the service")
    env_file: Optional[str] = Field(None, description="Path to environment file")
    verbose: bool = Field(False, description="Verbose output")
    output: str = Field("text", description="Output format: text, json")
    dry_run: bool = Field(False, description="Dry run mode")
    compose_file: Optional[str] = Field(None, description="Path to the compose file")

    @field_validator("env_file")
    @classmethod
    def validate_env_file(cls, env_file: str) -> Optional[str]:
        if not env_file:
            return None
        stripped_env_file = env_file.strip()
        if not stripped_env_file:
            return None
        if not os.path.exists(stripped_env_file):
            raise ValueError(environment_file_not_found.format(path=stripped_env_file))
        return stripped_env_file

    @field_validator("compose_file")
    @classmethod
    def validate_compose_file(cls, compose_file: str) -> Optional[str]:
        if not compose_file:
            return None
        stripped_compose_file = compose_file.strip()
        if not stripped_compose_file:
            return None
        if not os.path.exists(stripped_compose_file):
            raise ValueError(compose_file_not_found.format(path=stripped_compose_file))
        return stripped_compose_file


class BaseResult(BaseModel):
    name: str
    env_file: Optional[str]
    verbose: bool
    output: str
    success: bool = False
    error: Optional[str] = None
    docker_output: Optional[str] = None


class BaseService(Generic[TConfig, TResult]):
    def __init__(self, config: TConfig, logger: LoggerProtocol = None, docker_service: DockerServiceProtocol = None):
        self.config = config
        self.logger = logger or Logger(verbose=config.verbose)
        self.docker_service = docker_service
        self.formatter = None

    def _create_result(self, success: bool, error: str = None) -> TResult:
        raise NotImplementedError

    def execute(self) -> TResult:
        raise NotImplementedError

    def execute_and_format(self) -> str:
        raise NotImplementedError


class BaseAction(Generic[TConfig, TResult]):
    def __init__(self, logger: LoggerProtocol = None):
        self.logger = logger
        self.formatter = None

    def execute(self, config: TConfig) -> TResult:
        raise NotImplementedError

    def format_output(self, result: TResult, output: str) -> str:
        raise NotImplementedError
