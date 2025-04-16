import os
import secrets
import string
from pathlib import Path
import subprocess
import socket
import json
import time

class EnvironmentSetup:
    def __init__(self, domain):
        self.domain = domain
        self.project_root = Path(__file__).parent.parent
        self.env_file = self.project_root / ".env"
        self.env_sample = self.project_root / ".env.sample"
        self.config_dir = Path("/etc/nixopus")
        self.docker_certs_dir = self.config_dir / "docker-certs"
        self.ssh_dir = self.config_dir / "ssh"

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
                "-subj", "/CN=nixopus"
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
                "openssl", "req", "-subj", "/CN=client", "-new",
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

    def setup_docker_tls(self):
        docker_config_dir = Path("/etc/docker")
        docker_config_dir.mkdir(parents=True, exist_ok=True)

        daemon_config = {
            "tls": True,
            "tlscacert": str(self.docker_certs_dir / "ca.pem"),
            "tlscert": str(self.docker_certs_dir / "server-cert.pem"),
            "tlskey": str(self.docker_certs_dir / "server-key.pem"),
            "hosts": [
                "unix:///var/run/docker.sock",
                "tcp://0.0.0.0:2376"
            ],
            "tlsverify": True
        }

        with open(docker_config_dir / "daemon.json", "w") as f:
            json.dump(daemon_config, f, indent=2)

        try:
            result = subprocess.run(["systemctl", "stop", "docker"], capture_output=True, text=True)
            if result.returncode != 0:
                print("Error stopping Docker service:")
                print(result.stderr)
                raise Exception("Failed to stop Docker service")

            result = subprocess.run(["systemctl", "start", "docker"], capture_output=True, text=True)
            if result.returncode != 0:
                print("Error starting Docker service:")
                print(result.stderr)
                raise Exception("Failed to start Docker service")

            time.sleep(5)
            result = subprocess.run(["systemctl", "status", "docker"], capture_output=True, text=True)
            if result.returncode != 0:
                print("Error checking Docker service status:")
                print(result.stderr)
                raise Exception("Failed to check Docker service status")

            # print("\nDocker service status:")
            # print(result.stdout)

        except Exception as e:
            result = subprocess.run(["journalctl", "-u", "docker", "-n", "50"], capture_output=True, text=True)
            print("\nDocker service error logs:")
            print(result.stdout)
            raise Exception(f"Failed to manage Docker service. Error: {str(e)}\nJournal logs: {result.stdout}")

    def setup_environment(self):
        db_name = f"nixopus_{self.generate_random_string(8)}"
        username = f"nixopus_{self.generate_random_string(8)}"
        password = self.generate_random_string(16)
        
        api_port = 8443
        next_public_port = 7443 
        db_port = 5432 

        private_key_path, public_key_path = self.generate_ssh_key()
        self.setup_authorized_keys()
        local_ip = self.get_local_ip()

        self.setup_docker_certs()
        self.setup_docker_tls()

        env_vars = {
            "DB_NAME": db_name,
            "USERNAME": username,
            "PASSWORD": password,
            "HOST_NAME": "nixopus-db",
            "DB_PORT": str(db_port),
            "SSL_MODE": "disable",
            "API_PORT": str(api_port),
            "API_URL": f"https://api.{self.domain}/api",
            "WEBSOCKET_URL": f"wss://api.{self.domain}/ws",
            "WEBHOOK_URL": f"https://api.{self.domain}/webhook",
            "NEXT_PUBLIC_PORT": str(next_public_port),
            "MOUNT_PATH": "/etc/nixopus/configs",
            "PORT": str(api_port),
            "SSH_HOST": local_ip,
            "SSH_PORT": "22",
            "SSH_USER": "root",
            "SSH_PRIVATE_KEY": str(private_key_path),
            "DOCKER_HOST": f"tcp://{local_ip}:2376",
            "DOCKER_TLS_VERIFY": "1",
            "DOCKER_CERT_PATH": str(self.docker_certs_dir),
            "CADDY_ENDPOINT": "http://nixopus-caddy:2019",
            "CADDY_DATA_VOLUME": str(self.config_dir / "caddy" / "data"),
            "CADDY_CONFIG_VOLUME": str(self.config_dir / "caddy" / "config"),
            "DB_VOLUME": str(self.config_dir / "db"),
            "ALLOWED_ORIGIN": f"https://app.{self.domain}"
        }

        with open(self.env_file, 'w') as f:
            for key, value in env_vars.items():
                f.write(f"{key}={value}\n")

        # copy to api/.env
        api_env_file = self.project_root / "api" / ".env"
        with open(api_env_file, 'w') as f:
            for key, value in env_vars.items():
                f.write(f"{key}={value}\n")
        
        # copy to view/.env
        view_env_file = self.project_root / "view" / ".env"
        with open(view_env_file, 'w') as f:
            for key, value in env_vars.items():
                if key == "PORT":
                    f.write(f"PORT={next_public_port}\n")
                else:
                    f.write(f"{key}={value}\n")

        self.env_file.chmod(0o600)
        private_key_path.chmod(0o600)
        public_key_path.chmod(0o644)
        self.docker_certs_dir.chmod(0o700)
        for cert_file in self.docker_certs_dir.glob("*"):
            cert_file.chmod(0o600)

        return env_vars