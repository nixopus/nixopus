import platform
import shutil

import requests

from app.utils.message import FAILED_TO_GET_PUBLIC_IP_MESSAGE
from app.utils.supported import SupportedOS, SupportedPackageManager


def get_os_name() -> str:
    """Get the operating system name"""
    return platform.system().lower()


def command_exists(command: str) -> bool:
    """Check if a command exists in the system PATH"""
    return shutil.which(command) is not None


def get_package_manager() -> str:
    """Detect and return the package manager for the current system"""
    os_name = get_os_name()

    if os_name == SupportedOS.MACOS.value:
        return SupportedPackageManager.BREW.value

    package_managers = [pm.value for pm in SupportedPackageManager if pm != SupportedPackageManager.BREW]

    for pm in package_managers:
        if command_exists(pm):
            return pm
    raise RuntimeError("No supported package manager found on this system. Please install one or specify it manually.")


def get_public_ip() -> str:
    """Get the public IP address of the current system"""
    try:
        response = requests.get("https://api.ipify.org", timeout=10)
        response.raise_for_status()  # fail on non-2xx
        return response.text.strip()
    except requests.RequestException:
        raise Exception(FAILED_TO_GET_PUBLIC_IP_MESSAGE)

