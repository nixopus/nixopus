import ipaddress
import re


def validate_domains(api_domain: str = None, view_domain: str = None) -> None:
    if (api_domain is None) != (view_domain is None):
        raise ValueError("Both api_domain and view_domain must be provided together, or neither should be provided")

    if api_domain and view_domain:
        domain_pattern = re.compile(
            r"^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?))*$"
        )
        if not domain_pattern.match(api_domain) or not domain_pattern.match(view_domain):
            raise ValueError("Invalid domain format. Domains must be valid hostnames")


def validate_repo(repo: str) -> None:
    if repo:
        if not (
            repo.startswith(("http://", "https://", "git://", "ssh://"))
            or (repo.endswith(".git") and not repo.startswith("github.com:"))
            or ("@" in repo and ":" in repo and repo.count("@") == 1)
        ):
            raise ValueError("Invalid repository URL format")


def validate_host_ip(host_ip: str) -> None:
    if host_ip:
        try:
            ipaddress.ip_address(host_ip)
        except ValueError:
            raise ValueError(f"Invalid IP address format: {host_ip}")

