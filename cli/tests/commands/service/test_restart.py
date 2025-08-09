import os
import subprocess
from unittest.mock import Mock, patch

import pytest
from pydantic import ValidationError

from app.commands.service.messages import (
    dry_run_command,
    dry_run_command_would_be_executed,
    dry_run_env_file,
    dry_run_mode,
    dry_run_service,
    end_dry_run,
    service_restart_failed,
    services_restarted_successfully,
    unknown_error,
)
from app.commands.service.restart import (
    DockerCommandBuilder,
    DockerService,
    Restart,
    RestartConfig,
    RestartFormatter,
    RestartResult,
    RestartService,
)
from app.utils.logger import Logger


class TestDockerCommandBuilder:
    def test_build_restart_command_default(self):
        cmd = DockerCommandBuilder.build_restart_command()
        assert cmd == ["docker", "compose", "restart"]

    def test_build_restart_command_with_service_name(self):
        cmd = DockerCommandBuilder.build_restart_command("web")
        assert cmd == ["docker", "compose", "restart", "web"]

    def test_build_restart_command_with_env_file(self):
        cmd = DockerCommandBuilder.build_restart_command("all", "/path/to/.env")
        assert cmd == ["docker", "compose", "restart", "--env-file", "/path/to/.env"]

    def test_build_restart_command_with_compose_file(self):
        cmd = DockerCommandBuilder.build_restart_command("all", None, "/path/to/docker-compose.yml")
        assert cmd == ["docker", "compose", "-f", "/path/to/docker-compose.yml", "restart"]

    def test_build_restart_command_with_all_parameters(self):
        cmd = DockerCommandBuilder.build_restart_command("api", "/path/to/.env", "/path/to/docker-compose.yml")
        assert cmd == [
            "docker",
            "compose",
            "-f",
            "/path/to/docker-compose.yml",
            "restart",
            "--env-file",
            "/path/to/.env",
            "api",
        ]


