import os
import subprocess
from unittest.mock import MagicMock, Mock, patch

import pytest
from pydantic import ValidationError

from app.commands.service.messages import (
    dry_run_command,
    dry_run_detach_mode,
    dry_run_env_file,
    dry_run_mode,
    dry_run_service,
    services_started_successfully,
)
from app.commands.service.up import DockerCommandBuilder, DockerService, Up, UpConfig, UpFormatter, UpResult, UpService
from app.utils.logger import Logger


class TestDockerCommandBuilder:
    def test_build_up_command_default(self):
        cmd = DockerCommandBuilder.build_up_command()
        assert cmd == ["docker", "compose", "up", "-d"]

    def test_build_up_command_with_service_name(self):
        cmd = DockerCommandBuilder.build_up_command("web")
        assert cmd == ["docker", "compose", "up", "-d", "web"]

    def test_build_up_command_without_detach(self):
        cmd = DockerCommandBuilder.build_up_command("all", detach=False)
        assert cmd == ["docker", "compose", "up"]

    def test_build_up_command_with_env_file(self):
        cmd = DockerCommandBuilder.build_up_command("all", True, "/path/to/.env")
        assert cmd == ["docker", "compose", "up", "-d", "--env-file", "/path/to/.env"]

    def test_build_up_command_with_compose_file(self):
        cmd = DockerCommandBuilder.build_up_command("all", True, None, "/path/to/docker-compose.yml")
        assert cmd == ["docker", "compose", "-f", "/path/to/docker-compose.yml", "up", "-d"]

    def test_build_up_command_with_all_parameters(self):
        cmd = DockerCommandBuilder.build_up_command("api", False, "/path/to/.env", "/path/to/docker-compose.yml")
        assert cmd == ["docker", "compose", "-f", "/path/to/docker-compose.yml", "up", "--env-file", "/path/to/.env", "api"]


