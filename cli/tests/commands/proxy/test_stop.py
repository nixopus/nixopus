from unittest.mock import patch

import pytest
from typer.testing import CliRunner

from app.commands.proxy.command import proxy_app

runner = CliRunner()


def test_stop_success():
    with patch("app.commands.proxy.stop.CaddyService.stop_caddy", return_value=(True, "Caddy stopped successfully")):
        result = runner.invoke(proxy_app, ["stop"])
        assert result.exit_code == 0
        assert "stopped successfully" in result.output


def test_stop_error():
    with patch("app.commands.proxy.stop.CaddyService.stop_caddy", return_value=(False, "fail")):
        result = runner.invoke(proxy_app, ["stop"])
        assert result.exit_code != 0
        assert "fail" in result.output
