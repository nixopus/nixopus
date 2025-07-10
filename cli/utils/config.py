import os

class Config:
    def __init__(self, default_env="PRODUCTION"):
        self.default_env = default_env
    
    def get_env(self):
        return os.environ.get("ENV", self.default_env)
    
    def is_development(self):
        return self.get_env().upper() == "DEVELOPMENT" 
