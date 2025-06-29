import os
import subprocess
import json
import time
import shutil
from pathlib import Path
import socket
import requests

class DockerSetup:
    def __init__(self, env="staging", debug=False):
        self.env = env
        self.debug = debug
        self.context_name = f"nixopus-{self.env}"
        # store all the docker config in their respective docker-certs dir
        self.config_dir = Path("/etc/nixopus-staging") if env == "staging" else Path("/etc/nixopus")
        self.docker_certs_dir = self.config_dir / "docker-certs"

    def debug_print(self, message):
        if self.debug:
            print(f"[DEBUG] {message}")

    def get_public_ip(self):
        try:
            response = requests.get('https://api.ipify.org', timeout=10)
            response.raise_for_status()  # fail on non-2xx
            return response.text.strip()
        except requests.RequestException:
            self.debug_print("Failed to get public IP, falling back to localhost")
            return "localhost"
    
    def get_local_ip(self):
        try:
            response = socket.gethostbyname(socket.gethostname())
            return response
        except socket.gaierror:
            self.debug_print("Failed to get local IP, falling back to localhost")
            return "localhost"

    def setup_docker_certs(self):
        self.debug_print("Setting up Docker certificates...")
        self.docker_certs_dir.mkdir(parents=True, exist_ok=True)
        local_ip = self.get_local_ip()
        
        try:
            self.debug_print("Generating CA key...")
            result = subprocess.run([
                "openssl", "genrsa", "-out", str(self.docker_certs_dir / "ca-key.pem"), "4096"
            ], capture_output=True, text=True)
            if result.returncode != 0:
                raise Exception("Failed to generate CA key")

            self.debug_print("Generating CA certificate...")
            result = subprocess.run([
                "openssl", "req", "-new", "-x509", "-days", "365",
                "-key", str(self.docker_certs_dir / "ca-key.pem"),
                "-sha256", "-out", str(self.docker_certs_dir / "ca.pem"),
                "-subj", f"/CN={self.context_name}"
            ], capture_output=True, text=True)
            if result.returncode != 0:
                raise Exception("Failed to generate CA certificate")

            self.debug_print("Generating server key...")
            result = subprocess.run([
                "openssl", "genrsa", "-out", str(self.docker_certs_dir / "server-key.pem"), "4096"
            ], capture_output=True, text=True)
            if result.returncode != 0:
                raise Exception("Failed to generate server key")

            with open(self.docker_certs_dir / "extfile.cnf", "w") as f:
                f.write(f"subjectAltName = DNS:localhost,IP:{local_ip},IP:127.0.0.1\n")

            self.debug_print("Generating server CSR...")
            result = subprocess.run([
                "openssl", "req", "-subj", f"/CN={local_ip}", "-new",
                "-key", str(self.docker_certs_dir / "server-key.pem"),
                "-out", str(self.docker_certs_dir / "server.csr")
            ], capture_output=True, text=True)
            if result.returncode != 0:
                raise Exception("Failed to generate server CSR")

            self.debug_print("Generating server certificate...")
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

            self.debug_print("Generating client key...")
            result = subprocess.run([
                "openssl", "genrsa", "-out", str(self.docker_certs_dir / "key.pem"), "4096"
            ], capture_output=True, text=True)
            if result.returncode != 0:
                raise Exception("Failed to generate client key")

            self.debug_print("Generating client CSR...")
            result = subprocess.run([
                "openssl", "req", "-subj", f"/CN={self.context_name}", "-new",
                "-key", str(self.docker_certs_dir / "key.pem"),
                "-out", str(self.docker_certs_dir / "client.csr")
            ], capture_output=True, text=True)
            if result.returncode != 0:
                raise Exception("Failed to generate client CSR")

            self.debug_print("Generating client certificate...")
            result = subprocess.run([
                "openssl", "x509", "-req", "-days", "365",
                "-in", str(self.docker_certs_dir / "client.csr"),
                "-CA", str(self.docker_certs_dir / "ca.pem"),
                "-CAkey", str(self.docker_certs_dir / "ca-key.pem"),
                "-CAcreateserial", "-out", str(self.docker_certs_dir / "cert.pem")
            ], capture_output=True, text=True)
            if result.returncode != 0:
                raise Exception("Failed to generate client certificate")

            self.debug_print("Setting certificate permissions...")
            for cert_file in self.docker_certs_dir.glob("*"):
                cert_file.chmod(0o600)

        except Exception as e:
            raise Exception(f"Error setting up Docker certificates: {str(e)}")

    def setup_docker_systemd_override(self):
        self.debug_print("Setting up Docker systemd override...")
        override_dir = Path("/etc/systemd/system/docker.service.d")
        override_file = override_dir / "override.conf"
        
        try:
            override_dir.mkdir(parents=True, exist_ok=True)
            
            override_content = """# Disable flags to dockerd, all settings are done in /etc/docker/daemon.json
[Service]
ExecStart=
ExecStart=/usr/bin/dockerd"""
            
            override_file.write_text(override_content)
            
            self.debug_print("Reloading systemd daemon...")
            subprocess.run(["systemctl", "daemon-reload"], check=True)
            self.debug_print("Restarting Docker service...")
            subprocess.run(["systemctl", "restart", "docker"], check=True)
            
        except Exception as e:
            raise Exception(f"Failed to setup Docker systemd override: {str(e)}")

    def setup_docker_daemon_for_tcp(self):
        self.debug_print("Setting up Docker daemon for TCP...")
        docker_config_dir = Path("/etc/docker")
        docker_config_dir.mkdir(parents=True, exist_ok=True)
        
        docker_port = 2376 if self.env == "production" else 2377
        
        self.debug_print(f"Checking if port {docker_port} is available...")
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
        
        if daemon_json_path.exists():
            self.debug_print("Backing up existing daemon.json...")
            backup_path = daemon_json_path.with_suffix('.json.bak')
            shutil.copy2(daemon_json_path, backup_path)
        
        self.debug_print("Writing new daemon.json configuration...")
        with open(daemon_json_path, "w") as f:
            json.dump(daemon_config, f, indent=2)
        
        try:
            self.debug_print("Stopping Docker service...")
            subprocess.run(["systemctl", "stop", "docker"], 
                         capture_output=True, text=True, check=True)

            time.sleep(2)
            
            self.debug_print("Starting Docker service...")
            result = subprocess.run(["systemctl", "start", "docker"], 
                                  capture_output=True, text=True)
            if result.returncode != 0:
                raise Exception("Failed to start Docker service")
            
            time.sleep(5)
            
            self.debug_print("Checking Docker service status...")
            result = subprocess.run(["systemctl", "is-active", "docker"], 
                                  capture_output=True, text=True)
            if result.returncode != 0:
                if daemon_json_path.with_suffix('.json.bak').exists():
                    self.debug_print("Restoring backup configuration...")
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
        self.debug_print("Creating Docker context...")
        docker_port = 2376 if self.env == "production" else 2377
        local_ip = self.get_local_ip()
        
        try:
            self.debug_print("Checking Docker context support...")
            result = subprocess.run(["docker", "context", "ls"], 
                                  capture_output=True, text=True)
            if result.returncode != 0:
                raise Exception("Docker not available or doesn't support contexts")
                
            self.debug_print(f"Removing existing context {self.context_name} if it exists...")
            subprocess.run(["docker", "context", "rm", "-f", self.context_name], 
                         capture_output=True, check=False)
            
            self.debug_print("Creating new Docker context...")
            result = subprocess.run([
                "docker", "context", "create", self.context_name,
                "--docker", f"host=tcp://{local_ip}:{docker_port}",
                "--docker", f"ca={self.docker_certs_dir / 'ca.pem'}",
                "--docker", f"cert={self.docker_certs_dir / 'cert.pem'}",
                "--docker", f"key={self.docker_certs_dir / 'key.pem'}"
            ], capture_output=True, text=True)
            
            if result.returncode != 0:
                raise Exception(f"Failed to create Docker context: {result.stderr}")
                
            self.debug_print(f"Switching to context {self.context_name}...")
            subprocess.run(["docker", "context", "use", self.context_name],
                         capture_output=True, text=True)
            
            self.debug_print("Verifying context switch...")
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
            
            self.debug_print("Checking Docker contexts...")
            context_result = subprocess.run(
                ["docker", "context", "ls"],
                capture_output=True,
                text=True
            )
            
            self.debug_print("Checking systemd status...")
            systemd_result = subprocess.run(
                ["systemctl", "status", "docker"],
                capture_output=True,
                text=True
            )
            
            self.debug_print("Checking Docker daemon status...")
            daemon_result = subprocess.run(
                ["docker", "info"],
                capture_output=True,
                text=True
            )
            
            if context_result.returncode != 0:
                self.debug_print(f"Error listing Docker contexts:\n{context_result.stderr}")
                return {
                    "status": "error",
                    "message": "Failed to list Docker contexts",
                    "error": context_result.stderr
                }
                
            if daemon_result.returncode != 0:
                self.debug_print(f"Error checking Docker daemon:\n{daemon_result.stderr}")
                return {
                    "status": "error",
                    "message": "Docker daemon is not running",
                    "error": daemon_result.stderr
                }
            
            self.debug_print("Docker contexts:")
            self.debug_print(context_result.stdout)
            self.debug_print("Docker daemon info:")
            self.debug_print(daemon_result.stdout)
                
            return {
                "status": "success",
                "contexts": context_result.stdout,
                "daemon_info": daemon_result.stdout
            }
            
        except Exception as e:
            self.debug_print(f"Error testing Docker context: {str(e)}")
            return {
                "status": "error",
                "message": "Failed to test Docker context",
                "error": str(e)
            }

    def setup(self):
        self.debug_print("Starting Docker setup...")
        self.setup_docker_certs()
        self.setup_docker_systemd_override()
        self.setup_docker_daemon_for_tcp()
        self.create_docker_context()
        time.sleep(20)
        return self.test_docker_context_output()