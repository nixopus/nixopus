import unittest
from unittest.mock import Mock, patch, call
import subprocess
from app.commands.conflict.models import ConflictConfig
from app.commands.conflict.conflict import (
    ToolVersionChecker,
    ConflictChecker,
)
from app.utils.logger import Logger


class TestVersionChecker(unittest.TestCase):
    """Test version checking and comparison logic"""

    def setUp(self):
        self.logger = Logger(verbose=False)
        self.config = ConflictConfig(
            config_file="test_config.yaml", verbose=False, output="text"
        )

    @patch("subprocess.run")
    def test_tool_version_checker_successful(self, mock_run):
        """Test ToolVersionChecker with successful version check"""
        mock_result = Mock()
        mock_result.returncode = 0
        mock_result.stdout = "Docker version 20.10.5, build 55c4c88"
        mock_run.return_value = mock_result

        checker = ToolVersionChecker(self.logger, timeout=5)
        version = checker.get_tool_version("docker")

        self.assertEqual(version, "20.10.5")
        mock_run.assert_called_once_with(["docker", "--version"], capture_output=True, text=True, timeout=5)

    @patch("subprocess.run")
    def test_tool_version_checker_not_found(self, mock_run):
        """Test ToolVersionChecker with tool not found"""
        mock_result = Mock()
        mock_result.returncode = 1
        mock_result.stdout = ""
        mock_run.return_value = mock_result

        checker = ToolVersionChecker(self.logger, timeout=5)
        version = checker.get_tool_version("nonexistent")

        self.assertIsNone(version)

    @patch("subprocess.run")
    def test_tool_version_checker_timeout(self, mock_run):
        """Test ToolVersionChecker with timeout"""
        mock_run.side_effect = subprocess.TimeoutExpired("cmd", 5)

        checker = ToolVersionChecker(self.logger, timeout=5)
        version = checker.get_tool_version("slow_tool")

        self.assertIsNone(version)

    @patch("app.commands.conflict.conflict.ConflictChecker._load_user_config")
    def test_tool_mapping(self, mock_load_config):
        """Test tool name mapping for system commands"""
        # Provide a dummy config for ConflictChecker
        mock_load_config.return_value = {"deps": {"docker": {"version": "20.10.0"}, "go": {"version": "1.18.0"}, "python": {"version": "3.9.0"}}}
        deps = {"docker": {"version": "20.10.0"}, "go": {"version": "1.18.0"}, "python": {"version": "3.9.0"}}

        conflict_checker = ConflictChecker(self.config, self.logger)

        # Mock the version checker to simulate tool responses
        with patch.object(conflict_checker.version_checker, "get_tool_version") as mock_get_version:
            mock_get_version.return_value = "20.10.5"

            results = conflict_checker._check_version_conflicts(deps)

            # Should have called get_tool_version for each tool
            self.assertEqual(mock_get_version.call_count, 3)
            # Check that we got results for all tools
            self.assertEqual(len(results), 3)
            # Check that the results have the expected structure
            for result in results:
                self.assertIn(result.tool, ["docker", "go", "python"])
                self.assertIsNotNone(result.current)
                self.assertIsNotNone(result.expected)
                self.assertIsInstance(result.conflict, bool)

    def test_version_requirement_none_or_empty(self):
        """Test handling of tools with no version requirements"""
        import yaml
        import tempfile
        import os
        
        config_data = {"deps": {"docker": {"version": ""}, "git": {"version": None}, "python": {}}}  # No version key

        with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
            yaml.dump(config_data, f)
            temp_path = f.name

        try:
            config = ConflictConfig(config_file=temp_path, verbose=False, output="text")

            checker = ConflictChecker(config, self.logger)

            # Mock version checker to return versions
            with patch.object(checker.version_checker, "get_tool_version") as mock_get_version:
                mock_get_version.return_value = "1.0.0"

                results = checker._check_version_conflicts(config_data["deps"])

                # Only docker and git should be checked (they have version keys)
                # python should not be checked (no version key)
                self.assertEqual(len(results), 2)

                # All should be compatible (no version requirement)
                for result in results:
                    self.assertFalse(result.conflict)
                    self.assertEqual(result.expected, "present")
        finally:
            os.unlink(temp_path)

    def test_tool_version_check_integration(self):
        """Test the integration of tool version checking"""
        checker = ToolVersionChecker(self.logger, timeout=5)
        
        # Test that the tool version checking works with mocked subprocess
        with patch("subprocess.run") as mock_run:
            mock_result = Mock()
            mock_result.returncode = 0
            mock_result.stdout = "Test version 1.0.0"
            mock_run.return_value = mock_result
            
            version = checker.get_tool_version("test_tool")
            
            # Should extract version from output
            self.assertEqual(version, "1.0.0")

    def test_version_commands_mapping(self):
        """Test that different tools use correct version commands"""
        deps_config = {
            "docker": {"version-command": ["docker", "--version"]},
            "go": {"version-command": ["go", "version"]},
            "ssh": {"version-command": ["ssh", "-V"]},
        }
        checker = ToolVersionChecker(self.logger, deps_config, timeout=5)
        with patch("subprocess.run") as mock_run:
            mock_result = Mock()
            mock_result.returncode = 0
            mock_result.stdout = "version 1.0.0"
            mock_run.return_value = mock_result
            # Test Docker uses correct command
            checker.get_tool_version("docker")
            mock_run.assert_called_with(["docker", "--version"], capture_output=True, text=True, timeout=5)
            # Test Go uses correct command
            checker.get_tool_version("go")
            mock_run.assert_called_with(["go", "version"], capture_output=True, text=True, timeout=5)
            # Test SSH uses correct command
            checker.get_tool_version("ssh")
            mock_run.assert_called_with(["ssh", "-V"], capture_output=True, text=True, timeout=5)


if __name__ == "__main__":
    unittest.main()
