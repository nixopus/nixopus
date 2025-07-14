import pytest
import json
from unittest.mock import Mock, patch
from pydantic import ValidationError

from app.commands.conf.set import (
    EnvironmentManager,
    SetResult,
    SetConfig,
    SetService,
    Set
)
from app.commands.conf.messages import (
    configuration_set,
    configuration_set_failed,
    key_required,
    value_required,
    dry_run_mode,
    dry_run_set_config,
    end_dry_run
)
from app.utils.logger import Logger


class TestEnvironmentManager:
    def setup_method(self):
        self.logger = Mock(spec=Logger)
        self.manager = EnvironmentManager(self.logger)
    
    @patch('app.commands.conf.base.BaseEnvironmentManager.read_env_file')
    @patch('app.commands.conf.base.BaseEnvironmentManager.write_env_file')
    def test_set_config_success(self, mock_write_env_file, mock_read_env_file):
        mock_read_env_file.return_value = (True, {"KEY1": "value1"}, None)
        mock_write_env_file.return_value = (True, None)
        
        success, error = self.manager.set_config("api", "KEY2", "value2")
        
        assert success is True
        assert error is None
        mock_read_env_file.assert_called_once_with("/etc/nixopus/source/api/.env")
        mock_write_env_file.assert_called_once_with("/etc/nixopus/source/api/.env", {"KEY1": "value1", "KEY2": "value2"})
    
    @patch('app.commands.conf.base.BaseEnvironmentManager.read_env_file')
    def test_set_config_read_failure(self, mock_read_env_file):
        mock_read_env_file.return_value = (False, {}, "File not found")
        
        success, error = self.manager.set_config("api", "KEY1", "value1")
        
        assert success is False
        assert error == "File not found"
    
    @patch('app.commands.conf.base.BaseEnvironmentManager.read_env_file')
    @patch('app.commands.conf.base.BaseEnvironmentManager.write_env_file')
    def test_set_config_write_failure(self, mock_write_env_file, mock_read_env_file):
        mock_read_env_file.return_value = (True, {"KEY1": "value1"}, None)
        mock_write_env_file.return_value = (False, "Write error")
        
        success, error = self.manager.set_config("api", "KEY2", "value2")
        
        assert success is False
        assert error == "Write error"
    
    @patch('app.commands.conf.base.BaseEnvironmentManager.get_service_env_file')
    def test_set_config_with_custom_env_file(self, mock_get_service_env_file):
        mock_get_service_env_file.return_value = "/custom/.env"
        
        with patch('app.commands.conf.base.BaseEnvironmentManager.read_env_file') as mock_read:
            with patch('app.commands.conf.base.BaseEnvironmentManager.write_env_file') as mock_write:
                mock_read.return_value = (True, {}, None)
                mock_write.return_value = (True, None)
                
                self.manager.set_config("api", "KEY1", "value1", "/custom/.env")
                
                mock_get_service_env_file.assert_called_once_with("api", "/custom/.env")


class TestSetConfig:
    def test_valid_config_default(self):
        config = SetConfig(key="TEST_KEY", value="test_value")
        assert config.service == "api"
        assert config.key == "TEST_KEY"
        assert config.value == "test_value"
        assert config.verbose is False
        assert config.output == "text"
        assert config.dry_run is False
        assert config.env_file is None
    
    def test_valid_config_custom(self):
        with patch('os.path.exists') as mock_exists:
            mock_exists.return_value = True
            config = SetConfig(
                service="view",
                key="TEST_KEY",
                value="test_value",
                verbose=True,
                output="json",
                dry_run=True,
                env_file="/path/to/.env"
            )
            assert config.service == "view"
            assert config.key == "TEST_KEY"
            assert config.value == "test_value"
            assert config.verbose is True
            assert config.output == "json"
            assert config.dry_run is True
            assert config.env_file == "/path/to/.env"


class TestSetResult:
    def test_set_result_default(self):
        result = SetResult(
            service="api",
            key="TEST_KEY",
            value="test_value",
            verbose=False,
            output="text"
        )
        assert result.service == "api"
        assert result.key == "TEST_KEY"
        assert result.value == "test_value"
        assert result.config == {}
        assert result.verbose is False
        assert result.output == "text"
        assert result.success is False
        assert result.error is None
    
    def test_set_result_success(self):
        result = SetResult(
            service="view",
            key="TEST_KEY",
            value="test_value",
            config={"KEY1": "value1"},
            verbose=True,
            output="json",
            success=True
        )
        assert result.service == "view"
        assert result.key == "TEST_KEY"
        assert result.value == "test_value"
        assert result.config == {"KEY1": "value1"}
        assert result.verbose is True
        assert result.output == "json"
        assert result.success is True


