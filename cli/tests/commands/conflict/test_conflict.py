"""
Comprehensive test suite for the conflict command.

This module serves as the main test runner that imports and runs all
the separated test modules for better organization and separation of concerns.

The conflict command has been refactored into:
- models.py: Data models and configuration classes
- conflict.py: Core business logic and services
"""

import os
import sys
import unittest

# Add the tests directory to the path to import the separated test modules
sys.path.insert(0, os.path.dirname(__file__))

# Import all the separated test modules
from test_config_and_models import TestConfigAndModels
from test_service_integration import TestServiceIntegration
from test_version_checker import TestVersionChecker


def create_test_suite():
    """Create a comprehensive test suite with all conflict command tests."""
    suite = unittest.TestSuite()

    # Add all test classes
    suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestConfigAndModels))
    suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestVersionChecker))
    suite.addTests(unittest.TestLoader().loadTestsFromTestCase(TestServiceIntegration))

    return suite


class TestConflictCommand(unittest.TestCase):
    """Main test class that runs all separated tests."""

    def test_run_all_conflict_tests(self):
        """Run all separated test modules and ensure they pass."""
        suite = create_test_suite()
        runner = unittest.TextTestRunner(verbosity=2)
        result = runner.run(suite)

        # Ensure all tests passed
        self.assertEqual(result.errors, [])
        self.assertEqual(result.failures, [])
        self.assertTrue(result.wasSuccessful())


if __name__ == "__main__":
    unittest.main()
