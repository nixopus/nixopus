import os
import platform
import shutil
import stat
import subprocess
from concurrent.futures import ThreadPoolExecutor, as_completed
from enum import Enum
from typing import Callable, List, Optional, Tuple, TypeVar
import requests

from app.utils.message import (
    FAILED_TO_GET_PUBLIC_IP_MESSAGE,
    FAILED_TO_REMOVE_DIRECTORY_MESSAGE,
    REMOVED_DIRECTORY_MESSAGE,
    SETTING_PERMISSIONS_MESSAGE,
    PERMISSIONS_SET_SUCCESS_MESSAGE,
    FAILED_TO_SET_PERMISSIONS_MESSAGE,
    CREATED_DIRECTORY_MESSAGE,
    FAILED_TO_CREATE_DIRECTORY_MESSAGE,
    CONTENT_APPENDED_MESSAGE,
    FAILED_TO_APPEND_MESSAGE,
    FAILED_TO_READ_FILE_MESSAGE,
    FAILED_TO_CREATE_FILE_MESSAGE
)

T = TypeVar("T")
R = TypeVar("R")


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
            result = subprocess.run(["command", "-v", command], capture_output=True, text=True, check=False)
            return result.returncode == 0
        except Exception:
            return False
    
    @staticmethod
    def get_public_ip():
        try:
            response = requests.get('https://api.ipify.org', timeout=10)
            response.raise_for_status()  # fail on non-2xx
            return response.text.strip()
        except requests.RequestException:
            raise Exception(FAILED_TO_GET_PUBLIC_IP_MESSAGE)


class ParallelProcessor:
    @staticmethod
    def process_items(
        items: List[T],
        processor_func: Callable[[T], R],
        max_workers: int = 50,
        error_handler: Callable[[T, Exception], R] = None,
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


class FileManager:
    @staticmethod
    def create_file(file_path: str, mode: int = stat.S_IRUSR | stat.S_IWUSR | stat.S_IRGRP | stat.S_IROTH, logger=None) -> Tuple[bool, Optional[str]]:
        try:
            with open(file_path, "w") as f:
                f.write("")
            return True, None
        except Exception as e:
            error_msg = FAILED_TO_CREATE_FILE_MESSAGE.format(file_path=file_path, error=e)
            if logger:
                logger.error(error_msg)
            return False, error_msg

    @staticmethod
    def set_permissions(file_path: str, mode: int, logger=None) -> Tuple[bool, Optional[str]]:
        try:
            if logger:
                logger.debug(SETTING_PERMISSIONS_MESSAGE.format(mode=oct(mode), file_path=file_path))

            os.chmod(file_path, mode)

            if logger:
                logger.debug(PERMISSIONS_SET_SUCCESS_MESSAGE)
            return True, None
        except Exception as e:
            error_msg = FAILED_TO_SET_PERMISSIONS_MESSAGE.format(file_path=file_path, error=e)
            if logger:
                logger.error(error_msg)
            return False, error_msg

    @staticmethod
    def create_directory(
        path: str, mode: int = stat.S_IRUSR | stat.S_IWUSR | stat.S_IXUSR, logger=None
    ) -> Tuple[bool, Optional[str]]:
        try:
            if not os.path.exists(path):
                os.makedirs(path, mode=mode)
                if logger:
                    logger.debug(CREATED_DIRECTORY_MESSAGE.format(path=path))
            return True, None
        except Exception as e:
            error_msg = FAILED_TO_CREATE_DIRECTORY_MESSAGE.format(path=path, error=e)
            if logger:
                logger.error(error_msg)
            return False, error_msg

    @staticmethod
    def append_to_file(
        file_path: str, content: str, mode: int = stat.S_IRUSR | stat.S_IWUSR | stat.S_IRGRP | stat.S_IROTH, logger=None
    ) -> Tuple[bool, Optional[str]]:
        try:
            with open(file_path, "a") as f:
                f.write(f"\n{content}\n")

            FileManager.set_permissions(file_path, mode, logger)

            if logger:
                logger.debug(CONTENT_APPENDED_MESSAGE.format(file_path=file_path))
            return True, None
        except Exception as e:
            error_msg = FAILED_TO_APPEND_MESSAGE.format(file_path=file_path, error=e)
            if logger:
                logger.error(error_msg)
            return False, error_msg

    @staticmethod
    def read_file_content(file_path: str, logger=None) -> Tuple[bool, Optional[str], Optional[str]]:
        try:
            with open(file_path, "r") as f:
                content = f.read().strip()
            return True, content, None
        except Exception as e:
            error_msg = FAILED_TO_READ_FILE_MESSAGE.format(file_path=file_path, error=e)
            if logger:
                logger.error(error_msg)
            return False, None, error_msg

    @staticmethod
    def expand_user_path(path: str) -> str:
        return os.path.expanduser(path)

    @staticmethod
    def get_directory_path(file_path: str) -> str:
        return os.path.dirname(file_path)

    @staticmethod
    def get_public_key_path(private_key_path: str) -> str:
        return f"{private_key_path}.pub"
