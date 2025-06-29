import subprocess
import logging
from pathlib import Path
from typing import Tuple
from dataclasses import dataclass

@dataclass
class SSHConfig:
    port: int
    user: str
    key_bits: int
    key_type: str
    errors: dict

class SSHSetup:
    def __init__(self, config: SSHConfig, ssh_dir: Path):
        self.config = config
        self.ssh_dir = ssh_dir
        self.logger = logging.getLogger("nixopus")

    def generate_key(self) -> Tuple[Path, Path]:
        self.logger.debug(f"Generating SSH key in {self.ssh_dir}")
        self.ssh_dir.mkdir(parents=True, exist_ok=True)
        private_key_path = self.ssh_dir / f"id_{self.config.key_type}"
        public_key_path = self.ssh_dir / f"id_{self.config.key_type}.pub"
        
        # (Re)generate when either part of the keypair is missing
        if not private_key_path.exists() or not public_key_path.exists():
            self.logger.debug(f"Generating new key pair - Type: {self.config.key_type}, Bits: {self.config.key_bits}")
            try:
                subprocess.run(
                    [
                        "ssh-keygen",
                        "-t", self.config.key_type,
                        "-b", str(self.config.key_bits),
                        "-f", str(private_key_path),
                        "-N", ""
                    ],
                    check=True
                )
                self.logger.debug("SSH key pair generated successfully")
            except FileNotFoundError as err:
                self.logger.debug(f"ssh-keygen not found: {err}")
                raise Exception(
                    self.config.errors["ssh_keygen_not_found"].format(error=str(err))
                ) from err
            except subprocess.CalledProcessError as err:
                self.logger.debug(f"ssh-keygen failed: {err}")
                raise Exception(
                    self.config.errors["ssh_keygen_failed"]
                ) from err
            except Exception as err:
                self.logger.debug(f"Error generating SSH key: {err}")
                raise Exception(
                    self.config.errors["ssh_key_error"].format(error=str(err))
                ) from err
        else:
            self.logger.debug("Using existing SSH key pair")
        return private_key_path, public_key_path

    def setup_authorized_keys(self, public_key_path: Path, permissions: dict) -> None:
        self.logger.debug("Setting up authorized keys")
        try:
            ssh_config_dir = Path.home() / ".ssh"
            ssh_config_dir.mkdir(mode=0o700, parents=True, exist_ok=True)
            authorized_keys_path = ssh_config_dir / "authorized_keys"
            
            with open(public_key_path, 'r') as pk_file:
                public_key_content = pk_file.read().strip()
                
            if authorized_keys_path.exists():
                self.logger.debug("Checking existing authorized_keys file")
                with open(authorized_keys_path, 'r') as auth_file:
                    existing_content = auth_file.read()
                    if public_key_content in existing_content:
                        self.logger.debug("Public key already exists in authorized_keys")
                        return
                        
            self.logger.debug("Adding public key to authorized_keys")
            with open(authorized_keys_path, 'a+') as auth_file:
                auth_file.write(f"\n{public_key_content}\n")
                
            authorized_keys_path.chmod(int(permissions["authorized_keys"], 8))
            self.logger.debug("Authorized keys setup completed")
            
        except Exception as e:
            self.logger.debug(f"Error setting up authorized keys: {e}")
            raise Exception(self.config.errors["auth_keys_error"].format(error=str(e))) 