class TestRestartFormatter:
    def setup_method(self):
        self.formatter = RestartFormatter()

    def test_format_output_success(self):
        result = RestartResult(name="web", env_file=None, verbose=False, output="text", success=True)
        formatted = self.formatter.format_output(result, "text")
        assert formatted == ""

    def test_format_output_failure(self):
        result = RestartResult(
            name="web", env_file=None, verbose=False, output="text", success=False, error="Service not found"
        )
        formatted = self.formatter.format_output(result, "text")
        assert "Service not found" in formatted

    def test_format_output_json(self):
        result = RestartResult(name="web", env_file=None, verbose=False, output="json", success=True)
        formatted = self.formatter.format_output(result, "json")
        import json

        data = json.loads(formatted)
        assert data["success"] is True
        expected_message = services_restarted_successfully.format(services="web")
        assert expected_message in data["message"]

    def test_format_output_invalid(self):
        result = RestartResult(name="web", env_file=None, verbose=False, output="invalid", success=True)
        formatted = self.formatter.format_output(result, "invalid")
        assert formatted == ""

    def test_format_dry_run_default(self):
        config = RestartConfig(name="all", env_file=None, dry_run=True)
        formatted = self.formatter.format_dry_run(config)
        assert dry_run_mode in formatted
        assert dry_run_command in formatted
        assert dry_run_service.format(service="all") in formatted

    def test_format_dry_run_with_service(self):
        config = RestartConfig(name="web", env_file=None, dry_run=True)
        formatted = self.formatter.format_dry_run(config)
        assert dry_run_command in formatted
        assert dry_run_service.format(service="web") in formatted

    def test_format_dry_run_with_env_file(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = RestartConfig(name="all", env_file="/path/to/.env", dry_run=True)
            formatted = self.formatter.format_dry_run(config)
            assert dry_run_command in formatted
            assert dry_run_env_file.format(env_file="/path/to/.env") in formatted

    def test_format_dry_run_with_compose_file(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = RestartConfig(name="all", compose_file="/path/to/docker-compose.yml", dry_run=True)
            formatted = self.formatter.format_dry_run(config)
            assert dry_run_command in formatted
            assert "Command:" in formatted


class TestDockerService:
    def setup_method(self):
        self.logger = Mock(spec=Logger)
        self.docker_service = DockerService(self.logger)

    @patch("subprocess.run")
    def test_restart_services_success(self, mock_run):
        mock_result = Mock(returncode=0, stdout="", stderr="")
        mock_run.return_value = mock_result

        success, error = self.docker_service.restart_services("web")

        assert success is True
        assert error == ""

    @patch("subprocess.run")
    def test_restart_services_with_env_file(self, mock_run):
        mock_result = Mock(returncode=0, stdout="", stderr="")
        mock_run.return_value = mock_result

        success, error = self.docker_service.restart_services("all", "/path/to/.env")

        assert success is True
        assert error == ""
        mock_run.assert_called_once()
        cmd = mock_run.call_args[0][0]
        assert cmd == ["docker", "compose", "restart", "--env-file", "/path/to/.env"]

    @patch("subprocess.run")
    def test_restart_services_with_compose_file(self, mock_run):
        mock_result = Mock(returncode=0, stdout="", stderr="")
        mock_run.return_value = mock_result

        success, error = self.docker_service.restart_services("all", None, "/path/to/docker-compose.yml")

        assert success is True
        assert error == ""
        mock_run.assert_called_once()
        cmd = mock_run.call_args[0][0]
        assert cmd == ["docker", "compose", "-f", "/path/to/docker-compose.yml", "restart"]

    @patch("subprocess.run")
    def test_restart_services_failure(self, mock_run):
        mock_run.side_effect = subprocess.CalledProcessError(1, "docker compose restart", stderr="Service not found")

        success, error = self.docker_service.restart_services("web")

        assert success is False
        assert error == "Service not found"
        expected_error = "Service restart failed: Service not found"
        self.logger.error.assert_called_once_with(expected_error)

    @patch("subprocess.run")
    def test_restart_services_unexpected_error(self, mock_run):
        mock_run.side_effect = Exception("Unexpected error")

        success, error = self.docker_service.restart_services("web")

        assert success is False
        assert error == "Unexpected error"
        expected_error = "Unexpected error during restart: Unexpected error"
        self.logger.error.assert_called_once_with(expected_error)


class TestRestartConfig:
    def test_valid_config_default(self):
        config = RestartConfig()
        assert config.name == "all"
        assert config.env_file is None
        assert config.verbose is False
        assert config.output == "text"
        assert config.dry_run is False
        assert config.compose_file is None

    def test_valid_config_custom(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = RestartConfig(
                name="web",
                env_file="/path/to/.env",
                verbose=True,
                output="json",
                dry_run=True,
                compose_file="/path/to/docker-compose.yml",
            )
            assert config.name == "web"
            assert config.env_file == "/path/to/.env"
            assert config.verbose is True
            assert config.output == "json"
            assert config.dry_run is True
            assert config.compose_file == "/path/to/docker-compose.yml"

    @patch("os.path.exists")
    def test_validate_env_file_exists(self, mock_exists):
        mock_exists.return_value = True
        config = RestartConfig(env_file="/path/to/.env")
        assert config.env_file == "/path/to/.env"

    @patch("os.path.exists")
    def test_validate_env_file_not_exists(self, mock_exists):
        mock_exists.return_value = False
        with pytest.raises(ValidationError):
            RestartConfig(env_file="/path/to/.env")

    def test_validate_env_file_none(self):
        config = RestartConfig(env_file=None)
        assert config.env_file is None

    def test_validate_env_file_empty(self):
        config = RestartConfig(env_file="")
        assert config.env_file is None

    def test_validate_env_file_whitespace(self):
        config = RestartConfig(env_file="   ")
        assert config.env_file is None

    def test_validate_env_file_stripped(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = RestartConfig(env_file="  /path/to/.env  ")
            assert config.env_file == "/path/to/.env"

    @patch("os.path.exists")
    def test_validate_compose_file_exists(self, mock_exists):
        mock_exists.return_value = True
        config = RestartConfig(compose_file="/path/to/docker-compose.yml")
        assert config.compose_file == "/path/to/docker-compose.yml"

    @patch("os.path.exists")
    def test_validate_compose_file_not_exists(self, mock_exists):
        mock_exists.return_value = False
        with pytest.raises(ValidationError):
            RestartConfig(compose_file="/path/to/docker-compose.yml")

    def test_validate_compose_file_none(self):
        config = RestartConfig(compose_file=None)
        assert config.compose_file is None

    def test_validate_compose_file_empty(self):
        config = RestartConfig(compose_file="")
        assert config.compose_file is None

    def test_validate_compose_file_whitespace(self):
        config = RestartConfig(compose_file="   ")
        assert config.compose_file is None

    def test_validate_compose_file_stripped(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = RestartConfig(compose_file="  /path/to/docker-compose.yml  ")
            assert config.compose_file == "/path/to/docker-compose.yml"


class TestRestartService:
    def setup_method(self):
        self.config = RestartConfig(name="web", env_file=None, verbose=False, output="text", dry_run=False)
        self.logger = Mock(spec=Logger)
        self.docker_service = Mock()
        self.service = RestartService(self.config, self.logger, self.docker_service)

    def test_create_result_success(self):
        result = self.service._create_result(True)
        assert result.name == "web"
        assert result.success is True
        assert result.error is None
        assert result.output == "text"
        assert result.verbose is False

    def test_create_result_failure(self):
        result = self.service._create_result(False, "Service not found")
        assert result.success is False
        assert result.error == "Service not found"

    def test_restart_success(self):
        self.docker_service.restart_services.return_value = (True, "")

        result = self.service.restart()

        assert result.success is True
        assert result.error is None
        self.docker_service.restart_services.assert_called_once_with("web", None, None)

    def test_restart_failure(self):
        self.docker_service.restart_services.return_value = (False, "Service not found")

        result = self.service.restart()

        assert result.success is False
        assert result.error == "Service not found"

    def test_restart_and_format_dry_run(self):
        self.config.dry_run = True
        formatted = self.service.restart_and_format()
        assert dry_run_mode in formatted
        assert dry_run_command in formatted

    def test_restart_and_format_success(self):
        self.docker_service.restart_services.return_value = (True, "")
        formatted = self.service.restart_and_format()
        assert formatted == ""


class TestRestart:
    def setup_method(self):
        self.logger = Mock(spec=Logger)
        self.restart = Restart(self.logger)

    def test_restart_success(self):
        config = RestartConfig(name="web", env_file=None, verbose=False, output="text", dry_run=False)

        with patch(
            "app.commands.service.restart.RestartService.execute",
            return_value=RestartResult(
                name=config.name, env_file=config.env_file, verbose=config.verbose, output=config.output, success=True
            ),
        ):
            result = self.restart.restart(config)
            assert result.success is True

    def test_restart_failure(self):
        config = RestartConfig(name="web", env_file=None, verbose=False, output="text", dry_run=False)

        with patch(
            "app.commands.service.restart.RestartService.execute",
            return_value=RestartResult(
                name=config.name,
                env_file=config.env_file,
                verbose=config.verbose,
                output=config.output,
                success=False,
                error="Service not found",
            ),
        ):
            result = self.restart.restart(config)
            assert result.success is False
            assert result.error == "Service not found"

    def test_format_output(self):
        result = RestartResult(name="web", env_file=None, verbose=False, output="text", success=True)

        formatted = self.restart.format_output(result, "text")
        assert formatted == ""


class TestRestartResult:
    def test_restart_result_creation(self):
        result = RestartResult(name="web", env_file="/path/to/.env", verbose=True, output="json", success=True, error=None)

        assert result.name == "web"
        assert result.env_file == "/path/to/.env"
        assert result.verbose is True
        assert result.output == "json"
        assert result.success is True
        assert result.error is None

    def test_restart_result_default_success(self):
        result = RestartResult(name="web", env_file=None, verbose=False, output="text")

        assert result.name == "web"
        assert result.success is False
        assert result.error is None
