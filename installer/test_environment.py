import os
import pytest
from pathlib import Path
from unittest.mock import patch, MagicMock, mock_open
from .environment import EnvironmentConfig, EnvironmentSetup
from .docker_setup import DockerSetup

def test_environment_config_create():
    prod_config = EnvironmentConfig.create("production")
    assert prod_config.env == "production"
    assert prod_config.config_dir == Path("/etc/nixopus")
    assert prod_config.api_port == 8443
    assert prod_config.next_public_port == 7443
    assert prod_config.db_port == 5432
    assert prod_config.host_name == "nixopus-db"
    assert prod_config.redis_url == "redis://nixopus-redis:6379"
    assert prod_config.mount_path == "/etc/nixopus/configs"
    assert prod_config.docker_host == "tcp://{ip}:2376"
    assert prod_config.docker_port == 2376

    staging_config = EnvironmentConfig.create("staging")
    assert staging_config.env == "staging"
    assert staging_config.config_dir == Path("/etc/nixopus-staging")
    assert staging_config.api_port == 8444
    assert staging_config.next_public_port == 7444
    assert staging_config.db_port == 5433
    assert staging_config.host_name == "nixopus-staging-db"
    assert staging_config.redis_url == "redis://nixopus-staging-redis:6380"
    assert staging_config.mount_path == "/etc/nixopus-staging/configs"
    assert staging_config.docker_host == "tcp://{ip}:2377"
    assert staging_config.docker_port == 2377

def test_environment_config_create_invalid_env():
    with pytest.raises(ValueError) as exc_info:
        EnvironmentConfig.create("invalid_env")
    assert "Invalid environment" in str(exc_info.value)

def test_generate_random_string():
    env_setup = EnvironmentSetup(None)
    
    random_str = env_setup.generate_random_string()
    assert len(random_str) == 12
    assert all(c.isalnum() for c in random_str)
    
    random_str = env_setup.generate_random_string(length=20)
    assert len(random_str) == 20
    assert all(c.isalnum() for c in random_str)

def test_generate_random_string_invalid_length():
    env_setup = EnvironmentSetup(None)
    with pytest.raises(ValueError) as exc_info:
        env_setup.generate_random_string(length=0)
    assert "Length must be positive" in str(exc_info.value)

@patch('installer.environment.subprocess.run')
@patch('pathlib.Path.exists')
def test_generate_ssh_key(mock_exists, mock_run):
    mock_exists.return_value = False
    mock_run.return_value = MagicMock(returncode=0)
    
    env_setup = EnvironmentSetup(None)
    private_key_path, public_key_path = env_setup.generate_ssh_key()
    
    assert private_key_path.name == "id_rsa"
    assert public_key_path.name == "id_rsa.pub"
    assert private_key_path.parent == env_setup.ssh_dir
    assert public_key_path.parent == env_setup.ssh_dir
    
    mock_run.assert_called_once_with(
        ["ssh-keygen", "-t", "rsa", "-b", "4096", "-f", str(private_key_path), "-N", ""],
        capture_output=True,
        text=True
    )

@patch('installer.environment.subprocess.run')
@patch('pathlib.Path.exists')
def test_generate_ssh_key_failure(mock_exists, mock_run):
    mock_exists.return_value = False
    mock_run.return_value = MagicMock(returncode=1)
    
    env_setup = EnvironmentSetup(None)
    with pytest.raises(Exception) as exc_info:
        env_setup.generate_ssh_key()
    
    assert "Failed to generate SSH key" in str(exc_info.value)

@patch('installer.environment.subprocess.run')
@patch('pathlib.Path.exists')
def test_generate_ssh_key_command_not_found(mock_exists, mock_run):
    mock_exists.return_value = False
    mock_run.side_effect = FileNotFoundError("ssh-keygen not found")
    
    env_setup = EnvironmentSetup(None)
    with pytest.raises(Exception) as exc_info:
        env_setup.generate_ssh_key()
    
    assert "ssh-keygen not found" in str(exc_info.value)

