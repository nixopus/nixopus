from dataclasses import dataclass
from typing import Dict, List
from pathlib import Path

@dataclass
class ServiceConfig:
    config_dir: Path
    docker: Dict[str, str]
    source: str
    compose: Dict[str, str]
    containers: Dict[str, str]
    caddy: Dict[str, str]
    api: Dict[str, str]
    system: Dict[str, List[str]]

    def __post_init__(self):
        if isinstance(self.config_dir, str):
            self.config_dir = Path(self.config_dir) 