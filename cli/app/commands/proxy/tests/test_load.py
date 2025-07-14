import pytest
from typer.testing import CliRunner
from app.commands.proxy.command import proxy_app
from unittest.mock import patch

runner = CliRunner()

def test_load_success(tmp_path):
    config_file = tmp_path / "caddy.json"
    config_file.write_text("{}")
    with patch("app.commands.proxy.load.CaddyService.load_config_file", return_value=(True, "ok")):
        result = runner.invoke(proxy_app, ["load", "--config-file", str(config_file)])
        assert result.exit_code == 0
        assert "successfully" in result.output

def test_load_missing_config():
    result = runner.invoke(proxy_app, ["load"])
    assert result.exit_code != 0
    assert "Configuration file is required" in result.output

def test_load_error(tmp_path):
    config_file = tmp_path / "caddy.json"
    config_file.write_text("{}")
    with patch("app.commands.proxy.load.CaddyService.load_config_file", return_value=(False, "fail")):
        result = runner.invoke(proxy_app, ["load", "--config-file", str(config_file)])
        assert result.exit_code != 0
        assert "fail" in result.output 
