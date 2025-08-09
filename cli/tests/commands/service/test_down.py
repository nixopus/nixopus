import subprocess
from unittest.mock import Mock, patch

import pytest
from pydantic import ValidationError

from app.commands.service.down import (
    DockerCommandBuilder,
    DockerService,
    Down,
    DownConfig,
    DownFormatter,
    DownResult,
    DownService,
)
from app.commands.service.messages import (
    dry_run_command,
    dry_run_env_file,
    dry_run_mode,
    dry_run_service,
    services_stopped_successfully,
)
from app.utils.logger import Logger


class TestDockerCommandBuilder:
    def test_build_down_command_default(self):
        cmd = DockerCommandBuilder.build_down_command()
        assert cmd == ["docker", "compose", "down"]

    def test_build_down_command_with_service_name(self):
        cmd = DockerCommandBuilder.build_down_command("web")
        assert cmd == ["docker", "compose", "down", "web"]

    def test_build_down_command_with_env_file(self):
        cmd = DockerCommandBuilder.build_down_command("all", "/path/to/.env")
        assert cmd == ["docker", "compose", "down", "--env-file", "/path/to/.env"]

    def test_build_down_command_with_compose_file(self):
        cmd = DockerCommandBuilder.build_down_command("all", None, "/path/to/docker-compose.yml")
        assert cmd == ["docker", "compose", "-f", "/path/to/docker-compose.yml", "down"]

    def test_build_down_command_with_all_parameters(self):
        cmd = DockerCommandBuilder.build_down_command("api", "/path/to/.env", "/path/to/docker-compose.yml")
        assert cmd == ["docker", "compose", "-f", "/path/to/docker-compose.yml", "down", "--env-file", "/path/to/.env", "api"]


