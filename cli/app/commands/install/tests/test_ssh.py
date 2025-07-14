import unittest
from unittest.mock import Mock, patch, MagicMock
import tempfile
import os
from app.commands.install.ssh import SSH, SSHConfig, SSHKeyManager, SSHCommandBuilder

class TestSSHKeyGeneration(unittest.TestCase):
    def setUp(self):
        self.mock_logger = Mock()
        self.temp_dir = tempfile.mkdtemp()
        self.test_key_path = os.path.join(self.temp_dir, "test_key")
    
    def tearDown(self):
        import shutil
        shutil.rmtree(self.temp_dir)
    
    def test_ssh_command_builder_rsa(self):
        cmd = SSHCommandBuilder.build_ssh_keygen_command(
            self.test_key_path, "rsa", 4096, "testpass"
        )
        expected = ["ssh-keygen", "-t", "rsa", "-f", self.test_key_path, "-N", "-b", "4096", "testpass"]
        self.assertEqual(cmd, expected)
    
    def test_ssh_command_builder_ed25519_no_passphrase(self):
        cmd = SSHCommandBuilder.build_ssh_keygen_command(
            self.test_key_path, "ed25519", 256
        )
        expected = ["ssh-keygen", "-t", "ed25519", "-f", self.test_key_path, "-N", ""]
        self.assertEqual(cmd, expected)
    
    def test_ssh_command_builder_ecdsa(self):
        cmd = SSHCommandBuilder.build_ssh_keygen_command(
            self.test_key_path, "ecdsa", 256
        )
        expected = ["ssh-keygen", "-t", "ecdsa", "-f", self.test_key_path, "-N", "-b", "256", ""]
        self.assertEqual(cmd, expected)
    
    def test_ssh_command_builder_dsa(self):
        cmd = SSHCommandBuilder.build_ssh_keygen_command(
            self.test_key_path, "dsa", 1024
        )
        expected = ["ssh-keygen", "-t", "dsa", "-f", self.test_key_path, "-N", "-b", "1024", ""]
        self.assertEqual(cmd, expected)
    
    def test_ssh_config_validation_valid_key_type(self):
        config = SSHConfig(
            path=self.test_key_path,
            key_type="ed25519",
            key_size=256
        )
        self.assertEqual(config.key_type, "ed25519")
    
    def test_ssh_config_validation_invalid_key_type(self):
        with self.assertRaises(ValueError):
            SSHConfig(
                path=self.test_key_path,
                key_type="invalid_type",
                key_size=256
            )
    
    def test_ssh_config_validation_valid_key_size(self):
        config = SSHConfig(
            path=self.test_key_path,
            key_type="rsa",
            key_size=4096
        )
        self.assertEqual(config.key_size, 4096)
    
    def test_ssh_config_validation_invalid_key_size(self):
        with self.assertRaises(ValueError):
            SSHConfig(
                path=self.test_key_path,
                key_type="rsa",
                key_size=512
            )
    
    def test_ssh_config_ed25519_key_size_always_256(self):
        config = SSHConfig(
            path=self.test_key_path,
            key_type="ed25519",
            key_size=512 
        )
        self.assertEqual(config.key_size, 256)
    
    @patch('subprocess.run')
    def test_ssh_key_manager_availability_check_success(self, mock_run):
        mock_result = Mock()
        mock_result.returncode = 0
        mock_run.return_value = mock_result
        
        manager = SSHKeyManager(self.mock_logger)
        available, error = manager._check_ssh_keygen_availability()
        
        self.assertTrue(available)
        self.assertIsNone(error)
        mock_run.assert_called_once_with(["ssh-keygen", "-h"], capture_output=True, text=True, check=False)
    
    @patch('subprocess.run')
    def test_ssh_key_manager_availability_check_failure(self, mock_run):
        mock_result = Mock()
        mock_result.returncode = 1
        mock_run.return_value = mock_result
        
        manager = SSHKeyManager(self.mock_logger)
        available, error = manager._check_ssh_keygen_availability()
        
        self.assertFalse(available)
        self.assertIsNone(error)
    
    @patch('subprocess.run')
    def test_ssh_key_manager_version_check(self, mock_run):
        mock_result = Mock()
        mock_result.returncode = 0
        mock_result.stdout = "OpenSSH_8.9p1"
        mock_run.return_value = mock_result
        
        manager = SSHKeyManager(self.mock_logger)
        success, error = manager._check_ssh_keygen_version()
        
        self.assertTrue(success)
        self.assertIsNone(error)
        self.mock_logger.debug.assert_called_with("SSH keygen version: OpenSSH_8.9p1")
    
    @patch('subprocess.run')
    def test_ssh_key_manager_success(self, mock_run):
        mock_avail_result = Mock()
        mock_avail_result.returncode = 0
        
        mock_version_result = Mock()
        mock_version_result.returncode = 0
        mock_version_result.stdout = "OpenSSH_8.9p1"
        
        mock_gen_result = Mock()
        mock_gen_result.returncode = 0
        
        mock_run.side_effect = [mock_avail_result, mock_version_result, mock_gen_result]
        
        manager = SSHKeyManager(self.mock_logger)
        success, error = manager.generate_ssh_key(self.test_key_path, "ed25519", 256)
        
        self.assertTrue(success)
        self.assertIsNone(error)
        self.assertEqual(mock_run.call_count, 3)
    
    @patch('subprocess.run')
    def test_ssh_key_manager_failure(self, mock_run):
        from subprocess import CalledProcessError
        
        mock_avail_result = Mock()
        mock_avail_result.returncode = 0
        
        mock_version_result = Mock()
        mock_version_result.returncode = 0
        mock_run.side_effect = [mock_avail_result, mock_version_result, CalledProcessError(1, "ssh-keygen", stderr="Permission denied")]
        
        manager = SSHKeyManager(self.mock_logger)
        success, error = manager.generate_ssh_key(self.test_key_path, "ed25519", 256)
        
        self.assertFalse(success)
        self.assertEqual(error, "Permission denied")
    
    @patch('subprocess.run')
    def test_ssh_key_manager_availability_failure(self, mock_run):
        mock_result = Mock()
        mock_result.returncode = 1
        mock_run.return_value = mock_result
        
        manager = SSHKeyManager(self.mock_logger)
        success, error = manager.generate_ssh_key(self.test_key_path, "ed25519", 256)
        
        self.assertFalse(success)
        self.assertIsNone(error)
    
    def test_ssh_service_dry_run(self):
        config = SSHConfig(
            path=self.test_key_path,
            key_type="ed25519",
            key_size=256,
            dry_run=True
        )
        
        ssh = SSH(self.mock_logger)
        result = ssh.generate(config)
        
        self.assertTrue(result.success)
        self.assertIsNotNone(result.error)
        self.assertIn("DRY RUN MODE", result.error)
    
    @patch('subprocess.run')
    def test_ssh_service_force_overwrite(self, mock_run):
        from subprocess import CalledProcessError
        
        with open(self.test_key_path, 'w') as f:
            f.write("existing key")
        
        mock_avail_result = Mock()
        mock_avail_result.returncode = 0
        
        mock_version_result = Mock()
        mock_version_result.returncode = 0
        
        mock_run.side_effect = [mock_avail_result, mock_version_result, CalledProcessError(1, "ssh-keygen", stderr="ssh-keygen failed")]
        
        config = SSHConfig(
            path=self.test_key_path,
            key_type="ed25519",
            key_size=256,
            force=True
        )
        
        ssh = SSH(self.mock_logger)
        result = ssh.generate(config)
        
        self.assertFalse(result.success)
        self.assertIn("ssh-keygen", result.error)
    
    @patch('subprocess.run')
    def test_ssh_key_manager_with_permissions(self, mock_run):
        mock_result = Mock()
        mock_result.returncode = 0
        mock_run.return_value = mock_result
        
        manager = SSHKeyManager(self.mock_logger)
        
        with open(self.test_key_path, 'w') as f:
            f.write("private key content")
        
        with open(f"{self.test_key_path}.pub", 'w') as f:
            f.write("public key content")
        
        success, error = manager.set_key_permissions(self.test_key_path, f"{self.test_key_path}.pub")
        
        self.assertTrue(success)
        self.assertIsNone(error)
    
    def test_ssh_key_manager_create_ssh_directory(self):
        manager = SSHKeyManager(self.mock_logger)
        test_ssh_dir = os.path.join(self.temp_dir, "test_ssh")
        
        success, error = manager.create_ssh_directory(test_ssh_dir)
        
        self.assertTrue(success)
        self.assertIsNone(error)
        self.assertTrue(os.path.exists(test_ssh_dir))
    
    @patch('builtins.open', create=True)
    def test_ssh_key_manager_add_to_authorized_keys(self, mock_open):
        manager = SSHKeyManager(self.mock_logger)
        
        public_key_path = f"{self.test_key_path}.pub"
        with open(public_key_path, 'w') as f:
            f.write("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... test@example.com")
        
        success, error = manager.add_to_authorized_keys(public_key_path)
        
        self.assertTrue(success)
        self.assertIsNone(error)
    
    def test_ssh_config_with_new_options(self):
        config = SSHConfig(
            path=self.test_key_path,
            key_type="ed25519",
            key_size=256,
            set_permissions=True,
            add_to_authorized_keys=True,
            create_ssh_directory=True
        )
        
        self.assertTrue(config.set_permissions)
        self.assertTrue(config.add_to_authorized_keys)
        self.assertTrue(config.create_ssh_directory)
    
    def test_ssh_config_ed25519_key_size_validation(self):
        config = SSHConfig(
            path=self.test_key_path,
            key_type="ed25519",
            key_size=512
        )
        self.assertEqual(config.key_size, 256)
    
    def test_ssh_config_ecdsa_key_size_validation(self):
        valid_sizes = [256, 384, 521]
        for size in valid_sizes:
            config = SSHConfig(
                path=self.test_key_path,
                key_type="ecdsa",
                key_size=size
            )
            self.assertEqual(config.key_size, size)
        
        with self.assertRaises(ValueError):
            SSHConfig(
                path=self.test_key_path,
                key_type="ecdsa",
                key_size=512
            )

if __name__ == '__main__':
    unittest.main() 
