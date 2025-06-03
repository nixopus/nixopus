import json
from pathlib import Path
from dataclasses import dataclass
from typing import Dict, Any, TypeVar, Generic, Type

T = TypeVar('T')

@dataclass
class BaseConfig(Generic[T]):
    config_path: Path
    env: str
    required_keys: list[str]
    valid_environments: list[str]

    def load_config(self) -> Dict[str, Any]:
        try:
            with open(self.config_path, 'r') as f:
                return json.load(f)
        except FileNotFoundError:
            raise Exception(f"Configuration file not found at {self.config_path}")
        except json.JSONDecodeError:
            raise Exception(f"Invalid JSON in configuration file at {self.config_path}")

    def validate_environment(self) -> None:
        if self.env not in self.valid_environments:
            raise ValueError(f"Invalid environment: {self.env}. Must be one of {', '.join(self.valid_environments)}")

    def validate_config(self, config: Dict[str, Any]) -> None:
        env_config = config.get(self.env)
        if not env_config:
            raise Exception(f"Configuration for environment '{self.env}' not found in config file")

        missing_keys = [key for key in self.required_keys if key not in env_config]
        if missing_keys:
            raise Exception(f"Missing required configuration keys for environment '{self.env}': {', '.join(missing_keys)}")

    def create(self, config_class: Type[T]) -> T:
        self.validate_environment()
        config = self.load_config()
        self.validate_config(config)
        
        try:
            return config_class(**config[self.env])
        except (ValueError, TypeError) as e:
            raise Exception(f"Invalid configuration type for environment '{self.env}': {str(e)}") from e 