class TestDownFormatter:
    def setup_method(self):
        self.formatter = DownFormatter()

    def test_format_output_success(self):
        result = DownResult(name="web", env_file=None, verbose=False, output="text", success=True)
        formatted = self.formatter.format_output(result, "text")
        assert formatted == ""

    def test_format_output_failure(self):
        result = DownResult(name="web", env_file=None, verbose=False, output="text", success=False, error="Service not found")
        formatted = self.formatter.format_output(result, "text")
        assert "Service not found" in formatted

    def test_format_output_json(self):
        result = DownResult(name="web", env_file=None, verbose=False, output="json", success=True)
        formatted = self.formatter.format_output(result, "json")
        import json

        data = json.loads(formatted)
        assert data["success"] is True
        expected_message = services_stopped_successfully.format(services="web")
        assert expected_message in data["message"]

    def test_format_output_invalid(self):
        result = DownResult(name="web", env_file=None, verbose=False, output="invalid", success=True)
        # The formatter doesn't validate output format, so no ValueError is raised
        formatted = self.formatter.format_output(result, "invalid")
        assert formatted == ""

    def test_format_dry_run_default(self):
        config = DownConfig(name="all", env_file=None, dry_run=True)
        formatted = self.formatter.format_dry_run(config)
        assert dry_run_mode in formatted
        assert dry_run_command in formatted
        assert dry_run_service.format(service="all") in formatted

    def test_format_dry_run_with_service(self):
        config = DownConfig(name="web", env_file=None, dry_run=True)
        formatted = self.formatter.format_dry_run(config)
        assert dry_run_command in formatted
        assert dry_run_service.format(service="web") in formatted

    def test_format_dry_run_with_env_file(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = DownConfig(name="all", env_file="/path/to/.env", dry_run=True)
            formatted = self.formatter.format_dry_run(config)
            assert dry_run_command in formatted
            assert dry_run_env_file.format(env_file="/path/to/.env") in formatted

    def test_format_dry_run_with_compose_file(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = DownConfig(name="all", compose_file="/path/to/docker-compose.yml", dry_run=True)
            formatted = self.formatter.format_dry_run(config)
            assert dry_run_command in formatted
            assert "Command:" in formatted


class TestDockerService:
    def setup_method(self):
        self.logger = Mock(spec=Logger)
        self.docker_service = DockerService(self.logger)

    @patch("subprocess.run")
    def test_stop_services_success(self, mock_run):
        mock_result = Mock(returncode=0, stdout="", stderr="")
        mock_run.return_value = mock_result

        success, error = self.docker_service.stop_services("web")

        assert success is True
        assert error == ""

    @patch("subprocess.run")
    def test_stop_services_with_env_file(self, mock_run):
        mock_result = Mock(returncode=0, stdout="", stderr="")
        mock_run.return_value = mock_result

        success, error = self.docker_service.stop_services("all", "/path/to/.env")

        assert success is True
        assert error == ""
        mock_run.assert_called_once()
        cmd = mock_run.call_args[0][0]
        assert cmd == ["docker", "compose", "down", "--env-file", "/path/to/.env"]

    @patch("subprocess.run")
    def test_stop_services_with_compose_file(self, mock_run):
        mock_result = Mock(returncode=0, stdout="", stderr="")
        mock_run.return_value = mock_result

        success, error = self.docker_service.stop_services("all", None, "/path/to/docker-compose.yml")

        assert success is True
        assert error == ""
        mock_run.assert_called_once()
        cmd = mock_run.call_args[0][0]
        assert cmd == ["docker", "compose", "-f", "/path/to/docker-compose.yml", "down"]

    @patch("subprocess.run")
    def test_stop_services_failure(self, mock_run):
        mock_run.side_effect = subprocess.CalledProcessError(1, "docker compose down", stderr="Service not found")

        success, error = self.docker_service.stop_services("web")

        assert success is False
        assert error == "Service not found"
        expected_error = "Service down failed: Service not found"
        self.logger.error.assert_called_once_with(expected_error)

    @patch("subprocess.run")
    def test_stop_services_unexpected_error(self, mock_run):
        mock_run.side_effect = Exception("Unexpected error")

        success, error = self.docker_service.stop_services("web")

        assert success is False
        assert error == "Unexpected error"
        expected_error = "Unexpected error during down: Unexpected error"
        self.logger.error.assert_called_once_with(expected_error)


class TestDownConfig:
    def test_valid_config_default(self):
        config = DownConfig()
        assert config.name == "all"
        assert config.env_file is None
        assert config.verbose is False
        assert config.output == "text"
        assert config.dry_run is False
        assert config.compose_file is None

    def test_valid_config_custom(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = DownConfig(
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
        config = DownConfig(env_file="/path/to/.env")
        assert config.env_file == "/path/to/.env"

    @patch("os.path.exists")
    def test_validate_env_file_not_exists(self, mock_exists):
        mock_exists.return_value = False
        with pytest.raises(ValidationError):
            DownConfig(env_file="/path/to/.env")

    def test_validate_env_file_none(self):
        config = DownConfig(env_file=None)
        assert config.env_file is None

    def test_validate_env_file_empty(self):
        config = DownConfig(env_file="")
        assert config.env_file is None

    def test_validate_env_file_whitespace(self):
        config = DownConfig(env_file="   ")
        assert config.env_file is None

    def test_validate_env_file_stripped(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = DownConfig(env_file="  /path/to/.env  ")
            assert config.env_file == "/path/to/.env"

    @patch("os.path.exists")
    def test_validate_compose_file_exists(self, mock_exists):
        mock_exists.return_value = True
        config = DownConfig(compose_file="/path/to/docker-compose.yml")
        assert config.compose_file == "/path/to/docker-compose.yml"

    @patch("os.path.exists")
    def test_validate_compose_file_not_exists(self, mock_exists):
        mock_exists.return_value = False
        with pytest.raises(ValidationError):
            DownConfig(compose_file="/path/to/docker-compose.yml")

    def test_validate_compose_file_none(self):
        config = DownConfig(compose_file=None)
        assert config.compose_file is None

    def test_validate_compose_file_empty(self):
        config = DownConfig(compose_file="")
        assert config.compose_file is None

    def test_validate_compose_file_whitespace(self):
        config = DownConfig(compose_file="   ")
        assert config.compose_file is None

    def test_validate_compose_file_stripped(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = DownConfig(compose_file="  /path/to/docker-compose.yml  ")
            assert config.compose_file == "/path/to/docker-compose.yml"


class TestDownService:
    def setup_method(self):
        self.config = DownConfig(name="web", env_file=None, verbose=False, output="text", dry_run=False)
        self.logger = Mock(spec=Logger)
        self.docker_service = Mock()
        self.service = DownService(self.config, self.logger, self.docker_service)

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

    def test_down_success(self):
        self.docker_service.stop_services.return_value = (True, None)

        result = self.service.down()

        assert result.success is True
        assert result.error is None
        self.docker_service.stop_services.assert_called_once_with("web", None, None)

    def test_down_failure(self):
        self.docker_service.stop_services.return_value = (False, "Service not found")

        result = self.service.down()

        assert result.success is False
        assert result.error == "Service not found"

    def test_down_and_format_dry_run(self):
        self.config.dry_run = True
        formatted = self.service.down_and_format()
        assert dry_run_mode in formatted
        assert dry_run_command in formatted

    def test_down_and_format_success(self):
        self.docker_service.stop_services.return_value = (True, "")
        formatted = self.service.down_and_format()
        assert formatted == ""


class TestDown:
    def setup_method(self):
        self.logger = Mock(spec=Logger)
        self.down = Down(self.logger)

    def test_down_success(self):
        config = DownConfig(name="web", env_file=None, verbose=False, output="text", dry_run=False)

        with patch("app.commands.service.down.DockerService") as mock_docker_service_class:
            mock_docker_service = Mock()
            mock_docker_service.stop_services.return_value = (True, "")
            mock_docker_service_class.return_value = mock_docker_service

            result = self.down.down(config)

            assert result.success is True
            assert result.error is None
            assert result.name == "web"

    def test_down_failure(self):
        config = DownConfig(name="web", env_file=None, verbose=False, output="text", dry_run=False)

        with patch("app.commands.service.down.DockerService") as mock_docker_service_class:
            mock_docker_service = Mock()
            mock_docker_service.stop_services.return_value = (False, "Service not found")
            mock_docker_service_class.return_value = mock_docker_service

            result = self.down.down(config)

            assert result.success is False
            assert result.error == "Service not found"

    def test_format_output(self):
        result = DownResult(name="web", env_file=None, verbose=False, output="text", success=True)

        formatted = self.down.format_output(result, "text")
        assert formatted == ""


class TestDownResult:
    def test_down_result_creation(self):
        result = DownResult(name="web", env_file="/path/to/.env", verbose=True, output="json", success=True, error=None)

        assert result.name == "web"
        assert result.env_file == "/path/to/.env"
        assert result.verbose is True
        assert result.output == "json"
        assert result.success is True
        assert result.error is None

    def test_down_result_default_success(self):
        result = DownResult(name="web", env_file=None, verbose=False, output="text")

        assert result.name == "web"
        assert result.success is False
        assert result.error is None
