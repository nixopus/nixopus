import typer
import subprocess
from app.utils.config import Config
from app.utils.logger import Logger
from .messages import development_only_error, running_command

class TestCommand:
    def __init__(self):
        self.config = Config()
        self.logger = Logger()

    def run(self, target: str = typer.Argument(None, help="Test target (e.g., version)")):
        if not self.config.is_development():
            self.logger.error(development_only_error)
            raise typer.Exit(1)
        cmd = ["make", "test"]
        if target:
            cmd.append(f"test-{target}")
        self.logger.info(running_command.format(command=' '.join(cmd)))
        result = subprocess.run(cmd)
        raise typer.Exit(result.returncode) 
