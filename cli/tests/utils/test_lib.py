import os
import platform
import shutil
import stat
import tempfile
import unittest
from unittest.mock import Mock, patch, mock_open
import requests

from app.utils.lib import (
    SupportedOS,
    SupportedDistribution,
    SupportedPackageManager,
    Supported,
    HostInformation,
    ParallelProcessor,
    DirectoryManager,
    FileManager,
)
from app.utils.message import (
    FAILED_TO_GET_PUBLIC_IP_MESSAGE,
    FAILED_TO_REMOVE_DIRECTORY_MESSAGE,
    REMOVED_DIRECTORY_MESSAGE,
)


class TestSupportedOS(unittest.TestCase):
    def test_supported_os_values(self):
        self.assertEqual(SupportedOS.LINUX.value, "linux")
        self.assertEqual(SupportedOS.MACOS.value, "darwin")


class TestSupportedDistribution(unittest.TestCase):
    def test_supported_distribution_values(self):
        self.assertEqual(SupportedDistribution.DEBIAN.value, "debian")
        self.assertEqual(SupportedDistribution.UBUNTU.value, "ubuntu")
        self.assertEqual(SupportedDistribution.CENTOS.value, "centos")
        self.assertEqual(SupportedDistribution.FEDORA.value, "fedora")
        self.assertEqual(SupportedDistribution.ALPINE.value, "alpine")


class TestSupportedPackageManager(unittest.TestCase):
    def test_supported_package_manager_values(self):
        self.assertEqual(SupportedPackageManager.APT.value, "apt")
        self.assertEqual(SupportedPackageManager.YUM.value, "yum")
        self.assertEqual(SupportedPackageManager.DNF.value, "dnf")
        self.assertEqual(SupportedPackageManager.PACMAN.value, "pacman")
        self.assertEqual(SupportedPackageManager.APK.value, "apk")
        self.assertEqual(SupportedPackageManager.BREW.value, "brew")


class TestSupported(unittest.TestCase):
    def test_os_supported(self):
        self.assertTrue(Supported.os("linux"))
        self.assertTrue(Supported.os("darwin"))

    def test_os_not_supported(self):
        self.assertFalse(Supported.os("windows"))
        self.assertFalse(Supported.os("freebsd"))
        self.assertFalse(Supported.os(""))

    def test_os_case_sensitive(self):
        self.assertFalse(Supported.os("Linux"))
        self.assertFalse(Supported.os("DARWIN"))

    def test_distribution_supported(self):
        self.assertTrue(Supported.distribution("debian"))
        self.assertTrue(Supported.distribution("ubuntu"))
        self.assertTrue(Supported.distribution("centos"))

    def test_distribution_not_supported(self):
        self.assertFalse(Supported.distribution("arch"))
        self.assertFalse(Supported.distribution("gentoo"))
        self.assertFalse(Supported.distribution(""))

    def test_package_manager_supported(self):
        self.assertTrue(Supported.package_manager("apt"))
        self.assertTrue(Supported.package_manager("yum"))
        self.assertTrue(Supported.package_manager("brew"))

    def test_package_manager_not_supported(self):
        self.assertFalse(Supported.package_manager("pip"))
        self.assertFalse(Supported.package_manager("npm"))
        self.assertFalse(Supported.package_manager(""))

    def test_get_os(self):
        os_list = Supported.get_os()
        self.assertIsInstance(os_list, list)
        self.assertIn("linux", os_list)
        self.assertIn("darwin", os_list)
        self.assertEqual(len(os_list), 2)

    def test_get_distributions(self):
        dist_list = Supported.get_distributions()
        self.assertIsInstance(dist_list, list)
        self.assertIn("debian", dist_list)
        self.assertIn("ubuntu", dist_list)
        self.assertIn("centos", dist_list)
        self.assertIn("fedora", dist_list)
        self.assertIn("alpine", dist_list)
        self.assertEqual(len(dist_list), 5)


