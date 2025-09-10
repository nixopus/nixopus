import os
import sys
import tempfile
import unittest
from unittest.mock import Mock, mock_open, patch

from app.utils.config import Config, expand_env_placeholders
from app.utils.message import MISSING_CONFIG_KEY_MESSAGE


class TestConfig(unittest.TestCase):
    def setUp(self):
        self.temp_dir = tempfile.mkdtemp()
        self.test_config_path = os.path.join(self.temp_dir, "test_config.yaml")
        self.sample_config = {
            "services": {
                "api": {"env": {"PORT": "${API_PORT:-8443}", "DB_NAME": "${DB_NAME:-postgres}"}},
                "view": {"env": {"PORT": "${VIEW_PORT:-7443}"}},
            },
            "clone": {"repo": "https://github.com/test/repo", "branch": "main", "source-path": "/tmp/source"},
            "deps": {"curl": {"package": "curl", "command": "curl"}, "docker": {"package": "docker.io", "command": "docker"}},
            "ports": [2019, 80, 443, 7443, 8443],
        }

    def tearDown(self):
        import shutil

        shutil.rmtree(self.temp_dir, ignore_errors=True)

    def test_get_env_default(self):
        if "ENV" in os.environ:
            del os.environ["ENV"]
        config = Config()
        self.assertEqual(config.get_env(), "PRODUCTION")

    @patch("os.environ.get")
    def test_get_env_custom(self, mock_environ_get):
        mock_environ_get.return_value = "DEVELOPMENT"
        config = Config()
        self.assertEqual(config.get_env(), "DEVELOPMENT")

    @patch("os.environ.get")
    def test_is_development_true(self, mock_environ_get):
        mock_environ_get.return_value = "DEVELOPMENT"
        config = Config()
        self.assertTrue(config.is_development())

    @patch("os.environ.get")
    def test_is_development_false(self, mock_environ_get):
        mock_environ_get.return_value = "PRODUCTION"
        config = Config()
        self.assertFalse(config.is_development())

    @patch("os.environ.get")
    def test_is_development_case_insensitive(self, mock_environ_get):
        mock_environ_get.return_value = "development"
        config = Config()
        self.assertTrue(config.is_development())

    @patch("builtins.open", new_callable=mock_open)
    @patch("yaml.safe_load")
    def test_load_yaml_config_success(self, mock_yaml_load, mock_file):
        mock_yaml_load.return_value = self.sample_config
        config = Config()
        result = config.load_yaml_config()
        self.assertEqual(result, self.sample_config)
        mock_file.assert_called_once()

    @patch("builtins.open")
    def test_load_yaml_config_file_not_found(self, mock_open):
        mock_open.side_effect = FileNotFoundError("File not found")
        config = Config()
        with self.assertRaises(FileNotFoundError):
            config.load_yaml_config()

    @patch("builtins.open", new_callable=mock_open)
    @patch("yaml.safe_load")
    def test_load_yaml_config_cached(self, mock_yaml_load, mock_file):
        mock_yaml_load.return_value = self.sample_config
        config = Config()
        result1 = config.load_yaml_config()
        result2 = config.load_yaml_config()
        self.assertEqual(result1, result2)
        self.assertEqual(mock_yaml_load.call_count, 1)

    @patch("builtins.open", new_callable=mock_open)
    @patch("yaml.safe_load")
    @patch("app.utils.config.expand_env_placeholders")
    def test_get_yaml_value_success(self, mock_expand, mock_yaml_load, mock_file):
        mock_yaml_load.return_value = self.sample_config
        mock_expand.return_value = "8443"
        config = Config()
        result = config.get_yaml_value("services.api.env.PORT")
        self.assertEqual(result, "8443")
        mock_expand.assert_called_once_with("${API_PORT:-8443}")

    @patch("builtins.open", new_callable=mock_open)
    @patch("yaml.safe_load")
    def test_get_yaml_value_non_string(self, mock_yaml_load, mock_file):
        mock_yaml_load.return_value = self.sample_config
        config = Config()
        result = config.get_yaml_value("ports")
        self.assertEqual(result, [2019, 80, 443, 7443, 8443])

    @patch("builtins.open", new_callable=mock_open)
    @patch("yaml.safe_load")
    def test_get_yaml_value_missing_key(self, mock_yaml_load, mock_file):
        mock_yaml_load.return_value = self.sample_config
        config = Config()
        with self.assertRaises(KeyError) as context:
            config.get_yaml_value("services.api.env.NONEXISTENT")
        expected_message = MISSING_CONFIG_KEY_MESSAGE.format(path="services.api.env.NONEXISTENT", key="NONEXISTENT")
        self.assertEqual(context.exception.args[0], expected_message)

    @patch("builtins.open", new_callable=mock_open)
    @patch("yaml.safe_load")
    def test_get_yaml_value_missing_path(self, mock_yaml_load, mock_file):
        mock_yaml_load.return_value = self.sample_config
        config = Config()
        with self.assertRaises(KeyError) as context:
            config.get_yaml_value("nonexistent.path")
        expected_message = MISSING_CONFIG_KEY_MESSAGE.format(path="nonexistent.path", key="nonexistent")
        self.assertEqual(context.exception.args[0], expected_message)

    @patch("builtins.open", new_callable=mock_open)
    @patch("yaml.safe_load")
    @patch("app.utils.config.expand_env_placeholders")
    def test_get_service_env_values(self, mock_expand, mock_yaml_load, mock_file):
        mock_yaml_load.return_value = self.sample_config
        mock_expand.side_effect = lambda x: x.replace("${API_PORT:-8443}", "8443")
        config = Config()
        result = config.get_service_env_values("services.api.env")
        expected = {"PORT": "8443", "DB_NAME": "${DB_NAME:-postgres}"}
        self.assertEqual(result, expected)

    @patch("yaml.safe_load")
    def test_load_user_config_success(self, mock_yaml_load):
        user_config = {"services": {"api": {"env": {"PORT": "9000"}}}}
        mock_yaml_load.return_value = user_config
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
            f.write("dummy content")
            config_file = f.name
        try:
            config = Config()
            result = config.load_user_config(config_file)
            expected = {"services.api.env.PORT": "9000"}
            self.assertEqual(result, expected)
        finally:
            os.unlink(config_file)

    def test_load_user_config_empty_file(self):
        config = Config()
        result = config.load_user_config(None)
        self.assertEqual(result, {})

    def test_load_user_config_file_not_found(self):
        config = Config()
        with self.assertRaises(FileNotFoundError) as context:
            config.load_user_config("/nonexistent/file.yaml")
        self.assertIn("Config file not found", str(context.exception))

    def test_flatten_config_simple(self):
        config = Config()
        nested = {"a": 1, "b": 2}
        flattened = {}
        config.flatten_config(nested, flattened)
        self.assertEqual(flattened, {"a": 1, "b": 2})

    def test_flatten_config_nested(self):
        config = Config()
        nested = {"services": {"api": {"env": {"PORT": "8443"}}}}
        flattened = {}
        config.flatten_config(nested, flattened)
        expected = {"services.api.env.PORT": "8443"}
        self.assertEqual(flattened, expected)

    def test_flatten_config_with_prefix(self):
        config = Config()
        nested = {"a": 1}
        flattened = {}
        config.flatten_config(nested, flattened, "prefix")
        self.assertEqual(flattened, {"prefix.a": 1})

    def test_unflatten_config_simple(self):
        config = Config()
        flattened = {"a": 1, "b": 2}
        result = config.unflatten_config(flattened)
        expected = {"a": 1, "b": 2}
        self.assertEqual(result, expected)

    def test_unflatten_config_nested(self):
        config = Config()
        flattened = {"a.b.c": 1, "a.b.d": 2, "a.e": 3, "f": 4}
        result = config.unflatten_config(flattened)
        expected = {"a": {"b": {"c": 1, "d": 2}, "e": 3}, "f": 4}
        self.assertEqual(result, expected)

    def test_get_config_value_cached(self):
        config = Config()
        user_config = {"test.key": "value"}
        defaults = {"test.key": "default"}
        result1 = config.get_config_value("test.key", user_config, defaults)
        result2 = config.get_config_value("test.key", user_config, defaults)
        self.assertEqual(result1, "value")
        self.assertEqual(result2, "value")
        self.assertEqual(result1, result2)

    def test_get_config_value_user_config_priority(self):
        config = Config()
        user_config = {"services.caddy.env.PROXY_PORT": "2020"}
        defaults = {"proxy_port": "2019"}
        result = config.get_config_value("proxy_port", user_config, defaults)
        self.assertEqual(result, "2020")

    def test_get_config_value_defaults_fallback(self):
        config = Config()
        user_config = {}
        defaults = {"proxy_port": "2019"}
        result = config.get_config_value("proxy_port", user_config, defaults)
        self.assertEqual(result, "2019")

    def test_get_config_value_missing_no_default(self):
        config = Config()
        user_config = {}
        defaults = {}
        with self.assertRaises(ValueError) as context:
            config.get_config_value("missing_key", user_config, defaults)
        self.assertIn("Configuration key 'missing_key' has no default value", str(context.exception))

    def test_get_config_value_ssh_passphrase_optional(self):
        config = Config()
        user_config = {}
        defaults = {}
        result = config.get_config_value("ssh_passphrase", user_config, defaults)
        self.assertIsNone(result)

    def test_get_config_value_key_mappings(self):
        config = Config()
        user_config = {
            "clone.repo": "https://github.com/test/repo",
            "clone.branch": "main",
            "clone.source-path": "/tmp/source",
        }
        defaults = {}
        repo_result = config.get_config_value("repo_url", user_config, defaults)
        branch_result = config.get_config_value("branch_name", user_config, defaults)
        path_result = config.get_config_value("source_path", user_config, defaults)
        self.assertEqual(repo_result, "https://github.com/test/repo")
        self.assertEqual(branch_result, "main")
        self.assertEqual(path_result, "/tmp/source")

    def test_config_pyinstaller_bundle(self):
        sys.frozen = True
        sys._MEIPASS = "/bundle"
        with patch("os.path.join") as mock_join:
            mock_join.return_value = "/bundle/helpers/config.prod.yaml"
            config = Config()
            self.assertEqual(config._yaml_path, "/bundle/helpers/config.prod.yaml")
        del sys.frozen
        del sys._MEIPASS

    def test_config_normal_python(self):
        if hasattr(sys, "frozen"):
            del sys.frozen
        if hasattr(sys, "_MEIPASS"):
            del sys._MEIPASS
        with patch("os.path.abspath") as mock_abspath:
            mock_abspath.return_value = "/normal/path/helpers/config.prod.yaml"
            config = Config()
            self.assertNotIn("_MEIPASS", config._yaml_path)


