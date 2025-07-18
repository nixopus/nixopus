import typer
from app.utils.protocols import LoggerProtocol
from .messages import installing_nixopus, failed_to_run_preflight, failed_to_run_conf, failed_to_run_deps, failed_to_run_clone  
from app.utils.config import PORTS, DEPS, NIXOPUS_CONFIG_DIR, DEFAULT_REPO, DEFAULT_BRANCH, DEFAULT_PATH
from app.commands.preflight.run import preflight_run
from app.utils.config import Config
from app.commands.conf.run import conf_run
from app.commands.install.deps import install_all_deps
from app.commands.clone.clone import Clone, CloneConfig

config = Config()
ports_config = config.get_yaml_value(PORTS)
ports_list = [int(port) for port in ports_config]
nixopus_config_dir = config.get_yaml_value(NIXOPUS_CONFIG_DIR)
deps_list = list(config.get_yaml_value(DEPS).keys())
repo = config.get_yaml_value(DEFAULT_REPO)
branch = config.get_yaml_value(DEFAULT_BRANCH)
path = nixopus_config_dir + "/" + config.get_yaml_value(DEFAULT_PATH)

class Install:
    def __init__(self, logger: LoggerProtocol):
        self.logger = logger

    def run(self):
        try:
            # Run preflight checks and make sure the system is ready to install Nixopus, it will check for ports, dependencies, and create the config directory for nixopus
            success = preflight_run(
                logger=self.logger,
                ports_list=ports_list,
                deps_list=deps_list,
                nixopus_config_dir=nixopus_config_dir,
                output="text",
                host="localhost",
                timeout=1,
                verbose=False
            )
            if not success:
                raise Exception(failed_to_run_preflight)

            # Run the conf command to set the environment variables for the api and view services
            success = conf_run(verbose=False, output="text")
            if not success:
                raise Exception(failed_to_run_conf)
            
            # Run the dependency install command to install the dependencies for the api and view services
            success = install_all_deps(verbose=False, output="text", dry_run=False)
            if not success:
                raise Exception(failed_to_run_deps)
            
            # Run the clone command to clone the nixopus repository
            config = CloneConfig(repo=repo, branch=branch, path=path, force=False, verbose=False, output="text", dry_run=False)
            clone_operation = Clone(logger=self.logger)
            result = clone_operation.clone(config)
            
            # ssh run
            
            # service up
            
            # proxy
            
            # doctor
            
            # success
                        
        except Exception as e:
            self.logger.error(f"Something went wrong while installing Nixopus: {str(e)}")
            raise typer.Exit(1)
