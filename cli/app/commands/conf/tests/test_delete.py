import json
from unittest.mock import Mock, patch

import pytest
from pydantic import ValidationError

from app.commands.conf.delete import Delete, DeleteConfig, DeleteResult, DeleteService, EnvironmentManager
from app.commands.conf.messages import (
    configuration_delete_failed,
    configuration_deleted,
    dry_run_delete_config,
    dry_run_mode,
    end_dry_run,
    key_required_delete,
)
from app.utils.logger import Logger


class TestEnvironmentManager:
    def setup_method(self):
        self.logger = Mock(spec=Logger)
        self.logger.verbose = False  # Add verbose attribute to mock
        self.manager = EnvironmentManager(self.logger)

    @patch("app.commands.conf.base.BaseEnvironmentManager.read_env_file")
    @patch("app.commands.conf.base.BaseEnvironmentManager.write_env_file")
    def test_delete_config_success(self, mock_write_env_file, mock_read_env_file):
        mock_read_env_file.return_value = (True, {"KEY1": "value1", "KEY2": "value2"}, None)
        mock_write_env_file.return_value = (True, None)

        success, error = self.manager.delete_config("api", "KEY1")

        assert success is True
        assert error is None
        mock_read_env_file.assert_called_once_with("/etc/nixopus/source/api/.env")
        mock_write_env_file.assert_called_once_with("/etc/nixopus/source/api/.env", {"KEY2": "value2"})

    @patch("app.commands.conf.base.BaseEnvironmentManager.read_env_file")
    def test_delete_config_read_failure(self, mock_read_env_file):
        mock_read_env_file.return_value = (False, {}, "File not found")

        success, error = self.manager.delete_config("api", "KEY1")

        assert success is False
        assert error == "File not found"

    @patch("app.commands.conf.base.BaseEnvironmentManager.read_env_file")
    def test_delete_config_key_not_found(self, mock_read_env_file):
        mock_read_env_file.return_value = (True, {"KEY1": "value1"}, None)

        success, error = self.manager.delete_config("api", "KEY2")

        assert success is False
        assert "Configuration key 'KEY2' not found" in error

    @patch("app.commands.conf.base.BaseEnvironmentManager.read_env_file")
    @patch("app.commands.conf.base.BaseEnvironmentManager.write_env_file")
    def test_delete_config_write_failure(self, mock_write_env_file, mock_read_env_file):
        mock_read_env_file.return_value = (True, {"KEY1": "value1"}, None)
        mock_write_env_file.return_value = (False, "Write error")

        success, error = self.manager.delete_config("api", "KEY1")

        assert success is False
        assert error == "Write error"

    @patch("app.commands.conf.base.BaseEnvironmentManager.get_service_env_file")
    def test_delete_config_with_custom_env_file(self, mock_get_service_env_file):
        mock_get_service_env_file.return_value = "/custom/.env"

        with patch("app.commands.conf.base.BaseEnvironmentManager.read_env_file") as mock_read:
            with patch("app.commands.conf.base.BaseEnvironmentManager.write_env_file") as mock_write:
                mock_read.return_value = (True, {"KEY1": "value1"}, None)
                mock_write.return_value = (True, None)

                self.manager.delete_config("api", "KEY1", "/custom/.env")

                mock_get_service_env_file.assert_called_once_with("api", "/custom/.env")


class TestDeleteConfig:
    def test_valid_config_default(self):
        config = DeleteConfig(key="TEST_KEY")
        assert config.service == "api"
        assert config.key == "TEST_KEY"
        assert config.value is None
        assert config.verbose is False
        assert config.output == "text"
        assert config.dry_run is False
        assert config.env_file is None

    def test_valid_config_custom(self):
        with patch("os.path.exists") as mock_exists:
            mock_exists.return_value = True
            config = DeleteConfig(
                service="view", key="TEST_KEY", verbose=True, output="json", dry_run=True, env_file="/path/to/.env"
            )
            assert config.service == "view"
            assert config.key == "TEST_KEY"
            assert config.verbose is True
            assert config.output == "json"
            assert config.dry_run is True
            assert config.env_file == "/path/to/.env"


class TestDeleteResult:
    def test_delete_result_default(self):
        result = DeleteResult(service="api", key="TEST_KEY", verbose=False, output="text")
        assert result.service == "api"
        assert result.key == "TEST_KEY"
        assert result.value is None
        assert result.config == {}
        assert result.verbose is False
        assert result.output == "text"
        assert result.success is False
        assert result.error is None

    def test_delete_result_success(self):
        result = DeleteResult(
            service="view", key="TEST_KEY", config={"KEY1": "value1"}, verbose=True, output="json", success=True
        )
        assert result.service == "view"
        assert result.key == "TEST_KEY"
        assert result.config == {"KEY1": "value1"}
        assert result.verbose is True
        assert result.output == "json"
        assert result.success is True


