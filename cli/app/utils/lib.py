from enum import Enum
import platform
import subprocess
import os
import shutil
from typing import TypeVar, Callable, List
from concurrent.futures import ThreadPoolExecutor, as_completed
from app.utils.message import REMOVED_DIRECTORY_MESSAGE, FAILED_TO_REMOVE_DIRECTORY_MESSAGE

T = TypeVar('T')
R = TypeVar('R')

class SupportedOS(str, Enum):
    LINUX = "linux"
    MACOS = "darwin"
    
class SupportedDistribution(str, Enum):
    DEBIAN = "debian"
    UBUNTU = "ubuntu"
    CENTOS = "centos"
    FEDORA = "fedora"
    ALPINE = "alpine"
    
class SupportedPackageManager(str, Enum):
    APT = "apt"
    YUM = "yum"
    DNF = "dnf"
    PACMAN = "pacman"
    APK = "apk"
    BREW = "brew"

class Supported:
    @staticmethod
    def os(os_name: str) -> bool:
        return os_name in [os.value for os in SupportedOS]
    
    @staticmethod
    def distribution(distribution: str) -> bool:
        return distribution in [dist.value for dist in SupportedDistribution]
    
    @staticmethod
    def package_manager(package_manager: str) -> bool:
        return package_manager in [pm.value for pm in SupportedPackageManager]

    @staticmethod
    def get_os():
        return [os.value for os in SupportedOS]
    
    @staticmethod
    def get_distributions():
        return [dist.value for dist in SupportedDistribution]

class HostInformation:
    @staticmethod
    def get_os_name():
        return platform.system().lower()
    
    @staticmethod
    def get_package_manager():
        os_name = HostInformation.get_os_name()
        
        if os_name == SupportedOS.MACOS.value:
            return SupportedPackageManager.BREW.value
        
        package_managers = [pm.value for pm in SupportedPackageManager if pm != SupportedPackageManager.BREW]
        
        for pm in package_managers:
            if HostInformation.command_exists(pm):
                return pm
        
        return None
    
    @staticmethod
    def command_exists(command):
        try:
            result = subprocess.run(["command", "-v", command], 
                                  capture_output=True, text=True, check=False)
            return result.returncode == 0
        except Exception:
            return False

class ParallelProcessor:
    @staticmethod
    def process_items(
        items: List[T],
        processor_func: Callable[[T], R],
        max_workers: int = 50,
        error_handler: Callable[[T, Exception], R] = None
    ) -> List[R]:
        if not items:
            return []
        
        results = []
        max_workers = min(len(items), max_workers)
        
        with ThreadPoolExecutor(max_workers=max_workers) as executor:
            futures = {executor.submit(processor_func, item): item for item in items}
            
            for future in as_completed(futures):
                try:
                    result = future.result()
                    results.append(result)
                except Exception as e:
                    item = futures[future]
                    if error_handler:
                        error_result = error_handler(item, e)
                        results.append(error_result)
        return results

class DirectoryManager:
    @staticmethod
    def path_exists(path: str) -> bool:
        return os.path.exists(path)
    
    @staticmethod
    def path_exists_and_not_force(path: str, force: bool) -> bool:
        return os.path.exists(path) and not force
    
    @staticmethod
    def remove_directory(path: str, logger=None) -> bool:
        try:
            shutil.rmtree(path)
            if logger:
                logger.info(REMOVED_DIRECTORY_MESSAGE.format(path=path))
            return True
        except Exception as e:
            if logger:
                logger.error(FAILED_TO_REMOVE_DIRECTORY_MESSAGE.format(path=path, error=e))
            return False