@patch('pathlib.Path.exists')
@patch('pathlib.Path.mkdir')
@patch('builtins.open', new_callable=mock_open)
def test_setup_authorized_keys(mock_open, mock_mkdir, mock_exists):
    mock_exists.return_value = False
    
    env_setup = EnvironmentSetup(None)
    with patch.object(env_setup, 'generate_ssh_key') as mock_generate:
        mock_generate.return_value = (Path("/test/private"), Path("/test/public"))
        env_setup.setup_authorized_keys()
    
    mock_mkdir.assert_called_once()
    mock_open.assert_called()

@patch('pathlib.Path.exists')
@patch('pathlib.Path.mkdir')
@patch('builtins.open', new_callable=mock_open)
def test_setup_authorized_keys_existing_key(mock_open, mock_mkdir, mock_exists):
    mock_exists.return_value = True
    mock_file = MagicMock()
    mock_file.read.return_value = "existing_key"
    mock_open.return_value.__enter__.return_value = mock_file
    
    env_setup = EnvironmentSetup(None)
    with patch.object(env_setup, 'generate_ssh_key') as mock_generate:
        mock_generate.return_value = (Path("/test/private"), Path("/test/public"))
        env_setup.setup_authorized_keys()
    
    mock_mkdir.assert_called_once()
    mock_open.assert_called()

@patch('pathlib.Path.exists')
@patch('pathlib.Path.mkdir')
@patch('builtins.open', new_callable=mock_open)
def test_setup_authorized_keys_permission_error(mock_open, mock_mkdir, mock_exists):
    mock_exists.return_value = False
    mock_mkdir.side_effect = PermissionError("Permission denied")
    
    env_setup = EnvironmentSetup(None)
    with pytest.raises(Exception) as exc_info:
        env_setup.setup_authorized_keys()
    
    assert "Permission denied" in str(exc_info.value)

@patch('pathlib.Path.exists')
@patch('builtins.open', new_callable=mock_open)
def test_get_version(mock_open, mock_exists):
    mock_exists.return_value = True
    mock_file = MagicMock()
    mock_file.read.return_value = "1.0.0"
    mock_open.return_value.__enter__.return_value = mock_file
    
    env_setup = EnvironmentSetup(None)
    version = env_setup.get_version()
    assert version == "1.0.0"

@patch('pathlib.Path.exists')
def test_get_version_unknown(mock_exists):
    mock_exists.return_value = False
    
    env_setup = EnvironmentSetup(None)
    version = env_setup.get_version()
    assert version == "unknown"

@patch('pathlib.Path.exists')
@patch('builtins.open', new_callable=mock_open)
def test_get_version_file_error(mock_open, mock_exists):
    mock_exists.return_value = True
    mock_open.side_effect = IOError("File read error")
    
    env_setup = EnvironmentSetup(None)
    with pytest.raises(Exception) as exc_info:
        env_setup.get_version()
    
    assert "File read error" in str(exc_info.value)

@patch('installer.environment.DockerSetup')
@patch('pathlib.Path.mkdir')
@patch('pathlib.Path.chmod')
@patch('builtins.open', new_callable=mock_open)
@patch('installer.environment.subprocess.run')
def test_setup_environment_with_domains(mock_run, mock_open, mock_chmod, mock_mkdir, mock_docker_setup):
    mock_run.return_value = MagicMock(returncode=0)
    mock_docker = MagicMock()
    mock_docker.get_public_ip.return_value = "127.0.0.1"
    mock_docker.setup.return_value = "test-context"
    mock_docker.docker_certs_dir = Path("/test/certs")
    mock_docker_setup.return_value = mock_docker
    
    domains = {
        "api_domain": "api.example.com",
        "app_domain": "app.example.com"
    }
    
    env_setup = EnvironmentSetup(domains)
    env_vars = env_setup.setup_environment()
    
    assert env_vars["API_URL"] == "https://api.example.com/api"
    assert env_vars["WEBSOCKET_URL"] == "wss://api.example.com/ws"
    assert env_vars["WEBHOOK_URL"] == "https://api.example.com/api/v1/webhook"
    assert env_vars["ALLOWED_ORIGIN"] == "https://app.example.com"
    assert env_vars["DOCKER_HOST"] == "tcp://127.0.0.1:2377"
    assert env_vars["DOCKER_CONTEXT"] == "test-context"