class TestExpandEnvPlaceholders(unittest.TestCase):
    def setUp(self):
        self.original_environ = os.environ.copy()

    def tearDown(self):
        os.environ.clear()
        os.environ.update(self.original_environ)

    def test_expand_env_placeholders_no_placeholders(self):
        result = expand_env_placeholders("simple string")
        self.assertEqual(result, "simple string")

    def test_expand_env_placeholders_simple_variable(self):
        os.environ["TEST_VAR"] = "test_value"
        result = expand_env_placeholders("${TEST_VAR}")
        self.assertEqual(result, "test_value")

    def test_expand_env_placeholders_with_default(self):
        result = expand_env_placeholders("${TEST_VAR:-default_value}")
        self.assertEqual(result, "default_value")

    def test_expand_env_placeholders_variable_overrides_default(self):
        os.environ["TEST_VAR"] = "actual_value"
        result = expand_env_placeholders("${TEST_VAR:-default_value}")
        self.assertEqual(result, "actual_value")

    def test_expand_env_placeholders_multiple_placeholders(self):
        os.environ["VAR1"] = "value1"
        os.environ["VAR2"] = "value2"
        result = expand_env_placeholders("${VAR1} and ${VAR2}")
        self.assertEqual(result, "value1 and value2")

    def test_expand_env_placeholders_mixed_content(self):
        os.environ["PORT"] = "8443"
        result = expand_env_placeholders("http://localhost:${PORT:-8080}/api")
        self.assertEqual(result, "http://localhost:8443/api")

    def test_expand_env_placeholders_empty_default(self):
        result = expand_env_placeholders("${TEST_VAR:-}")
        self.assertEqual(result, "")

    def test_expand_env_placeholders_complex_default(self):
        result = expand_env_placeholders("${TEST_VAR:-http://localhost:8080}")
        self.assertEqual(result, "http://localhost:8080")

    def test_expand_env_placeholders_special_characters_in_default(self):
        result = expand_env_placeholders("${TEST_VAR:-/path/with/special/chars}")
        self.assertEqual(result, "/path/with/special/chars")

    def test_expand_env_placeholders_numeric_default(self):
        result = expand_env_placeholders("${TEST_VAR:-123}")
        self.assertEqual(result, "123")

    def test_expand_env_placeholders_underscore_in_variable_name(self):
        os.environ["TEST_VAR_NAME"] = "test_value"
        result = expand_env_placeholders("${TEST_VAR_NAME}")
        self.assertEqual(result, "test_value")

    def test_expand_env_placeholders_case_sensitive(self):
        os.environ["test_var"] = "lowercase"
        os.environ["TEST_VAR"] = "uppercase"
        result = expand_env_placeholders("${test_var} and ${TEST_VAR}")
        self.assertEqual(result, "lowercase and uppercase")

    def test_expand_env_placeholders_invalid_variable_name(self):
        result = expand_env_placeholders("${123INVALID}")
        self.assertEqual(result, "${123INVALID}")

    def test_expand_env_placeholders_malformed_placeholder(self):
        result = expand_env_placeholders("${MISSING_BRACE")
        self.assertEqual(result, "${MISSING_BRACE")

    def test_expand_env_placeholders_empty_variable_name(self):
        result = expand_env_placeholders("${}")
        self.assertEqual(result, "${}")

    def test_expand_env_placeholders_nested_braces(self):
        result = expand_env_placeholders("${TEST_VAR:-{nested}}")
        self.assertEqual(result, "{nested}")

    def test_expand_env_placeholders_multiple_defaults(self):
        result = expand_env_placeholders("${VAR1:-default1} and ${VAR2:-default2}")
        self.assertEqual(result, "default1 and default2")

    def test_expand_env_placeholders_real_world_example(self):
        os.environ["API_PORT"] = "9000"
        os.environ["DB_NAME"] = "production_db"
        result = expand_env_placeholders("${API_PORT:-8443} and ${DB_NAME:-postgres}")
        self.assertEqual(result, "9000 and production_db")


if __name__ == "__main__":
    unittest.main()