class TestHostInformation(unittest.TestCase):
    @patch("platform.system")
    def test_get_os_name(self, mock_system):
        mock_system.return_value = "Linux"
        self.assertEqual(HostInformation.get_os_name(), "linux")

        mock_system.return_value = "Darwin"
        self.assertEqual(HostInformation.get_os_name(), "darwin")

        mock_system.return_value = "Windows"
        self.assertEqual(HostInformation.get_os_name(), "windows")

    @patch("app.utils.lib.HostInformation.get_os_name")
    @patch("app.utils.lib.HostInformation.command_exists")
    def test_get_package_manager_macos(self, mock_command_exists, mock_get_os_name):
        mock_get_os_name.return_value = "darwin"
        mock_command_exists.return_value = True

        result = HostInformation.get_package_manager()
        self.assertEqual(result, "brew")

    @patch("app.utils.lib.HostInformation.get_os_name")
    @patch("app.utils.lib.HostInformation.command_exists")
    def test_get_package_manager_linux_apt(self, mock_command_exists, mock_get_os_name):
        mock_get_os_name.return_value = "linux"
        
        def command_exists_side_effect(command):
            return command == "apt"

        mock_command_exists.side_effect = command_exists_side_effect

        result = HostInformation.get_package_manager()
        self.assertEqual(result, "apt")

    @patch("app.utils.lib.HostInformation.get_os_name")
    @patch("app.utils.lib.HostInformation.command_exists")
    def test_get_package_manager_linux_yum(self, mock_command_exists, mock_get_os_name):
        mock_get_os_name.return_value = "linux"
        
        def command_exists_side_effect(command):
            return command == "yum"

        mock_command_exists.side_effect = command_exists_side_effect

        result = HostInformation.get_package_manager()
        self.assertEqual(result, "yum")

    @patch("app.utils.lib.HostInformation.get_os_name")
    @patch("app.utils.lib.HostInformation.command_exists")
    def test_get_package_manager_no_supported_manager(self, mock_command_exists, mock_get_os_name):
        mock_get_os_name.return_value = "linux"
        mock_command_exists.return_value = False

        with self.assertRaises(RuntimeError) as context:
            HostInformation.get_package_manager()
        
        self.assertIn("No supported package manager found", str(context.exception))

    @patch("shutil.which")
    def test_command_exists_true(self, mock_which):
        mock_which.return_value = "/usr/bin/apt"
        self.assertTrue(HostInformation.command_exists("apt"))

    @patch("shutil.which")
    def test_command_exists_false(self, mock_which):
        mock_which.return_value = None
        self.assertFalse(HostInformation.command_exists("nonexistent"))

    @patch("requests.get")
    def test_get_public_ip_success(self, mock_get):
        mock_response = Mock()
        mock_response.text = "192.168.1.1"
        mock_response.raise_for_status.return_value = None
        mock_get.return_value = mock_response

        result = HostInformation.get_public_ip()
        self.assertEqual(result, "192.168.1.1")
        mock_get.assert_called_once_with("https://api.ipify.org", timeout=10)

    @patch("requests.get")
    def test_get_public_ip_http_error(self, mock_get):
        mock_get.side_effect = requests.HTTPError("404 Not Found")

        with self.assertRaises(Exception) as context:
            HostInformation.get_public_ip()
        
        self.assertEqual(str(context.exception), FAILED_TO_GET_PUBLIC_IP_MESSAGE)

    @patch("requests.get")
    def test_get_public_ip_connection_error(self, mock_get):
        mock_get.side_effect = requests.ConnectionError("Connection failed")

        with self.assertRaises(Exception) as context:
            HostInformation.get_public_ip()
        
        self.assertEqual(str(context.exception), FAILED_TO_GET_PUBLIC_IP_MESSAGE)

    @patch("requests.get")
    def test_get_public_ip_timeout(self, mock_get):
        mock_get.side_effect = requests.Timeout("Request timeout")

        with self.assertRaises(Exception) as context:
            HostInformation.get_public_ip()
        
        self.assertEqual(str(context.exception), FAILED_TO_GET_PUBLIC_IP_MESSAGE)


class TestParallelProcessor(unittest.TestCase):
    def test_process_items_empty_list(self):
        def processor(x):
            return x * 2

        results = ParallelProcessor.process_items([], processor)
        self.assertEqual(results, [])

    def test_process_items_single_item(self):
        def processor(x):
            return x * 2

        results = ParallelProcessor.process_items([5], processor)
        self.assertEqual(results, [10])

    def test_process_items_multiple_items(self):
        def processor(x):
            return x * 2

        results = ParallelProcessor.process_items([1, 2, 3, 4, 5], processor)
        self.assertEqual(len(results), 5)
        self.assertEqual(set(results), {2, 4, 6, 8, 10})

    def test_process_items_with_error_handler(self):
        def processor(x):
            if x == 3:
                raise ValueError("Test error")
            return x * 2

        def error_handler(item, error):
            return f"Error processing {item}: {str(error)}"

        results = ParallelProcessor.process_items([1, 2, 3, 4, 5], processor, error_handler=error_handler)
        self.assertEqual(len(results), 5)
        
        error_results = [r for r in results if "Error processing 3" in str(r)]
        normal_results = [r for r in results if isinstance(r, int)]
        
        self.assertEqual(len(error_results), 1)
        self.assertEqual(set(normal_results), {2, 4, 8, 10})

    def test_process_items_without_error_handler(self):
        def processor(x):
            if x == 3:
                raise ValueError("Test error")
            return x * 2

        results = ParallelProcessor.process_items([1, 2, 3, 4, 5], processor)
        self.assertEqual(len(results), 4)
        self.assertEqual(set(results), {2, 4, 8, 10})

    def test_process_items_max_workers_limit(self):
        def processor(x):
            return x * 2

        results = ParallelProcessor.process_items([1, 2, 3, 4, 5], processor, max_workers=2)
        self.assertEqual(len(results), 5)
        self.assertEqual(set(results), {2, 4, 6, 8, 10})

    def test_process_items_max_workers_exceeds_items(self):
        def processor(x):
            return x * 2

        results = ParallelProcessor.process_items([1, 2], processor, max_workers=10)
        self.assertEqual(len(results), 2)
        self.assertEqual(set(results), {2, 4})


