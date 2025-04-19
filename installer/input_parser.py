import argparse
from validation import Validation
import secrets
import string

class InputParser:
    def __init__(self):
        self.parser = self._setup_arg_parser()
    
    def _setup_arg_parser(self):
        parser = argparse.ArgumentParser(description='Nixopus Installation Wizard')
        parser.add_argument('--api-domain', help='The domain where the nixopus api will be accessible (e.g. nixopusapi.example.com)')
        parser.add_argument('--app-domain', help='The domain where the nixopus app will be accessible (e.g. nixopus.example.com)')
        parser.add_argument('--email', '-e', help='The email to create the admin account with')
        parser.add_argument('--password', '-p', help='The password to create the admin account with')
        return parser
    
    def generate_strong_password(self):
        while True:
            password = ''.join(secrets.choice(
                string.ascii_letters + string.digits + string.punctuation
            ) for _ in range(16))
            if (any(c.isupper() for c in password) and
                any(c.islower() for c in password) and
                any(c.isdigit() for c in password) and
                any(c in string.punctuation for c in password)):
                return password
    
    def get_domains_from_args(self, args):
        """
        Get the domains from the command line arguments
        """
        if args.api_domain or args.app_domain:
            validation = Validation()
            
            try:
                # If both api and app domains are provided, validate them and we're good to go
                if args.api_domain and args.app_domain:
                    validation.validate_domain(args.api_domain)
                    validation.validate_domain(args.app_domain)
                    return {
                        "api_domain": args.api_domain,
                        "app_domain": args.app_domain,
                    }
                # If only api domain is provided, validate it and generate the app domain as default
                elif args.api_domain:
                    validation.validate_domain(args.api_domain)
                    return {
                        "api_domain": args.api_domain,
                        "app_domain": f"nixopus.{args.api_domain.split('.', 1)[1]}",
                    }
                # If only app domain is provided, validate it and generate the api domain as default
                elif args.app_domain:
                    validation.validate_domain(args.app_domain)
                    return {
                        "api_domain": f"nixopusapi.{args.app_domain.split('.', 1)[1]}",
                        "app_domain": args.app_domain,
                    }
            except SystemExit:
                print("Invalid domain provided. Please try again with valid domains.")
                return None
        return None
    
    def get_admin_credentials_from_args(self, args):
        """
        Get the admin credentials from the command line arguments
        """
        if args.email or args.password:
            validation = Validation()
            
            # If both email and password are provided, validate them and we're good to go
            if args.email and args.password:
                validation.validate_email(args.email)
                validation.validate_password(args.password)
                return args.email, args.password
            # If only email is provided, validate it and ask for password
            elif args.email:
                validation.validate_email(args.email)
                password = input("Please enter the password for the admin(generates a strong password if left blank): ")
                if not password:
                    password = self.generate_strong_password()
                validation.validate_password(password)
                return args.email, password
            # If only password is provided, validate it and ask for email
            elif args.password:
                validation.validate_password(args.password)
                while True:
                    email = input("Please enter the email for the admin: ")
                    try:
                        validation.validate_email(email)
                        return email, args.password
                    except SystemExit:
                        print("Please enter a valid email address")
                        continue
        return None, None
    
    # will be used if no args are provided
    def ask_admin_credentials(self):
        """
        Ask for admin credentials
        """
        validation = Validation()
        while True:
            email = input("Please enter the email for the admin: ")
            try:
                validation.validate_email(email)
                break
            except SystemExit:
                print("Please enter a valid email address")
                continue
                
        password = input("Please enter the password for the admin(generates a strong password if left blank): ")
        if not password:
            password = self.generate_strong_password()
        validation.validate_password(password)
        return email, password 
    
    def ask_for_domain(self):
        """
        Ask for the domain
        """
        validation = Validation()
        while True:
            domain = input("Please enter the base domain (if domain is example.com, then api domain will be nixopusapi.example.com and app domain will be nixopus.example.com) : ")
            try:
                validation.validate_domain(domain)
                domains = {
                    "api_domain": f"nixopusapi.{domain}",
                    "app_domain": f"nixopus.{domain}",
                }
                break
            except SystemExit:
                print("Please enter a valid domain name")
                continue
        return domains