class TestSetService:
    def setup_method(self):
        self.config = SetConfig(key="TEST_KEY", value="test_value")
        self.logger = Mock(spec=Logger)
        self.environment_service = Mock()
        self.service = SetService(self.config, self.logger, self.environment_service)
    
    def test_set_service_init(self):
        assert self.service.config == self.config
        assert self.service.logger == self.logger
        assert self.service.environment_service == self.environment_service
    
    def test_set_service_init_defaults(self):
        service = SetService(self.config)
        assert service.config == self.config
        assert service.logger is not None
        assert service.environment_service is not None
    
    def test_create_result_success(self):
        result = self.service._create_result(True, config_dict={"KEY1": "value1"})
        
        assert result.service == "api"
        assert result.key == "TEST_KEY"
        assert result.value == "test_value"
        assert result.config == {"KEY1": "value1"}
        assert result.success is True
        assert result.error is None
    
    def test_create_result_failure(self):
        result = self.service._create_result(False, error="Test error")
        
        assert result.service == "api"
        assert result.key == "TEST_KEY"
        assert result.value == "test_value"
        assert result.config == {}
        assert result.success is False
        assert result.error == "Test error"
    
    def test_set_missing_key(self):
        self.config.key = None
        
        result = self.service.set()
        
        assert result.success is False
        assert result.error == key_required
    
    def test_set_missing_value(self):
        self.config.value = None
        
        result = self.service.set()
        
        assert result.success is False
        assert result.error == value_required
    
    def test_set_success(self):
        self.environment_service.set_config.return_value = (True, None)
        
        result = self.service.set()
        
        assert result.success is True
        assert result.error is None
        self.logger.info.assert_called_once_with(configuration_set.format(
            service="api", key="TEST_KEY", value="test_value"
        ))
        self.environment_service.set_config.assert_called_once_with(
            "api", "TEST_KEY", "test_value", None
        )
    
    def test_set_failure(self):
        self.environment_service.set_config.return_value = (False, "Write error")
        
        result = self.service.set()
        
        assert result.success is False
        assert result.error == "Write error"
        self.logger.error.assert_called_once_with(configuration_set_failed.format(
            service="api", error="Write error"
        ))
    
    def test_set_dry_run(self):
        self.config.dry_run = True
        
        result = self.service.set()
        
        assert result.success is True
        assert result.error is None
        self.environment_service.set_config.assert_not_called()
    
    def test_set_and_format_success(self):
        self.environment_service.set_config.return_value = (True, None)
        
        output = self.service.set_and_format()
        
        assert configuration_set.format(service="api", key="TEST_KEY", value="test_value") in output
    
    def test_set_and_format_failure(self):
        self.environment_service.set_config.return_value = (False, "Write error")
        
        output = self.service.set_and_format()
        
        assert configuration_set_failed.format(service="api", error="Write error") in output
    
    def test_set_and_format_dry_run(self):
        self.config.dry_run = True
        
        output = self.service.set_and_format()
        
        assert dry_run_mode in output
        assert dry_run_set_config.format(
            service="api", key="TEST_KEY", value="test_value"
        ) in output
        assert end_dry_run in output
    
    def test_format_output_json(self):
        result = SetResult(
            service="api",
            key="TEST_KEY",
            value="test_value",
            success=True,
            verbose=False,
            output="json"
        )
        
        output = self.service._format_output(result, "json")
        data = json.loads(output)
        
        assert data["service"] == "api"
        assert data["key"] == "TEST_KEY"
        assert data["value"] == "test_value"
        assert data["success"] is True
    
    def test_format_output_text_success(self):
        result = SetResult(
            service="api",
            key="TEST_KEY",
            value="test_value",
            success=True,
            verbose=False,
            output="text"
        )
        
        output = self.service._format_output(result, "text")
        
        assert configuration_set.format(service="api", key="TEST_KEY", value="test_value") in output
    
    def test_format_output_text_failure(self):
        result = SetResult(
            service="api",
            key="TEST_KEY",
            value="test_value",
            success=False,
            error="Test error",
            verbose=False,
            output="text"
        )
        
        output = self.service._format_output(result, "text")
        
        assert configuration_set_failed.format(service="api", error="Test error") in output


class TestSet:
    def setup_method(self):
        self.logger = Mock(spec=Logger)
        self.action = Set(self.logger)
    
    def test_set_action_init(self):
        assert self.action.logger == self.logger
    
    def test_set_action_init_default(self):
        action = Set()
        assert action.logger is None
    
    def test_set_success(self):
        config = SetConfig(key="TEST_KEY", value="test_value")
        
        with patch('app.commands.conf.set.SetService') as mock_service_class:
            mock_service = Mock()
            mock_service.execute.return_value = SetResult(
                service="api",
                key="TEST_KEY",
                value="test_value",
                success=True,
                verbose=False,
                output="text"
            )
            mock_service_class.return_value = mock_service
            
            result = self.action.set(config)
            
            assert result.success is True
            assert result.key == "TEST_KEY"
            assert result.value == "test_value"
    
    def test_format_output(self):
        result = SetResult(
            service="api",
            key="TEST_KEY",
            value="test_value",
            success=True,
            verbose=False,
            output="text"
        )
        
        with patch('app.commands.conf.set.SetService') as mock_service_class:
            mock_service = Mock()
            mock_service._format_output.return_value = "formatted output"
            mock_service_class.return_value = mock_service
            
            output = self.action.format_output(result, "text")
            
            assert output == "formatted output" 
