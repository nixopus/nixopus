import os
import subprocess
from unittest.mock import Mock, patch

import pytest
from pydantic import ValidationError

from app.commands.service.base import (
    BaseAction,
    BaseConfig,
    BaseDockerCommandBuilder,
    BaseDockerService,
    BaseFormatter,
    BaseResult,
    BaseService,
)
from app.commands.service.up import UpConfig
from app.utils.logger import Logger


class TestBaseDockerCommandBuilder:
    def test_build_command_up_default(self):
        cmd = BaseDockerCommandBuilder.build_command("up", "all", None, None, detach=True)
        assert cmd == ["docker", "compose", "up", "-d"]

    def test_build_command_up_with_service(self):
        cmd = BaseDockerCommandBuilder.build_command("up", "web", None, None, detach=True)
        assert cmd == ["docker", "compose", "up", "-d", "web"]

    def test_build_command_up_without_detach(self):
        cmd = BaseDockerCommandBuilder.build_command("up", "all", None, None, detach=False)
        assert cmd == ["docker", "compose", "up"]

    def test_build_command_down_default(self):
        cmd = BaseDockerCommandBuilder.build_command("down", "all", None, None)
        assert cmd == ["docker", "compose", "down"]

    def test_build_command_down_with_service(self):
        cmd = BaseDockerCommandBuilder.build_command("down", "web", None, None)
        assert cmd == ["docker", "compose", "down", "web"]

    def test_build_command_with_env_file(self):
        cmd = BaseDockerCommandBuilder.build_command("up", "all", "/path/to/.env", None, detach=True)
        assert cmd == ["docker", "compose", "up", "-d", "--env-file", "/path/to/.env"]

    def test_build_command_with_compose_file(self):
        cmd = BaseDockerCommandBuilder.build_command("up", "all", None, "/path/to/docker-compose.yml", detach=True)
        assert cmd == ["docker", "compose", "-f", "/path/to/docker-compose.yml", "up", "-d"]

    def test_build_command_with_all_parameters(self):
        cmd = BaseDockerCommandBuilder.build_command("up", "web", "/path/to/.env", "/path/to/docker-compose.yml", detach=False)
        assert cmd == ["docker", "compose", "-f", "/path/to/docker-compose.yml", "up", "--env-file", "/path/to/.env", "web"]