@patch('installer.environment.DockerSetup')
@patch('pathlib.Path.mkdir')
@patch('pathlib.Path.chmod')
@patch('builtins.open', new_callable=mock_open)
@patch('installer.environment.subprocess.run')
def test_setup_environment_without_domains(mock_run, mock_open, mock_chmod, mock_mkdir, mock_docker_setup):
    mock_run.return_value = MagicMock(returncode=0)
    mock_docker = MagicMock()
    mock_docker.get_public_ip.return_value = "127.0.0.1"
    mock_docker.setup.return_value = "test-context"
    mock_docker.docker_certs_dir = Path("/test/certs")
    mock_docker_setup.return_value = mock_docker
    
    env_setup = EnvironmentSetup(None)
    env_vars = env_setup.setup_environment()
    
    assert env_vars["API_URL"] == "http://127.0.0.1:8444/api"
    assert env_vars["WEBSOCKET_URL"] == "ws://127.0.0.1:8444/ws"
    assert env_vars["WEBHOOK_URL"] == "http://127.0.0.1:8444/api/v1/webhook"
    assert env_vars["ALLOWED_ORIGIN"] == "http://127.0.0.1:7444"
    assert env_vars["DOCKER_HOST"] == "tcp://127.0.0.1:2377"
    assert env_vars["DOCKER_CONTEXT"] == "test-context"

@patch('installer.environment.DockerSetup')
@patch('pathlib.Path.mkdir')
@patch('pathlib.Path.chmod')
@patch('builtins.open', new_callable=mock_open)
def test_setup_environment_docker_setup_failure(mock_open, mock_chmod, mock_mkdir, mock_docker_setup):
    mock_docker = MagicMock()
    mock_docker.get_public_ip.side_effect = Exception("Docker setup failed")
    mock_docker_setup.return_value = mock_docker
    
    env_setup = EnvironmentSetup(None)
    with pytest.raises(Exception) as exc_info:
        env_setup.setup_environment()
    
    assert "Docker setup failed" in str(exc_info.value)

@patch('installer.environment.DockerSetup')
@patch('pathlib.Path.mkdir')
@patch('pathlib.Path.chmod')
@patch('builtins.open', new_callable=mock_open)
def test_setup_environment_file_write_error(mock_open, mock_chmod, mock_mkdir, mock_docker_setup):
    mock_docker = MagicMock()
    mock_docker.get_public_ip.return_value = "127.0.0.1"
    mock_docker.setup.return_value = "test-context"
    mock_docker.docker_certs_dir = Path("/test/certs")
    mock_docker_setup.return_value = mock_docker
    
    mock_open.side_effect = IOError("File write error")
    
    env_setup = EnvironmentSetup(None)
    with pytest.raises(Exception) as exc_info:
        env_setup.setup_environment()
    
    assert "File write error" in str(exc_info.value)

@patch('installer.environment.DockerSetup')
@patch('pathlib.Path.mkdir')
@patch('pathlib.Path.chmod')
@patch('builtins.open', new_callable=mock_open)
@patch('installer.environment.subprocess.run')
def test_setup_environment_invalid_domains(mock_run, mock_open, mock_chmod, mock_mkdir, mock_docker_setup):
    mock_run.return_value = MagicMock(returncode=0)
    mock_docker = MagicMock()
    mock_docker.get_public_ip.return_value = "127.0.0.1"
    mock_docker.setup.return_value = "test-context"
    mock_docker.docker_certs_dir = Path("/test/certs")
    mock_docker_setup.return_value = mock_docker
    
    domains = {
        "api_domain": "invalid domain",  # Invalid domain format
        "app_domain": "app.example.com"
    }
    
    env_setup = EnvironmentSetup(domains)
    with pytest.raises(ValueError) as exc_info:
        env_setup.setup_environment()
    
    assert "Invalid domain format" in str(exc_info.value) 