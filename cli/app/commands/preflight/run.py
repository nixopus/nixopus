from app.utils.config import DEPS, NIXOPUS_CONFIG_DIR, PORTS
from app.utils.lib import FileManager, HostInformation
from app.utils.logger import Logger
from app.commands.preflight.port import PortConfig, PortService
from app.commands.preflight.deps import DepsConfig, Deps

def preflight_run(logger: Logger, ports_list: list[int], deps_list: list[str], nixopus_config_dir: str, output: str = "text", host: str = "localhost", timeout: int = 1, verbose: bool = False) -> bool:
    try:
        try:
            ports_config = PortConfig(ports=ports_list, host=host, timeout=timeout, verbose=verbose)
            port_service = PortService(ports_config, logger=logger)
            port_results = port_service.check_ports()
            
            unavailable_ports = [r for r in port_results if not r["is_available"]]
            if unavailable_ports:
                return False
                
        except Exception as e:
            return False
        
        try:
            deps_config = DepsConfig(
                deps=deps_list,
                timeout=timeout,
                verbose=verbose,
                output=output,
                os=HostInformation.get_os_name(),
                package_manager=HostInformation.get_package_manager(),
            )
            deps_checker = Deps(logger=logger)
            deps_results = deps_checker.check(deps_config)
            
            missing_deps = [r for r in deps_results if not r.is_available]
            if missing_deps:
                return False
                
        except Exception as e:
            return False
        
        file_manager = FileManager()
        created, error = file_manager.create_directory(nixopus_config_dir)
        if not created:
            return False
        set_permissions, error = file_manager.set_permissions(nixopus_config_dir, 0o755)
        if not set_permissions:
            return False
            
        return True
        
    except Exception as e:
        return False