class TestUpFormatter:
    def setup_method(self):
        self.formatter = UpFormatter()

    def test_format_output_success(self):
        result = UpResult(name="web", detach=True, env_file=None, verbose=False, output="text", success=True)
        formatted = self.formatter.format_output(result, "text")
        expected_message = services_started_successfully.format(services="web")
        assert expected_message in formatted

    def test_format_output_failure(self):
        result = UpResult(
            name="web", detach=True, env_file=None, verbose=False, output="text", success=False, error="Service not found"
        )
        formatted = self.formatter.format_output(result, "text")
        assert "Service not found" in formatted

    def test_format_output_json(self):
        result = UpResult(name="web", detach=True, env_file=None, verbose=False, output="json", success=True)
        formatted = self.formatter.format_output(result, "json")
        import json

        data = json.loads(formatted)
        assert data["success"] is True
        expected_message = services_started_successfully.format(services="web")
        assert expected_message in data["message"]

    def test_format_output_invalid(self):
        result = UpResult(name="web", detach=True, env_file=None, verbose=False, output="invalid", success=True)
        with pytest.raises(ValueError):
            self.formatter.format_output(result, "invalid")

    def test_format_dry_run_default(self):
        config = UpConfig(name="all", detach=True, env_file=None, dry_run=True)
        formatted = self.formatter.format_dry_run(config)
        assert dry_run_mode in formatted
        assert dry_run_command in formatted
        assert dry_run_service.format(service="all") in formatted
        assert dry_run_detach_mode.format(detach=True) in formatted

    def test_format_dry_run_with_service(self):
        config = UpConfig(name="web", detach=False, env_file=None, dry_run=True)
        formatted = self.formatter.format_dry_run(config)
        assert dry_run_command in formatted
        assert dry_run_service.format(service="web") in formatted
        assert dry_run_detach_mode.format(detach=False) in formatted

    def test_format_dry_run_with_env_file(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = UpConfig(name="all", detach=True, env_file="/path/to/.env", dry_run=True)
            formatted = self.formatter.format_dry_run(config)
            assert dry_run_command in formatted
            assert dry_run_env_file.format(env_file="/path/to/.env") in formatted

    def test_format_dry_run_with_compose_file(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = UpConfig(name="all", detach=True, compose_file="/path/to/docker-compose.yml", dry_run=True)
            formatted = self.formatter.format_dry_run(config)
            assert dry_run_command in formatted
            assert "Command:" in formatted


class TestDockerService:
    def setup_method(self):
        self.logger = Mock(spec=Logger)
        self.docker_service = DockerService(self.logger)

    @patch("subprocess.run")
    def test_start_services_success(self, mock_run):
        mock_run.return_value = Mock(returncode=0)

        success, error = self.docker_service.start_services("web")

        assert success is True
        assert error is None
        self.logger.info.assert_called_once_with("up services: web")
        self.logger.success.assert_called_once_with("Service up successful: web")

    @patch("subprocess.run")
    def test_start_services_with_env_file(self, mock_run):
        mock_run.return_value = Mock(returncode=0)

        success, error = self.docker_service.start_services("all", True, "/path/to/.env")

        assert success is True
        assert error is None
        mock_run.assert_called_once()
        cmd = mock_run.call_args[0][0]
        assert cmd == ["docker", "compose", "up", "-d", "--env-file", "/path/to/.env"]

    @patch("subprocess.run")
    def test_start_services_with_compose_file(self, mock_run):
        mock_run.return_value = Mock(returncode=0)

        success, error = self.docker_service.start_services("all", True, None, "/path/to/docker-compose.yml")

        assert success is True
        assert error is None
        mock_run.assert_called_once()
        cmd = mock_run.call_args[0][0]
        assert cmd == ["docker", "compose", "-f", "/path/to/docker-compose.yml", "up", "-d"]

    @patch("subprocess.run")
    def test_start_services_failure(self, mock_run):
        mock_run.side_effect = subprocess.CalledProcessError(1, "docker compose", stderr="Service not found")
        success, error = self.docker_service.start_services("web")
        assert success is False
        assert error == "Service not found"
        expected_error = "Service up failed: Service not found"
        self.logger.error.assert_called_once_with(expected_error)

    @patch("subprocess.run")
    def test_start_services_unexpected_error(self, mock_run):
        mock_run.side_effect = Exception("Unexpected error")
        success, error = self.docker_service.start_services("web")
        assert success is False
        assert error == "Unexpected error"
        expected_error = "Unexpected error during up: Unexpected error"
        self.logger.error.assert_called_once_with(expected_error)


class TestUpConfig:
    def test_valid_config_default(self):
        config = UpConfig()
        assert config.name == "all"
        assert config.detach is True
        assert config.env_file is None
        assert config.verbose is False
        assert config.output == "text"
        assert config.dry_run is False

    def test_valid_config_custom(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = UpConfig(
                name="web",
                detach=False,
                env_file="/path/to/.env",
                verbose=True,
                output="json",
                dry_run=True,
                compose_file="/path/to/docker-compose.yml",
            )
            assert config.name == "web"
            assert config.detach is False
            assert config.env_file == "/path/to/.env"
            assert config.verbose is True
            assert config.output == "json"
            assert config.dry_run is True
            assert config.compose_file == "/path/to/docker-compose.yml"

    @patch("os.path.exists")
    def test_validate_env_file_exists(self, mock_exists):
        mock_exists.return_value = True
        config = UpConfig(env_file="/path/to/.env")
        assert config.env_file == "/path/to/.env"

    @patch("os.path.exists")
    def test_validate_env_file_not_exists(self, mock_exists):
        mock_exists.return_value = False
        with pytest.raises(ValidationError):
            UpConfig(env_file="/path/to/.env")

    def test_validate_env_file_none(self):
        config = UpConfig(env_file=None)
        assert config.env_file is None

    def test_validate_env_file_empty(self):
        config = UpConfig(env_file="")
        assert config.env_file is None

    def test_validate_env_file_whitespace(self):
        config = UpConfig(env_file="   ")
        assert config.env_file is None

    def test_validate_env_file_stripped(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = UpConfig(env_file="  /path/to/.env  ")
            assert config.env_file == "/path/to/.env"

    @patch("os.path.exists")
    def test_validate_compose_file_exists(self, mock_exists):
        mock_exists.return_value = True
        config = UpConfig(compose_file="/path/to/docker-compose.yml")
        assert config.compose_file == "/path/to/docker-compose.yml"

    @patch("os.path.exists")
    def test_validate_compose_file_not_exists(self, mock_exists):
        mock_exists.return_value = False
        with pytest.raises(ValidationError):
            UpConfig(compose_file="/path/to/docker-compose.yml")

    def test_validate_compose_file_none(self):
        config = UpConfig(compose_file=None)
        assert config.compose_file is None

    def test_validate_compose_file_empty(self):
        config = UpConfig(compose_file="")
        assert config.compose_file is None

    def test_validate_compose_file_whitespace(self):
        config = UpConfig(compose_file="   ")
        assert config.compose_file is None

    def test_validate_compose_file_stripped(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = UpConfig(compose_file="  /path/to/docker-compose.yml  ")
            assert config.compose_file == "/path/to/docker-compose.yml"


class TestUpService:
    def setup_method(self):
        self.config = UpConfig(name="web", detach=True, env_file=None, compose_file=None)
        self.logger = Mock(spec=Logger)
        self.docker_service = Mock()
        self.service = UpService(self.config, self.logger, self.docker_service)

    def test_create_result_success(self):
        result = self.service._create_result(True)

        assert result.name == self.config.name
        assert result.detach == self.config.detach
        assert result.env_file == self.config.env_file
        assert result.verbose == self.config.verbose
        assert result.output == self.config.output
        assert result.success is True
        assert result.error is None

    def test_create_result_failure(self):
        result = self.service._create_result(False, "Test error")

        assert result.success is False
        assert result.error == "Test error"

    def test_up_success(self):
        self.docker_service.start_services.return_value = (True, None)

        result = self.service.up()

        assert result.success is True
        self.docker_service.start_services.assert_called_once_with(
            self.config.name, self.config.detach, self.config.env_file, self.config.compose_file
        )

    def test_up_failure(self):
        self.docker_service.start_services.return_value = (False, "Test error")

        result = self.service.up()

        assert result.success is False
        assert result.error == "Test error"

    def test_up_and_format_dry_run(self):
        self.config.dry_run = True

        result = self.service.up_and_format()

        assert dry_run_mode in result

    def test_up_and_format_success(self):
        self.docker_service.start_services.return_value = (True, None)

        result = self.service.up_and_format()

        expected_message = services_started_successfully.format(services="web")
        assert expected_message in result


class TestUp:
    def setup_method(self):
        self.logger = Mock(spec=Logger)
        self.up = Up(self.logger)

    def test_up_success(self):
        config = UpConfig(name="web", detach=True, env_file=None)
        with patch(
            "app.commands.service.up.UpService.execute",
            return_value=UpResult(
                name=config.name,
                detach=config.detach,
                env_file=config.env_file,
                verbose=config.verbose,
                output=config.output,
                success=True,
            ),
        ):
            result = self.up.up(config)
            assert result.success is True

    def test_format_output(self):
        result = UpResult(name="web", detach=True, env_file=None, verbose=False, output="text", success=True)

        formatted = self.up.format_output(result, "text")

        expected_message = services_started_successfully.format(services="web")
        assert expected_message in formatted


class TestUpResult:
    def test_up_result_creation(self):
        result = UpResult(
            name="web", detach=True, env_file="/path/to/.env", verbose=True, output="json", success=True, error=None
        )

        assert result.name == "web"
        assert result.detach is True
        assert result.env_file == "/path/to/.env"
        assert result.verbose is True
        assert result.output == "json"
        assert result.success is True
        assert result.error is None

    def test_up_result_default_success(self):
        result = UpResult(name="web", detach=True, env_file=None, verbose=False, output="text")

        assert result.success is False
        assert result.error is None
