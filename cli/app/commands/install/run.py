from app.utils.protocols import LoggerProtocol
from .messages import installing_nixopus

class Install:
    def __init__(self, logger: LoggerProtocol):
        self.logger = logger

    def run(self):
        self.logger.debug(installing_nixopus)