class TestBaseFormatter:
    def setup_method(self):
        self.formatter = BaseFormatter()

    def test_format_output_success(self):
        result = BaseResult(name="web", env_file=None, verbose=False, output="text", success=True)
        formatted = self.formatter.format_output(result, "text", "Services started: {services}", "Service failed: {error}")
        assert formatted == ""

    def test_format_output_failure(self):
        result = BaseResult(name="web", env_file=None, verbose=False, output="text", success=False, error="Service not found")
        formatted = self.formatter.format_output(result, "text", "Services started: {services}", "Service failed: {error}")
        assert "Service not found" in formatted

    def test_format_output_json(self):
        result = BaseResult(name="web", env_file=None, verbose=False, output="json", success=True)
        formatted = self.formatter.format_output(result, "json", "Services started: {services}", "Service failed: {error}")
        import json

        data = json.loads(formatted)
        assert data["success"] is True
        assert "Services started: web" in data["message"]

    def test_format_dry_run(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = UpConfig(name="web", env_file="/path/to/.env", dry_run=True, detach=True)

            class MockCommandBuilder:
                def build_up_command(self, name, detach, env_file, compose_file):
                    return ["docker", "compose", "up", "-d", "web"]

            dry_run_messages = {
                "mode": "=== DRY RUN MODE ===",
                "command_would_be_executed": "The following commands would be executed:",
                "command": "Command:",
                "service": "Service:",
                "env_file": "Environment file:",
                "detach_mode": "Detach mode:",
                "end": "=== END DRY RUN ===",
            }

            formatted = self.formatter.format_dry_run(config, MockCommandBuilder(), dry_run_messages)
            assert "=== DRY RUN MODE ===" in formatted
            assert "Command:" in formatted
            assert "Service: web" in formatted
            assert "Environment file: /path/to/.env" in formatted
            assert "Detach mode: True" in formatted


class TestBaseDockerService:
    def setup_method(self):
        self.logger = Mock(spec=Logger)

    @patch("subprocess.Popen")
    def test_execute_services_success(self, mock_popen):
        mock_process = Mock()
        mock_process.stdout = ["line1\n", "line2\n"]
        mock_process.wait.return_value = 0
        mock_popen.return_value = mock_process

        docker_service = BaseDockerService(self.logger, "up")

        success, error = docker_service.execute_services("web")

        assert success is True
        assert error == "line1\nline2"

    @patch("subprocess.run")
    def test_execute_services_failure(self, mock_run):
        mock_run.side_effect = subprocess.CalledProcessError(1, "docker compose", stderr="Service not found")
        docker_service = BaseDockerService(self.logger, "down")

        success, error = docker_service.execute_services("web")

        assert success is False
        assert error == "Service not found"
        self.logger.error.assert_called_once_with("Service down failed: Service not found")

    @patch("subprocess.Popen")
    def test_execute_services_unexpected_error(self, mock_popen):
        mock_popen.side_effect = Exception("Unexpected error")
        docker_service = BaseDockerService(self.logger, "up")

        success, error = docker_service.execute_services("web")

        assert success is False
        assert error == "Unexpected error"


class TestBaseConfig:
    def test_valid_config_default(self):
        config = BaseConfig()
        assert config.name == "all"
        assert config.env_file is None
        assert config.verbose is False
        assert config.output == "text"
        assert config.dry_run is False
        assert config.compose_file is None

    def test_valid_config_custom(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = BaseConfig(
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
        config = BaseConfig(env_file="/path/to/.env")
        assert config.env_file == "/path/to/.env"

    @patch("os.path.exists")
    def test_validate_env_file_not_exists(self, mock_exists):
        mock_exists.return_value = False
        with pytest.raises(ValidationError):
            BaseConfig(env_file="/path/to/.env")

    def test_validate_env_file_none(self):
        config = BaseConfig(env_file=None)
        assert config.env_file is None

    def test_validate_env_file_empty(self):
        config = BaseConfig(env_file="")
        assert config.env_file is None

    def test_validate_env_file_whitespace(self):
        config = BaseConfig(env_file="   ")
        assert config.env_file is None

    def test_validate_env_file_stripped(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = BaseConfig(env_file="  /path/to/.env  ")
            assert config.env_file == "/path/to/.env"

    @patch("os.path.exists")
    def test_validate_compose_file_exists(self, mock_exists):
        mock_exists.return_value = True
        config = BaseConfig(compose_file="/path/to/docker-compose.yml")
        assert config.compose_file == "/path/to/docker-compose.yml"

    @patch("os.path.exists")
    def test_validate_compose_file_not_exists(self, mock_exists):
        mock_exists.return_value = False
        with pytest.raises(ValidationError):
            BaseConfig(compose_file="/path/to/docker-compose.yml")

    def test_validate_compose_file_none(self):
        config = BaseConfig(compose_file=None)
        assert config.compose_file is None

    def test_validate_compose_file_empty(self):
        config = BaseConfig(compose_file="")
        assert config.compose_file is None

    def test_validate_compose_file_whitespace(self):
        config = BaseConfig(compose_file="   ")
        assert config.compose_file is None

    def test_validate_compose_file_stripped(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = BaseConfig(compose_file="  /path/to/docker-compose.yml  ")
            assert config.compose_file == "/path/to/docker-compose.yml"


class TestBaseResult:
    def test_base_result_creation(self):
        result = BaseResult(name="web", env_file="/path/to/.env", verbose=True, output="json", success=True, error=None)

        assert result.name == "web"
        assert result.env_file == "/path/to/.env"
        assert result.verbose is True
        assert result.output == "json"
        assert result.success is True
        assert result.error is None

    def test_base_result_default_success(self):
        result = BaseResult(name="web", env_file=None, verbose=False, output="text")

        assert result.name == "web"
        assert result.success is False
        assert result.error is None


class TestBaseService:
    def setup_method(self):
        self.config = BaseConfig(name="web", env_file=None, verbose=False, output="text", dry_run=False)
        self.logger = Mock(spec=Logger)
        self.docker_service = Mock()
        self.service = BaseService(self.config, self.logger, self.docker_service)

    def test_create_result_not_implemented(self):
        with pytest.raises(NotImplementedError):
            self.service._create_result(True)

    def test_execute_not_implemented(self):
        with pytest.raises(NotImplementedError):
            self.service.execute()

    def test_execute_and_format_not_implemented(self):
        with pytest.raises(NotImplementedError):
            self.service.execute_and_format()


class TestBaseAction:
    def setup_method(self):
        self.logger = Mock(spec=Logger)
        self.action = BaseAction(self.logger)

    def test_execute_not_implemented(self):
        config = BaseConfig(name="web")
        with pytest.raises(NotImplementedError):
            self.action.execute(config)

    def test_format_output_not_implemented(self):
        result = BaseResult(name="web", env_file=None, verbose=False, output="text")
        with pytest.raises(NotImplementedError):
            self.action.format_output(result, "text")
