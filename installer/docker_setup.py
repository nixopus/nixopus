import os
import subprocess
import json
import time
import shutil
from pathlib import Path
import socket
import requests

class DockerSetup:
    def __init__(self, env="staging"):
        self.env = env
        self.context_name = f"nixopus-{self.env}"
        # store all the docker config in their respective docker-certs dir
        self.config_dir = Path("/etc/nixopus-staging") if env == "staging" else Path("/etc/nixopus")
        self.docker_certs_dir = self.config_dir / "docker-certs"

    def get_public_ip(self):
        try:
            response = requests.get('https://api.ipify.org', timeout=10)
            response.raise_for_status()  # fail on non-2xx
            return response.text.strip()
        except requests.RequestException:
            print("Failed to get public IP, falling back to localhost")
            return "localhost"
    
    def get_local_ip(self):
        try:
            response = socket.gethostbyname(socket.gethostname())
            return response
        except socket.gaierror:
            print("Failed to get local IP, falling back to localhost")
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

    def setup_docker_systemd_override(self):
        override_dir = Path("/etc/systemd/system/docker.service.d")
        override_file = override_dir / "override.conf"
        
        try:
            override_dir.mkdir(parents=True, exist_ok=True)
            
            override_content = """# Disable flags to dockerd, all settings are done in /etc/docker/daemon.json
[Service]
ExecStart=
ExecStart=/usr/bin/dockerd"""
            
            override_file.write_text(override_content)
            
            subprocess.run(["systemctl", "daemon-reload"], check=True)
            subprocess.run(["systemctl", "restart", "docker"], check=True)
            
        except Exception as e:
            raise Exception(f"Failed to setup Docker systemd override: {str(e)}")

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
            
            current_context = subprocess.run(["docker", "context", "ls", "--format", "{{.Name}} {{.Current}}"],
                                           capture_output=True, text=True)
            
            if f"{self.context_name} true" not in current_context.stdout:
                raise Exception(f"Failed to switch to context {self.context_name}. Current contexts:\n{current_context.stdout}")
                
            return self.context_name
            
        except Exception as e:
            raise Exception(f"Error setting up Docker context: {str(e)}")
    
    def test_docker_context_output(self):
        try:
            local_ip = self.get_local_ip()
            docker_port = 2376 if self.env == "production" else 2377
            
            os.environ["DOCKER_HOST"] = f"tcp://{local_ip}:{docker_port}"
            os.environ["DOCKER_TLS_VERIFY"] = "1"
            os.environ["DOCKER_CERT_PATH"] = str(self.docker_certs_dir)
            print("\nChecking Docker contexts...")
            context_result = subprocess.run(
                ["docker", "context", "ls"],
                capture_output=True,
                text=True
            )
            
            print("Checking systemd")
            systemd_result = subprocess.run(
                ["systemctl", "status", "docker"],
                capture_output=True,
                text=True
            )
            print(f"\n{systemd_result.stdout}")
            
            print("\nChecking Docker daemon status...")
            daemon_result = subprocess.run(
                ["docker", "info"],
                capture_output=True,
                text=True
            )
            
            if context_result.returncode != 0:
                print(f"\nError listing Docker contexts:\n{context_result.stderr}")
                return {
                    "status": "error",
                    "message": "Failed to list Docker contexts",
                    "error": context_result.stderr
                }
                
            if daemon_result.returncode != 0:
                print(f"\nError checking Docker daemon:\n{daemon_result.stderr}")
                return {
                    "status": "error",
                    "message": "Docker daemon is not running",
                    "error": daemon_result.stderr
                }
            
            print("\nDocker contexts:")
            print(context_result.stdout)
            print("\nDocker daemon info:")
            print(daemon_result.stdout)
                
            return {
                "status": "success",
                "contexts": context_result.stdout,
                "daemon_info": daemon_result.stdout
            }
            
        except Exception as e:
            print(f"\nError testing Docker context: {str(e)}")
            return {
                "status": "error",
                "message": "Failed to test Docker context",
                "error": str(e)
            }

    def setup(self):
        self.setup_docker_certs()
        self.setup_docker_systemd_override()
        self.setup_docker_daemon_for_tcp()
        self.create_docker_context()
        time.sleep(20)
        return self.test_docker_context_output()