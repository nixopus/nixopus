import unittest
from unittest.mock import patch
import yaml
import tempfile
import os
from app.commands.conflict.models import (
    ConflictConfig,
    ConflictCheckResult,
)
from app.commands.conflict.conflict import (
    ConflictService,
)
from app.utils.logger import Logger


class TestServiceIntegration(unittest.TestCase):
    """Test service integration and formatting"""

    def setUp(self):
        self.logger = Logger(verbose=False)
        self.config = ConflictConfig(
            config_file="test_config.yaml", verbose=False, output="text"
        )

    def test_conflict_service_integration(self):
        """Test ConflictService integration with YAML config"""
        config_data = {"deps": {"docker": {"version": "20.10.0"}, "go": {"version": "1.18.0"}, "python": {"version": "3.9.0"}}}

        with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
            yaml.dump(config_data, f)
            temp_path = f.name

        try:
            config = ConflictConfig(config_file=temp_path, verbose=False, output="text")

            service = ConflictService(config, self.logger)

            # This would normally make real system calls
            # In a real test, we'd mock all the checkers
            with patch.object(service.checker, "check_conflicts") as mock_check:
                mock_check.return_value = [
                    ConflictCheckResult(
                        tool="docker", expected="20.10.0", current="20.10.5", status="compatible", conflict=False
                    )
                ]

                results = service.check_conflicts()
                self.assertEqual(len(results), 1)
                self.assertFalse(results[0].conflict)
        finally:
            os.unlink(temp_path)

    def test_empty_deps_service_handling(self):
        """Test ConflictService handling of empty or missing deps section"""
        config_data = {}

        with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
            yaml.dump(config_data, f)
            temp_path = f.name

        try:
            config = ConflictConfig(config_file=temp_path, verbose=False, output="text")

            service = ConflictService(config, self.logger)
            results = service.check_conflicts()

            # Should return empty results for empty deps
            self.assertEqual(len(results), 0)
        finally:
            os.unlink(temp_path)

    def test_conflict_formatter_output(self):
        """Test ConflictFormatter output formatting"""
        from app.commands.conflict.conflict import ConflictFormatter

        formatter = ConflictFormatter()

        results = [
            ConflictCheckResult(tool="docker", expected="20.10.0", current="20.10.5", status="compatible", conflict=False),
            ConflictCheckResult(tool="python", expected="3.9.0", current="3.8.0", status="conflict", conflict=True),
        ]

        output = formatter.format_output(results, "text")

        # Should contain both tools
        self.assertIn("docker", output)
        self.assertIn("python", output)

        # Should indicate status
        self.assertIn("compatible", output)

    def test_conflict_formatter_json_output(self):
        """Test ConflictFormatter JSON output formatting"""
        from app.commands.conflict.conflict import ConflictFormatter

        formatter = ConflictFormatter()

        results = [
            ConflictCheckResult(tool="docker", expected="20.10.0", current="20.10.5", status="compatible", conflict=False)
        ]

        output = formatter.format_output(results, "json")

        # Should be valid JSON structure
        self.assertIn("docker", output)
        self.assertIn("compatible", output)
        self.assertIn("20.10.5", output)

    def test_service_with_multiple_tools(self):
        """Test ConflictService with multiple tool configurations"""
        config_data = {
            "deps": {
                "docker": {"version": "20.10.0"},
                "go": {"version": "1.18.0"},
                "python": {"version": "3.9.0"},
                "nodejs": {"version": "16.0.0"}
            }
        }

        with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
            yaml.dump(config_data, f)
            temp_path = f.name

        try:
            config = ConflictConfig(config_file=temp_path, verbose=False, output="text")

            service = ConflictService(config, self.logger)

            # Mock the checker to return mixed results
            with patch.object(service.checker, "check_conflicts") as mock_check:
                mock_check.return_value = [
                    ConflictCheckResult(tool="docker", expected="20.10.0", current="20.10.5", status="compatible", conflict=False),
                    ConflictCheckResult(tool="go", expected="1.18.0", current="1.17.0", status="conflict", conflict=True),
                    ConflictCheckResult(tool="python", expected="3.9.0", current="3.9.2", status="compatible", conflict=False),
                    ConflictCheckResult(tool="nodejs", expected="16.0.0", current=None, status="missing", conflict=True),
                ]

                results = service.check_conflicts()
                self.assertEqual(len(results), 4)
                
                # Check that we have both compatible and conflict results
                compatible_results = [r for r in results if not r.conflict]
                conflict_results = [r for r in results if r.conflict]
                
                self.assertEqual(len(compatible_results), 2)
                self.assertEqual(len(conflict_results), 2)
        finally:
            os.unlink(temp_path)


if __name__ == "__main__":
    unittest.main()
