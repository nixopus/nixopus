import os
import secrets
import string
from pathlib import Path
import subprocess
import socket
import json
import time

class EnvironmentSetup:
    def __init__(self, domains, env="staging"):
        self.domains = domains
        self.project_root = Path(__file__).parent.parent
        self.env_file = self.project_root / ".env"
        self.env_sample = self.project_root / ".env.sample"
        # we will use different config directories for staging and production
        self.config_dir = Path("/etc/nixopus-staging") if env == "staging" else Path("/etc/nixopus")
        self.docker_certs_dir = self.config_dir / "docker-certs"
        self.ssh_dir = self.config_dir / "ssh"
        self.env = env
        self.context_name = f"nixopus-{self.env}"

    def generate_random_string(self, length=12):
        alphabet = string.ascii_letters + string.digits
        return ''.join(secrets.choice(alphabet) for _ in range(length))
    
    def setup_authorized_keys(self):
        """Add the generated SSH public key to authorized_keys file."""
        try:
            ssh_config_dir = Path.home() / ".ssh"
            ssh_config_dir.mkdir(mode=0o700, parents=True, exist_ok=True)
            authorized_keys_path = ssh_config_dir / "authorized_keys"
            
            _, public_key_path = self.generate_ssh_key()
            with open(public_key_path, 'r') as pk_file:
                public_key_content = pk_file.read().strip()
                
            if authorized_keys_path.exists():
                with open(authorized_keys_path, 'r') as auth_file:
                    existing_content = auth_file.read()
                    if public_key_content in existing_content:
                        # if the key is already in the file, return
                        return
                        
            with open(authorized_keys_path, 'a+') as auth_file:
                auth_file.write(f"\n{public_key_content}\n")
                
            authorized_keys_path.chmod(0o600)
            # print(f"Added SSH key to {authorized_keys_path}")
            
        except Exception as e:
            print(f"Error setting up authorized_keys: {str(e)}")
            raise

    def generate_ssh_key(self):
        self.ssh_dir.mkdir(parents=True, exist_ok=True)
        private_key_path = self.ssh_dir / "id_rsa"
        public_key_path = self.ssh_dir / "id_rsa.pub"
        
        if not private_key_path.exists():
            try:
                result = subprocess.run(
                    ["ssh-keygen", "-t", "rsa", "-b", "4096", "-f", str(private_key_path), "-N", ""],
                    capture_output=True,
                    text=True
                )
                if result.returncode != 0:
                    print("Error generating SSH key:")
                    print(result.stderr)
                    raise Exception("Failed to generate SSH key")
            except Exception as e:
                print(f"Error generating SSH key: {str(e)}")
                raise
        
        return private_key_path, public_key_path

    def setup_docker_certs(self):
        self.docker_certs_dir.mkdir(parents=True, exist_ok=True)
        
        local_ip = self.get_local_ip()
        
        try:
            result = subprocess.run([
                "openssl", "genrsa", "-out", str(self.docker_certs_dir / "ca-key.pem"), "4096"
            ], capture_output=True, text=True)
            if result.returncode != 0:
                print("Error generating CA key:")
                print(result.stderr)
                raise Exception("Failed to generate CA key")

            result = subprocess.run([
                "openssl", "req", "-new", "-x509", "-days", "365",
                "-key", str(self.docker_certs_dir / "ca-key.pem"),
                "-sha256", "-out", str(self.docker_certs_dir / "ca.pem"),
                "-subj", f"/CN={self.context_name}"
            ], capture_output=True, text=True)
            if result.returncode != 0:
                print("Error generating CA certificate:")
                print(result.stderr)
                raise Exception("Failed to generate CA certificate")

            result = subprocess.run([
                "openssl", "genrsa", "-out", str(self.docker_certs_dir / "server-key.pem"), "4096"
            ], capture_output=True, text=True)
            if result.returncode != 0:
                print("Error generating server key:")
                print(result.stderr)
                raise Exception("Failed to generate server key")

            with open(self.docker_certs_dir / "extfile.cnf", "w") as f:
                f.write(f"subjectAltName = DNS:localhost,IP:{local_ip},IP:127.0.0.1\n")

            result = subprocess.run([
                "openssl", "req", "-subj", f"/CN={local_ip}", "-new",
                "-key", str(self.docker_certs_dir / "server-key.pem"),
                "-out", str(self.docker_certs_dir / "server.csr")
            ], capture_output=True, text=True)
            if result.returncode != 0:
                print("Error generating server CSR:")
                print(result.stderr)
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
                print("Error generating server certificate:")
                print(result.stderr)
                raise Exception("Failed to generate server certificate")

            result = subprocess.run([
                "openssl", "genrsa", "-out", str(self.docker_certs_dir / "key.pem"), "4096"
            ], capture_output=True, text=True)
            if result.returncode != 0:
                print("Error generating client key:")
                print(result.stderr)
                raise Exception("Failed to generate client key")

            result = subprocess.run([
                "openssl", "req", "-subj", f"/CN={self.context_name}", "-new",
                "-key", str(self.docker_certs_dir / "key.pem"),
                "-out", str(self.docker_certs_dir / "client.csr")
            ], capture_output=True, text=True)
            if result.returncode != 0:
                print("Error generating client CSR:")
                print(result.stderr)
                raise Exception("Failed to generate client CSR")

            result = subprocess.run([
                "openssl", "x509", "-req", "-days", "365",
                "-in", str(self.docker_certs_dir / "client.csr"),
                "-CA", str(self.docker_certs_dir / "ca.pem"),
                "-CAkey", str(self.docker_certs_dir / "ca-key.pem"),
                "-CAcreateserial", "-out", str(self.docker_certs_dir / "cert.pem")
            ], capture_output=True, text=True)
            if result.returncode != 0:
                print("Error generating client certificate:")
                print(result.stderr)
                raise Exception("Failed to generate client certificate")

            for cert_file in self.docker_certs_dir.glob("*"):
                cert_file.chmod(0o600)

        except Exception as e:
            print(f"Error setting up Docker certificates: {str(e)}")
            raise

    def get_local_ip(self):
        try:
            s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
            s.connect(("8.8.8.8", 80))
            local_ip = s.getsockname()[0]
            s.close()
            return local_ip
        except Exception:
            return "localhost"
    
    def create_docker_context(self):
        """Create a Docker context for this environment instead of modifying the daemon."""
        docker_port = 2376 if self.env == "production" else 2377
        local_ip = self.get_local_ip()
        
        try:
            result = subprocess.run(["docker", "context", "ls"], 
                                    capture_output=True, text=True)
            if result.returncode != 0:
                print("Error: Docker doesn't seem to be running or doesn't support contexts")
                print(result.stderr)
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
                print(f"Error creating Docker context: {result.stderr}")
                raise Exception(f"Failed to create Docker context: {result.stderr}")
                
            print(f"Successfully created Docker context '{self.context_name}'")
            
            test_result = subprocess.run(
                ["docker", "--context", self.context_name, "version", "--format", "{{.Server.Version}}"],
                capture_output=True, text=True
            )
            
            if test_result.returncode != 0:
                print(f"Warning: Docker context created but connection test failed: {test_result.stderr}")
                print("Make sure Docker daemon is configured to listen on TCP port")
                print(f"Please check that port {docker_port} is open and Docker is properly configured")
                
            return self.context_name
            
        except Exception as e:
            print(f"Error setting up Docker context: {str(e)}")
            raise
        

    def setup_docker_daemon_for_tcp(self):
        """Configure Docker daemon to listen on TCP for the context to connect."""
        docker_config_dir = Path("/etc/docker")
        docker_config_dir.mkdir(parents=True, exist_ok=True)
        
        docker_port = 2376 if self.env == "production" else 2377
        
        daemon_config = {}
        daemon_json_path = docker_config_dir / "daemon.json"
        
        if daemon_json_path.exists():
            with open(daemon_json_path, "r") as f:
                try:
                    daemon_config = json.load(f)
                except json.JSONDecodeError:
                    print("Warning: Existing daemon.json is invalid. Starting with empty config.")
        
        hosts = daemon_config.get("hosts", ["unix:///var/run/docker.sock"])
        tcp_endpoint = f"tcp://0.0.0.0:{docker_port}"
        
        if tcp_endpoint not in hosts:
            hosts.append(tcp_endpoint)
            daemon_config["hosts"] = hosts
            
            with open(daemon_json_path, "w") as f:
                json.dump(daemon_config, f, indent=2)
            
            try:
                result = subprocess.run(["systemctl", "reload", "docker"], 
                                        capture_output=True, text=True)
                if result.returncode != 0:
                    result = subprocess.run(["systemctl", "restart", "docker"], 
                                            capture_output=True, text=True)
                    if result.returncode != 0:
                        print("Error restarting Docker service:")
                        print(result.stderr)
                        raise Exception("Failed to restart Docker service")
                
                time.sleep(3)
                
                result = subprocess.run(["systemctl", "status", "docker"], 
                                        capture_output=True, text=True)
                if result.returncode != 0:
                    print("Error checking Docker service status:")
                    print(result.stderr)
                    raise Exception("Failed to check Docker service status")
                
            except Exception as e:
                result = subprocess.run(["journalctl", "-u", "docker", "-n", "50"], 
                                        capture_output=True, text=True)
                print("\nDocker service error logs:")
                print(result.stdout)
                raise Exception(f"Failed to manage Docker service. Error: {str(e)}")

    def setup_environment(self):
        db_name = f"nixopus_{self.generate_random_string(8)}"
        username = f"nixopus_{self.generate_random_string(8)}"
        password = self.generate_random_string(16)
        
        # we will use different ports for staging and production
        api_port = 8443 if self.env == "production" else 8444
        next_public_port = 7443 if self.env == "production" else 7444
        db_port = 5432 if self.env == "production" else 5433

        private_key_path, public_key_path = self.generate_ssh_key()
        self.setup_authorized_keys()
        local_ip = self.get_local_ip()

        self.setup_docker_certs()
        self.setup_docker_daemon_for_tcp()
        self.create_docker_context()

        base_env_vars = {
            "DB_NAME": db_name,
            "USERNAME": username,
            "PASSWORD": password,
            "HOST_NAME": "nixopus-db" if self.env == "production" else f"nixopus-{self.env}-db",
            "DB_PORT": str(db_port),
            "SSL_MODE": "disable",
            "API_PORT": str(api_port),
            "API_URL": f"https://{self.domains['api_domain']}/api",
            "WEBSOCKET_URL": f"wss://{self.domains['api_domain']}/ws",
            "WEBHOOK_URL": f"https://{self.domains['api_domain']}/api/v1/webhook",
            "NEXT_PUBLIC_PORT": str(next_public_port),
            "MOUNT_PATH": "/etc/nixopus/configs" if self.env == "production" else "/etc/nixopus-staging/configs",
            "PORT": str(api_port),
            "SSH_HOST": local_ip,
            "SSH_PORT": "22",
            "SSH_USER": "root",
            "SSH_PRIVATE_KEY": str(private_key_path),
            "DOCKER_HOST": f"tcp://{local_ip}:2376" if self.env == "production" else f"tcp://{local_ip}:2377",
            "DOCKER_TLS_VERIFY": "1",
            "DOCKER_CERT_PATH": str(self.docker_certs_dir),
            "DOCKER_CONTEXT": self.context_name,
            "CADDY_ENDPOINT": "http://nixopus-caddy:2019",
            "CADDY_DATA_VOLUME": str(self.config_dir / "caddy" / "data"),
            "CADDY_CONFIG_VOLUME": str(self.config_dir / "caddy" / "config"),
            "DB_VOLUME": str(self.config_dir / "db"),
            "ALLOWED_ORIGIN": f"https://{self.domains['app_domain']}"
        }

        with open(self.env_file, 'w') as f:
            for key, value in base_env_vars.items():
                f.write(f"{key}={value}\n")

        api_env_vars = base_env_vars.copy()
        api_env_vars["PORT"] = str(api_port)

        api_env_file = self.project_root / "api" / ".env"
        with open(api_env_file, 'w') as f:
            for key, value in api_env_vars.items():
                f.write(f"{key}={value}\n")
        
        view_env_vars = base_env_vars.copy()
        view_env_vars["PORT"] = str(next_public_port)

        view_env_file = self.project_root / "view" / ".env"
        with open(view_env_file, 'w') as f:
            for key, value in view_env_vars.items():
                f.write(f"{key}={value}\n")

        self.env_file.chmod(0o600)
        private_key_path.chmod(0o600)
        public_key_path.chmod(0o644)
        self.docker_certs_dir.chmod(0o700)
        return base_env_vars