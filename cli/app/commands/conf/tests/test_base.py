import pytest
import os
import tempfile
import shutil
from unittest.mock import Mock, patch, mock_open
from pydantic import ValidationError

from app.commands.conf.base import (
    BaseEnvironmentManager,
    BaseConfig,
    BaseResult,
    BaseService,
    BaseAction
)
from app.utils.logger import Logger


class TestBaseEnvironmentManager:
    def setup_method(self):
        self.logger = Mock(spec=Logger)
        self.manager = BaseEnvironmentManager(self.logger)
    
    @patch('os.path.exists')
    def test_read_env_file_exists(self, mock_exists):
        mock_exists.return_value = True
        
        with patch('builtins.open', mock_open(read_data="KEY1=value1\nKEY2=value2\n")):
            success, config, error = self.manager.read_env_file("/path/to/.env")
            
            assert success is True
            assert config == {"KEY1": "value1", "KEY2": "value2"}
            assert error is None
    
    @patch('os.path.exists')
    def test_read_env_file_not_exists(self, mock_exists):
        mock_exists.return_value = False
        
        success, config, error = self.manager.read_env_file("/path/to/.env")
        
        assert success is False
        assert config == {}
        assert "Environment file not found" in error
    
    @patch('os.path.exists')
    def test_read_env_file_with_comments_and_empty_lines(self, mock_exists):
        mock_exists.return_value = True
        
        content = "# Comment line\nKEY1=value1\n\nKEY2=value2\n# Another comment"
        with patch('builtins.open', mock_open(read_data=content)):
            success, config, error = self.manager.read_env_file("/path/to/.env")
            
            assert success is True
            assert config == {"KEY1": "value1", "KEY2": "value2"}
            assert error is None
    
    @patch('os.path.exists')
    def test_read_env_file_with_invalid_line(self, mock_exists):
        mock_exists.return_value = True
        
        content = "KEY1=value1\nINVALID_LINE\nKEY2=value2"
        with patch('builtins.open', mock_open(read_data=content)):
            success, config, error = self.manager.read_env_file("/path/to/.env")
            
            assert success is True
            assert config == {"KEY1": "value1", "KEY2": "value2"}
            assert error is None
            self.logger.warning.assert_called_once()
    
    @patch('os.path.exists')
    def test_create_backup_file_exists(self, mock_exists):
        mock_exists.return_value = True
        
        with patch('shutil.copy2') as mock_copy:
            success, backup_path, error = self.manager._create_backup("/path/to/.env")
            
            assert success is True
            assert backup_path == "/path/to/.env.backup"
            assert error is None
            mock_copy.assert_called_once_with("/path/to/.env", "/path/to/.env.backup")
    
    @patch('os.path.exists')
    def test_create_backup_file_not_exists(self, mock_exists):
        mock_exists.return_value = False
        
        success, backup_path, error = self.manager._create_backup("/path/to/.env")
        
        assert success is True
        assert backup_path is None
        assert error is None
    
    @patch('os.path.exists')
    def test_create_backup_failure(self, mock_exists):
        mock_exists.return_value = True
        
        with patch('shutil.copy2', side_effect=Exception("Copy failed")):
            success, backup_path, error = self.manager._create_backup("/path/to/.env")
            
            assert success is False
            assert backup_path is None
            assert "Failed to create backup" in error
    
    @patch('os.path.exists')
    def test_restore_backup_success(self, mock_exists):
        mock_exists.return_value = True
        
        with patch('shutil.copy2') as mock_copy:
            with patch('os.remove') as mock_remove:
                success, error = self.manager._restore_backup("/path/to/.env.backup", "/path/to/.env")
                
                assert success is True
                assert error is None
                mock_copy.assert_called_once_with("/path/to/.env.backup", "/path/to/.env")
                mock_remove.assert_called_once_with("/path/to/.env.backup")
    
    @patch('os.path.exists')
    def test_restore_backup_not_exists(self, mock_exists):
        mock_exists.return_value = False
        
        success, error = self.manager._restore_backup("/path/to/.env.backup", "/path/to/.env")
        
        assert success is False
        assert error == "Backup file not found"
    
    @patch('os.path.exists')
    def test_restore_backup_failure(self, mock_exists):
        mock_exists.return_value = True
        
        with patch('shutil.copy2', side_effect=Exception("Copy failed")):
            success, error = self.manager._restore_backup("/path/to/.env.backup", "/path/to/.env")
            
            assert success is False
            assert "Failed to restore from backup" in error
    
    @patch('os.makedirs')
    @patch('tempfile.NamedTemporaryFile')
    @patch('os.replace')
    @patch('os.fsync')
    def test_atomic_write_success(self, mock_fsync, mock_replace, mock_tempfile, mock_makedirs):
        config = {"KEY2": "value2", "KEY1": "value1"}
        
        mock_temp = Mock()
        mock_temp.name = "/tmp/temp_file"
        mock_temp.fileno.return_value = 123
        mock_tempfile.return_value.__enter__.return_value = mock_temp
        mock_tempfile.return_value.__exit__.return_value = None
        
        success, error = self.manager._atomic_write("/path/to/.env", config)
        
        assert success is True
        assert error is None
        mock_makedirs.assert_called_once_with("/path/to", exist_ok=True)
        mock_temp.write.assert_called()
        mock_temp.flush.assert_called_once()
        mock_temp.fileno.assert_called_once()
        mock_replace.assert_called_once_with("/tmp/temp_file", "/path/to/.env")
    
    @patch('os.makedirs')
    @patch('tempfile.NamedTemporaryFile')
    def test_atomic_write_failure(self, mock_tempfile, mock_makedirs):
        config = {"KEY1": "value1"}
        
        mock_tempfile.side_effect = Exception("Temp file creation failed")
        
        success, error = self.manager._atomic_write("/path/to/.env", config)
        
        assert success is False
        assert "Failed to write environment file" in error
    
    @patch('os.makedirs')
    @patch('tempfile.NamedTemporaryFile')
    @patch('os.replace')
    @patch('os.fsync')
    def test_atomic_write_simple(self, mock_fsync, mock_replace, mock_tempfile, mock_makedirs):
        config = {"KEY1": "value1"}
        
        mock_temp = Mock()
        mock_temp.name = "/tmp/temp_file"
        mock_temp.fileno.return_value = 123
        mock_tempfile.return_value.__enter__.return_value = mock_temp
        mock_tempfile.return_value.__exit__.return_value = None
        
        success, error = self.manager._atomic_write("/path/to/.env", config)
        
        assert success is True
        assert error is None
    
    @patch('os.path.exists')
    @patch('shutil.copy2')
    @patch('tempfile.NamedTemporaryFile')
    @patch('os.replace')
    @patch('os.fsync')
    @patch('os.makedirs')
    def test_write_env_file_success_with_backup(self, mock_makedirs, mock_fsync, mock_replace, mock_tempfile, mock_copy, mock_exists):
        mock_exists.return_value = True
        config = {"KEY2": "value2", "KEY1": "value1"}
        
        mock_temp = Mock()
        mock_temp.name = "/tmp/temp_file"
        mock_temp.fileno.return_value = 123
        mock_tempfile.return_value.__enter__.return_value = mock_temp
        mock_tempfile.return_value.__exit__.return_value = None
        
        with patch('os.remove') as mock_remove:
            success, error = self.manager.write_env_file("/path/to/.env", config)
            
            assert success is True
            assert error is None
            mock_copy.assert_called_once_with("/path/to/.env", "/path/to/.env.backup")
            mock_remove.assert_called_once_with("/path/to/.env.backup")
            self.logger.info.assert_called()
    
    @patch('os.path.exists')
    @patch('tempfile.NamedTemporaryFile')
    @patch('os.replace')
    @patch('os.fsync')
    @patch('os.makedirs')
    def test_write_env_file_success_no_backup_needed(self, mock_makedirs, mock_fsync, mock_replace, mock_tempfile, mock_exists):
        mock_exists.return_value = False
        config = {"KEY1": "value1"}
        
        mock_temp = Mock()
        mock_temp.name = "/tmp/temp_file"
        mock_temp.fileno.return_value = 123
        mock_tempfile.return_value.__enter__.return_value = mock_temp
        mock_tempfile.return_value.__exit__.return_value = None
        
        success, error = self.manager.write_env_file("/path/to/.env", config)
        
        assert success is True
        assert error is None
        mock_replace.assert_called_once_with("/tmp/temp_file", "/path/to/.env")
    
    @patch('os.path.exists')
    @patch('shutil.copy2')
    def test_write_env_file_backup_failure(self, mock_copy, mock_exists):
        mock_exists.return_value = True
        mock_copy.side_effect = Exception("Backup failed")
        config = {"KEY1": "value1"}
        
        success, error = self.manager.write_env_file("/path/to/.env", config)
        
        assert success is False
        assert "Failed to create backup" in error
    
    @patch('os.path.exists')
    @patch('shutil.copy2')
    @patch('tempfile.NamedTemporaryFile')
    def test_write_env_file_write_failure_with_restore(self, mock_tempfile, mock_copy, mock_exists):
        mock_exists.return_value = True
        config = {"KEY1": "value1"}
        
        mock_tempfile.side_effect = Exception("Write failed")
        
        with patch.object(self.manager, '_restore_backup') as mock_restore:
            mock_restore.return_value = (True, None)
            
            success, error = self.manager.write_env_file("/path/to/.env", config)
            
            assert success is False
            assert "Failed to write environment file" in error
            mock_restore.assert_called_once_with("/path/to/.env.backup", "/path/to/.env")
            self.logger.warning.assert_called()
            self.logger.info.assert_called()
    
    def test_get_service_env_file_with_custom_env_file(self):
        env_file = self.manager.get_service_env_file("api", "/custom/.env")
        assert env_file == "/custom/.env"
    
    def test_get_service_env_file_api_service(self):
        env_file = self.manager.get_service_env_file("api")
        assert env_file == "/etc/nixopus/source/api/.env"
    
    def test_get_service_env_file_view_service(self):
        env_file = self.manager.get_service_env_file("view")
        assert env_file == "/etc/nixopus/source/view/.env"
    
    def test_get_service_env_file_invalid_service(self):
        with pytest.raises(ValueError, match="Invalid service: invalid"):
            self.manager.get_service_env_file("invalid")


