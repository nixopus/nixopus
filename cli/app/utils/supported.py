from enum import Enum


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


def is_supported_os(os_name: str) -> bool:
    """Check if OS is supported"""
    return os_name in [os.value for os in SupportedOS]


def is_supported_distribution(distribution: str) -> bool:
    """Check if distribution is supported"""
    return distribution in [dist.value for dist in SupportedDistribution]


def is_supported_package_manager(package_manager: str) -> bool:
    """Check if package manager is supported"""
    return package_manager in [pm.value for pm in SupportedPackageManager]


def get_supported_os_list() -> list[str]:
    """Get list of supported OS names"""
    return [os.value for os in SupportedOS]


def get_supported_distributions_list() -> list[str]:
    """Get list of supported distribution names"""
    return [dist.value for dist in SupportedDistribution]

