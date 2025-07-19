import typer
from app.utils.protocols import LoggerProtocol
from .messages import installing_nixopus, failed_to_run_preflight, failed_to_run_conf, failed_to_run_deps, failed_to_run_clone, failed_to_run_up, failed_to_run_ssh
from app.utils.config import PORTS, DEPS, NIXOPUS_CONFIG_DIR, DEFAULT_REPO, DEFAULT_BRANCH, DEFAULT_PATH, DEFAULT_COMPOSE_FILE, PROXY_PORT
from app.commands.preflight.run import preflight_run
from app.utils.config import Config
from app.commands.conf.run import conf_run
from app.commands.install.deps import install_all_deps
from app.commands.clone.clone import Clone, CloneConfig
from app.commands.service.command import Up, UpConfig
from app.commands.install.ssh import SSH, SSHConfig
from app.commands.proxy.command import Load, LoadConfig

config = Config()
ports_config = config.get_yaml_value(PORTS)
ports_list = [int(port) for port in ports_config]
nixopus_config_dir = config.get_yaml_value(NIXOPUS_CONFIG_DIR)
deps_list = list(config.get_yaml_value(DEPS).keys())
repo = config.get_yaml_value(DEFAULT_REPO)
branch = config.get_yaml_value(DEFAULT_BRANCH)
path = nixopus_config_dir + "/" + config.get_yaml_value(DEFAULT_PATH)
compose_file = config.get_yaml_value(DEFAULT_COMPOSE_FILE)
proxy_port = config.get_yaml_value(PROXY_PORT)

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
            
            # ssh setup 
            config = SSHConfig(
                path=nixopus_config_dir + "/ssh/nixopus_ed25519",
                key_type="ed25519",
                key_size=4096,
                passphrase=None,
                verbose=False,
                output="text",
                dry_run=False,
                force=False,
                set_permissions=True,
                add_to_authorized_keys=False,
                create_ssh_directory=True,
                )
            ssh_operation = SSH(logger=self.logger)
            result = ssh_operation.generate(config)
            if not result.success:
                raise Exception(failed_to_run_ssh)
            else:
                self.logger.success(ssh_operation.format_output(result, "text"))

            # service up
            config = UpConfig(
                name="all",
                detach=False,
                env_file=None,
                verbose=False,
                output="text",
                dry_run=False,
                compose_file=nixopus_config_dir + "/" + compose_file
            )

            up_service = Up(logger=self.logger)
            result = up_service.up(config)
            if not result.success:
                raise Exception(failed_to_run_up)
            else:
                self.logger.success(up_service.format_output(result, "text"))

            # setup caddy proxy to redirect to the api and view services
            config = LoadConfig(proxy_port=proxy_port, verbose=False, output="text", dry_run=False, config_file=None)

            load_service = Load(logger=self.logger)
            result = load_service.load(config)

            if result.success:
                self.logger.success(load_service.format_output(result, "text"))
            else:
                self.logger.error(result.error)
                raise typer.Exit(1)
            
            # doctor check for any issues after the installation
            # success
            
        except Exception as e:
            self.logger.error(f"Something went wrong while installing Nixopus: {str(e)}")
            raise typer.Exit(1)