class TestDeleteService:
    def setup_method(self):
        self.config = DeleteConfig(key="TEST_KEY")
        self.logger = Mock(spec=Logger)
        self.environment_service = Mock()
        self.service = DeleteService(self.config, self.logger, self.environment_service)

    def test_delete_service_init(self):
        assert self.service.config == self.config
        assert self.service.logger == self.logger
        assert self.service.environment_service == self.environment_service

    def test_delete_service_init_defaults(self):
        service = DeleteService(self.config)
        assert service.config == self.config
        assert service.logger is not None
        assert service.environment_service is not None

    def test_create_result_success(self):
        result = self.service._create_result(True, config_dict={"KEY1": "value1"})

        assert result.service == "api"
        assert result.key == "TEST_KEY"
        assert result.config == {"KEY1": "value1"}
        assert result.success is True
        assert result.error is None

    def test_create_result_failure(self):
        result = self.service._create_result(False, error="Test error")

        assert result.service == "api"
        assert result.key == "TEST_KEY"
        assert result.config == {}
        assert result.success is False
        assert result.error == "Test error"

    def test_delete_missing_key(self):
        self.config.key = None

        result = self.service.delete()

        assert result.success is False
        assert result.error == key_required_delete

    def test_delete_success(self):
        self.environment_service.delete_config.return_value = (True, None)

        result = self.service.delete()

        assert result.success is True
        assert result.error is None
        self.environment_service.delete_config.assert_called_once_with("api", "TEST_KEY", None)

    def test_delete_failure(self):
        self.environment_service.delete_config.return_value = (False, "Delete error")

        result = self.service.delete()

        assert result.success is False
        assert result.error == "Delete error"

    def test_delete_dry_run(self):
        self.config.dry_run = True

        result = self.service.delete()

        assert result.success is True
        assert result.error is None
        self.environment_service.delete_config.assert_not_called()

    def test_delete_and_format_success(self):
        self.environment_service.delete_config.return_value = (True, None)

        output = self.service.delete_and_format()

        assert configuration_deleted.format(service="api", key="TEST_KEY") in output

    def test_delete_and_format_failure(self):
        self.environment_service.delete_config.return_value = (False, "Delete error")

        output = self.service.delete_and_format()

        assert configuration_delete_failed.format(service="api", error="Delete error") in output

    def test_delete_and_format_dry_run(self):
        self.config.dry_run = True

        output = self.service.delete_and_format()

        assert dry_run_mode in output
        assert dry_run_delete_config.format(service="api", key="TEST_KEY") in output
        assert end_dry_run in output

    def test_format_output_json(self):
        result = DeleteResult(service="api", key="TEST_KEY", success=True, verbose=False, output="json")

        output = self.service._format_output(result, "json")
        data = json.loads(output)

        assert data["service"] == "api"
        assert data["key"] == "TEST_KEY"
        assert data["success"] is True

    def test_format_output_text_success(self):
        result = DeleteResult(service="api", key="TEST_KEY", success=True, verbose=False, output="text")

        output = self.service._format_output(result, "text")

        assert configuration_deleted.format(service="api", key="TEST_KEY") in output

    def test_format_output_text_failure(self):
        result = DeleteResult(service="api", key="TEST_KEY", success=False, error="Test error", verbose=False, output="text")

        output = self.service._format_output(result, "text")

        assert configuration_delete_failed.format(service="api", error="Test error") in output


class TestDelete:
    def setup_method(self):
        self.logger = Mock(spec=Logger)
        self.action = Delete(self.logger)

    def test_delete_action_init(self):
        assert self.action.logger == self.logger

    def test_delete_action_init_default(self):
        action = Delete()
        assert action.logger is None

    def test_delete_success(self):
        config = DeleteConfig(key="TEST_KEY")

        with patch("app.commands.conf.delete.DeleteService") as mock_service_class:
            mock_service = Mock()
            mock_service.execute.return_value = DeleteResult(
                service="api", key="TEST_KEY", success=True, verbose=False, output="text"
            )
            mock_service_class.return_value = mock_service

            result = self.action.delete(config)

            assert result.success is True
            assert result.key == "TEST_KEY"

    def test_format_output(self):
        result = DeleteResult(service="api", key="TEST_KEY", success=True, verbose=False, output="text")

        with patch("app.commands.conf.delete.DeleteService") as mock_service_class:
            mock_service = Mock()
            mock_service._format_output.return_value = "formatted output"
            mock_service_class.return_value = mock_service

            output = self.action.format_output(result, "text")

            assert output == "formatted output"
