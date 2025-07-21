import json
from unittest.mock import Mock, patch

import pytest
from pydantic import ValidationError

from app.commands.conf.list import EnvironmentManager, List, ListConfig, ListResult, ListService
from app.commands.conf.messages import (
    configuration_list_failed,
    configuration_listed,
    dry_run_list_config,
    dry_run_mode,
    end_dry_run,
    no_configuration_found,
)
from app.utils.logger import Logger


class TestEnvironmentManager:
    def setup_method(self):
        self.logger = Mock(spec=Logger)
        self.logger.verbose = False  # Add verbose attribute to mock
        self.manager = EnvironmentManager(self.logger)

    @patch("app.commands.conf.base.BaseEnvironmentManager.read_env_file")
    def test_list_config_success(self, mock_read_env_file):
        mock_read_env_file.return_value = (True, {"KEY1": "value1", "KEY2": "value2"}, None)

        success, config, error = self.manager.list_config("api")

        assert success is True
        assert config == {"KEY1": "value1", "KEY2": "value2"}
        assert error is None
        mock_read_env_file.assert_called_once_with("/etc/nixopus/source/api/.env")

    @patch("app.commands.conf.base.BaseEnvironmentManager.read_env_file")
    def test_list_config_failure(self, mock_read_env_file):
        mock_read_env_file.return_value = (False, {}, "File not found")

        success, config, error = self.manager.list_config("api")

        assert success is False
        assert config == {}
        assert error == "File not found"

    @patch("app.commands.conf.base.BaseEnvironmentManager.get_service_env_file")
    def test_list_config_with_custom_env_file(self, mock_get_service_env_file):
        mock_get_service_env_file.return_value = "/custom/.env"

        self.manager.list_config("api", "/custom/.env")

        mock_get_service_env_file.assert_called_once_with("api", "/custom/.env")


class TestListConfig:
    def test_valid_config_default(self):
        config = ListConfig()
        assert config.service == "api"
        assert config.key is None
        assert config.value is None
        assert config.verbose is False
        assert config.output == "text"
        assert config.dry_run is False
        assert config.env_file is None

    def test_valid_config_custom(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = ListConfig(service="view", verbose=True, output="json", dry_run=True, env_file="/path/to/.env")
            assert config.service == "view"
            assert config.verbose is True
            assert config.output == "json"
            assert config.dry_run is True
            assert config.env_file == "/path/to/.env"


class TestListResult:
    def test_list_result_default(self):
        result = ListResult(service="api", verbose=False, output="text")
        assert result.service == "api"
        assert result.key is None
        assert result.value is None
        assert result.config == {}
        assert result.verbose is False
        assert result.output == "text"
        assert result.success is False
        assert result.error is None

    def test_list_result_with_config(self):
        result = ListResult(
            service="view", config={"KEY1": "value1", "KEY2": "value2"}, verbose=True, output="json", success=True
        )
        assert result.service == "view"
        assert result.config == {"KEY1": "value1", "KEY2": "value2"}
        assert result.verbose is True
        assert result.output == "json"
        assert result.success is True


class TestListService:
    def setup_method(self):
        self.config = ListConfig()
        self.logger = Mock(spec=Logger)
        self.environment_service = Mock()
        self.service = ListService(self.config, self.logger, self.environment_service)

    def test_list_service_init(self):
        assert self.service.config == self.config
        assert self.service.logger == self.logger
        assert self.service.environment_service == self.environment_service

    def test_list_service_init_defaults(self):
        service = ListService(self.config)
        assert service.config == self.config
        assert service.logger is not None
        assert service.environment_service is not None

    def test_create_result_success(self):
        result = self.service._create_result(True, config_dict={"KEY1": "value1"})

        assert result.service == "api"
        assert result.config == {"KEY1": "value1"}
        assert result.success is True
        assert result.error is None

    def test_create_result_failure(self):
        result = self.service._create_result(False, error="Test error")

        assert result.service == "api"
        assert result.config == {}
        assert result.success is False
        assert result.error == "Test error"

    def test_list_success(self):
        self.environment_service.list_config.return_value = (True, {"KEY1": "value1"}, None)

        result = self.service.list()

        assert result.success is True
        assert result.config == {"KEY1": "value1"}
        assert result.error is None

    def test_list_failure(self):
        self.environment_service.list_config.return_value = (False, {}, "File not found")

        result = self.service.list()

        assert result.success is False
        assert result.error == "File not found"
        self.logger.error.assert_called_once_with(configuration_list_failed.format(service="api", error="File not found"))

    def test_list_dry_run(self):
        self.config.dry_run = True

        result = self.service.list()

        assert result.success is True
        assert result.error is None
        self.environment_service.list_config.assert_not_called()

    def test_list_and_format_success(self):
        self.environment_service.list_config.return_value = (True, {"KEY1": "value1"}, None)

        output = self.service.list_and_format()

        assert "KEY1" in output
        assert "value1" in output
        assert "Key" in output
        assert "Value" in output

    def test_list_and_format_failure(self):
        self.environment_service.list_config.return_value = (False, {}, "File not found")

        output = self.service.list_and_format()

        assert configuration_list_failed.format(service="api", error="File not found") in output

    def test_list_and_format_dry_run(self):
        self.config.dry_run = True

        output = self.service.list_and_format()

        assert dry_run_mode in output
        assert dry_run_list_config.format(service="api") in output
        assert end_dry_run in output

    def test_format_output_json(self):
        result = ListResult(service="api", config={"KEY1": "value1"}, success=True, verbose=False, output="json")

        output = self.service._format_output(result, "json")
        data = json.loads(output)

        assert data["success"] is True
        assert data["message"] == "Configuration listed successfully for service: api"
        assert data["data"] == {"KEY1": "value1"}

    def test_format_output_text_success(self):
        result = ListResult(
            service="api", config={"KEY1": "value1", "KEY2": "value2"}, success=True, verbose=False, output="text"
        )

        output = self.service._format_output(result, "text")

        assert "KEY1" in output
        assert "value1" in output
        assert "KEY2" in output
        assert "value2" in output
        assert "Key" in output
        assert "Value" in output

    def test_format_output_text_failure(self):
        result = ListResult(service="api", success=False, error="Test error", verbose=False, output="text")

        output = self.service._format_output(result, "text")

        assert configuration_list_failed.format(service="api", error="Test error") in output

    def test_format_output_text_no_config(self):
        result = ListResult(service="api", config={}, success=True, verbose=False, output="text")

        output = self.service._format_output(result, "text")

        assert no_configuration_found.format(service="api") in output


class TestList:
    def setup_method(self):
        self.logger = Mock(spec=Logger)
        self.action = List(self.logger)

    def test_list_action_init(self):
        assert self.action.logger == self.logger

    def test_list_action_init_default(self):
        action = List()
        assert action.logger is None

    def test_list_success(self):
        config = ListConfig(service="api")

        with patch("app.commands.conf.list.ListService") as mock_service_class:
            mock_service = Mock()
            mock_service.execute.return_value = ListResult(
                service="api", config={"KEY1": "value1"}, success=True, verbose=False, output="text"
            )
            mock_service_class.return_value = mock_service

            result = self.action.list(config)

            assert result.success is True
            assert result.config == {"KEY1": "value1"}

    def test_format_output(self):
        result = ListResult(service="api", config={"KEY1": "value1"}, success=True, verbose=False, output="text")

        with patch("app.commands.conf.list.ListService") as mock_service_class:
            mock_service = Mock()
            mock_service._format_output.return_value = "formatted output"
            mock_service_class.return_value = mock_service

            output = self.action.format_output(result, "text")

            assert output == "formatted output"
