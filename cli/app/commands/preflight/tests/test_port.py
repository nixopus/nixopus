import pytest
from typing import List
from app.commands.preflight.port import PortConfig, PortCheckResult, PortService

class TestPort:
    def test_valid_ports(self):
        ports = [80, 443, 8080]
        config = PortConfig(ports=ports)
        assert config.ports == [80, 443, 8080]

    def test_empty_ports_list(self):
        with pytest.raises(ValueError):
            PortConfig(ports=[])

    def test_valid_host_localhost(self):
        config = PortConfig(ports=[80], host="localhost")
        assert config.host == "localhost"

    def test_valid_host_ipv4(self):
        config = PortConfig(ports=[80], host="192.168.1.1")
        assert config.host == "192.168.1.1"

    def test_valid_host_ipv4_loopback(self):
        config = PortConfig(ports=[80], host="127.0.0.1")
        assert config.host == "127.0.0.1"

    def test_valid_host_domain(self):
        config = PortConfig(ports=[80], host="example.com")
        assert config.host == "example.com"

    def test_valid_host_subdomain(self):
        config = PortConfig(ports=[80], host="api.example.com")
        assert config.host == "api.example.com"

    def test_valid_host_domain_with_hyphens(self):
        config = PortConfig(ports=[80], host="my-domain.com")
        assert config.host == "my-domain.com"

    def test_invalid_host_invalid_ip(self):
        with pytest.raises(ValueError, match="Host must be 'localhost', a valid IP address, or a valid domain name"):
            PortConfig(ports=[80], host="256.256.256.256")

    def test_invalid_host_empty(self):
        with pytest.raises(ValueError):
            PortConfig(ports=[80], host="")

    def test_invalid_host_invalid_domain(self):
        with pytest.raises(ValueError, match="Host must be 'localhost', a valid IP address, or a valid domain name"):
            PortConfig(ports=[80], host="invalid..domain")

    def test_check_ports_basic(self):
        config = PortConfig(ports=[80, 443], host="localhost", timeout=1, verbose=False)
        port_service = PortService(config)
        results = port_service.check_ports()
        assert len(results) == 2
        assert all("port" in result for result in results)
        assert all("status" in result for result in results)
        assert all(result["host"] is None for result in results)
        assert all(result["error"] is None for result in results)
        assert all(result["is_available"] is True for result in results)

    def test_check_ports_verbose(self):
        config = PortConfig(ports=[80, 443], host="localhost", timeout=1, verbose=True)
        port_service = PortService(config)
        results = port_service.check_ports()
        assert len(results) == 2
        assert all("port" in result for result in results)
        assert all("status" in result for result in results)
        assert all("host" in result for result in results)
        assert all(result["host"] == "localhost" for result in results)
        assert all(result["error"] is None for result in results)
        assert all(result["is_available"] is True for result in results)

def test_port_check_result_type():
    """Test that PortCheckResult has correct structure"""
    result: PortCheckResult = {
        "port": 8080,
        "status": "available",
        "host": "localhost",
        "error": None,
        "is_available": True
    }
    
    assert isinstance(result["port"], int)
    assert isinstance(result["status"], str)
    assert isinstance(result["host"], str) or result["host"] is None
    assert isinstance(result["error"], str) or result["error"] is None
    assert isinstance(result["is_available"], bool)

def test_check_ports_return_type():
    """Test that check_ports returns correct type"""
    config = PortConfig(ports=[8080, 3000], host="localhost", timeout=1, verbose=False)
    port_service = PortService(config)
    results: List[PortCheckResult] = port_service.check_ports()
    
    assert isinstance(results, list)
    for result in results:
        assert isinstance(result, dict)
        assert "port" in result
        assert "status" in result
        assert "host" in result
        assert "error" in result
        assert "is_available" in result
        assert isinstance(result["port"], int)
        assert isinstance(result["status"], str)
        assert isinstance(result["host"], str) or result["host"] is None
        assert isinstance(result["error"], str) or result["error"] is None
        assert isinstance(result["is_available"], bool)
