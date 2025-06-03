import subprocess
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

    def generate_key(self) -> Tuple[Path, Path]:
        self.ssh_dir.mkdir(parents=True, exist_ok=True)
        private_key_path = self.ssh_dir / f"id_{self.config.key_type}"
        public_key_path = self.ssh_dir / f"id_{self.config.key_type}.pub"
        
        # (Re)generate when either part of the keypair is missing
        if not private_key_path.exists() or not public_key_path.exists():
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
            except FileNotFoundError as err:
                raise Exception(
                    self.config.errors["ssh_keygen_not_found"].format(error=str(err))
                ) from err
            except subprocess.CalledProcessError as err:
                raise Exception(
                    self.config.errors["ssh_keygen_failed"]
                ) from err
            except Exception as err:
                raise Exception(
                    self.config.errors["ssh_key_error"].format(error=str(err))
                ) from err
        return private_key_path, public_key_path

    def setup_authorized_keys(self, public_key_path: Path, permissions: dict) -> None:
        try:
            ssh_config_dir = Path.home() / ".ssh"
            ssh_config_dir.mkdir(mode=0o700, parents=True, exist_ok=True)
            authorized_keys_path = ssh_config_dir / "authorized_keys"
            
            with open(public_key_path, 'r') as pk_file:
                public_key_content = pk_file.read().strip()
                
            if authorized_keys_path.exists():
                with open(authorized_keys_path, 'r') as auth_file:
                    existing_content = auth_file.read()
                    if public_key_content in existing_content:
                        return
                        
            with open(authorized_keys_path, 'a+') as auth_file:
                auth_file.write(f"\n{public_key_content}\n")
                
            authorized_keys_path.chmod(int(permissions["authorized_keys"], 8))
            
        except Exception as e:
            raise Exception(self.config.errors["auth_keys_error"].format(error=str(e))) 