class TestDirectoryManager(unittest.TestCase):
    @patch("os.path.exists")
    def test_path_exists_true(self, mock_exists):
        mock_exists.return_value = True
        self.assertTrue(DirectoryManager.path_exists("/test/path"))

    @patch("os.path.exists")
    def test_path_exists_false(self, mock_exists):
        mock_exists.return_value = False
        self.assertFalse(DirectoryManager.path_exists("/test/path"))

    @patch("os.path.exists")
    def test_path_exists_and_not_force_true(self, mock_exists):
        mock_exists.return_value = True
        self.assertTrue(DirectoryManager.path_exists_and_not_force("/test/path", False))

    @patch("os.path.exists")
    def test_path_exists_and_not_force_false_when_force(self, mock_exists):
        mock_exists.return_value = True
        self.assertFalse(DirectoryManager.path_exists_and_not_force("/test/path", True))

    @patch("os.path.exists")
    def test_path_exists_and_not_force_false_when_not_exists(self, mock_exists):
        mock_exists.return_value = False
        self.assertFalse(DirectoryManager.path_exists_and_not_force("/test/path", False))

    @patch("shutil.rmtree")
    @patch("os.path.exists")
    @patch("os.path.isdir")
    def test_remove_directory_success(self, mock_isdir, mock_exists, mock_rmtree):
        mock_exists.return_value = True
        mock_isdir.return_value = True
        mock_logger = Mock()

        result = DirectoryManager.remove_directory("/test/path", mock_logger)
        
        self.assertTrue(result)
        mock_rmtree.assert_called_once_with("/test/path")
        mock_logger.debug.assert_called()

    @patch("shutil.rmtree")
    @patch("os.path.exists")
    def test_remove_directory_success_no_logger(self, mock_exists, mock_rmtree):
        mock_exists.return_value = True

        result = DirectoryManager.remove_directory("/test/path")
        
        self.assertTrue(result)
        mock_rmtree.assert_called_once_with("/test/path")

    @patch("shutil.rmtree")
    @patch("os.path.exists")
    def test_remove_directory_failure(self, mock_exists, mock_rmtree):
        mock_exists.return_value = True
        mock_rmtree.side_effect = PermissionError("Permission denied")
        mock_logger = Mock()

        result = DirectoryManager.remove_directory("/test/path", mock_logger)
        
        self.assertFalse(result)
        mock_logger.debug.assert_called()
        mock_logger.error.assert_called_once()

    @patch("shutil.rmtree")
    @patch("os.path.exists")
    def test_remove_directory_failure_no_logger(self, mock_exists, mock_rmtree):
        mock_exists.return_value = True
        mock_rmtree.side_effect = OSError("Directory not found")

        result = DirectoryManager.remove_directory("/test/path")
        
        self.assertFalse(result)


