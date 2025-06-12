import re
import sys
import socket
import string

class Validation:
    SPECIAL_CHARS = '!@#$%^&*()_+-=[]{}|;:,.<>?'
    
    def __init__(self): 
        pass
    
    def validate_email(self, email):
        if not email:
            print("Error: Email is required")
            sys.exit(1)
        if not re.match(r"[^@]+@[^@]+\.[^@]+", email):
            print("Error: Invalid email address")
            sys.exit(1)
        return email

    def validate_password(self, password):
        if not password or len(password) < 8:
            print(f"Error: Password must be at least 8 characters long")
            sys.exit(1)
    
        has_uppercase = any(char.isupper() for char in password)
        has_lowercase = any(char.islower() for char in password)
        has_digit = any(char.isdigit() for char in password)
        has_special = any(char in self.SPECIAL_CHARS for char in password)
        
        if not (has_uppercase and has_lowercase and has_digit and has_special):
            print("Error: Password must contain at least one uppercase letter, one lowercase letter, one number, and one special character")
            sys.exit(1)
    
        return password
    
    def validate_domain(self, domain):
        if not domain:
            print("Error: Domain is required")
            sys.exit(1)
        if not re.match(r"^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$", domain):
            print("Error: Invalid domain name")
            sys.exit(1)
        
        try:
            hostname = socket.gethostname()
            server_ip = socket.gethostbyname(hostname)
            domain_ip = socket.gethostbyname(domain)
            
            if server_ip != domain_ip:
                print(f"Warning: Domain {domain} does not point to this server's IP ({server_ip})")
                print("Please ensure your DNS records are properly configured before proceeding.")
        except socket.gaierror:
            print(f"Warning: Could not resolve domain {domain}")
            print("Please ensure your DNS records are properly configured before proceeding.")
            
        return domain
