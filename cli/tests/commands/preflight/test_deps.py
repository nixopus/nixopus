import json
import subprocess
import unittest
from typing import List
from unittest.mock import MagicMock, Mock, patch

from app.commands.preflight.deps import (
    DependencyChecker,
    DependencyFormatter,
    DependencyValidator,
    Deps,
    DepsCheckResult,
    DepsConfig,
    DepsService,
)
from app.utils.lib import Supported
from app.utils.logger import Logger
from app.utils.protocols import LoggerProtocol


class MockLogger:
    def __init__(self):
        self.debug_calls = []
        self.error_calls = []
        self.info_calls = []
        self.warning_calls = []
        self.success_calls = []
        self.highlight_calls = []
        self.verbose = True

    def debug(self, message: str) -> None:
        self.debug_calls.append(message)

    def error(self, message: str) -> None:
        self.error_calls.append(message)

    def info(self, message: str) -> None:
        self.info_calls.append(message)

    def warning(self, message: str) -> None:
        self.warning_calls.append(message)

    def success(self, message: str) -> None:
        self.success_calls.append(message)

    def highlight(self, message: str) -> None:
        self.highlight_calls.append(message)


class TestDependencyChecker(unittest.TestCase):

    def setUp(self):
        self.mock_logger = MockLogger()
        self.checker = DependencyChecker(logger=self.mock_logger)

    @patch("shutil.which")
    def test_check_dependency_available(self, mock_which):
        mock_which.return_value = "/usr/bin/docker"

        result = self.checker.check_dependency("docker")

        self.assertTrue(result)
        mock_which.assert_called_once_with("docker")
        self.assertEqual(len(self.mock_logger.debug_calls), 1)
        self.assertIn("docker", self.mock_logger.debug_calls[0])

    @patch("shutil.which")
    def test_check_dependency_not_available(self, mock_which):
        mock_which.return_value = None

        result = self.checker.check_dependency("nonexistent")

        self.assertFalse(result)
        mock_which.assert_called_once_with("nonexistent")

    @patch("shutil.which")
    def test_check_dependency_timeout(self, mock_which):
        mock_which.side_effect = subprocess.TimeoutExpired("command", 5)

        result = self.checker.check_dependency("slow_command")

        self.assertFalse(result)
        self.assertEqual(len(self.mock_logger.error_calls), 1)
        self.assertIn("slow_command", self.mock_logger.error_calls[0])

    @patch("shutil.which")
    def test_check_dependency_exception(self, mock_which):
        mock_which.side_effect = Exception("Test exception")

        result = self.checker.check_dependency("failing_command")

        self.assertFalse(result)
        self.assertEqual(len(self.mock_logger.error_calls), 1)
        self.assertIn("failing_command", self.mock_logger.error_calls[0])


class TestDependencyValidator(unittest.TestCase):

    def setUp(self):
        self.validator = DependencyValidator()

    def test_validate_os_valid(self):
        result = self.validator.validate_os("linux")
        self.assertEqual(result, "linux")

        result = self.validator.validate_os("darwin")
        self.assertEqual(result, "darwin")

    def test_validate_os_invalid(self):
        with self.assertRaises(ValueError) as context:
            self.validator.validate_os("windows")

        self.assertIn("windows", str(context.exception))

    def test_validate_package_manager_valid(self):
        valid_managers = ["apt", "yum", "dnf", "pacman", "apk", "brew"]
        for manager in valid_managers:
            result = self.validator.validate_package_manager(manager)
            self.assertEqual(result, manager)

    def test_validate_package_manager_invalid(self):
        with self.assertRaises(ValueError) as context:
            self.validator.validate_package_manager("invalid_manager")

        self.assertIn("invalid_manager", str(context.exception))


class TestDependencyFormatter(unittest.TestCase):

    def setUp(self):
        self.formatter = DependencyFormatter()
        self.sample_results = [
            DepsCheckResult(
                dependency="docker",
                verbose=False,
                output="text",
                os="linux",
                package_manager="apt",
                is_available=True,
            ),
            DepsCheckResult(
                dependency="kubectl",
                verbose=False,
                output="text",
                os="linux",
                package_manager="apt",
                is_available=False,
            ),
        ]

    def test_format_output_text(self):
        result = self.formatter.format_output(self.sample_results, "text")
        self.assertIn("docker", result)
        self.assertIn("kubectl", result)
        self.assertIn("available", result)
        self.assertIn("not available", result)

    def test_format_output_json(self):
        result = self.formatter.format_output(self.sample_results, "json")
        parsed = json.loads(result)
        self.assertEqual(len(parsed), 2)
        self.assertTrue(parsed[0]["is_available"])
        self.assertFalse(parsed[1]["is_available"])

    def test_format_output_invalid(self):
        pass


