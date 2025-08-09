from importlib.metadata import version
from unittest.mock import MagicMock, patch

import pytest

from app.commands.version.version import VersionCommand


class TestVersionCommand:
    """Test cases for the VersionCommand class"""

    @patch("app.commands.version.version.Console")
    @patch("app.commands.version.version.version")
    def test_version_command_success(self, mock_version, mock_console_class):
        """Test successful version display"""
        mock_version.return_value = "1.0.0"
        mock_console = MagicMock()
        mock_console_class.return_value = mock_console

        version_command = VersionCommand()
        version_command.run()

        mock_version.assert_called_once_with("nixopus")
        mock_console.print.assert_called_once()

        call_args = mock_console.print.call_args[0][0]
        assert call_args.title == "[bold white]Version Info[/bold white]"
        assert call_args.border_style == "blue"
        assert call_args.padding == (0, 1)

    @patch("app.commands.version.version.Console")
    @patch("app.commands.version.version.version")
    def test_version_command_with_different_versions(self, mock_version, mock_console_class):
        """Test version display with different version numbers"""
        test_versions = ["0.1.0", "2.3.4", "1.0.0-beta"]
        mock_console = MagicMock()
        mock_console_class.return_value = mock_console

        for test_version in test_versions:
            mock_version.return_value = test_version
            mock_console.reset_mock()

            version_command = VersionCommand()
            version_command.run()

            mock_version.assert_called_with("nixopus")
            mock_console.print.assert_called_once()

    @patch("app.commands.version.version.Console")
    @patch("app.commands.version.version.version")
    def test_version_command_panel_content(self, mock_version, mock_console_class):
        """Test that panel contains correct text content"""
        mock_version.return_value = "1.2.3"
        mock_console = MagicMock()
        mock_console_class.return_value = mock_console

        version_command = VersionCommand()
        version_command.run()

        call_args = mock_console.print.call_args[0][0]
        panel_content = call_args.renderable

        assert "Nixopus CLI" in str(panel_content)
        assert "v1.2.3" in str(panel_content)

    @patch("app.commands.version.version.Console")
    @patch("app.commands.version.version.version")
    def test_version_command_handles_version_error(self, mock_version, mock_console_class):
        """Test handling of version import error"""
        mock_version.side_effect = Exception("Version not found")
        mock_console = MagicMock()
        mock_console_class.return_value = mock_console

        with pytest.raises(Exception):
            version_command = VersionCommand()
            version_command.run()

        mock_version.assert_called_once_with("nixopus")

    @patch("app.commands.version.version.Console")
    @patch("app.commands.version.version.version")
    def test_version_command_console_error_handling(self, mock_version, mock_console_class):
        """Test handling of console print errors"""
        mock_version.return_value = "1.0.0"
        mock_console = MagicMock()
        mock_console.print.side_effect = Exception("Console error")
        mock_console_class.return_value = mock_console

        with pytest.raises(Exception):
            version_command = VersionCommand()
            version_command.run()

        mock_version.assert_called_once_with("nixopus")
        mock_console.print.assert_called_once()


class TestVersionCommandClass:
    """Test cases for VersionCommand class structure"""

    def test_version_command_initialization(self):
        """Test that VersionCommand can be instantiated"""
        with patch("app.commands.version.version.Console"):
            version_command = VersionCommand()
            assert hasattr(version_command, "console")

    def test_version_command_run_method(self):
        """Test that VersionCommand has a run method"""
        with patch("app.commands.version.version.Console"):
            version_command = VersionCommand()
            assert hasattr(version_command, "run")
            assert callable(version_command.run)

    def test_version_command_run_returns_none(self):
        """Test that run method returns None"""
        with patch("app.commands.version.version.Console"):
            with patch("app.commands.version.version.version", return_value="1.0.0"):
                version_command = VersionCommand()
                result = version_command.run()
                assert result is None


class TestVersionModuleImports:
    """Test cases for module imports and dependencies"""

    def test_import_metadata_version(self):
        """Test that importlib.metadata.version is available"""
        try:
            from importlib.metadata import version

            assert callable(version)
        except ImportError:
            pytest.skip("importlib.metadata not available")

    def test_rich_console_import(self):
        """Test that rich.console.Console is available"""
        try:
            from rich.console import Console

            assert callable(Console)
        except ImportError:
            pytest.skip("rich.console not available")

    def test_rich_panel_import(self):
        """Test that rich.panel.Panel is available"""
        try:
            from rich.panel import Panel

            assert callable(Panel)
        except ImportError:
            pytest.skip("rich.panel not available")

    def test_rich_text_import(self):
        """Test that rich.text.Text is available"""
        try:
            from rich.text import Text

            assert callable(Text)
        except ImportError:
            pytest.skip("rich.text not available")


class TestVersionCommandSignature:
    """Test cases for class method signature and behavior"""

    def test_version_command_is_instantiable(self):
        """Test that VersionCommand can be instantiated"""
        with patch("app.commands.version.version.Console"):
            version_command = VersionCommand()
            assert isinstance(version_command, VersionCommand)

    def test_run_method_no_parameters(self):
        """Test that run method takes no parameters"""
        import inspect

        with patch("app.commands.version.version.Console"):
            version_command = VersionCommand()
            sig = inspect.signature(version_command.run)
            assert len(sig.parameters) == 0
