from app.utils.config import API_SERVICE, API_ENV_FILE, VIEW_SERVICE, VIEW_ENV_FILE, Config
from app.utils.lib import FileManager
from app.utils.logger import Logger
from app.commands.conf.base import BaseEnvironmentManager

config = Config()
api_env_file = config.get_yaml_value(API_ENV_FILE)
view_env_file = config.get_yaml_value(VIEW_ENV_FILE)

def conf_run(verbose: bool = False, output: str = "text") -> bool:
    logger = Logger(verbose=verbose)
    try:
        services = [
            ("api", API_SERVICE, api_env_file),
            ("view", VIEW_SERVICE, view_env_file),
        ]
        env_manager = BaseEnvironmentManager(logger)
        for service_name, service_key, env_file in services:
            env_values = config.get_service_env_values(service_key)
            success, error = env_manager.write_env_file(env_file, env_values)
            if not success:
                return False
            file_perm_success, file_perm_error = FileManager.set_permissions(env_file, 0o755)
            if not file_perm_success:
                return False
        return True
    except Exception:
            return False