class TestDepsCheckResult(unittest.TestCase):

    def test_deps_check_result_creation(self):
        result = DepsCheckResult(
            dependency="docker",
            verbose=True,
            output="json",
            os="linux",
            package_manager="apt",
            is_available=True,
            error=None,
        )

        self.assertEqual(result.dependency, "docker")
        self.assertTrue(result.verbose)
        self.assertEqual(result.output, "json")
        self.assertEqual(result.os, "linux")
        self.assertEqual(result.package_manager, "apt")
        self.assertTrue(result.is_available)
        self.assertIsNone(result.error)

    def test_deps_check_result_with_error(self):
        result = DepsCheckResult(
            dependency="failing_dep",
            timeout=5,
            verbose=False,
            output="text",
            os="darwin",
            package_manager="brew",
            is_available=False,
            error="Command not found",
        )

        self.assertFalse(result.is_available)
        self.assertEqual(result.error, "Command not found")


class TestDepsConfig(unittest.TestCase):

    def test_valid_config(self):
        config = DepsConfig(
            deps=["docker", "kubectl"], verbose=True, output="json", os="linux", package_manager="apt"
        )

        self.assertEqual(config.deps, ["docker", "kubectl"])
        self.assertTrue(config.verbose)
        self.assertEqual(config.output, "json")
        self.assertEqual(config.os, "linux")
        self.assertEqual(config.package_manager, "apt")

    def test_config_validation_os(self):
        with self.assertRaises(ValueError):
            DepsConfig(deps=["docker"], os="invalid_os", package_manager="apt")

    def test_config_validation_package_manager(self):
        with self.assertRaises(ValueError):
            DepsConfig(deps=["docker"], os="linux", package_manager="invalid_manager")

    def test_config_timeout_validation(self):
        pass

    def test_config_deps_validation(self):
        with self.assertRaises(ValueError):
            DepsConfig(deps=[], os="linux", package_manager="apt")


class TestDepsService(unittest.TestCase):

    def setUp(self):
        self.config = DepsConfig(
            deps=["docker", "kubectl"], verbose=False, output="text", os="linux", package_manager="apt"
        )
        self.mock_logger = MockLogger()
        self.mock_checker = Mock()
        self.service = DepsService(config=self.config, logger=self.mock_logger, checker=self.mock_checker)

    def test_create_result(self):
        result = self.service._create_result("docker", True)

        self.assertEqual(result.dependency, "docker")
        self.assertFalse(result.verbose)
        self.assertEqual(result.output, "text")
        self.assertEqual(result.os, "linux")
        self.assertEqual(result.package_manager, "apt")
        self.assertTrue(result.is_available)
        self.assertIsNone(result.error)

    def test_create_result_with_error(self):
        result = self.service._create_result("failing_dep", False, "Command not found")

        self.assertFalse(result.is_available)
        self.assertEqual(result.error, "Command not found")

    def test_check_single_dependency_success(self):
        self.mock_checker.check_dependency.return_value = True

        result = self.service._check_dependency("docker")

        self.assertTrue(result.is_available)
        self.mock_checker.check_dependency.assert_called_once_with("docker")

    def test_check_single_dependency_failure(self):
        self.mock_checker.check_dependency.return_value = False

        result = self.service._check_dependency("nonexistent")

        self.assertFalse(result.is_available)
        self.mock_checker.check_dependency.assert_called_once_with("nonexistent")

    def test_check_single_dependency_exception(self):
        self.mock_checker.check_dependency.side_effect = Exception("Test error")

        result = self.service._check_dependency("failing_dep")

        self.assertFalse(result.is_available)
        self.assertEqual(result.error, "Test error")

    @patch("app.commands.preflight.deps.ParallelProcessor")
    def test_check_dependencies(self, mock_parallel_processor):
        mock_results = [self.service._create_result("docker", True), self.service._create_result("kubectl", False)]
        mock_parallel_processor.process_items.return_value = mock_results

        results = self.service.check_dependencies()

        self.assertEqual(len(results), 2)
        mock_parallel_processor.process_items.assert_called_once()

    def test_check_and_format(self):
        mock_results = [self.service._create_result("docker", True), self.service._create_result("kubectl", False)]

        with patch.object(self.service, "check_dependencies", return_value=mock_results):
            result = self.service.check_and_format()

        self.assertIn("docker", result)
        self.assertIn("kubectl", result)
        self.assertIn("available", result)
        self.assertIn("not available", result)


class TestDeps(unittest.TestCase):

    def setUp(self):
        self.mock_logger = MockLogger()
        self.deps = Deps(logger=self.mock_logger)

    def test_check(self):
        config = DepsConfig(deps=["docker"], os="linux", package_manager="apt")

        with patch("app.commands.preflight.deps.DepsService") as mock_service_class:
            mock_service = Mock()
            mock_results = [Mock()]
            mock_service.check_dependencies.return_value = mock_results
            mock_service_class.return_value = mock_service

            results = self.deps.check(config)

            self.assertEqual(results, mock_results)
            mock_service_class.assert_called_once_with(config, logger=self.mock_logger)

    def test_format_output(self):
        mock_results = [Mock()]

        with patch.object(self.deps.formatter, "format_output", return_value="formatted") as mock_format:
            result = self.deps.format_output(mock_results, "text")

            self.assertEqual(result, "formatted")
            mock_format.assert_called_once_with(mock_results, "text")


if __name__ == "__main__":
    unittest.main()
