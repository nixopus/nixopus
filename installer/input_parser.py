import argparse
import secrets
import string
import sys
import getpass
import logging

from validation import Validation


class InputParser:
    def __init__(self):
        self.parser = self._setup_arg_parser()
        self.validation = Validation()
        self.logger = logging.getLogger("nixopus")
    
    def _setup_arg_parser(self):
        parser = argparse.ArgumentParser(description='Nixopus Installation Wizard')
        parser.add_argument('--api-domain', help='The domain where the nixopus api will be accessible (e.g. nixopusapi.example.com)')
        parser.add_argument('--app-domain', help='The domain where the nixopus app will be accessible (e.g. nixopus.example.com)')
        parser.add_argument('--email', '-e', help='The email to create the admin account with')
        parser.add_argument('--password', '-p', help='The password to create the admin account with')
        parser.add_argument('--env', choices=['production', 'staging'], default='production', help='The environment to install in (production or staging)')
        parser.add_argument("--debug", action='store_true', help='Enable debug mode')
        return parser
    
    def parse_args(self):
        args = self.parser.parse_args()
        if args.debug:
            self.logger.setLevel(logging.DEBUG)
        return args
    
    def generate_strong_password(self):
        while True:
            password = ''.join(secrets.choice(
                string.ascii_letters + string.digits + self.validation.SPECIAL_CHARS
            ) for _ in range(16))
            if (any(c.isupper() for c in password) and
                any(c.islower() for c in password) and
                any(c.isdigit() for c in password) and
                any(c in self.validation.SPECIAL_CHARS for c in password)):
                self.logger.debug(f"Generated password: {password}")
                return password

    def get_env_from_args(self, args):
        """
        Get the environment from the command line arguments
        """
        if args.env:
            if args.env not in ['production', 'staging']:
                print("Error: Environment must be either 'production' or 'staging'")
                sys.exit(1)
            self.logger.debug(f"Using environment: {args.env}")
            return args.env
        else:
            self.logger.debug("No environment specified, defaulting to production")
            return "production"
    
    def get_domains_from_args(self, args):
        if args.api_domain and args.app_domain:
            try:
                self.logger.debug(f"Validating domains - API: {args.api_domain}, App: {args.app_domain}")
                self.validation.validate_domain(args.api_domain)
                self.validation.validate_domain(args.app_domain)
                return {
                    "api_domain": args.api_domain,
                    "app_domain": args.app_domain,
                }
            except SystemExit:
                return None
        self.logger.debug("No domains provided")
        return None
    
    def get_admin_credentials_from_args(self, args):
        # if email and password are provided, validate them and return them
        if args.email and args.password:
            self.logger.debug(f"Validating admin credentials for email: {args.email}")
            self.validation.validate_email(args.email)
            self.validation.validate_password(args.password)
            return args.email, args.password

        # return None if only one of email or password is provided or if both are not provided
        self.logger.debug("No admin credentials provided")
        return None, None