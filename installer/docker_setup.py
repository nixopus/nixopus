import os
import subprocess
import json
import time
import shutil
from pathlib import Path
import socket

class DockerSetup:
    def __init__(self, env="staging"):
        self.env = env
        self.context_name = f"nixopus-{self.env}"
        self.config_dir = Path("/etc/nixopus-staging") if env == "staging" else Path("/etc/nixopus")
        self.docker_certs_dir = self.config_dir / "docker-certs"

    def get_local_ip(self):
        try:
            s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
            s.connect(("8.8.8.8", 80))
            local_ip = s.getsockname()[0]
            s.close()
            return local_ip
        except Exception:
            return "localhost"

    def setup_docker_certs(self):
        self.docker_certs_dir.mkdir(parents=True, exist_ok=True)
        local_ip = self.get_local_ip()
        
        try:
            result = subprocess.run([
                "openssl", "genrsa", "-out", str(self.docker_certs_dir / "ca-key.pem"), "4096"
            ], capture_output=True, text=True)
            if result.returncode != 0:
                raise Exception("Failed to generate CA key")

            result = subprocess.run([
                "openssl", "req", "-new", "-x509", "-days", "365",
                "-key", str(self.docker_certs_dir / "ca-key.pem"),
                "-sha256", "-out", str(self.docker_certs_dir / "ca.pem"),
                "-subj", f"/CN={self.context_name}"
            ], capture_output=True, text=True)
            if result.returncode != 0:
                raise Exception("Failed to generate CA certificate")

            result = subprocess.run([
                "openssl", "genrsa", "-out", str(self.docker_certs_dir / "server-key.pem"), "4096"
            ], capture_output=True, text=True)
            if result.returncode != 0:
                raise Exception("Failed to generate server key")

            with open(self.docker_certs_dir / "extfile.cnf", "w") as f:
                f.write(f"subjectAltName = DNS:localhost,IP:{local_ip},IP:127.0.0.1\n")

            result = subprocess.run([
                "openssl", "req", "-subj", f"/CN={local_ip}", "-new",
                "-key", str(self.docker_certs_dir / "server-key.pem"),
                "-out", str(self.docker_certs_dir / "server.csr")
            ], capture_output=True, text=True)
            if result.returncode != 0:
                raise Exception("Failed to generate server CSR")

            result = subprocess.run([
                "openssl", "x509", "-req", "-days", "365",
                "-in", str(self.docker_certs_dir / "server.csr"),
                "-CA", str(self.docker_certs_dir / "ca.pem"),
                "-CAkey", str(self.docker_certs_dir / "ca-key.pem"),
                "-CAcreateserial", "-out", str(self.docker_certs_dir / "server-cert.pem"),
                "-extfile", str(self.docker_certs_dir / "extfile.cnf")
            ], capture_output=True, text=True)
            if result.returncode != 0:
                raise Exception("Failed to generate server certificate")

            result = subprocess.run([
                "openssl", "genrsa", "-out", str(self.docker_certs_dir / "key.pem"), "4096"
            ], capture_output=True, text=True)
            if result.returncode != 0:
                raise Exception("Failed to generate client key")

            result = subprocess.run([
                "openssl", "req", "-subj", f"/CN={self.context_name}", "-new",
                "-key", str(self.docker_certs_dir / "key.pem"),
                "-out", str(self.docker_certs_dir / "client.csr")
            ], capture_output=True, text=True)
            if result.returncode != 0:
                raise Exception("Failed to generate client CSR")

            result = subprocess.run([
                "openssl", "x509", "-req", "-days", "365",
                "-in", str(self.docker_certs_dir / "client.csr"),
                "-CA", str(self.docker_certs_dir / "ca.pem"),
                "-CAkey", str(self.docker_certs_dir / "ca-key.pem"),
                "-CAcreateserial", "-out", str(self.docker_certs_dir / "cert.pem")
            ], capture_output=True, text=True)
            if result.returncode != 0:
                raise Exception("Failed to generate client certificate")

            for cert_file in self.docker_certs_dir.glob("*"):
                cert_file.chmod(0o600)

        except Exception as e:
            raise Exception(f"Error setting up Docker certificates: {str(e)}")

    def copy_certs_to_docker_dir(self):
        docker_certs_dir = Path("/etc/docker")
        docker_certs_dir.mkdir(parents=True, exist_ok=True)
        
        certs_to_copy = ["ca.pem", "server-cert.pem", "server-key.pem", "cert.pem", "key.pem"]
        
        for cert in certs_to_copy:
            src = self.docker_certs_dir / cert
            dst = docker_certs_dir / cert
            if src.exists():
                shutil.copy2(src, dst)
                dst.chmod(0o600)

    def setup_docker_daemon_for_tcp(self):
        docker_config_dir = Path("/etc/docker")
        docker_config_dir.mkdir(parents=True, exist_ok=True)
        
        docker_port = 2376 if self.env == "production" else 2377
        
        # Check if port is already in use
        try:
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.bind(('0.0.0.0', docker_port))
            sock.close()
        except socket.error:
            raise Exception(f"Port {docker_port} is already in use")
        
        daemon_config = {
            "hosts": [f"tcp://0.0.0.0:{docker_port}", "unix:///var/run/docker.sock"],
            "tls": True,
            "tlsverify": True,
            "tlscacert": str(self.docker_certs_dir / "ca.pem"),
            "tlscert": str(self.docker_certs_dir / "server-cert.pem"),
            "tlskey": str(self.docker_certs_dir / "server-key.pem")
        }
        
        daemon_json_path = docker_config_dir / "daemon.json"
        
        # Backup existing config if it exists
        if daemon_json_path.exists():
            backup_path = daemon_json_path.with_suffix('.json.bak')
            shutil.copy2(daemon_json_path, backup_path)
        
        with open(daemon_json_path, "w") as f:
            json.dump(daemon_config, f, indent=2)
        
        try:
            subprocess.run(["systemctl", "stop", "docker"], 
                         capture_output=True, text=True, check=True)

            time.sleep(2)
            
            result = subprocess.run(["systemctl", "start", "docker"], 
                                  capture_output=True, text=True)
            if result.returncode != 0:
                raise Exception("Failed to start Docker service")
            
            time.sleep(5)
            
            result = subprocess.run(["systemctl", "is-active", "docker"], 
                                  capture_output=True, text=True)
            if result.returncode != 0:
                # If service failed, restore backup and restart
                if daemon_json_path.with_suffix('.json.bak').exists():
                    shutil.copy2(daemon_json_path.with_suffix('.json.bak'), daemon_json_path)
                    subprocess.run(["systemctl", "start", "docker"], 
                                 capture_output=True, text=True)
                raise Exception("Docker service failed to start properly")
            
        except Exception as e:
            result = subprocess.run(["journalctl", "-u", "docker", "-n", "50"], 
                                  capture_output=True, text=True)
            error_logs = result.stdout if result.returncode == 0 else "Failed to get logs"
            raise Exception(f"Failed to manage Docker service. Error: {str(e)}\nLogs: {error_logs}")

    def create_docker_context(self):
        docker_port = 2376 if self.env == "production" else 2377
        local_ip = self.get_local_ip()
        
        try:
            result = subprocess.run(["docker", "context", "ls"], 
                                  capture_output=True, text=True)
            if result.returncode != 0:
                raise Exception("Docker not available or doesn't support contexts")
                
            subprocess.run(["docker", "context", "rm", "-f", self.context_name], 
                         capture_output=True, check=False)
            
            self.copy_certs_to_docker_dir()
            
            result = subprocess.run([
                "docker", "context", "create", self.context_name,
                "--docker", f"host=tcp://{local_ip}:{docker_port}",
                "--docker", f"ca={self.docker_certs_dir / 'ca.pem'}",
                "--docker", f"cert={self.docker_certs_dir / 'cert.pem'}",
                "--docker", f"key={self.docker_certs_dir / 'key.pem'}"
            ], capture_output=True, text=True)
            
            if result.returncode != 0:
                raise Exception(f"Failed to create Docker context: {result.stderr}")
                
            subprocess.run(["docker", "context", "use", self.context_name],
                         capture_output=True, text=True)
            
            test_result = subprocess.run(
                ["docker", "version", "--format", "{{.Server.Version}}"],
                capture_output=True, text=True
            )
            
            if test_result.returncode != 0:
                for cert_file in self.docker_certs_dir.glob("*"):
                    cert_file.chmod(0o600)
                
                subprocess.run(["systemctl", "restart", "docker"], 
                             capture_output=True, check=False)
                
                time.sleep(5)
                
                test_result = subprocess.run(
                    ["docker", "version", "--format", "{{.Server.Version}}"],
                    capture_output=True, text=True
                )
                
                if test_result.returncode != 0:
                    raise Exception("Failed to establish secure connection to Docker daemon")
                
            return self.context_name
            
        except Exception as e:
            raise Exception(f"Error setting up Docker context: {str(e)}")

    def setup(self):
        self.setup_docker_certs()
        self.copy_certs_to_docker_dir()
        self.setup_docker_daemon_for_tcp()
        return self.create_docker_context() 