class TestFileManager(unittest.TestCase):
    def setUp(self):
        self.temp_dir = tempfile.mkdtemp()
        self.test_file = os.path.join(self.temp_dir, "test.txt")

    def tearDown(self):
        shutil.rmtree(self.temp_dir, ignore_errors=True)

    @patch("os.chmod")
    def test_set_permissions_success(self, mock_chmod):
        mock_logger = Mock()
        
        with open(self.test_file, "w") as f:
            f.write("test content")

        success, error = FileManager.set_permissions(self.test_file, 0o644, mock_logger)
        
        self.assertTrue(success)
        self.assertIsNone(error)
        mock_chmod.assert_called_once_with(self.test_file, 0o644)
        mock_logger.debug.assert_called()

    @patch("os.chmod")
    def test_set_permissions_failure(self, mock_chmod):
        mock_chmod.side_effect = PermissionError("Permission denied")
        mock_logger = Mock()

        success, error = FileManager.set_permissions(self.test_file, 0o644, mock_logger)
        
        self.assertFalse(success)
        self.assertIn("Failed to set permissions", error)
        mock_logger.error.assert_called_once()

    @patch("os.chmod")
    def test_set_permissions_success_no_logger(self, mock_chmod):
        with open(self.test_file, "w") as f:
            f.write("test content")

        success, error = FileManager.set_permissions(self.test_file, 0o644)
        
        self.assertTrue(success)
        self.assertIsNone(error)
        mock_chmod.assert_called_once_with(self.test_file, 0o644)

    @patch("os.makedirs")
    def test_create_directory_success_new(self, mock_makedirs):
        mock_logger = Mock()
        test_dir = os.path.join(self.temp_dir, "new_dir")

        success, error = FileManager.create_directory(test_dir, 0o755, mock_logger)
        
        self.assertTrue(success)
        self.assertIsNone(error)
        mock_makedirs.assert_called_once_with(test_dir, mode=0o755)
        mock_logger.debug.assert_called_once()

    @patch("os.makedirs")
    @patch("os.path.exists")
    def test_create_directory_success_exists(self, mock_exists, mock_makedirs):
        mock_logger = Mock()
        test_dir = os.path.join(self.temp_dir, "existing_dir")
        mock_exists.return_value = True

        success, error = FileManager.create_directory(test_dir, 0o755, mock_logger)
        
        self.assertTrue(success)
        self.assertIsNone(error)
        mock_makedirs.assert_not_called()

    @patch("os.makedirs")
    def test_create_directory_failure(self, mock_makedirs):
        mock_makedirs.side_effect = PermissionError("Permission denied")
        mock_logger = Mock()
        test_dir = "/root/restricted_dir"

        success, error = FileManager.create_directory(test_dir, 0o755, mock_logger)
        
        self.assertFalse(success)
        self.assertIn("Failed to create directory", error)
        mock_logger.error.assert_called_once()

    def test_append_to_file_success(self):
        mock_logger = Mock()
        content = "new content"

        success, error = FileManager.append_to_file(self.test_file, content, 0o644, mock_logger)
        
        self.assertTrue(success)
        self.assertIsNone(error)
        
        with open(self.test_file, "r") as f:
            file_content = f.read()
        
        self.assertIn(content, file_content)
        mock_logger.debug.assert_called()

    def test_append_to_file_failure_permission(self):
        mock_logger = Mock()
        content = "new content"
        
        with patch("builtins.open", side_effect=PermissionError("Permission denied")):
            success, error = FileManager.append_to_file(self.test_file, content, 0o644, mock_logger)
        
        self.assertFalse(success)
        self.assertIn("Failed to append to", error)
        mock_logger.error.assert_called_once()

    def test_read_file_content_success(self):
        content = "test content"
        with open(self.test_file, "w") as f:
            f.write(content)

        success, file_content, error = FileManager.read_file_content(self.test_file)
        
        self.assertTrue(success)
        self.assertEqual(file_content, content)
        self.assertIsNone(error)

    def test_read_file_content_failure(self):
        mock_logger = Mock()
        
        with patch("builtins.open", side_effect=FileNotFoundError("File not found")):
            success, file_content, error = FileManager.read_file_content(self.test_file, mock_logger)
        
        self.assertFalse(success)
        self.assertIsNone(file_content)
        self.assertIn("Failed to read", error)
        mock_logger.error.assert_called_once()

    def test_read_file_content_strips_whitespace(self):
        content = "  test content  \n"
        with open(self.test_file, "w") as f:
            f.write(content)

        success, file_content, error = FileManager.read_file_content(self.test_file)
        
        self.assertTrue(success)
        self.assertEqual(file_content, "test content")
        self.assertIsNone(error)

    @patch("os.path.expanduser")
    def test_expand_user_path(self, mock_expanduser):
        mock_expanduser.return_value = "/home/user/test"
        
        result = FileManager.expand_user_path("~/test")
        
        self.assertEqual(result, "/home/user/test")
        mock_expanduser.assert_called_once_with("~/test")

    @patch("os.path.dirname")
    def test_get_directory_path(self, mock_dirname):
        mock_dirname.return_value = "/path/to"
        
        result = FileManager.get_directory_path("/path/to/file.txt")
        
        self.assertEqual(result, "/path/to")
        mock_dirname.assert_called_once_with("/path/to/file.txt")

    def test_get_public_key_path(self):
        private_key_path = "/path/to/id_rsa"
        expected_public_key_path = "/path/to/id_rsa.pub"
        
        result = FileManager.get_public_key_path(private_key_path)
        
        self.assertEqual(result, expected_public_key_path)

    def test_get_public_key_path_empty_string(self):
        result = FileManager.get_public_key_path("")
        self.assertEqual(result, ".pub")

    def test_get_public_key_path_with_spaces(self):
        private_key_path = "/path with spaces/id_rsa"
        expected_public_key_path = "/path with spaces/id_rsa.pub"
        
        result = FileManager.get_public_key_path(private_key_path)
        
        self.assertEqual(result, expected_public_key_path)


if __name__ == "__main__":
    unittest.main() 