from unittest.mock import patch

import pytest
from typer.testing import CliRunner

from app.commands.proxy.command import proxy_app

runner = CliRunner()


def test_status_running():
    with patch("app.commands.proxy.status.CaddyService.get_status", return_value=(True, "Caddy is running")):
        result = runner.invoke(proxy_app, ["status"])
        assert result.exit_code == 0
        assert "running" in result.output


def test_status_not_running():
    with patch("app.commands.proxy.status.CaddyService.get_status", return_value=(False, "not running")):
        result = runner.invoke(proxy_app, ["status"])
        assert result.exit_code != 0
        assert "not running" in result.output