class TestBaseConfig:
    def test_valid_config_default(self):
        config = BaseConfig()
        assert config.service == "api"
        assert config.key is None
        assert config.value is None
        assert config.verbose is False
        assert config.output == "text"
        assert config.dry_run is False
        assert config.env_file is None
    
    def test_valid_config_custom(self):
        with patch('os.path.exists') as mock_exists:
            mock_exists.return_value = True
            config = BaseConfig(
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
    
    @patch('os.path.exists')
    def test_validate_env_file_exists(self, mock_exists):
        mock_exists.return_value = True
        config = BaseConfig(env_file="/path/to/.env")
        assert config.env_file == "/path/to/.env"
    
    @patch('os.path.exists')
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
        with patch('os.path.exists') as mock_exists:
            mock_exists.return_value = True
            config = BaseConfig(env_file="  /path/to/.env  ")
            assert config.env_file == "/path/to/.env"


class TestBaseResult:
    def test_base_result_default(self):
        result = BaseResult(
            service="api",
            verbose=False,
            output="text"
        )
        assert result.service == "api"
        assert result.key is None
        assert result.value is None
        assert result.config == {}
        assert result.verbose is False
        assert result.output == "text"
        assert result.success is False
        assert result.error is None
    
    def test_base_result_custom(self):
        result = BaseResult(
            service="view",
            key="TEST_KEY",
            value="test_value",
            config={"KEY1": "value1"},
            verbose=True,
            output="json",
            success=True,
            error="test error"
        )
        assert result.service == "view"
        assert result.key == "TEST_KEY"
        assert result.value == "test_value"
        assert result.config == {"KEY1": "value1"}
        assert result.verbose is True
        assert result.output == "json"
        assert result.success is True
        assert result.error == "test error"


class TestBaseService:
    def setup_method(self):
        self.config = BaseConfig()
        self.logger = Mock(spec=Logger)
        self.environment_service = Mock()
    
    def test_base_service_init(self):
        service = BaseService(self.config, self.logger, self.environment_service)
        assert service.config == self.config
        assert service.logger == self.logger
        assert service.environment_service == self.environment_service
        assert service.formatter is None
    
    def test_base_service_init_defaults(self):
        service = BaseService(self.config)
        assert service.config == self.config
        assert service.logger is not None
        assert service.environment_service is None
        assert service.formatter is None


class TestBaseAction:
    def setup_method(self):
        self.logger = Mock(spec=Logger)
    
    def test_base_action_init(self):
        action = BaseAction(self.logger)
        assert action.logger == self.logger
        assert action.formatter is None
    
    def test_base_action_init_default(self):
        action = BaseAction()
        assert action.logger is None
        assert action.formatter is None


def mock_open(read_data=""):
    """Helper function to create a mock open function"""
    from unittest.mock import mock_open as _mock_open
    return _mock_open(read_data=read_data) 
