import argparse
import secrets
import string
import sys
import getpass

from validation import Validation


class InputParser:
    def __init__(self):
        self.parser = self._setup_arg_parser()
        self.validation = Validation()
    
    def _setup_arg_parser(self):
        parser = argparse.ArgumentParser(description='Nixopus Installation Wizard')
        parser.add_argument('--api-domain', help='The domain where the nixopus api will be accessible (e.g. nixopusapi.example.com)')
        parser.add_argument('--app-domain', help='The domain where the nixopus app will be accessible (e.g. nixopus.example.com)')
        parser.add_argument('--email', '-e', help='The email to create the admin account with')
        parser.add_argument('--password', '-p', help='The password to create the admin account with')
        parser.add_argument('--env', choices=['production', 'staging'], default='production', help='The environment to install in (production or staging)')
        return parser
    
    def generate_strong_password(self):
        while True:
            password = ''.join(secrets.choice(
                string.ascii_letters + string.digits + self.validation.SPECIAL_CHARS
            ) for _ in range(16))
            if (any(c.isupper() for c in password) and
                any(c.islower() for c in password) and
                any(c.isdigit() for c in password) and
                any(c in self.validation.SPECIAL_CHARS for c in password)):
                return password

    def get_env_from_args(self, args):
        """
        Get the environment from the command line arguments
        """
        if args.env:
            if args.env not in ['production', 'staging']:
                print("Error: Environment must be either 'production' or 'staging'")
                sys.exit(1)
            return args.env
        else:
            # default to production environment if no environment is specified
            return "production"
    
    def get_domains_from_args(self, args):
        if args.api_domain and args.app_domain:
            try:
                self.validation.validate_domain(args.api_domain)
                self.validation.validate_domain(args.app_domain)
                return {
                    "api_domain": args.api_domain,
                    "app_domain": args.app_domain,
                }
            except SystemExit:
                return None
        return None
    
    def get_admin_credentials_from_args(self, args):
        if not args.email and not args.password:
            return None, None

        if args.email and args.password:
            self.validation.validate_email(args.email)
            self.validation.validate_password(args.password)
            return args.email, args.password

        if args.email:
            self.validation.validate_email(args.email)
            password = getpass.getpass("Please enter the password for the admin(generates a strong password if left blank): ")
            if not password:
                password = self.generate_strong_password()
            self.validation.validate_password(password)
            return args.email, password

        if args.password:
            self.validation.validate_password(args.password)
            email = input("Please enter the email for the admin: ")
            if not email:
                print("Error: Email is required")
                sys.exit(1)
            self.validation.validate_email(email)
            return email, args.password

        return None, None
    
    def ask_admin_credentials(self):
        """
        Ask for admin credentials
        """
        while True:
            email = input("Please enter the email for the admin: ")
            try:
                self.validation.validate_email(email)
                break
            except SystemExit:
                print("Please enter a valid email address")
                continue
                
        password = input("Please enter the password for the admin(generates a strong password if left blank): ")
        if not password:
            password = self.generate_strong_password()
        self.validation.validate_password(password)
        return email, password
    
    # def ask_for_domain(self):
    #     """
    #     Ask for the domain
    #     """
    #     validation = Validation()
    #     while True:
    #         domain = input("Please enter the base domain (if domain is example.com, then api domain will be nixopusapi.example.com and app domain will be nixopus.example.com) : ")
    #         if not domain:
    #             continue
    #         try:
    #             validation.validate_domain(domain)
    #             return {
    #                 "api_domain": f"nixopusapi.{domain}",
    #                 "app_domain": f"nixopus.{domain}",
    #             }
    #         except SystemExit:
    #             print("Please enter a valid domain name")
    #             continue
