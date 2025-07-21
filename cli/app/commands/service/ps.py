import json
import subprocess

from app.utils.protocols import DockerServiceProtocol, LoggerProtocol

from .base import BaseAction, BaseConfig, BaseDockerCommandBuilder, BaseDockerService, BaseFormatter, BaseResult, BaseService
from .messages import (
    dry_run_command,
    dry_run_command_would_be_executed,
    dry_run_env_file,
    dry_run_mode,
    dry_run_service,
    end_dry_run,
    service_status_failed,
    services_status_retrieved,
    docker_command_executing,
    docker_command_completed,
    docker_command_failed,
    docker_command_stdout,
    docker_command_stderr,
    docker_unexpected_error,
    service_action_info,
    service_action_failed,
    service_action_unexpected_error,
)


class DockerCommandBuilder(BaseDockerCommandBuilder):
    @staticmethod
    def build_ps_command(name: str = "all", env_file: str = None, compose_file: str = None) -> list[str]:
        cmd = ["docker", "compose"]
        if compose_file:
            cmd.extend(["-f", compose_file])
        cmd.extend(["config", "--format", "json"])
        if env_file:
            cmd.extend(["--env-file", env_file])
        return cmd


class PsFormatter(BaseFormatter):
    def format_output(self, result: "PsResult", output: str) -> str:
        if result.success:
            if output == "json":
                message = services_status_retrieved.format(services=result.name)
                output_message = self.output_formatter.create_success_message(message, result.model_dump())
                return self.output_formatter.format_output(output_message, output)
            else:
                if result.docker_output and result.docker_output.strip():
                    try:
                        config_data = json.loads(result.docker_output)
                        services = config_data.get("services", {})
                        
                        if services:
                            table_data = []
                            for service_name, service_config in services.items():
                                ports = service_config.get("ports", [])
                                port_mappings = []
                                for port in ports:
                                    if isinstance(port, dict):
                                        published = port.get("published", "")
                                        target = port.get("target", "")
                                        port_mappings.append(f"{published}:{target}")
                                    else:
                                        port_mappings.append(str(port))
                                
                                networks = list(service_config.get("networks", {}).keys())
                                
                                table_data.append({
                                    "Service": service_name,
                                    "Image": service_config.get("image", ""),
                                    "Ports": ", ".join(port_mappings) if port_mappings else "",
                                    "Networks": ", ".join(networks) if networks else "default",
                                    "Command": str(service_config.get("command", "")) if service_config.get("command") else "",
                                    "Entrypoint": str(service_config.get("entrypoint", "")) if service_config.get("entrypoint") else "",
                                })
                            
                            if result.name != "all":
                                table_data = [row for row in table_data if row["Service"] == result.name]
                            
                            if table_data:
                                headers = ["Service", "Image", "Ports", "Networks", "Command", "Entrypoint"]
                                return self.output_formatter.create_table(
                                    data=table_data,
                                    title="Docker Compose Services Configuration",
                                    headers=headers,
                                    show_header=True,
                                    show_lines=True
                                ).strip()
                            else:
                                return f"No service found with name: {result.name}" if result.name != "all" else "No services found"
                        else:
                            return "No services found in compose file"
                    except json.JSONDecodeError as e:
                        return result.docker_output.strip()
                else:
                    return "No configuration found"
        else:
            return super().format_output(result, output, services_status_retrieved, service_status_failed)

    def format_dry_run(self, config: "PsConfig") -> str:
        dry_run_messages = {
            "mode": dry_run_mode,
            "command_would_be_executed": dry_run_command_would_be_executed,
            "command": dry_run_command,
            "service": dry_run_service,
            "env_file": dry_run_env_file,
            "end": end_dry_run,
        }
        return super().format_dry_run(config, DockerCommandBuilder(), dry_run_messages)


class DockerService(BaseDockerService):
    def __init__(self, logger: LoggerProtocol):
        super().__init__(logger, "config")

    def show_services_status(self, name: str = "all", env_file: str = None, compose_file: str = None) -> tuple[bool, str]:
        cmd = DockerCommandBuilder.build_ps_command(name, env_file, compose_file)
        
        self.logger.debug(docker_command_executing.format(command=' '.join(cmd)))
        
        try:
            self.logger.info(service_action_info.format(action="ps", name=name))
            result = subprocess.run(cmd, capture_output=True, text=True, check=True)
            
            self.logger.debug(docker_command_completed.format(action="ps"))
            
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
            
            self.logger.error(service_action_failed.format(action="ps", error=e.stderr or str(e)))
            return False, e.stderr or e.stdout or str(e)
        except Exception as e:
            self.logger.debug(docker_unexpected_error.format(action="ps", error=str(e)))
            self.logger.error(service_action_unexpected_error.format(action="ps", error=e))
            return False, str(e)


class PsResult(BaseResult):
    pass


class PsConfig(BaseConfig):
    pass


class PsService(BaseService[PsConfig, PsResult]):
    def __init__(self, config: PsConfig, logger: LoggerProtocol = None, docker_service: DockerServiceProtocol = None):
        super().__init__(config, logger, docker_service)
        self.docker_service = docker_service or DockerService(self.logger)
        self.formatter = PsFormatter()

    def _create_result(self, success: bool, error: str = None, docker_output: str = None) -> PsResult:
        return PsResult(
            name=self.config.name,
            env_file=self.config.env_file,
            verbose=self.config.verbose,
            output=self.config.output,
            success=success,
            error=error,
            docker_output=docker_output,
        )

    def ps(self) -> PsResult:
        return self.execute()

    def execute(self) -> PsResult:
        self.logger.debug(f"Checking status of services: {self.config.name}")

        success, docker_output = self.docker_service.show_services_status(
            self.config.name, self.config.env_file, self.config.compose_file
        )
        
        error = None if success else docker_output
        return self._create_result(success, error, docker_output)

    def ps_and_format(self) -> str:
        return self.execute_and_format()

    def execute_and_format(self) -> str:
        if self.config.dry_run:
            return self.formatter.format_dry_run(self.config)

        result = self.execute()
        return self.formatter.format_output(result, self.config.output)


class Ps(BaseAction[PsConfig, PsResult]):
    def __init__(self, logger: LoggerProtocol = None):
        super().__init__(logger)
        self.formatter = PsFormatter()

    def ps(self, config: PsConfig) -> PsResult:
        return self.execute(config)

    def execute(self, config: PsConfig) -> PsResult:
        service = PsService(config, logger=self.logger)
        return service.execute()

    def format_output(self, result: PsResult, output: str) -> str:
        return self.formatter.format_output(result, output)
    
    def format_dry_run(self, config: PsConfig) -> str:
        return self.formatter.format_dry_run(config)
