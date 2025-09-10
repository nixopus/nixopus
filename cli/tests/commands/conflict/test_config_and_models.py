import os
import tempfile
import unittest

import yaml

from app.commands.conflict.conflict import (
    ConflictChecker,
)
from app.commands.conflict.models import (
    ConflictCheckResult,
    ConflictConfig,
)
from app.utils.logger import Logger


class TestConfigAndModels(unittest.TestCase):
    """Test configuration loading and data models"""

    def setUp(self):
        self.logger = Logger(verbose=False)
        self.config = ConflictConfig(config_file="test_config.yaml", verbose=False, output="text")

    def test_conflict_check_result_creation(self):
        """Test ConflictCheckResult model creation"""
        result = ConflictCheckResult(tool="docker", expected="20.10.0", current="20.10.5", status="compatible", conflict=False)

        self.assertEqual(result.tool, "docker")
        self.assertEqual(result.expected, "20.10.0")
        self.assertEqual(result.current, "20.10.5")
        self.assertFalse(result.conflict)

    def test_conflict_checker_config_loading(self):
        """Test ConflictChecker config loading with valid YAML config"""
        config_data = {"deps": {"docker": {"version": "20.10.0"}, "go": {"version": "1.18.0"}, "python": {"version": "3.9.0"}}}

        with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
            yaml.dump(config_data, f)
            temp_path = f.name

        try:
            conflict_config = ConflictConfig(config_file=temp_path, verbose=False, output="text")

            # Create ConflictChecker which will load the config internally
            checker = ConflictChecker(conflict_config, self.logger)

            # Test that the config was loaded correctly by checking internal state
            # We can verify this by calling _load_user_config directly
            result = checker._load_user_config(temp_path)

            self.assertEqual(result, config_data)
            self.assertIn("deps", result)
            self.assertIn("docker", result["deps"])
            self.assertEqual(result["deps"]["docker"]["version"], "20.10.0")
        finally:
            os.unlink(temp_path)

    def test_config_loading_missing_file(self):
        """Test ConflictChecker config loading with missing file"""
        conflict_config = ConflictConfig(config_file="nonexistent.yaml", verbose=False, output="text")

        # ConflictChecker initialization should fail with missing config file
        with self.assertRaises(FileNotFoundError):
            ConflictChecker(conflict_config, self.logger)

    def test_config_loading_invalid_yaml(self):
        """Test ConflictChecker config loading with invalid YAML"""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
            f.write("invalid: yaml: content: [")
            temp_path = f.name

        try:
            conflict_config = ConflictConfig(config_file=temp_path, verbose=False, output="text")

            # ConflictChecker initialization should fail with invalid YAML
            with self.assertRaises(Exception):
                ConflictChecker(conflict_config, self.logger)
        finally:
            os.unlink(temp_path)

    def test_empty_deps_handling(self):
        """Test handling of empty or missing deps section"""
        config_data = {}

        with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
            yaml.dump(config_data, f)
            temp_path = f.name

        try:
            config = ConflictConfig(config_file=temp_path, verbose=False, output="text")

            # Test that config is created successfully even with empty deps
            self.assertEqual(config.config_file, temp_path)
            self.assertFalse(config.verbose)
            self.assertEqual(config.output, "text")
        finally:
            os.unlink(temp_path)


if __name__ == "__main__":
    unittest.main()
